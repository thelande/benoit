/*
Copyright © 2026 Tom Helander <thomas.helander@gmail.com>
*/
package benoit

import (
	"bytes"
	"fmt"
	"os"
	"time"

	"golang.org/x/crypto/ssh"
)

// runSSHCommand uses SSH to connect to the target and run the specified command.
// The resulting exit code, stdout, stderr, and any transport/connection error
// are returned. Defaults to port 22 if no port is provided in the target.
func runSSHCommand(t Target, command string, timeout time.Duration, env map[string]string) (int, string, string, error) {
	key, err := os.ReadFile(t.KeyFile)
	if err != nil {
		return 0, "", "", err
	}

	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		return 0, "", "", err
	}

	cfg := &ssh.ClientConfig{
		User:            t.User,
		Auth:            []ssh.AuthMethod{ssh.PublicKeys(signer)},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         timeout,
	}

	addr := fmt.Sprintf("%s:%d", t.Host, valueOrDefaultInt(t.Port, 22))
	client, err := ssh.Dial("tcp", addr, cfg)
	if err != nil {
		return 0, "", "", err
	}
	defer client.Close()

	sess, err := client.NewSession()
	if err != nil {
		return 0, "", "", err
	}
	defer sess.Close()

	for k, v := range env {
		sess.Setenv(k, v)
	}

	var stdoutBuf, stderrBuf bytes.Buffer
	sess.Stdout = &stdoutBuf
	sess.Stderr = &stderrBuf

	err = sess.Run(command)

	exitCode := 0
	if err != nil {
		if ee, ok := err.(*ssh.ExitError); ok {
			exitCode = ee.ExitStatus()
		} else {
			return 0, stdoutBuf.String(), stderrBuf.String(), err
		}
	}
	return exitCode, stdoutBuf.String(), stderrBuf.String(), nil
}
