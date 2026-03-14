/*
Copyright © 2026 Tom Helander <thomas.helander@gmail.com>
*/
package benoit

type Engine struct {
	targets TargetMap
	checks  CheckMap
}

type TemplateContext struct {
	Target Target
	Check  Check
}

type Spec struct {
	Metadata Metadata `yaml:"metadata"`
	Targets  []Target `yaml:"targets"`
	Checks   []Check  `yaml:"checks"`
}

type Metadata struct {
	ID          string `yaml:"id"`
	Description string `yaml:"description"`
	Start       string `yaml:"start"`
}

type TargetKind string

const (
	TargetKubernetes TargetKind = "kubernetes"
	TargetSSH        TargetKind = "ssh"
	TargetLocal      TargetKind = "local"
)

type Target struct {
	ID   string            `yaml:"id"`
	Kind TargetKind        `yaml:"kind"`
	Env  map[string]string `yaml:"env,omitempty"`
	// SSH
	Host    string `yaml:"host,omitempty"`
	User    string `yaml:"user,omitempty"`
	KeyFile string `yaml:"keyFile,omitempty"`
	Port    int    `yaml:"port,omitempty"`
	// Kubernetes
	Kubeconfig string `yaml:"kubeconfig,omitempty"`
	Context    string `yaml:"context,omitempty"`
}

type Check struct {
	ID         string            `yaml:"id"`
	Target     string            `yaml:"target"`
	Command    []string          `yaml:"command"`
	Timeout    string            `yaml:"timeout,omitempty"`
	OnSuccess  []string          `yaml:"on_success,omitempty"`
	OnFailure  []string          `yaml:"on_failure,omitempty"`
	RunLocal   bool              `yaml:"run_local,omitempty"`
	HideStdout bool              `yaml:"hide_stdout,omitempty"`
	Env        map[string]string `yaml:"env,omitempty"`

	// Kubernetes specific
	Namespace string `yaml:"namespace,omitempty"`
	Pod       string `yaml:"pod,omitempty"`
	Container string `yaml:"container,omitempty"`
}

type TargetMap map[string]Target
type CheckMap map[string]Check
