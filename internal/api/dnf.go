package api

type PackageManager interface {
	ListAvailablePackages() ([]PackageInfo, error)
	ListInstalledPackages() ([]PackageInfo, error)
	Install(packages []string, batchMode, dryRun bool) error
	Remove(packages []string, batchMode, dryRun bool) error
}

type PackageInfo struct {
	Name       string
	Epoch      string
	Version    string
	Release    string
	Arch       string
	SourceName string
	Repo       string
}

func (p PackageInfo) NEVRA() string {
	epochStr := ""
	if p.Epoch != "" && p.Epoch != "0" {
		epochStr = p.Epoch + ":"
	}
	return p.Name + "-" + epochStr + p.Version + "-" + p.Release + "." + p.Arch
}
