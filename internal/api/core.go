package api

import (
	"context"
)

type InstallOptions struct {
	AutoDetect bool
	DryRun     bool
	Force      bool
}
type RemoveOptions struct {
	DryRun bool
	All    bool
}

type RepositoryManager interface {
	EnsureRepositoriesEnabled(ctx context.Context) error
}

type DriverID struct {
	ProviderID string
	Version    string
}

type Provider interface {
	GetID() string
	GetName() string
	Install(ctx context.Context, drivers []DriverID) ([]string, error)
	Remove(ctx context.Context, drivers []DriverID) ([]string, error)
	ListAvailable(ctx context.Context) ([]DriverID, error)
	ListInstalled(ctx context.Context) ([]DriverID, error)
	DetectHardware(ctx context.Context) (bool, error)
}

type CoreDeps struct {
	PM           PackageManager
	RepoVerifier RepositoryManager
	Providers    []Provider
}

type DriverStatus struct {
	ID         DriverID
	Available  bool
	Installed  bool
	Compatible bool
}
