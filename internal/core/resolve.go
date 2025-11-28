package core

import (
	"fmt"
	"strings"

	"github.com/mizdebsk/rhel-drivers/internal/api"
)

func parseDriverID(input string) (api.DriverID, error) {
	parts := strings.Split(input, ":")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return api.DriverID{}, fmt.Errorf("invalid driver ID format: %q (expected 'vendor:version')", input)
	}
	return api.DriverID{
		ProviderID: parts[0],
		Version:    parts[1],
	}, nil
}

func lookupProvider(deps api.CoreDeps, driver api.DriverID) (api.Provider, error) {
	for _, provider := range deps.Providers {
		provID := provider.GetID()
		if driver.ProviderID == provID {
			return provider, nil
		}
	}
	return nil, fmt.Errorf("unknown provider for driver: %s:%s", driver.ProviderID, driver.Version)
}

func resolveDriver(deps api.CoreDeps, driverStr string) (api.DriverID, api.Provider, error) {
	driver, err := parseDriverID(driverStr)
	if err != nil {
		return driver, nil, err
	}
	provider, err := lookupProvider(deps, driver)
	if err != nil {
		return driver, nil, err
	}
	return driver, provider, nil
}
