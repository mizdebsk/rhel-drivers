package api

//go:generate mockgen -source=provider.go -destination=../mocks/provider_mock.go -package=mocks

type Provider interface {
	GetID() string
	GetName() string
	Install(drivers []DriverID) ([]string, error)
	Remove(drivers []DriverID) ([]string, error)
	ListAvailable() ([]DriverID, error)
	ListInstalled() ([]DriverID, error)
	DetectHardware() (bool, error)
}
