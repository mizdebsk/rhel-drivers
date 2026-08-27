package rhsm

import (
	"fmt"
	"reflect"
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
			name:    "EnableReposWithSupplementary",
			sysInfo: sysinfo.SysInfo{IsRhel: true, OsVersion: 5, Arch: "sparc"},
			testFunc: func(t *testing.T) error {
				mockExec.EXPECT().
					Run(rm.rhsmExecPath, []string{
						"repos",
						"--enable", "rhel-5-for-sparc-baseos-rpms",
						"--enable", "rhel-5-for-sparc-appstream-rpms",
						"--enable", "rhel-5-for-sparc-extensions-rpms",
						"--enable", "rhel-5-for-sparc-supplementary-rpms",
					}).
					Return(nil)
				return rm.EnsureRepositoriesEnabled(true)
			},
		},
		{
			name:    "EnableReposWithoutSupplementary",
			sysInfo: sysinfo.SysInfo{IsRhel: true, OsVersion: 5, Arch: "sparc"},
			testFunc: func(t *testing.T) error {
				mockExec.EXPECT().
					Run(rm.rhsmExecPath, []string{
						"repos",
						"--enable", "rhel-5-for-sparc-baseos-rpms",
						"--enable", "rhel-5-for-sparc-appstream-rpms",
						"--enable", "rhel-5-for-sparc-extensions-rpms",
					}).
					Return(nil)
				return rm.EnsureRepositoriesEnabled(false)
			},
		},
		{
			name:    "EnableReposFailure",
			sysInfo: sysinfo.SysInfo{IsRhel: true, OsVersion: 5, Arch: "sparc"},
			testFunc: func(t *testing.T) error {
				mockExec.EXPECT().
					Run(rm.rhsmExecPath, []string{
						"repos",
						"--enable", "rhel-5-for-sparc-baseos-rpms",
						"--enable", "rhel-5-for-sparc-appstream-rpms",
						"--enable", "rhel-5-for-sparc-extensions-rpms",
						"--enable", "rhel-5-for-sparc-supplementary-rpms",
					}).
					Return(fmt.Errorf("hey, you don't have a valid subscription"))
				return rm.EnsureRepositoriesEnabled(true)
			},
			expectErr: true,
		},
		{
			name:    "ReopsAlreadyEnabled",
			sysInfo: sysinfo.SysInfo{IsRhel: true, OsVersion: 10, Arch: "x86_64"},
			testFunc: func(t *testing.T) error {
				return rm.EnsureRepositoriesEnabled(true)
			},
		},
		{
			name:    "SubscriptionManagerAbsent",
			sysInfo: sysinfo.SysInfo{IsRhel: true},
			testFunc: func(t *testing.T) error {
				rm.rhsmExecPath = "testdata/rhsm-absent-xxx"
				return rm.EnsureRepositoriesEnabled(true)
			},
		},
		{
			name: "NonRhelSystem",
			testFunc: func(t *testing.T) error {
				return rm.EnsureRepositoriesEnabled(true)
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

func TestGetRepoIDs(t *testing.T) {
	tests := []struct {
		name              string
		sysInfo           sysinfo.SysInfo
		needSupplementary bool
		want              []string
	}{
		{
			name:              "RhelWithSupplementary",
			sysInfo:           sysinfo.SysInfo{IsRhel: true, OsVersion: 10, Arch: "x86_64"},
			needSupplementary: true,
			want: []string{
				"rhel-10-for-x86_64-baseos-rpms",
				"rhel-10-for-x86_64-appstream-rpms",
				"rhel-10-for-x86_64-extensions-rpms",
				"rhel-10-for-x86_64-supplementary-rpms",
			},
		},
		{
			name:              "RhelWithoutSupplementary",
			sysInfo:           sysinfo.SysInfo{IsRhel: true, OsVersion: 10, Arch: "x86_64"},
			needSupplementary: false,
			want: []string{
				"rhel-10-for-x86_64-baseos-rpms",
				"rhel-10-for-x86_64-appstream-rpms",
				"rhel-10-for-x86_64-extensions-rpms",
			},
		},
		{
			name:              "NonRhelReturnsNil",
			sysInfo:           sysinfo.SysInfo{IsRhel: false},
			needSupplementary: true,
			want:              nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rm := repoMgr{systemInfo: tt.sysInfo}
			got, err := rm.GetRepoIDs(tt.needSupplementary)
			if err != nil {
				t.Fatalf("GetRepoIDs() unexpected error: %v", err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetRepoIDs() = %v, want %v", got, tt.want)
			}
		})
	}
}
