package rhsm

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"github.com/mizdebsk/rhel-drivers/internal/api"
	"github.com/mizdebsk/rhel-drivers/internal/log"
	"github.com/mizdebsk/rhel-drivers/internal/sysinfo"
	"strings"
)

const (
	redhatRepoPath = "/etc/yum.repos.d/redhat.repo"
	rhsmExecPath   = "/usr/sbin/subscription-manager"
)

type Verifier struct {
	SysInfo sysinfo.SysInfo
}

var _ api.RepoVerifier = (*Verifier)(nil)

func NewVerifier(sysinfo sysinfo.SysInfo) *Verifier {
	return &Verifier{
		SysInfo: sysinfo,
	}
}

func (v *Verifier) VerifyAndEnable(ctx context.Context) error {
	if v.SysInfo.IsRhel {
		log.Logf("detected RHEL %d", v.SysInfo.OsVersion)
		if v.SubscriptionManagerPresent() {
			log.Logf("Subscription Manager is present")
			channels := []string{"BaseOS", "AppStream", "Extensions", "Supplementary"}
			return v.EnsureChannelsEnabled(ctx, channels)
		} else {
			log.Warnf("Subscription Manager is absent.")
			log.Warnf("You may need to enable appropriate repositories yourself.")
		}
	} else {
		log.Warnf("This system is not RHEL.")
		log.Warnf("You may need to enable appropriate repositories yourself.")
	}
	return nil
}

func (v *Verifier) SubscriptionManagerPresent() bool {
	stat, err := os.Stat(rhsmExecPath)
	if err != nil || stat == nil {
		log.Debugf("stat %s failed: %v", rhsmExecPath, err)
		return false
	}
	log.Debugf("stat %s: isRegular=%v mode=0%o", rhsmExecPath, stat.Mode().IsRegular(), stat.Mode().Perm())
	return stat.Mode().IsRegular() && stat.Mode().Perm()&0111 != 0
}

func (v *Verifier) EnsureChannelsEnabled(ctx context.Context, channels []string) error {
	log.Logf("checking repository status")
	allEnabled := true
	args := []string{"repos"}
	for _, channel := range channels {
		repo := fmt.Sprintf("rhel-%d-for-%s-%s-rpms", v.SysInfo.OsVersion, v.SysInfo.Arch, strings.ToLower(channel))
		log.Logf("mapped RHEL channel %s to repo ID %s", channel, repo)
		if !repoEnabled(redhatRepoPath, repo) {
			log.Infof("enabling channel %s, repository %s", channel, repo)
			args = append(args, "--enable", repo)
			allEnabled = false
		} else {
			log.Logf("repository %s is already enabled", repo)
		}
	}

	if allEnabled {
		log.Logf("all required repositories are already enabled")
		return nil
	}

	log.Logf("running subscription-manager to enable repositories")
	cmd := exec.CommandContext(ctx, rhsmExecPath, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to enable repositories: %w", err)
	}

	log.Logf("repositories were enabled successfully")
	return nil
}
