package rhsm

import (
	"fmt"
	"testing"

	"github.com/golang/mock/gomock"

	"github.com/mizdebsk/radii/internal/mocks"
	"github.com/mizdebsk/radii/internal/sysinfo"
)

func TestRhsm(t *testing.T) {
	var rm repoMgr
	var mockExec *mocks.MockExecutor
	tests := []struct {
		name      string
		sysInfo   sysinfo.SysInfo
		testFunc  func(t *testing.T) error
		expectErr bool
	}{
		{
			name:    "EnableReposSuccess",
			sysInfo: sysinfo.SysInfo{IsRhel: true, OsVersion: 5, Arch: "sparc"},
			testFunc: func(t *testing.T) error {
				mockExec.EXPECT().
					Run(rm.rhsmExecPath, []string{"status"}).
					Return(nil)
				mockExec.EXPECT().
					Run(rm.rhsmExecPath, []string{
						"repos",
						"--enable", "rhel-5-for-sparc-baseos-rpms",
						"--enable", "rhel-5-for-sparc-appstream-rpms",
						"--enable", "rhel-5-for-sparc-extensions-rpms",
						"--enable", "rhel-5-for-sparc-supplementary-rpms",
					}).
					Return(nil)
				return rm.EnsureRepositoriesEnabled()
			},
		},
		{
			name:    "EnableReposFailure",
			sysInfo: sysinfo.SysInfo{IsRhel: true, OsVersion: 5, Arch: "sparc"},
			testFunc: func(t *testing.T) error {
				mockExec.EXPECT().
					Run(rm.rhsmExecPath, []string{"status"}).
					Return(nil)
				mockExec.EXPECT().
					Run(rm.rhsmExecPath, []string{
						"repos",
						"--enable", "rhel-5-for-sparc-baseos-rpms",
						"--enable", "rhel-5-for-sparc-appstream-rpms",
						"--enable", "rhel-5-for-sparc-extensions-rpms",
						"--enable", "rhel-5-for-sparc-supplementary-rpms",
					}).
					Return(fmt.Errorf("hey, you don't have a valid subscription"))
				return rm.EnsureRepositoriesEnabled()
			},
			expectErr: true,
		},
		{
			name:    "ReopsAlreadyEnabled",
			sysInfo: sysinfo.SysInfo{IsRhel: true, OsVersion: 10, Arch: "x86_64"},
			testFunc: func(t *testing.T) error {
				mockExec.EXPECT().
					Run(rm.rhsmExecPath, []string{"status"}).
					Return(nil)
				return rm.EnsureRepositoriesEnabled()
			},
		},
		{
			name:    "SubscriptionManagerNotConfigured",
			sysInfo: sysinfo.SysInfo{IsRhel: true, OsVersion: 10, Arch: "x86_64"},
			testFunc: func(t *testing.T) error {
				mockExec.EXPECT().
					Run(rm.rhsmExecPath, []string{"status"}).
					Return(fmt.Errorf("subscription-manager command failed: exit status 1"))
				return rm.EnsureRepositoriesEnabled()
			},
			expectErr: true,
		},
		{
			name:    "SubscriptionManagerAbsent",
			sysInfo: sysinfo.SysInfo{IsRhel: true},
			testFunc: func(t *testing.T) error {
				rm.rhsmExecPath = "testdata/rhsm-absent-xxx"
				return rm.EnsureRepositoriesEnabled()
			},
		},
		{
			name: "NonRhelSystem",
			testFunc: func(t *testing.T) error {
				return rm.EnsureRepositoriesEnabled()
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			mockExec = mocks.NewMockExecutor(ctrl)
			rm = repoMgr{
				systemInfo:     tt.sysInfo,
				executor:       mockExec,
				redhatRepoPath: "testdata/rhel10.repo",
				rhsmExecPath:   "testdata/rhsm-exec",
			}

			err := tt.testFunc(t)
			if (err != nil) != tt.expectErr {
				t.Errorf("Expected error: %v, but got: %v", tt.expectErr, err)
			}
		})
	}
}
