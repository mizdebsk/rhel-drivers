package cli

import (
	"fmt"
	"github.com/mizdebsk/rhel-drivers/internal/log"
	"strings"

	"github.com/mizdebsk/rhel-drivers/internal/api"
	"github.com/mizdebsk/rhel-drivers/internal/core"

	"github.com/spf13/cobra"
)

var (
	flagVerbose bool
	flagQuiet   bool
	flagDebug   bool
	flagVersion bool
)

func NewRootCmd(deps api.CoreDeps, version string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "rhel-drivers",
		Short: "Install and manage RHEL hardware drivers",
		RunE: func(cmd *cobra.Command, args []string) error {
			if flagVersion {
				printVersion(version)
				return nil
			}
			return cmd.Help()
		},
	}

	cmd.SetHelpCommand(&cobra.Command{})
	cmd.SilenceUsage = true
	cmd.SilenceErrors = false
	cmd.CompletionOptions.DisableDefaultCmd = true

	cmd.PersistentFlags().BoolVar(&flagVerbose, "verbose", false, "Increase verbosity")
	cmd.PersistentFlags().BoolVar(&flagQuiet, "quiet", false, "Suppress non-error output")
	cmd.PersistentFlags().BoolVar(&flagDebug, "debug", false, "Activate debug mode")
	cmd.PersistentFlags().BoolVar(&flagVersion, "version", false, "Show version and exit")

	cobra.OnInitialize(func() {
		log.Quiet = flagQuiet
		log.Verbose = flagVerbose
		log.Debug = flagDebug
	})

	cmd.AddCommand(
		newInstallCmd(deps),
		newRemoveCmd(deps),
		newListCmd(deps),
	)

	return cmd
}

func printVersion(version string) {
	v := strings.TrimSpace(version)
	if v == "" {
		v = "unknown"
	}
	fmt.Println("rhel-drivers version", v)
}

func newInstallCmd(deps api.CoreDeps) *cobra.Command {
	var (
		flagAutoDetect bool
		flagDryRun     bool
		flagForce      bool
	)

	cmd := &cobra.Command{
		Use:     "install [OPTIONS] [DRIVER...]",
		Short:   "Install hardware drivers",
		Aliases: []string{"in"},
		Args:    cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			opts := api.InstallOptions{
				AutoDetect: flagAutoDetect,
				DryRun:     flagDryRun,
				Force:      flagForce,
			}

			if len(args) == 0 && !flagAutoDetect {
				return fmt.Errorf("not specified what to install (use --auto-detect or provide drivers)")
			}
			if len(args) > 0 && flagAutoDetect {
				return fmt.Errorf("both --auto-detect and specific drivers given")
			}

			return core.Install(ctx, deps, opts, args)
		},
	}

	cmd.Flags().BoolVar(&flagAutoDetect, "auto-detect", false, "Auto-detect drivers to install")
	cmd.Flags().BoolVar(&flagDryRun, "dry-run", false, "Show what would happen, don't change anything")
	cmd.Flags().BoolVar(&flagForce, "force", false, "Force install (ignore checks)")

	return cmd
}

func newRemoveCmd(deps api.CoreDeps) *cobra.Command {
	var (
		flagDryRun bool
		flagAll    bool
	)

	cmd := &cobra.Command{
		Use:     "remove [OPTIONS] [DRIVER...]",
		Short:   "Remove hardware drivers",
		Aliases: []string{"rm"},
		Args:    cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			opts := api.RemoveOptions{
				DryRun: flagDryRun,
				All:    flagAll,
			}

			if len(args) == 0 && !flagAll {
				return fmt.Errorf("not specified what to remove (use --all or provide drivers)")
			}
			if len(args) > 0 && flagAll {
				return fmt.Errorf("both --all and specific drivers given")
			}

			return core.Remove(ctx, deps, opts, args)
		},
	}

	cmd.Flags().BoolVar(&flagDryRun, "dry-run", false, "Show what would happen, don't change anything")
	cmd.Flags().BoolVar(&flagAll, "all", false, "Remove all installed drivers")

	return cmd
}

func newListCmd(deps api.CoreDeps) *cobra.Command {
	var (
		flagAvailable bool
		flagInstalled bool
	)

	cmd := &cobra.Command{
		Use:     "list [OPTIONS]",
		Short:   "List drivers",
		Aliases: []string{"ls"},
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			if flagAvailable || (!flagAvailable && !flagInstalled) {
				res, err := core.List(ctx, deps, true, true, true)
				if err != nil {
					return err
				}
				if len(res) > 0 {
					fmt.Println("Available drivers:")
					for _, dev := range res {
						markInstalled := " "
						if dev.Installed {
							markInstalled = "*"
						}
						markAuto := " "
						if dev.Compatible {
							markAuto = ">"
						}
						fmt.Printf("%s%s %s:%s\n", markInstalled, markAuto, dev.ID.ProviderID, dev.ID.Version)
					}
				} else {
					fmt.Println("Available drivers:\n  (none)")
				}
			}

			if flagInstalled {
				res, err := core.List(ctx, deps, true, false, false)
				if err != nil {
					return err
				}
				fmt.Print("Installed drivers:")
				for _, dev := range res {
					if dev.Installed {
						fmt.Printf("\n%s:%s", dev.ID.ProviderID, dev.ID.Version)
					}
				}
				fmt.Println()
			}

			return nil
		},
	}

	cmd.Flags().BoolVar(&flagAvailable, "available", false, "List available drivers")
	cmd.Flags().BoolVar(&flagInstalled, "installed", false, "List installed drivers")

	return cmd
}
