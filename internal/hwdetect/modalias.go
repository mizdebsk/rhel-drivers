package hwdetect

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/mizdebsk/radii/internal/log"
)

const (
	DefaultModaliasRoot = "/sys/devices"

	PCIClassDisplay = "03"
)

type PCIModalias struct {
	VendorID  string
	DeviceID  string
	BaseClass string
}

func NormalizeID(id string) string {
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

var pciModaliasRe = regexp.MustCompile(
	`^(pci):` + // bus
		`v([0-9A-Fa-f]{8})` + // vendor
		`d([0-9A-Fa-f]{8})` + // device
		`sv([0-9A-Fa-f]{8})` + // subvendor
		`sd([0-9A-Fa-f]{8})` + // subdevice
		`bc([0-9A-Fa-f]{2})` + // base class
		`sc([0-9A-Fa-f]{2})` + // subclass
		`i([0-9A-Fa-f]{2})$`, // interface
)

func ParsePCIModalias(modalias string) (PCIModalias, bool) {
	modalias = strings.TrimSpace(modalias)
	modalias = strings.ToLower(modalias)
	if !strings.HasPrefix(modalias, "pci:") {
		return PCIModalias{}, false
	}
	m := pciModaliasRe.FindStringSubmatch(modalias)
	if m == nil {
		log.Debugf("invalid modalias: %s", modalias)
		return PCIModalias{}, false
	}

	return PCIModalias{
		VendorID:  NormalizeID(m[2]),
		DeviceID:  NormalizeID(m[3]),
		BaseClass: strings.ToLower(m[6]),
	}, true
}

func ScanPCIModaliases(root, providerName string, match func(PCIModalias) bool) (bool, error) {
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
		modalias, valid := ParsePCIModalias(modal)
		if !valid || !match(modalias) {
			return nil
		}

		log.Logf("modalias path: %s", path)
		log.Logf("modalias entry: %s", modal)
		found = true
		return nil
	}

	log.Logf("scanning modalias files in %s", root)
	err := filepath.WalkDir(root, walkFn)
	if err != nil {
		return false, fmt.Errorf("error scanning modalias files in %s: %w", root, err)
	}
	if found {
		log.Logf("compatible %s hardware was found", providerName)
	} else {
		log.Logf("compatible %s hardware was NOT found", providerName)
	}
	return found, nil
}
