package agent_installer

import (
	"context"
)

type Provider interface {
	Versions() ([]string, error)
	Installer(selector Selector) (Installable, error)
}

type Installable interface {
	Version() string
	Install(ctx context.Context) error
	Uninstall(ctx context.Context) error
}

type Selector interface {
	// TODO (dbugrik): WIP Decouple selection with receiving InstallValues.
	ConstraintsString() string
}
