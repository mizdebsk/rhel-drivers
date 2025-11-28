package api

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
	Executor          Executor
}

type DriverStatus struct {
	ID         DriverID
	Available  bool
	Installed  bool
	Compatible bool
}
