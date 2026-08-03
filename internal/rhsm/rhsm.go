package rhsm

import (
	"fmt"
	"os"
	"strings"

	"github.com/mizdebsk/radii/internal/api"
	"github.com/mizdebsk/radii/internal/log"
	"github.com/mizdebsk/radii/internal/sysinfo"
)

const (
	defaultRedhatRepoPath = "/etc/yum.repos.d/redhat.repo"
	defaultRhsmExecPath   = "/usr/sbin/subscription-manager"
)

type repoMgr struct {
	systemInfo     sysinfo.SysInfo
	executor       api.Executor
	redhatRepoPath string
	rhsmExecPath   string
}

var _ api.RepositoryManager = (*repoMgr)(nil)

func NewRepositoryManager(executor api.Executor, systemInfo sysinfo.SysInfo) api.RepositoryManager {
	return &repoMgr{
		systemInfo:     systemInfo,
		executor:       executor,
		redhatRepoPath: defaultRedhatRepoPath,
		rhsmExecPath:   defaultRhsmExecPath,
	}
}

func (rm *repoMgr) EnsureRepositoriesEnabled() error {
	if rm.systemInfo.IsRhel {
		log.Logf("detected RHEL %d", rm.systemInfo.OsVersion)
		if rm.subscriptionManagerPresent() {
			log.Logf("Subscription Manager is present")
			channels := []string{"BaseOS", "AppStream", "Extensions", "Supplementary"}
			return rm.ensureChannelsEnabled(channels)
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

func (rm *repoMgr) subscriptionManagerPresent() bool {
	stat, err := os.Stat(rm.rhsmExecPath)
	if err != nil || stat == nil {
		log.Debugf("stat %s failed: %v", rm.rhsmExecPath, err)
		return false
	}
	log.Debugf("stat %s: isRegular=%v mode=0%o", rm.rhsmExecPath, stat.Mode().IsRegular(), stat.Mode().Perm())
	return stat.Mode().IsRegular() && stat.Mode().Perm()&0111 != 0
}

func (rm *repoMgr) ensureChannelsEnabled(channels []string) error {
	log.Logf("checking repository status")
	allEnabled := true
	args := []string{"repos"}
	for _, channel := range channels {
		repo := fmt.Sprintf("rhel-%d-for-%s-%s-rpms", rm.systemInfo.OsVersion, rm.systemInfo.Arch, strings.ToLower(channel))
		log.Logf("mapped RHEL channel %s to repo ID %s", channel, repo)
		if !repoEnabled(rm.redhatRepoPath, repo) {
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
	err := rm.executor.Run(rm.rhsmExecPath, args)
	if err != nil {
		return fmt.Errorf("failed to enable repositories, please ensure the system is registered and subscribed (see https://access.redhat.com/solutions/253273): %w", err)
	}

	log.Logf("repositories were enabled successfully")
	return nil
}
