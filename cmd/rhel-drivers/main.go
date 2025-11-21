package main

import (
	"context"
	"os"

	"github.com/mizdebsk/rhel-drivers/internal/api"
	"github.com/mizdebsk/rhel-drivers/internal/cli"
	"github.com/mizdebsk/rhel-drivers/internal/dnf"
	"github.com/mizdebsk/rhel-drivers/internal/provider/amd"
	"github.com/mizdebsk/rhel-drivers/internal/provider/nvidia"
	"github.com/mizdebsk/rhel-drivers/internal/rhsm"
	"github.com/mizdebsk/rhel-drivers/internal/sysinfo"
)

// set at build time via -ldflags, eg: go build -ldflags="-X main.version=1.0.0" ./cmd/rhel-drivers
var version = "dev"

func main() {
	ctx := context.Background()
	systemInfo := sysinfo.DetectSysInfo()

	pm := dnf.New()
	repoVerifier := rhsm.NewVerifier(systemInfo)
	providers := []api.Provider{nvidia.NewProvider(pm), amd.NewProvider(pm)}
	deps := api.CoreDeps{
		PM:           pm,
		RepoVerifier: repoVerifier,
		Providers:    providers,
	}

	root := cli.NewRootCmd(deps, version)

	if err := root.ExecuteContext(ctx); err != nil {
		os.Exit(1)
	}
}
