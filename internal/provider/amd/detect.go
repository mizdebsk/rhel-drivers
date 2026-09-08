package amd

import (
	"github.com/mizdebsk/radii/internal/hwdetect"
	"github.com/mizdebsk/radii/internal/log"
)

const amdVendor = "1002"

func detectHardware(modaliasRoot string) (bool, error) {
	return hwdetect.ScanPCIModaliases(modaliasRoot, "AMD", isCompatibleAMDHardware)
}

func isCompatibleAMDHardware(modalias hwdetect.PCIModalias) bool {
	if modalias.VendorID != amdVendor {
		return false
	}
	if modalias.BaseClass != hwdetect.PCIClassDisplay &&
		modalias.BaseClass != hwdetect.PCIClassProcessingAccelerator {
		return false
	}

	log.Infof("found compatible AMD hardware: device=%s class=%s", modalias.DeviceID, modalias.BaseClass)
	return true
}
