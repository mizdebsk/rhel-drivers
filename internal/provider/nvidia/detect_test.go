package nvidia

import (
	"testing"

	"github.com/mizdebsk/radii/internal/hwdetect"
)

func TestHasFeature(t *testing.T) {
	tests := []struct {
		name     string
		features []string
		search   string
		want     bool
	}{
		{
			name:     "feature present",
			features: []string{"foo", "kernelopen", "bar"},
			search:   "kernelopen",
			want:     true,
		},
		{
			name:     "feature absent",
			features: []string{"foo", "bar"},
			search:   "kernelopen",
			want:     false,
		},
		{
			name:     "empty list",
			features: nil,
			search:   "kernelopen",
			want:     false,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			got := hasFeature(tt.features, tt.search)
			if got != tt.want {
				t.Fatalf("hasFeature(%v, %q) = %v, want %v", tt.features, tt.search, got, tt.want)
			}
		})
	}
}

func TestIsCompatibleNvidiaDisplay_Found(t *testing.T) {
	d := newAutoDetector()
	compatible := map[string]string{
		"31c2": "NVIDIA A100-PCIE-40GB",
	}
	modal := "pci:v000010DEd000031C2sv000010DEsd000013C2bc03sc00i00"
	parsed, valid := hwdetect.ParsePCIModalias(modal)
	if !valid {
		t.Fatalf("expected modalias %q to be valid", modal)
	}
	if !d.isCompatibleNvidiaDisplay(parsed, compatible) {
		t.Fatalf("expected modalias %q to be detected as compatible NVIDIA display", modal)
	}
}

func TestIsCompatibleNvidiaDisplay_NotFoundOrInvalid(t *testing.T) {
	d := newAutoDetector()
	compatible := map[string]string{
		"31c2": "NVIDIA A100-PCIE-40GB",
	}
	tests := []struct {
		name    string
		modal   string
		wantHit bool
	}{
		{
			name:    "non-nvidia vendor",
			modal:   "pci:v00001000d000031C2sv00001000sd000013C2bc03sc00i00",
			wantHit: false,
		},
		{
			name:    "non-display class",
			modal:   "pci:v000010DEd000031C2sv000010DEsd000013C2bc02sc00i00",
			wantHit: false,
		},
		{
			name:    "uncompatible device id",
			modal:   "pci:v000010DEd0000DEADsv000010DEsd000013C2bc03sc00i00",
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
			modal, valid := hwdetect.ParsePCIModalias(tt.modal)
			got := valid && d.isCompatibleNvidiaDisplay(modal, compatible)
			if got != tt.wantHit {
				t.Fatalf("isCompatibleNvidiaDisplay(%q) = %v, want %v", tt.modal, got, tt.wantHit)
			}
		})
	}
}

func TestLoadCompatibleDevices_UsesTestJSON(t *testing.T) {
	d := newAutoDetector()
	d.hwdataJsonPath = "testdata/hwdata.json"
	compatible, err := d.loadCompatibleDevices()
	if err != nil {
		t.Fatalf("loadCompatibleDevices() error = %v", err)
	}
	if len(compatible) == 0 {
		t.Fatalf("loadCompatibleDevices() returned 0 compatible devices, want > 0")
	}
}

func TestDetect_WithA100SysfsAndHwdata(t *testing.T) {
	d := newAutoDetector()
	d.hwdataJsonPath = "testdata/hwdata.json"
	d.modaliasRoot = "testdata/sysfs-A100-PCIE-40GB"
	found, err := d.Detect()
	if err != nil {
		t.Fatalf("Detect() error = %v", err)
	}
	if !found {
		t.Fatalf("Detect() = %v, want true (A100 PCIe 40GB should be detected)", found)
	}
}
