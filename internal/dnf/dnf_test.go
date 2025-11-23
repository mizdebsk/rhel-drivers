package dnf

import (
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/mizdebsk/rhel-drivers/internal/api"
	"github.com/mizdebsk/rhel-drivers/internal/mocks"
)

func TestInstall(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockExec := mocks.NewMockExecutor(ctrl)
	mockExec.EXPECT().Run("mydnf", []string{"install", "foo", "bar"}).Return(nil)
	pm := pkgMgr{
		bin:  "mydnf",
		exec: mockExec,
	}

	err := pm.Install([]string{"foo", "bar"}, api.InstallOptions{DryRun: false})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}
