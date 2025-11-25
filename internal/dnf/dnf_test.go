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
	dnfBin := "mydnf"
	var mockExec *mocks.MockExecutor
	expectDnfCall := func(args []string, err error) {
		mockExec.EXPECT().Run(dnfBin, args).Return(err)
	}
	expectDnfCallCapture := func(args []string, err error, out []string) {
		mockExec.EXPECT().RunCapture(dnfBin, args).Return(out, err)
	}
	expectRpmCallCapture := func(args []string, err error, out []string) {
		mockExec.EXPECT().RunCapture("rpm", args).Return(out, err)
	}
	tests := []struct {
		name      string
		funcToRun func(t *testing.T) error
		expectErr bool
	}{
		{
			name: "InstallSuccess",
			funcToRun: func(t *testing.T) error {
				expectDnfCall([]string{"install", "foo", "bar"}, nil)
				return pm.Install([]string{"foo", "bar"}, api.InstallOptions{DryRun: false})
			},
		},
		{
			name: "InstallFailure",
			funcToRun: func(t *testing.T) error {
				expectDnfCall([]string{"install", "foo", "bar"}, fmt.Errorf("something went wrong"))
				return pm.Install([]string{"foo", "bar"}, api.InstallOptions{DryRun: false})
			},
			expectErr: true,
		},
		{
			name: "InstallNothing",
			funcToRun: func(t *testing.T) error {
				return pm.Install([]string{}, api.InstallOptions{DryRun: false})
			},
		},
		{
			name: "InstallDryRun",
			funcToRun: func(t *testing.T) error {
				return pm.Install([]string{"foo", "bar"}, api.InstallOptions{DryRun: true})
			},
		},
		{
			name: "RemoveSuccess",
			funcToRun: func(t *testing.T) error {
				expectDnfCall([]string{"remove", "foo", "bar"}, nil)
				return pm.Remove([]string{"foo", "bar"}, api.RemoveOptions{DryRun: false})
			},
		},
		{
			name: "RemoveFailure",
			funcToRun: func(t *testing.T) error {
				expectDnfCall([]string{"remove", "foo", "bar"}, fmt.Errorf("something went wrong"))
				return pm.Remove([]string{"foo", "bar"}, api.RemoveOptions{DryRun: false})
			},
			expectErr: true,
		},
		{
			name: "RemoveNothing",
			funcToRun: func(t *testing.T) error {
				return pm.Remove([]string{}, api.RemoveOptions{DryRun: false})
			},
		},
		{
			name: "RemoveDryRun",
			funcToRun: func(t *testing.T) error {
				return pm.Remove([]string{"foo", "bar"}, api.RemoveOptions{DryRun: true})
			},
		},
		{
			name: "ListAvailableFailure",
			funcToRun: func(t *testing.T) error {
				expectDnfCallCapture([]string{
					"-q", "repoquery", "--qf",
					"QQQ|%{name}|%{epoch}|%{version}|%{release}|%{arch}|%{sourcerpm}|%{repoid}|YYY\n",
				}, fmt.Errorf("fatal error"), []string{
					"QQQ|ant-junit|0|1.10.15|32.fc43|noarch|ant-1.10.15-32.fc43.src.rpm|updates-testing|YYY",
					"oh well...",
				})
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
			funcToRun: func(t *testing.T) error {
				expectDnfCallCapture([]string{
					"-q", "repoquery", "--qf",
					"QQQ|%{name}|%{epoch}|%{version}|%{release}|%{arch}|%{sourcerpm}|%{repoid}|YYY\n",
				}, nil, []string{
					"QQQ|ant-junit|0|1.10.15|32.fc43|noarch|ant-1.10.15-32.fc43.src.rpm|updates-testing|YYY",
					"some trash",
					"QQQ|f|o|o|YYY",
					"QQQ|bash|0|5.3.0|2.fc43|x86_64|bash-5.3.0-2.fc43.src.rpm|fedora|YYY",
				})
				out, err := pm.ListAvailablePackages()
				assertTwoPackagesAntBash(out, t)
				return err
			},
		},
		{
			name: "ListAvailableCached",
			funcToRun: func(t *testing.T) error {
				out, err := pm.ListAvailablePackages()
				assertTwoPackagesAntBash(out, t)
				return err
			},
		},
		{
			name: "ListInstalledFailure",
			funcToRun: func(t *testing.T) error {
				expectRpmCallCapture([]string{
					"-qa", "--qf",
					"QQQ|%|NAME?{%{NAME}}||%|EPOCH?{%{EPOCH}}||%|VERSION?{%{VERSION}}||%|RELEASE?{%{RELEASE}}||%|ARCH?{%{ARCH}}||%|SOURCERPM?{%{SOURCERPM}}|||YYY\n",
				}, fmt.Errorf("fatal error"), []string{
					"QQQ|ant-junit|0|1.10.15|32.fc43|noarch|ant-1.10.15-32.fc43.src.rpm||YYY",
					"oh well...",
				})
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
			funcToRun: func(t *testing.T) error {
				expectRpmCallCapture([]string{
					"-qa", "--qf",
					"QQQ|%|NAME?{%{NAME}}||%|EPOCH?{%{EPOCH}}||%|VERSION?{%{VERSION}}||%|RELEASE?{%{RELEASE}}||%|ARCH?{%{ARCH}}||%|SOURCERPM?{%{SOURCERPM}}|||YYY\n",
				}, nil, []string{
					"QQQ|ant-junit|0|1.10.15|32.fc43|noarch|ant-1.10.15-32.fc43.src.rpm||YYY",
					"some trash",
					"QQQ|f|o|o|YYY",
					"QQQ|bash|0|5.3.0|2.fc43|x86_64|bash-5.3.0-2.fc43.src.rpm||YYY",
				})
				out, err := pm.ListInstalledPackages()
				assertTwoPackagesAntBash(out, t)
				return err
			},
		},
		{
			name: "ListInstalledCached",
			funcToRun: func(t *testing.T) error {
				out, err := pm.ListInstalledPackages()
				assertTwoPackagesAntBash(out, t)
				return err
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
			err := tt.funcToRun(t)
			if (err != nil) != tt.expectErr {
				t.Errorf("Expected error: %v, but got: %v", tt.expectErr, err)
			}
		})
	}
}
