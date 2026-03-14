/*
Copyright © 2026 Tom Helander <thomas.helander@gmail.com>
*/
package benoit

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// runLocalCommand executes a command argument list on the current system.
// Returns the exit code, stdout, and stderr. A non-nil error is returned if the
// specified command is not found, not executable, or some other execution error
// occurs besides a non-zero exit code from the command.
func runLocalCommand(command []string, timeout time.Duration, envMap map[string]string) (int, string, string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	var env []string
	for k, v := range envMap {
		env = append(env, fmt.Sprintf("%s=%s", k, os.ExpandEnv(v)))
	}

	path := os.ExpandEnv(command[0])

	// Command is not absolute, we may need to make it relative to the cwd.
	if !filepath.IsAbs(path) {
		// Check if the command resolves to something in the PATH.
		absPath, err := exec.LookPath(path)
		if err != nil {
			// Did not find it in the path
			cwd, err := os.Getwd()
			if err != nil {
				return -1, "", "", err
			}
			path = filepath.Join(cwd, "scripts", path)
		} else {
			path = absPath
		}
	}
	cmd := exec.CommandContext(ctx, path, command[1:]...)
	cmd.Env = append(cmd.Environ(), env...)

	var stdoutBuf, stderrBuf bytes.Buffer
	cmd.Stdout = &stdoutBuf
	cmd.Stderr = &stderrBuf

	err := cmd.Run()

	exitCode := 0
	if err != nil {
		if ee, ok := err.(*exec.ExitError); ok {
			exitCode = ee.ExitCode()
		} else if ctx.Err() == context.DeadlineExceeded {
			// timed out
			exitCode = 124 // or any convention you like
		} else {
			return 0, strings.TrimSpace(stdoutBuf.String()), strings.TrimSpace(stderrBuf.String()), err
		}
	}

	return exitCode, stdoutBuf.String(), stderrBuf.String(), nil
}
