package exec

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
)

func RunCommand(ctx context.Context, command string, args []string) error {
	cmd := exec.CommandContext(ctx, command, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%s command failed: %w", command, err)
	}
	return nil
}

func RunCommandCapture(ctx context.Context, command string, args ...string) ([]string, error) {
	cmd := exec.CommandContext(ctx, command, args...)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to get stdout for %s command: %w", command, err)
	}
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to start %s command: %w", command, err)
	}

	var lines []string
	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		_ = cmd.Wait()
		return nil, fmt.Errorf("error reading %s output: %w", command, err)
	}

	if err := cmd.Wait(); err != nil {
		return nil, fmt.Errorf("%s command failed: %w", command, err)
	}

	return lines, nil
}
