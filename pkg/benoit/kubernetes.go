/*
Copyright © 2026 Tom Helander <thomas.helander@gmail.com>
*/
package benoit

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/remotecommand"
)

// runKubeCommand executes the specified command argument list within the
// specified container, pod, and namespace. Returns the exit code, stdout, and
// stderr. Returns a non-nil error for non-execution based errors (e.g., invalid
// namespace, pod, or container).
func runKubeCommand(
	t Target,
	namespace, pod, container string,
	cmd []string,
	timeout time.Duration,
	envMap map[string]string,
) (int, string, string, error) {
	config, err := clientcmd.BuildConfigFromFlags("", os.ExpandEnv(t.Kubeconfig))
	if err != nil {
		return 0, "", "", err
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return 0, "", "", err
	}

	newCmd := []string{"env"}
	for k, v := range envMap {
		newCmd = append(newCmd, fmt.Sprintf("%s=%s", k, v))
	}
	newCmd = append(newCmd, cmd...)

	req := clientset.CoreV1().RESTClient().Post().
		Resource("pods").
		Namespace(namespace).
		Name(pod).
		SubResource("exec").
		VersionedParams(&corev1.PodExecOptions{
			Container: container,
			Command:   cmd,
			Stdout:    true,
			Stderr:    true,
			Stdin:     false,
			TTY:       false,
		}, scheme.ParameterCodec)

	executor, err := remotecommand.NewSPDYExecutor(config, "POST", req.URL())
	if err != nil {
		return 0, "", "", err
	}

	var stdoutBuf, stderrBuf bytes.Buffer
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	err = executor.StreamWithContext(ctx, remotecommand.StreamOptions{
		Stdout: &stdoutBuf,
		Stderr: &stderrBuf,
	})

	// kubectl-style exec doesn’t give you a bare exit code directly;
	// Need to infer success from err == nil or parse from error when non-zero.
	// TODO: Stick with the easy err == nil for now, but expand to examining
	// the error later.
	exitCode := 0
	if err != nil {
		exitCode = 1
	}
	return exitCode, stdoutBuf.String(), stderrBuf.String(), nil
}
