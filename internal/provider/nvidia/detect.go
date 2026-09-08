package nvidia

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/mizdebsk/radii/internal/hwdetect"
	"github.com/mizdebsk/radii/internal/log"
)

const (
	defaultHwdataJsonPath = "/usr/share/radii-hwdata-nv/hwdata-nv.json"
	nvidiaVendor          = "10de"
)

type autoDetector struct {
	hwdataJsonPath string
	modaliasRoot   string
}

func newAutoDetector() autoDetector {
	return autoDetector{
		hwdataJsonPath: defaultHwdataJsonPath,
		modaliasRoot:   hwdetect.DefaultModaliasRoot,
	}
}

func (d *autoDetector) Detect() (bool, error) {
	compatible, err := d.loadCompatibleDevices()
	if err != nil {
		return false, err
	}
	if len(compatible) == 0 {
		return false, nil
	}

	found, err := hwdetect.ScanPCIModaliases(d.modaliasRoot, "NVIDIA", func(modalias hwdetect.PCIModalias) bool {
		return d.isCompatibleNvidiaDisplay(modalias, compatible)
	})
	if err != nil {
		return false, err
	}
	return found, nil
}

type hardwareDatabase struct {
	Chips []struct {
		Name     string   `json:"name"`
		DevID    string   `json:"devid"`
		Features []string `json:"features"`
	} `json:"chips"`
}

func (d *autoDetector) loadCompatibleDevices() (map[string]string, error) {
	log.Logf("loading hardware database from %s", d.hwdataJsonPath)
	hwdataJson, err := os.ReadFile(d.hwdataJsonPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("cannot find hardware database file: %s", d.hwdataJsonPath)
		}
		return nil, fmt.Errorf("failed to read hardware database file %s: %w", d.hwdataJsonPath, err)
	}

	var hwdata hardwareDatabase
	if err := json.Unmarshal(hwdataJson, &hwdata); err != nil {
		return nil, fmt.Errorf("failed to parse hardware database file %s: %w", d.hwdataJsonPath, err)
	}

	compatible := make(map[string]string)
	for _, chip := range hwdata.Chips {
		if !hasFeature(chip.Features, "kernelopen") {
			continue
		}
		dev := hwdetect.NormalizeID(chip.DevID)
		if dev == "" {
			continue
		}
		compatible[dev] = chip.Name
	}

	return compatible, nil
}

func hasFeature(features []string, name string) bool {
	for _, f := range features {
		if f == name {
			return true
		}
	}
	return false
}

func (d *autoDetector) isCompatibleNvidiaDisplay(modalias hwdetect.PCIModalias, compatible map[string]string) bool {
	if modalias.VendorID != nvidiaVendor {
		return false
	}
	if modalias.BaseClass != hwdetect.PCIClassDisplay {
		return false
	}

	if name, ok := compatible[modalias.DeviceID]; ok {
		log.Infof("found compatible hardware: %s", name)
		return true
	}
	return false
}
