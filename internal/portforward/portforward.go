package portforward

import (
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// New creates a new and initialized tunnel.
func New(client kubernetes.Interface, config *rest.Config, namespace, podName string, remotePort int) (*Tunnel, error) {
	t := NewTunnel(client.CoreV1().RESTClient(), config, namespace, podName, remotePort)
	return t, t.ForwardPort()
}
