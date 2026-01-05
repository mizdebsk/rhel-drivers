package core

import (
	"fmt"

	"github.com/mizdebsk/rhel-drivers/internal/api"
	"github.com/mizdebsk/rhel-drivers/internal/log"
)

func RemoveSpecific(deps api.CoreDeps, drivers []string, batchMode, dryRun bool) error {
	var toRemove []api.DriverID

	if len(drivers) == 0 {
		return fmt.Errorf("not specified what to remove")
	}
outer:
	for _, driverStr := range drivers {
		driver, provider, err := resolveDriver(deps, driverStr)
		if err != nil {
			return err
		}
		installed, err := provider.ListInstalled()
		if err != nil {
			return fmt.Errorf("failed to list installed %s drivers: %w", provider.GetName(), err)
		}
		for _, inst := range installed {
			if inst.Version == driver.Version {
				toRemove = append(toRemove, inst)
				continue outer
			}
		}
		return fmt.Errorf("driver %s version %s is NOT installed", provider.GetName(), driver.Version)
	}
	return doRemove(deps, toRemove, batchMode, dryRun)
}

func RemoveAll(deps api.CoreDeps, batchMode, dryRun bool) error {
	var toRemove []api.DriverID

	for _, provider := range deps.Providers {
		installed, err := provider.ListInstalled()
		if err != nil {
			return fmt.Errorf("failed to list installed %s drivers: %w", provider.GetName(), err)
		}
		toRemove = append(toRemove, installed...)
	}
	if len(toRemove) == 0 {
		return fmt.Errorf("not found any installed drivers to remove")
	}
	return doRemove(deps, toRemove, batchMode, dryRun)
}

func doRemove(deps api.CoreDeps, toRemove []api.DriverID, batchMode, dryRun bool) error {
	var allPkgs []string
	for _, provider := range deps.Providers {
		provID := provider.GetID()
		var provToRemove []api.DriverID
		for _, driver := range toRemove {
			if driver.ProviderID == provID {
				provToRemove = append(provToRemove, driver)
			}
		}
		if len(provToRemove) != 0 {
			pkgs, err := provider.Remove(provToRemove)
			if err != nil {
				return fmt.Errorf("failed to remove %s driver: %w", provider.GetName(), err)
			}
			allPkgs = append(allPkgs, pkgs...)
		}
	}
	if len(allPkgs) == 0 {
		return fmt.Errorf("nothing to remove")
	}
	for _, pkg := range allPkgs {
		log.Logf("package will be removed: %v", pkg)
	}
	if err := deps.PackageManager.Remove(allPkgs, batchMode, dryRun); err != nil {
		return fmt.Errorf("failed to remove pacakges: %w", err)
	}
	return nil
}
