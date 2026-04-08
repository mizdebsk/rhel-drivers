package amd

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
	defaultModaliasRoot = "/sys/devices"

	modaliasBus         = "pci"
	amdVendor           = "1002"
	pciClassDisplay     = "03"
	pciClassAccelerator = "12"
)

type autoDetector struct {
	modaliasRoot string
}

func newAutoDetector() autoDetector {
	return autoDetector{
		modaliasRoot: defaultModaliasRoot,
	}
}

func (d *autoDetector) Detect() (bool, error) {
	found, err := d.scanModaliases()
	if err != nil {
		return false, err
	}
	return found, nil
}

func (d *autoDetector) scanModaliases() (bool, error) {
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
		if d.isCompatibleAMDHardware(modal) {
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
		log.Logf("compatible AMD hardware was found")
	} else {
		log.Logf("compatible AMD hardware was NOT found")
	}
	return found, nil
}

var modaliasRe = regexp.MustCompile(
	`^(pci):` +
		`v([0-9A-Fa-f]{8})` +
		`d([0-9A-Fa-f]{8})` +
		`sv([0-9A-Fa-f]{8})` +
		`sd([0-9A-Fa-f]{8})` +
		`bc([0-9A-Fa-f]{2})` +
		`sc([0-9A-Fa-f]{2})` +
		`i([0-9A-Fa-f]{2})$`,
)

func (d *autoDetector) isCompatibleAMDHardware(modalias string) bool {
	if !strings.HasPrefix(modalias, modaliasBus+":") {
		return false
	}
	m := modaliasRe.FindStringSubmatch(modalias)
	if m == nil {
		log.Debugf("invalid modalias: %s", modalias)
		return false
	}

	vendor := strings.ToLower(m[2])
	device := strings.ToLower(m[3])
	baseClass := strings.ToLower(m[6])

	vendor4 := vendor[len(vendor)-4:]
	device4 := device[len(device)-4:]

	if vendor4 != amdVendor {
		return false
	}
	if baseClass != pciClassDisplay && baseClass != pciClassAccelerator {
		return false
	}

	log.Infof("found compatible AMD hardware: device=%s class=%s", device4, baseClass)
	return true
}
