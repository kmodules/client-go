package exec

import (
	"bytes"
	"fmt"
	"strings"

	core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/remotecommand"
)

type Exec struct {
	ClientConfig *rest.Config
	KubeClient   kubernetes.Interface

	Options ExecOptions
}

type ExecOptions struct {
	EnableStdin  bool
	EnableStdout bool
	EnableStderr bool
	EnableTTY    bool

	Input string
}

func (o ExecOptions) GetPodExecOptions(containerName string, command ...string) runtime.Object {
	return &core.PodExecOptions{
		Container: containerName,
		Command:   command,
		Stdin:     o.EnableStdin,
		Stdout:    o.EnableStdout,
		Stderr:    o.EnableStderr,
		TTY:       o.EnableTTY,
	}
}

func (o ExecOptions) GetStreamOptions(execOut, execErr *bytes.Buffer) remotecommand.StreamOptions {
	opt := remotecommand.StreamOptions{}

	if o.EnableStdin {
		opt.Stdin = strings.NewReader(o.Input)
	}

	if o.EnableStdout {
		opt.Stdout = execOut
	}

	if o.EnableStderr {
		opt.Stderr = execErr
	}

	return opt
}

func NewExec(clientConfig *rest.Config, kubeClient kubernetes.Interface) *Exec {
	return &Exec{
		ClientConfig: clientConfig,
		KubeClient:   kubeClient,
	}
}

func NewExecWithDefaultOptions(clientConfig *rest.Config, kubeClient kubernetes.Interface) *Exec {
	exec := NewExec(clientConfig, kubeClient)
	exec.Options.EnableStdout = true
	exec.Options.EnableStderr = true

	return exec
}

func NewExecWithInputOptions(clientConfig *rest.Config, kubeClient kubernetes.Interface, input string) *Exec {
	exec := NewExecWithDefaultOptions(clientConfig, kubeClient)
	exec.Options.EnableStdin = true
	exec.Options.Input = input

	return exec
}

func (e *Exec) Run(pod *core.Pod, command ...string) (string, error) {
	var (
		execOut bytes.Buffer
		execErr bytes.Buffer
	)

	req := e.KubeClient.CoreV1().RESTClient().Post().
		Resource("pods").
		Name(pod.Name).
		Namespace(pod.Namespace).
		SubResource("exec")
	req.VersionedParams(e.Options.GetPodExecOptions(pod.Spec.Containers[0].Name, command...), scheme.ParameterCodec)

	exec, err := remotecommand.NewSPDYExecutor(e.ClientConfig, "POST", req.URL())
	if err != nil {
		return "", fmt.Errorf("failed to init executor: %v", err)
	}

	err = exec.Stream(e.Options.GetStreamOptions(&execOut, &execErr))

	if err != nil {
		return "", fmt.Errorf("could not execute: %v", err)
	}

	if execErr.Len() > 0 {
		return "", fmt.Errorf("stderr: %v", execErr.String())
	}

	return execOut.String(), nil
}
