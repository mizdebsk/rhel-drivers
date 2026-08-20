package core

import (
	"fmt"

	"github.com/mizdebsk/radii/internal/api"
	"github.com/mizdebsk/radii/internal/log"
)

func List(deps api.CoreDeps, listInst, listAvail, hwdetect, compatibleOnly bool) ([]api.DriverStatus, error) {
	var result []api.DriverStatus

	if compatibleOnly {
		hwdetect = true
	}

	if listAvail {
		if err := deps.RepositoryManager.EnsureRepositoriesEnabled(true); err != nil {
			return result, fmt.Errorf("failed to verify/enable repositories: %w", err)
		}
	}

	for _, provider := range deps.Providers {
		var compat bool
		if hwdetect {
			var err error
			compat, err = provider.DetectHardware()
			if err != nil {
				log.Warnf("hardware detection failed for %s failed: %v", provider.GetName(), err)
			}
		}
		var installed []api.DriverID
		if listInst {
			var err error
			installed, err = provider.ListInstalled()
			if err != nil {
				return result, fmt.Errorf("failed to check installed %s drivers: %w", provider.GetName(), err)
			}
			if len(installed) > 0 {
				log.Logf("Currently installed %d %s drivers", len(installed), provider.GetName())
			} else {
				log.Logf("%s driver is currently NOT installed", provider.GetName())
			}
		}
		var available []api.DriverID
		if listAvail {
			var err error
			available, err = provider.ListAvailable()
			if err != nil {
				return result, fmt.Errorf("failed to check available %s drivers: %w", provider.GetName(), err)
			}
			if len(available) > 0 {
				log.Logf("Currently available %d %s drivers", len(available), provider.GetName())
			} else {
				log.Logf("%s driver is currently NOT available", provider.GetName())
			}
		}
		var all []string
		installedSet := make(map[string]struct{})
		availableSet := make(map[string]struct{})
		for _, avail := range available {
			ver := avail.Version
			all = append(all, ver)
			availableSet[ver] = struct{}{}
		}
		for _, inst := range installed {
			ver := inst.Version
			if _, ok := availableSet[ver]; !ok {
				all = append(all, ver)
			}
			installedSet[ver] = struct{}{}
		}
		for _, ver := range all {
			_, inst := installedSet[ver]
			_, avail := availableSet[ver]
			result = append(result,
				api.DriverStatus{
					ID:         api.DriverID{ProviderID: provider.GetID(), Version: ver},
					Available:  avail,
					Installed:  inst,
					Compatible: compat,
				})
		}
	}

	if compatibleOnly {
		result = filterCompatible(result)
	}

	return result, nil
}

// filterCompatible returns only driver statuses that are compatible with the current hardware.
// It is used internally by List when compatibleOnly is true.
func filterCompatible(res []api.DriverStatus) []api.DriverStatus {
	var out []api.DriverStatus
	for _, dev := range res {
		if dev.Compatible {
			out = append(out, dev)
		}
	}
	return out
}
