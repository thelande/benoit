/*
Copyright © 2026 Tom Helander <thomas.helander@gmail.com>
*/
package benoit

import (
	"bytes"
	"strings"
	"text/template"
	"time"

	"al.essio.dev/pkg/shellescape"
	sprig "github.com/go-task/slim-sprig/v3"
)

// valueOrDefaultInt returns the provided value if it is non-zero, or the def
// value if v is 0.
func valueOrDefaultInt(v int, def int) int {
	if v == 0 {
		return def
	}
	return v
}

// parseTimeoutOrDefault converts a raw duration string into a time.Duration
// object. Returns a 30 second duration if no duration was provided.
func parseTimeoutOrDefault(raw string) time.Duration {
	// Default to 30 seconds
	if raw == "" {
		return 30 * time.Second
	}
	d, err := time.ParseDuration(raw)
	if err != nil {
		return 30 * time.Second
	}
	return d
}

// renderCommand takes a command list and passes each element through the
// template parser to render template variables into real values. The resulting
// list of rendered arguments is returned.
func renderCommand(tmplStr []string, ctx TemplateContext) ([]string, error) {
	result := make([]string, 0, len(tmplStr))

	for i := range tmplStr {
		tmpl, err := template.New("cmd").Funcs(sprig.FuncMap()).Parse(tmplStr[i])
		if err != nil {
			return []string{}, err
		}

		var buf bytes.Buffer
		if err := tmpl.Execute(&buf, ctx); err != nil {
			return []string{}, err
		}

		result = append(result, buf.String())
	}
	return result, nil
}

// argsToShellCommand takes a list of command arguments and converts them into
// an escaped shell command. Used to take the check command and convert it for
// use with remote execution via SSH.
func argsToShellCommand(args []string) string {
	if len(args) == 0 {
		return ""
	}
	quoted := make([]string, len(args))
	for i, a := range args {
		quoted[i] = shellescape.Quote(a)
	}
	return strings.Join(quoted, " ")
}
