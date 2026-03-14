/*
Copyright © 2026 Tom Helander <thomas.helander@gmail.com>
*/
package benoit

import (
	"fmt"
	"maps"
	"strings"
	"time"
)

// NewEngine validates the provided spec and returns the resulting Engine object
// when the spec is valid. Returns a non-nil error if any validation checks fail.
func NewEngine(spec Spec) (*Engine, error) {
	tmap := make(TargetMap)
	for _, t := range spec.Targets {
		if t.ID == "" {
			return nil, fmt.Errorf("target with empty id")
		}
		if _, exists := tmap[t.ID]; exists {
			return nil, fmt.Errorf("duplicate target id %q", t.ID)
		}
		tmap[t.ID] = t
	}

	cmap := make(CheckMap)
	for _, c := range spec.Checks {
		if c.ID == "" {
			return nil, fmt.Errorf("check with empty id")
		}
		if _, exists := cmap[c.ID]; exists {
			return nil, fmt.Errorf("duplicate check id %q", c.ID)
		}
		cmap[c.ID] = c
	}

	e := &Engine{
		targets: tmap,
		checks:  cmap,
	}

	if err := e.validateSpec(spec); err != nil {
		return nil, err
	}

	return e, nil
}

// validateSpec ensures that all referenced targets and checks exist. Returns
// an error when an undefined target or check is found, nil otherwise.
func (e *Engine) validateSpec(spec Spec) error {
	// Ensure each check's target exists
	for _, c := range spec.Checks {
		if _, ok := e.targets[c.Target]; !ok {
			return fmt.Errorf("check %q references unknown target %q", c.ID, c.Target)
		}
	}

	// Ensure all on_success / on_failure references exist
	for _, c := range spec.Checks {
		for _, nextID := range c.OnSuccess {
			if _, ok := e.checks[nextID]; !ok {
				return fmt.Errorf("check %q on_success references unknown check %q", c.ID, nextID)
			}
		}
		for _, nextID := range c.OnFailure {
			if _, ok := e.checks[nextID]; !ok {
				return fmt.Errorf("check %q on_failure references unknown check %q", c.ID, nextID)
			}
		}
	}

	return nil
}

// execCheck renders and executes the given Check against the provided Target,
// dispatching to the appropriate execution backend (local, SSH, or Kubernetes)
// and applying any per-check timeout and environment configuration. It returns
// true when the command completes successfully (exit code 0) and false when
// the command fails or a non-zero exit code is observed; in either case, a
// non-nil error is only returned for transport or execution-layer problems
// (e.g., connection failures, invalid configuration), not for normal command
// failures.
func (e *Engine) execCheck(chk Check, tgt Target) (bool, error) {
	timeout := parseTimeoutOrDefault(chk.Timeout)

	ctx := TemplateContext{
		Target: tgt,
		Check:  chk,
	}

	cmd, err := renderCommand(chk.Command, ctx)
	if err != nil {
		return false, fmt.Errorf("render command for %s: %w", chk.ID, err)
	}

	// Combine the environment variables from the Check and Target, with the
	// Target environment variables taking precedence.
	env := map[string]string{}
	if len(chk.Env) > 0 {
		maps.Copy(env, chk.Env)
	}
	if len(tgt.Env) > 0 {
		maps.Copy(env, tgt.Env)
	}

	if chk.RunLocal {
		exit, stdout, stderr, err := runLocalCommand(cmd, timeout, env)
		if len(strings.TrimSpace(stdout)) > 0 && !chk.HideStdout {
			PrintStdout(chk.ID, stdout)
		}
		if err != nil {
			styleSkip.Printf("%s [%-20s] error:\n%s\n", emojiInfo, chk.ID, stderr)
			return false, err
		}
		return exit == 0, nil
	}

	switch tgt.Kind {
	case TargetLocal:
		exit, stdout, stderr, err := runLocalCommand(cmd, timeout, env)
		if len(strings.TrimSpace(stdout)) > 0 && !chk.HideStdout {
			PrintStdout(chk.ID, stdout)
		}
		if err != nil {
			// treat as failure but keep error
			return false, err
		}
		if exit != 0 {
			PrintStderr(chk.ID, stderr)
		}
		return exit == 0, nil
	case TargetSSH:
		cmdStr := argsToShellCommand(cmd)
		exit, stdout, stderr, err := runSSHCommand(tgt, cmdStr, timeout, env)
		if len(strings.TrimSpace(stdout)) > 0 && !chk.HideStdout {
			PrintStdout(chk.ID, stdout)
		}
		if err != nil {
			PrintStderr(chk.ID, stderr)
			return false, err
		}
		return exit == 0, nil
	case TargetKubernetes:
		exit, stdout, stderr, err := runKubeCommand(
			tgt, chk.Namespace, chk.Pod, chk.Container, cmd, timeout, env,
		)
		if len(strings.TrimSpace(stdout)) > 0 && !chk.HideStdout {
			PrintStdout(chk.ID, stdout)
		}
		if err != nil {
			PrintStderr(chk.ID, stderr)
			return false, err
		}
		return exit == 0, nil
	default:
		return false, fmt.Errorf("unsupported target kind %q", tgt.Kind)
	}
}

// Run executes the diagnostic workflow starting from the check identified by
// startID, following each check's OnSuccess and OnFailure edges in a queue-
// based breadth-first manner. It logs and executes each check in turn, and
// propagates any unknown check or target references as an immediate error.
// If any check fails or returns an execution error, Run completes the workflow
// but returns a non-nil error indicating that one or more checks failed;
// otherwise it returns nil.
func (e *Engine) Run(startID string) error {
	queue := []string{startID}
	errors := false
	for len(queue) > 0 {
		id := queue[0]
		queue = queue[1:]

		chk, ok := e.checks[id]
		if !ok {
			return fmt.Errorf("unknown check %q", id)
		}
		tgt, ok := e.targets[chk.Target]
		if !ok {
			return fmt.Errorf("unknown target %q", chk.Target)
		}

		okExec, err := e.runCheckWithLogging(chk, tgt)
		if err != nil {
			errors = true
		}

		var next []string
		if okExec {
			next = chk.OnSuccess
		} else {
			next = chk.OnFailure
			errors = true
		}
		queue = append(queue, next...)
	}
	if errors {
		return fmt.Errorf("one or more checks failed")
	}
	return nil
}

// runCheckWithLogging wraps a given check on a target with timing and logging.
// Any failures or errors are propagated to the caller.
func (e *Engine) runCheckWithLogging(chk Check, tgt Target) (bool, error) {
	styleRunning.Printf("%s [%-20s] %s ", emojiRun, chk.ID, tgt.ID)
	fmt.Println("- starting")

	start := time.Now()
	ok, err := e.execCheck(chk, tgt)
	dur := time.Since(start)

	if err != nil {
		styleFail.Printf("%s [%-20s] %s ", emojiFail, chk.ID, tgt.ID)
		fmt.Printf("- error after %s: %v\n", dur.Truncate(time.Millisecond), err)
		return false, err
	}

	if ok {
		styleOK.Printf("%s [%-20s] %s ", emojiOK, chk.ID, tgt.ID)
		fmt.Printf("- success in %s\n", dur.Truncate(time.Millisecond))
	} else {
		styleFail.Printf("%s [%-20s] %s ", emojiFail, chk.ID, tgt.ID)
		fmt.Printf("- FAILED in %s\n", dur.Truncate(time.Millisecond))
	}

	return ok, nil
}
