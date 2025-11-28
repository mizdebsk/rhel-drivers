package dnf

import (
	"fmt"
	"testing"

	"github.com/golang/mock/gomock"

	"github.com/mizdebsk/rhel-drivers/internal/api"
	"github.com/mizdebsk/rhel-drivers/internal/mocks"
)

func assertTwoPackagesAntBash(out []api.PackageInfo, t *testing.T) {
	if len(out) != 2 {
		t.Errorf("Expected exactly 2 packages, got %d", len(out))
	}
	if len(out) >= 2 {
		if out[0].Name != "ant-junit" || out[0].SourceName != "ant" {
			t.Errorf("First package expected to be ant-junit from ant source, got %s from %s", out[0].Name, out[0].SourceName)
		}
		if out[1].Name != "bash" {
			t.Errorf("Second package expected to be bash, got %s", out[1].Name)
		}
	}
}

func TestDnf(t *testing.T) {
	var pm pkgMgr
	const dnfBin = "mydnf"
	var mockExec *mocks.MockExecutor
	tests := []struct {
		name      string
		testFunc  func(t *testing.T) error
		expectErr bool
	}{
		{
			name: "InstallSuccess",
			testFunc: func(t *testing.T) error {
				mockExec.EXPECT().
					Run(dnfBin, []string{"install", "foo", "bar"}).
					Return(nil)
				return pm.Install([]string{"foo", "bar"})
			},
		},
		{
			name: "InstallFailure",
			testFunc: func(t *testing.T) error {
				mockExec.EXPECT().
					Run(dnfBin, []string{"install", "foo", "bar"}).
					Return(fmt.Errorf("something went wrong"))
				return pm.Install([]string{"foo", "bar"})
			},
			expectErr: true,
		},
		{
			name: "InstallNothing",
			testFunc: func(t *testing.T) error {
				return pm.Install([]string{})
			},
		},
		{
			name: "RemoveSuccess",
			testFunc: func(t *testing.T) error {
				mockExec.EXPECT().
					Run(dnfBin, []string{"remove", "foo", "bar"}).
					Return(nil)
				return pm.Remove([]string{"foo", "bar"})
			},
		},
		{
			name: "RemoveFailure",
			testFunc: func(t *testing.T) error {
				mockExec.EXPECT().
					Run(dnfBin, []string{"remove", "foo", "bar"}).
					Return(fmt.Errorf("something went wrong"))
				return pm.Remove([]string{"foo", "bar"})
			},
			expectErr: true,
		},
		{
			name: "RemoveNothing",
			testFunc: func(t *testing.T) error {
				return pm.Remove([]string{})
			},
		},
		{
			name: "ListAvailableFailure",
			testFunc: func(t *testing.T) error {
				mockExec.EXPECT().
					RunCapture(dnfBin, []string{
						"-q", "repoquery", "--qf",
						"QQQ|%{name}|%{epoch}|%{version}|%{release}|%{arch}|%{sourcerpm}|%{repoid}|YYY\n",
					}).
					Return([]string{
						"QQQ|ant-junit|0|1.10.15|32.fc43|noarch|ant-1.10.15-32.fc43.src.rpm|updates-testing|YYY",
						"oh well...",
					}, fmt.Errorf("fatal error"))
				out, err := pm.ListAvailablePackages()
				if len(out) != 0 {
					t.Errorf("Expected exactly 0 packages, got %d", len(out))
				}
				return err
			},
			expectErr: true,
		},
		{
			name: "ListAvailableSuccess",
			testFunc: func(t *testing.T) error {
				mockExec.EXPECT().
					RunCapture(dnfBin, []string{
						"-q", "repoquery", "--qf",
						"QQQ|%{name}|%{epoch}|%{version}|%{release}|%{arch}|%{sourcerpm}|%{repoid}|YYY\n",
					}).
					Return([]string{
						"QQQ|ant-junit|0|1.10.15|32.fc43|noarch|ant-1.10.15-32.fc43.src.rpm|updates-testing|YYY",
						"some trash",
						"QQQ|f|o|o|YYY",
						"QQQ|bash|0|5.3.0|2.fc43|x86_64|bash-5.3.0-2.fc43.src.rpm|fedora|YYY",
					}, nil)
				out, err := pm.ListAvailablePackages()
				assertTwoPackagesAntBash(out, t)
				return err
			},
		},
		{
			name: "ListAvailableCached",
			testFunc: func(t *testing.T) error {
				out, err := pm.ListAvailablePackages()
				assertTwoPackagesAntBash(out, t)
				return err
			},
		},
		{
			name: "ListInstalledFailure",
			testFunc: func(t *testing.T) error {
				mockExec.EXPECT().
					RunCapture("rpm", []string{
						"-qa", "--qf",
						"QQQ|%|NAME?{%{NAME}}||%|EPOCH?{%{EPOCH}}||%|VERSION?{%{VERSION}}||%|RELEASE?{%{RELEASE}}||%|ARCH?{%{ARCH}}||%|SOURCERPM?{%{SOURCERPM}}|||YYY\n",
					}).
					Return([]string{
						"QQQ|ant-junit|0|1.10.15|32.fc43|noarch|ant-1.10.15-32.fc43.src.rpm||YYY",
						"oh well...",
					}, fmt.Errorf("fatal error"))
				out, err := pm.ListInstalledPackages()
				if len(out) != 0 {
					t.Errorf("Expected exactly 0 packages, got %d", len(out))
				}
				return err
			},
			expectErr: true,
		},
		{
			name: "ListInstalledSuccess",
			testFunc: func(t *testing.T) error {
				mockExec.EXPECT().
					RunCapture("rpm", []string{
						"-qa", "--qf",
						"QQQ|%|NAME?{%{NAME}}||%|EPOCH?{%{EPOCH}}||%|VERSION?{%{VERSION}}||%|RELEASE?{%{RELEASE}}||%|ARCH?{%{ARCH}}||%|SOURCERPM?{%{SOURCERPM}}|||YYY\n",
					}).
					Return([]string{
						"QQQ|ant-junit|0|1.10.15|32.fc43|noarch|ant-1.10.15-32.fc43.src.rpm||YYY",
						"some trash",
						"QQQ|f|o|o|YYY",
						"QQQ|bash|0|5.3.0|2.fc43|x86_64|bash-5.3.0-2.fc43.src.rpm||YYY",
					}, nil)
				out, err := pm.ListInstalledPackages()
				assertTwoPackagesAntBash(out, t)
				return err
			},
		},
		{
			name: "ListInstalledCached",
			testFunc: func(t *testing.T) error {
				out, err := pm.ListInstalledPackages()
				assertTwoPackagesAntBash(out, t)
				return err
			},
		},
		{
			name: "NewPackageManager",
			testFunc: func(t *testing.T) error {
				pm := NewPackageManager(mockExec)
				if pm == nil {
					t.Errorf("Expected PackageManager, got nil")
				}
				return nil
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			mockExec = mocks.NewMockExecutor(ctrl)
			pm = pkgMgr{
				bin:  dnfBin,
				exec: mockExec,
			}
			err := tt.testFunc(t)
			if (err != nil) != tt.expectErr {
				t.Errorf("Expected error: %v, but got: %v", tt.expectErr, err)
			}
		})
	}
}
