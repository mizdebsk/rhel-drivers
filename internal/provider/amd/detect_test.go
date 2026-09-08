package amd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/mizdebsk/radii/internal/hwdetect"
)

func TestIsCompatibleAMDHardware(t *testing.T) {
	tests := []struct {
		name     string
		modalias string
		want     bool
	}{
		{
			name:     "MI210 display class",
			modalias: "pci:v00001002d0000740Fsv00001002sd00000C34bc03sc80i00",
			want:     true,
		},
		{
			name:     "MI300X accelerator class",
			modalias: "pci:v00001002d000074A1sv00001002sd000074A1bc12sc00i00",
			want:     true,
		},
		{
			name:     "non-AMD vendor",
			modalias: "pci:v000010DEd00007440sv000010DEsd00000000bc03sc00i00",
			want:     false,
		},
		{
			name:     "unsupported class",
			modalias: "pci:v00001002d00007440sv00001002sd00000000bc02sc00i00",
			want:     false,
		},
		{
			name:     "invalid modalias",
			modalias: "this-is-not-a-pci-modalias",
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			modalias, valid := hwdetect.ParsePCIModalias(tt.modalias)
			got := valid && isCompatibleAMDHardware(modalias)
			if got != tt.want {
				t.Fatalf("compatible AMD modalias %q = %v, want %v", tt.modalias, got, tt.want)
			}
		})
	}
}

func TestDetectHardwareWithMI300XSysfs(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "pci0000:00", "0000:01:00.0", "modalias")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	if err := os.WriteFile(path, []byte("pci:v00001002d000074A1sv00001002sd000074A1bc12sc00i00\n"), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	found, err := detectHardware(root)
	if err != nil {
		t.Fatalf("detectHardware() error = %v", err)
	}
	if !found {
		t.Fatal("detectHardware() = false, want true")
	}
}
