package agent_installer

import "context"

type Provider interface {
	// TODO (dbugrik): WIP make version handling through injections.
	Versions() ([]string, error)
	Installer(versions string) (Installable, error)
}

type Installable interface {
	Version() string
	// TODO (dbugrik): WIP make receiving arguments typed operation (maybe injected component should be used).
	Install(ctx context.Context, kubeconfig string, args map[string]interface{}) error
}
