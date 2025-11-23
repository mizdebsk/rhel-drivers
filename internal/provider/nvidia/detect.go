package nvidia

import (
	"context"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/mizdebsk/rhel-drivers/internal/log"
)

const (
	defaultCompatibleGPUsPath = "/usr/share/rhel-drivers/nvidia/supported-gpus.json"
	defaultModaliasRoot       = "/sys/devices"

	modaliasBus     = "pci"
	pciClassDisplay = "03"
	nvidiaVendor    = "10de"
)

type autoDetector struct {
	compatibleGPUs string
	modaliasRoot   string
}

func newAutoDetector() autoDetector {
	return autoDetector{
		compatibleGPUs: defaultCompatibleGPUsPath,
		modaliasRoot:   defaultModaliasRoot,
	}
}

func (d *autoDetector) Detect(ctx context.Context) (bool, error) {
	compatible, err := d.loadCompatibleDevices()
	if err != nil {
		return false, err
	}
	if len(compatible) == 0 {
		return false, nil
	}

	found, err := d.scanModaliases(ctx, compatible)
	if err != nil {
		return false, err
	}
	return found, nil
}

type compatibleGPUFile struct {
	Chips []struct {
		Name     string   `json:"name"`
		DevID    string   `json:"devid"`
		Features []string `json:"features"`
	} `json:"chips"`
}

func (d *autoDetector) loadCompatibleDevices() (map[string]string, error) {
	log.Logf("loading compatible GPUs from %s", d.compatibleGPUs)
	data, err := os.ReadFile(d.compatibleGPUs)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("cannot find compatible GPUs file: %s", d.compatibleGPUs)
		}
		return nil, fmt.Errorf("failed to read compatible GPUs file %s: %w", d.compatibleGPUs, err)
	}

	var s compatibleGPUFile
	if err := json.Unmarshal(data, &s); err != nil {
		return nil, fmt.Errorf("failed to parse compatible GPUs file %s: %w", d.compatibleGPUs, err)
	}

	result := make(map[string]string)
	for _, chip := range s.Chips {
		if !hasFeature(chip.Features, "kernelopen") {
			continue
		}
		dev := normalizeDevID(chip.DevID)
		if dev == "" {
			continue
		}
		result[dev] = chip.Name
	}

	return result, nil
}

func hasFeature(features []string, name string) bool {
	for _, f := range features {
		if f == name {
			return true
		}
	}
	return false
}

func normalizeDevID(id string) string {
	id = strings.TrimSpace(id)
	id = strings.ToLower(id)
	id = strings.TrimPrefix(id, "0x")
	if len(id) == 0 {
		return ""
	}
	if len(id) > 4 {
		return id[len(id)-4:]
	}
	return id
}

func (d *autoDetector) scanModaliases(ctx context.Context, compatible map[string]string) (bool, error) {
	found := false

	walkFn := func(path string, de fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if de.IsDir() {
			return nil
		}
		if de.Name() != "modalias" {
			return nil
		}

		content, err := os.ReadFile(path)
		if err != nil {
			log.Errorf("failed to read modalias %s: %v", path, err)
			return nil
		}
		modal := strings.TrimSpace(strings.ToLower(string(content)))
		if d.isCompatibleNvidiaDisplay(modal, compatible) {
			log.Logf("modalias path: %s", path)
			log.Logf("modalias entry: %s", modal)
			found = true
		}

		return nil
	}

	log.Logf("scanning modalias files in %s", d.modaliasRoot)
	err := filepath.WalkDir(d.modaliasRoot, walkFn)
	if err != nil {
		return false, fmt.Errorf("error scanning modalias files in %s: %w", d.modaliasRoot, err)
	}
	if found {
		log.Logf("compatible NVIDIA hardware was found")
	} else {
		log.Logf("compatible NVIDIA hardware was NOT found")
	}
	return found, nil
}

var modaliasRe = regexp.MustCompile(
	`^(pci):` + // bus
		`v([0-9A-Fa-f]{8})` + // vendor
		`d([0-9A-Fa-f]{8})` + // device
		`sv([0-9A-Fa-f]{8})` + // subvendor
		`sd([0-9A-Fa-f]{8})` + // subdevice
		`bc([0-9A-Fa-f]{2})` + // base class
		`sc([0-9A-Fa-f]{2})` + // subclass
		`i([0-9A-Fa-f]{2})$`, // interface
)

func (d *autoDetector) isCompatibleNvidiaDisplay(modalias string, compatible map[string]string) bool {
	if !strings.HasPrefix(modalias, modaliasBus+":") {
		return false
	}
	m := modaliasRe.FindStringSubmatch(modalias)
	if m == nil {
		log.Debugf("invalid modalias: %s", modalias)
		return false
	}

	//bus := m[1]
	vendor := strings.ToLower(m[2])
	device := strings.ToLower(m[3])
	//subVendor := strings.ToLower(m[4])
	//subDevice := strings.ToLower(m[5])
	baseClass := strings.ToLower(m[6])
	//subClass := strings.ToLower(m[7])
	//iface := strings.ToLower(m[8])

	vendor4 := vendor[len(vendor)-4:]
	device4 := device[len(device)-4:]

	if vendor4 != nvidiaVendor {
		return false
	}
	if baseClass != pciClassDisplay {
		return false
	}

	if name, ok := compatible[device4]; ok {
		log.Infof("found compatible hardware: %s", name)
		return true
	}
	return false
}
