package hwdetect

import (
	"io"
	stdlog "log"
	"os"
	"testing"

	internalLog "github.com/mizdebsk/radii/internal/log"
)

func TestNormalizeID(t *testing.T) {
	tests := []struct {
		name string
		in   string
		out  string
	}{
		{
			name: "hex with 0x prefix upper",
			in:   "0x31C2",
			out:  "31c2",
		},
		{
			name: "hex with 0x prefix lower",
			in:   "0x1ad3",
			out:  "1ad3",
		},
		{
			name: "hex without prefix",
			in:   "1AD3",
			out:  "1ad3",
		},
		{
			name: "long id keeps last four",
			in:   "0x000031C2",
			out:  "31c2",
		},
		{
			name: "spaces trimmed",
			in:   "  0x31C2  ",
			out:  "31c2",
		},
		{
			name: "empty string",
			in:   "",
			out:  "",
		},
		{
			name: "only 0x prefix",
			in:   "0x",
			out:  "",
		},
		{
			name: "short id (less than four)",
			in:   "0x1a",
			out:  "1a",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			got := NormalizeID(tt.in)
			if got != tt.out {
				t.Fatalf("NormalizeID(%q) = %q, want %q", tt.in, got, tt.out)
			}
		})
	}
}

func TestParsePCIModaliasSilentlyRejectsNonPCIInput(t *testing.T) {
	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatalf("Pipe() error = %v", err)
	}
	originalStderr := os.Stderr
	originalDebug := internalLog.Debug
	os.Stderr = writer
	internalLog.Debug = true
	t.Cleanup(func() {
		internalLog.Debug = originalDebug
		os.Stderr = originalStderr
		stdlog.SetOutput(originalStderr)
		_ = reader.Close()
		_ = writer.Close()
	})

	if _, valid := ParsePCIModalias("usb:v1234p5678"); valid {
		t.Fatal("ParsePCIModalias() unexpectedly accepted a non-PCI modalias")
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}
	output, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("ReadAll() error = %v", err)
	}
	if len(output) != 0 {
		t.Fatalf("non-PCI modalias produced debug output: %s", output)
	}
}
