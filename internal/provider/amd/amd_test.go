package amd

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/mizdebsk/radii/internal/api"
)

type fakePackageManager struct {
	installed []api.PackageInfo
	available []api.PackageInfo
}

func (f fakePackageManager) Install(packages []string, batchMode, dryRun bool) error {
	return nil
}

func (f fakePackageManager) Remove(packages []string, batchMode, dryRun bool) error {
	return nil
}

func (f fakePackageManager) ListInstalledPackages() ([]api.PackageInfo, error) {
	return f.installed, nil
}

func (f fakePackageManager) ListAvailablePackages() ([]api.PackageInfo, error) {
	return f.available, nil
}

func TestInstallAndRemoveLatest(t *testing.T) {
	p := NewProvider(fakePackageManager{}).(*prov)

	gotInstall, err := p.Install([]api.DriverID{{ProviderID: "amdgpu", Version: variantLatest}})
	if err != nil {
		t.Fatalf("Install() error = %v", err)
	}
	wantInstall := []string{pkgKmodAmdgpu, pkgRocm}
	if !reflect.DeepEqual(gotInstall, wantInstall) {
		t.Fatalf("Install() = %v, want %v", gotInstall, wantInstall)
	}

	gotRemove, err := p.Remove([]api.DriverID{{ProviderID: "amdgpu", Version: variantLatest}})
	if err != nil {
		t.Fatalf("Remove() error = %v", err)
	}
	if !reflect.DeepEqual(gotRemove, wantInstall) {
		t.Fatalf("Remove() = %v, want %v", gotRemove, wantInstall)
	}
}

func TestInstallUnknownVariant(t *testing.T) {
	p := NewProvider(fakePackageManager{}).(*prov)
	if _, err := p.Install([]api.DriverID{{ProviderID: "amdgpu", Version: "unknown"}}); err == nil {
		t.Fatal("Install() error = nil, want non-nil")
	}
}

func TestListAvailableAndInstalledLatest(t *testing.T) {
	p := NewProvider(fakePackageManager{
		installed: []api.PackageInfo{
			{Name: pkgKmodAmdgpu},
			{Name: pkgRocm},
		},
		available: []api.PackageInfo{
			{Name: pkgKmodAmdgpu},
			{Name: pkgRocm},
		},
	}).(*prov)

	gotInstalled, err := p.ListInstalled()
	if err != nil {
		t.Fatalf("ListInstalled() error = %v", err)
	}
	want := []api.DriverID{{ProviderID: "amdgpu", Version: variantLatest}}
	if !reflect.DeepEqual(gotInstalled, want) {
		t.Fatalf("ListInstalled() = %v, want %v", gotInstalled, want)
	}

	gotAvailable, err := p.ListAvailable()
	if err != nil {
		t.Fatalf("ListAvailable() error = %v", err)
	}
	if !reflect.DeepEqual(gotAvailable, want) {
		t.Fatalf("ListAvailable() = %v, want %v", gotAvailable, want)
	}
}

func TestListInstalledWithPartialPackages(t *testing.T) {
	tests := []struct {
		name      string
		installed []api.PackageInfo
	}{
		{
			name:      "kernel package only",
			installed: []api.PackageInfo{{Name: pkgKmodAmdgpu}},
		},
		{
			name:      "rocm package only",
			installed: []api.PackageInfo{{Name: pkgRocm}},
		},
	}

	want := []api.DriverID{{ProviderID: "amdgpu", Version: variantLatest}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewProvider(fakePackageManager{
				installed: tt.installed,
			}).(*prov)

			gotInstalled, err := p.ListInstalled()
			if err != nil {
				t.Fatalf("ListInstalled() error = %v", err)
			}
			if !reflect.DeepEqual(gotInstalled, want) {
				t.Fatalf("ListInstalled() = %v, want %v", gotInstalled, want)
			}
		})
	}
}

func TestListAvailableRequiresBothPackages(t *testing.T) {
	p := NewProvider(fakePackageManager{
		available: []api.PackageInfo{{Name: pkgRocm}},
	}).(*prov)

	gotAvailable, err := p.ListAvailable()
	if err != nil {
		t.Fatalf("ListAvailable() error = %v", err)
	}
	if len(gotAvailable) != 0 {
		t.Fatalf("ListAvailable() = %v, want empty result", gotAvailable)
	}
}

func TestIsCompatibleAMDHardware_Found(t *testing.T) {
	d := newAutoDetector()

	tests := []struct {
		name  string
		modal string
	}{
		{
			name:  "display class MI210 real hardware sample",
			modal: "pci:v00001002d0000740Fsv00001002sd00000C34bc03sc80i00",
		},
		{
			name:  "display class real hardware sample",
			modal: "pci:v00001002d0000163Fsv00001002sd00000123bc03sc00i00",
		},
		{
			name:  "accelerator class MI300X real hardware sample",
			modal: "pci:v00001002d000074A1sv00001002sd000074A1bc12sc00i00",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !d.isCompatibleAMDHardware(tt.modal) {
				t.Fatalf("expected modalias %q to be detected as compatible AMD hardware", tt.modal)
			}
		})
	}
}

func TestIsCompatibleAMDHardware_NotFoundOrInvalid(t *testing.T) {
	d := newAutoDetector()

	tests := []struct {
		name    string
		modal   string
		wantHit bool
	}{
		{
			name:    "non-amd vendor",
			modal:   "pci:v000010DEd00007440sv000010DEsd00000000bc03sc00i00",
			wantHit: false,
		},
		{
			name:    "unsupported class",
			modal:   "pci:v00001002d00007440sv00001002sd00000000bc02sc00i00",
			wantHit: false,
		},
		{
			name:    "invalid format",
			modal:   "this-is-not-a-pci-modalias",
			wantHit: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := d.isCompatibleAMDHardware(tt.modal)
			if got != tt.wantHit {
				t.Fatalf("isCompatibleAMDHardware(%q) = %v, want %v", tt.modal, got, tt.wantHit)
			}
		})
	}
}

func TestDetect_WithCompatibleSysfs(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "pci0000:00", "0000:00:01.0"), 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	if err := os.WriteFile(
		filepath.Join(root, "pci0000:00", "0000:00:01.0", "modalias"),
		[]byte("pci:v00001002d000074A1sv00001002sd000074A1bc12sc00i00\n"),
		0o644,
	); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	d := newAutoDetector()
	d.modaliasRoot = root

	found, err := d.Detect()
	if err != nil {
		t.Fatalf("Detect() error = %v", err)
	}
	if !found {
		t.Fatalf("Detect() = %v, want true", found)
	}
}

func TestDetect_WithEightMI300XModaliases(t *testing.T) {
	root := t.TempDir()
	for i := range 8 {
		dir := filepath.Join(root, "pci0000:00", fmt.Sprintf("0000:%02x:00.0", i+1))
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatalf("MkdirAll() error = %v", err)
		}
		if err := os.WriteFile(
			filepath.Join(dir, "modalias"),
			[]byte("pci:v00001002d000074A1sv00001002sd000074A1bc12sc00i00\n"),
			0o644,
		); err != nil {
			t.Fatalf("WriteFile() error = %v", err)
		}
	}

	d := newAutoDetector()
	d.modaliasRoot = root

	found, err := d.Detect()
	if err != nil {
		t.Fatalf("Detect() error = %v", err)
	}
	if !found {
		t.Fatalf("Detect() = %v, want true", found)
	}
}
