package core

import (
	"fmt"

	"github.com/mizdebsk/rhel-drivers/internal/api"
	"github.com/mizdebsk/rhel-drivers/internal/log"
)

func Remove(deps api.CoreDeps, opts api.RemoveOptions, drivers []string) error {
	var toRemove []api.DriverID

	if len(drivers) == 0 && !opts.All {
		return fmt.Errorf("not specified what to remove")
	}
	if len(drivers) > 0 && opts.All {
		return fmt.Errorf("both --all and specific drivers specified")
	}
	if opts.All {
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
	} else {
	outer2:
		for _, driverStr := range drivers {
			driver, err := parseDriverID(driverStr)
			if err != nil {
				return fmt.Errorf("invalid driver ID %q: %w", driverStr, err)
			}
			for _, provider := range deps.Providers {
				provID := provider.GetID()
				if driver.ProviderID == provID {
					installed, err := provider.ListInstalled()
					if err != nil {
						return fmt.Errorf("failed to list installed %s drivers: %w", provID, err)
					}
					for _, inst := range installed {
						if inst.Version == driver.Version {
							toRemove = append(toRemove, inst)
							continue outer2
						}
					}
					return fmt.Errorf("driver %s version %s is NOT installed", provID, driver.Version)
				}
			}
			return fmt.Errorf("unknown provider for driver: %s", driver)
		}
	}
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
		log.Logf("package will be installed: %v", pkg)
	}
	if err := deps.PackageManager.Remove(allPkgs, opts); err != nil {
		return fmt.Errorf("failed to remove pacakges: %w", err)
	}
	return nil
}
