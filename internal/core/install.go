package core

import (
	"fmt"

	"github.com/mizdebsk/rhel-drivers/internal/api"
	"github.com/mizdebsk/rhel-drivers/internal/log"
)

func InstallSpecific(deps api.CoreDeps, drivers []string, dryRun, force bool) error {

	log.Debugf("options: dryRun=%v force=%v", dryRun, force)
	log.Debugf("arguments: %v", drivers)

	if len(drivers) == 0 {
		return fmt.Errorf("not specified what to install")
	}

	var toInstall []api.DriverID

outer:
	for _, driverStr := range drivers {
		driver, provider, err := resolveDriver(deps, driverStr)
		if err != nil {
			return err
		}
		available, err := provider.ListAvailable()
		if err != nil {
			return fmt.Errorf("failed to list available %s drivers: %w", provider.GetName(), err)
		}
		for _, avail := range available {
			if avail.Version == driver.Version {
				if !force {
					compat, err := provider.DetectHardware()
					if err != nil {
						log.Warnf("hardware detection failed for %s failed: %v", provider.GetName(), err)
					} else if !compat {
						return fmt.Errorf("no compatible %s hardware found", provider.GetName())
					} else {
						log.Infof("compatible hardware %s found", provider.GetName())
					}
				} else {
					log.Infof("not checking for %s hardware compatibility in force mode", provider.GetName())
				}
				toInstall = append(toInstall, driver)
				continue outer
			}
		}
		return fmt.Errorf("%s driver version %s is NOT available", provider.GetName(), driver.Version)
	}

	return doInstall(deps, toInstall, dryRun)
}

func InstallAutoDetect(deps api.CoreDeps, dryRun bool) error {

	log.Debugf("options: dryRun=%v", dryRun)

	var toInstall []api.DriverID

	hardwareDetected := false
	for _, provider := range deps.Providers {
		detected, err := provider.DetectHardware()
		if err != nil {
			log.Warnf("hardware detection failed for %s failed: %v", provider.GetName(), err)
			continue
		}
		if detected {
			hardwareDetected = true
			log.Logf("detected %s hardware", provider.GetName())
			available, err := provider.ListAvailable()
			if err != nil {
				return fmt.Errorf("failed to list available %s drivers: %w", provider.GetName(), err)
			}
			if len(available) > 0 {
				toInstall = append(toInstall, available[0])
			}
		}
	}
	if !hardwareDetected {
		return fmt.Errorf("no compatible hardware found")
	}
	if len(toInstall) == 0 {
		return fmt.Errorf("no drivers available for detected hardware")
	}

	return doInstall(deps, toInstall, dryRun)
}

func doInstall(deps api.CoreDeps, toInstall []api.DriverID, dryRun bool) error {
	if err := deps.RepositoryManager.EnsureRepositoriesEnabled(); err != nil {
		return fmt.Errorf("failed to verify/enable repositories: %w", err)
	}
	var allPkgs []string
	for _, provider := range deps.Providers {
		provID := provider.GetID()
		var provToInstall []api.DriverID
		for _, driver := range toInstall {
			if driver.ProviderID == provID {
				provToInstall = append(provToInstall, driver)
			}
		}
		if len(provToInstall) != 0 {
			pkgs, err := provider.Install(provToInstall)
			if err != nil {
				return fmt.Errorf("failed to install %s drivers: %w", provider.GetName(), err)
			}
			allPkgs = append(allPkgs, pkgs...)
		}
	}

	if len(allPkgs) == 0 {
		return fmt.Errorf("nothing to install")
	}
	for _, pkg := range allPkgs {
		log.Logf("package will be installed: %v", pkg)
	}
	if err := deps.PackageManager.Install(allPkgs, dryRun, false); err != nil {
		return fmt.Errorf("failed to install pacakges: %w", err)
	}
	return nil
}
