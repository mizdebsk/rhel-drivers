package dnf

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"github.com/mizdebsk/rhel-drivers/internal/api"
	"github.com/mizdebsk/rhel-drivers/internal/cache"
	"github.com/mizdebsk/rhel-drivers/internal/log"
	"strings"
)

const defaultDNFBinary = "dnf"

type Manager struct {
	Bin string
}

var _ api.PackageManager = (*Manager)(nil)

func New() *Manager {
	return &Manager{
		Bin: defaultDNFBinary,
	}
}

var availableCache = cache.Cache[[]api.PackageInfo]{}
var installedCache = cache.Cache[[]api.PackageInfo]{}

func (m *Manager) ListAvailablePackages(ctx context.Context) ([]api.PackageInfo, error) {
	return availableCache.Get(ctx, func(ctx context.Context) ([]api.PackageInfo, error) {
		return m.doListAvailablePackages(ctx)
	})
}

func (m *Manager) ListInstalledPackages(ctx context.Context) ([]api.PackageInfo, error) {
	return installedCache.Get(ctx, func(ctx context.Context) ([]api.PackageInfo, error) {
		return m.doListInstalledPackages(ctx)
	})
}

func (m *Manager) doListAvailablePackages(ctx context.Context) ([]api.PackageInfo, error) {
	tags := []string{"name", "epoch", "version", "release", "arch", "sourcerpm", "repoid"}
	// QQQ and YYY are there to make filtering spurious lines easier.
	format := "QQQ"
	for _, field := range tags {
		format += "|%{" + field + "}"
	}
	// Trailing NL is not required with DNF 4, but will be required with DNF 5.
	// With DNF 4 it will result in empty lines, but they are ignored anyway.
	format += "|YYY\n"

	args := []string{"-q", "repoquery", "--qf", format}
	cmd := exec.CommandContext(ctx, m.Bin, args...)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to get stdout for dnf repoquery: %w", err)
	}
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to start dnf repoquery: %w", err)
	}

	var infos []api.PackageInfo
	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if !strings.HasPrefix(line, "QQQ|") || !strings.HasSuffix(line, "|YYY") {
			continue
		}
		line = strings.TrimPrefix(line, "QQQ|")
		line = strings.TrimSuffix(line, "|YYY")
		fields := strings.SplitN(line, "|", len(tags))
		if len(fields) != len(tags) {
			continue
		}
		name := fields[0]
		epoch := fields[1]
		version := fields[2]
		release := fields[3]
		arch := fields[4]
		sourcerpm := fields[5]
		repoid := fields[6]

		sourceName := extractSourceName(sourcerpm)

		infos = append(infos, api.PackageInfo{
			Name:       name,
			Epoch:      epoch,
			Version:    version,
			Release:    release,
			Arch:       arch,
			SourceName: sourceName,
			Repo:       repoid,
		})
	}

	if err := scanner.Err(); err != nil {
		_ = cmd.Wait()
		return nil, fmt.Errorf("error reading dnf repoquery output: %w", err)
	}

	if err := cmd.Wait(); err != nil {
		return nil, fmt.Errorf("dnf repoquery failed: %w", err)
	}

	log.Logf("found %d available packages", len(infos))
	return infos, nil
}

func (m *Manager) doListInstalledPackages(ctx context.Context) ([]api.PackageInfo, error) {
	tags := []string{"NAME", "EPOCH", "VERSION", "RELEASE", "ARCH", "SOURCERPM"}
	format := "QQQ"
	for _, field := range tags {
		format += "|%|" + field + "?{%{" + field + "}}|"
	}
	format += "|YYY\n"

	cmd := exec.CommandContext(ctx, "rpm", "-qa", "--qf", format)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to get stdout for rpm -qa: %w", err)
	}
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to start rpm -qa: %w", err)
	}

	var infos []api.PackageInfo
	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if !strings.HasPrefix(line, "QQQ|") || !strings.HasSuffix(line, "|YYY") {
			continue
		}
		line = strings.TrimPrefix(line, "QQQ|")
		line = strings.TrimSuffix(line, "|YYY")
		fields := strings.SplitN(line, "|", len(tags))
		if len(fields) != len(tags) {
			continue
		}
		name := fields[0]
		epoch := fields[1]
		version := fields[2]
		release := fields[3]
		arch := fields[4]
		sourcerpm := fields[5]

		sourceName := extractSourceName(sourcerpm)

		infos = append(infos, api.PackageInfo{
			Name:       name,
			Epoch:      epoch,
			Version:    version,
			Release:    release,
			Arch:       arch,
			SourceName: sourceName,
			Repo:       "",
		})
	}

	if err := scanner.Err(); err != nil {
		_ = cmd.Wait()
		return nil, fmt.Errorf("error reading rpm -qa output: %w", err)
	}

	if err := cmd.Wait(); err != nil {
		return nil, fmt.Errorf("rpm -qa failed: %w", err)
	}

	log.Logf("found %d installed packages", len(infos))

	return infos, nil
}

func extractSourceName(sourcerpm string) string {
	sourcerpm = strings.TrimSpace(sourcerpm)
	if sourcerpm == "" {
		return ""
	}
	s := strings.TrimSuffix(sourcerpm, ".rpm")
	parts := strings.Split(s, "-")
	if len(parts) < 2 {
		return s
	}
	if len(parts) >= 3 {
		return strings.Join(parts[:len(parts)-2], "-")
	}
	return parts[0]
}

func (m *Manager) Install(ctx context.Context, packages []string, opts api.InstallOptions) error {
	if len(packages) == 0 {
		log.Warnf("no packages to install")
		return nil
	}

	log.Logf("installing packages: %v", packages)

	if opts.DryRun {
		return nil
	}

	args := append([]string{"install"}, packages...)
	cmd := exec.CommandContext(ctx, m.Bin, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("dnf install failed: %w", err)
	}
	return nil
}

func (m *Manager) Remove(ctx context.Context, packages []string, opts api.RemoveOptions) error {
	if len(packages) == 0 {
		log.Warnf("no packages to remove")
		return nil
	}

	log.Logf("removing packages: %v", packages)

	if opts.DryRun {
		return nil
	}

	args := append([]string{"remove"}, packages...)
	cmd := exec.CommandContext(ctx, m.Bin, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("dnf remove failed: %w", err)
	}
	return nil
}
