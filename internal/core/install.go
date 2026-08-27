package core

import (
	"fmt"
	"slices"
	"strings"

	"github.com/mizdebsk/radii/internal/api"
	"github.com/mizdebsk/radii/internal/log"
)

func InstallSpecific(deps api.CoreDeps, drivers []string, batchMode, dryRun, force bool) error {
	if len(drivers) == 0 {
		return fmt.Errorf("not specified what to install")
	}

	needSupplementary := slices.ContainsFunc(drivers, func(d string) bool {
		return strings.HasPrefix(d, "nvidia:")
	})
	if err := deps.RepositoryManager.EnsureRepositoriesEnabled(needSupplementary); err != nil {
		return fmt.Errorf("failed to verify/enable repositories: %w", err)
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

	return doInstall(deps, toInstall, batchMode, dryRun)
}

func InstallAutoDetect(deps api.CoreDeps, batchMode, dryRun bool) error {
	var detectedProviders []api.Provider
	for _, provider := range deps.Providers {
		detected, err := provider.DetectHardware()
		if err != nil {
			log.Warnf("hardware detection failed for %s failed: %v", provider.GetName(), err)
			continue
		}
		if detected {
			log.Logf("detected %s hardware", provider.GetName())
			detectedProviders = append(detectedProviders, provider)
		}
	}
	if len(detectedProviders) == 0 {
		return fmt.Errorf("no compatible hardware found")
	}

	needSupplementary := slices.ContainsFunc(detectedProviders, func(p api.Provider) bool {
		return p.GetID() == "nvidia"
	})
	if err := deps.RepositoryManager.EnsureRepositoriesEnabled(needSupplementary); err != nil {
		return fmt.Errorf("failed to verify/enable repositories: %w", err)
	}

	var toInstall []api.DriverID
	for _, provider := range detectedProviders {
		available, err := provider.ListAvailable()
		if err != nil {
			return fmt.Errorf("failed to list available %s drivers: %w", provider.GetName(), err)
		}
		if len(available) > 0 {
			// Pick the latest version (providers return sorted newest-first)
			toInstall = append(toInstall, available[0])
		}
	}
	if len(toInstall) == 0 {
		return fmt.Errorf("no drivers available for detected hardware")
	}

	return doInstall(deps, toInstall, batchMode, dryRun)
}

func doInstall(deps api.CoreDeps, toInstall []api.DriverID, batchMode, dryRun bool) error {
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
	if err := deps.PackageManager.Install(allPkgs, batchMode, dryRun); err != nil {
		return fmt.Errorf("failed to install pacakges: %w", err)
	}
	return nil
}
