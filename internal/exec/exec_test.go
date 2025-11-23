package exec

import (
	"context"
	"strings"
	"testing"
)

func TestRunCommand(t *testing.T) {
	tests := []struct {
		name      string
		command   string
		args      []string
		expectErr bool
	}{
		{
			name:      "True",
			command:   "true",
			args:      []string{},
			expectErr: false,
		},
		{
			name:      "False",
			command:   "false",
			args:      []string{},
			expectErr: true,
		},
		{
			name:      "CommandNotFound",
			command:   "blah-blah-blah-this-command-does-not-exist",
			args:      []string{},
			expectErr: true,
		},
		{
			name:      "ShellExit0",
			command:   "sh",
			args:      []string{"-c", "exit 0"},
			expectErr: false,
		},
		{
			name:      "ShellExit42",
			command:   "sh",
			args:      []string{"-c", "exit 42"},
			expectErr: true,
		},
		{
			name:      "ShellKillSigPipe",
			command:   "sh",
			args:      []string{"-c", "kill -PIPE $$ || :"},
			expectErr: true,
		},
		{
			name:      "InvalidArgument",
			command:   "sh",
			args:      []string{"-c", "exit 0;\000"},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			err := RunCommand(ctx, tt.command, tt.args)
			if (err != nil) != tt.expectErr {
				t.Errorf("Expected error: %v, but got: %v", tt.expectErr, err)
			}
		})
	}
}

func TestRunCommandCapture(t *testing.T) {
	tests := []struct {
		name      string
		command   string
		args      []string
		expectErr bool
		expectOut []string
	}{
		{
			name:      "EchoSingle",
			command:   "echo",
			args:      []string{"\t\r\fHello, World!  "},
			expectErr: false,
			expectOut: []string{"\t\r\fHello, World!  "},
		},
		{
			name:      "EchoMultiple",
			command:   "sh",
			args:      []string{"-c", "echo line1; echo line2; echo line3"},
			expectErr: false,
			expectOut: []string{"line1", "line2", "line3"},
		},
		{
			name:      "CommandNotFound",
			command:   "blah-blah-blah-this-command-does-not-exist",
			args:      []string{},
			expectErr: true,
			expectOut: nil,
		},
		{
			name:      "EmptyOutput",
			command:   "true",
			args:      []string{},
			expectErr: false,
			expectOut: []string{},
		},
		{
			name:      "FailingCommand",
			command:   "sh",
			args:      []string{"-c", "echo 'BOOM!' && exit 1"},
			expectErr: true,
			expectOut: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			output, err := RunCommandCapture(ctx, tt.command, tt.args...)

			if (err != nil) != tt.expectErr {
				t.Errorf("Expected error: %v, but got: %v", tt.expectErr, err)
			}

			if !strings.EqualFold(strings.Join(output, "\n"), strings.Join(tt.expectOut, "\n")) {
				t.Errorf("Expected output: %v, but got: %v", tt.expectOut, output)
			}
		})
	}
}
