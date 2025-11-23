package api

import (
	"github.com/mizdebsk/rhel-drivers/internal/exec"
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
	EnsureRepositoriesEnabled() error
}

type DriverID struct {
	ProviderID string
	Version    string
}

type Provider interface {
	GetID() string
	GetName() string
	Install(drivers []DriverID) ([]string, error)
	Remove(drivers []DriverID) ([]string, error)
	ListAvailable() ([]DriverID, error)
	ListInstalled() ([]DriverID, error)
	DetectHardware() (bool, error)
}

type CoreDeps struct {
	PackageManager    PackageManager
	RepositoryManager RepositoryManager
	Providers         []Provider
	Executor          exec.Executor
}

type DriverStatus struct {
	ID         DriverID
	Available  bool
	Installed  bool
	Compatible bool
}
