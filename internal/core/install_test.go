package core

import (
	"fmt"
	"testing"

	"github.com/golang/mock/gomock"

	"github.com/mizdebsk/rhel-drivers/internal/api"
	"github.com/mizdebsk/rhel-drivers/internal/mocks"
)

func TestInstallSpecific(t *testing.T) {
	tests := []struct {
		name      string
		drivers   []string
		batchMode bool
		dryRun    bool
		force     bool
		setup     func(*mocks.MockProvider, *mocks.MockPackageManager, *mocks.MockRepositoryManager)
		expectErr bool
	}{
		{
			name:      "EmptyDriversList",
			drivers:   []string{},
			expectErr: true,
			setup:     func(p *mocks.MockProvider, pm *mocks.MockPackageManager, rm *mocks.MockRepositoryManager) {},
		},
		{
			name:      "InvalidDriverFormat",
			drivers:   []string{"invalid-format"},
			expectErr: true,
			setup:     func(p *mocks.MockProvider, pm *mocks.MockPackageManager, rm *mocks.MockRepositoryManager) {},
		},
		{
			name:      "UnknownProvider",
			drivers:   []string{"unknown:1.0"},
			expectErr: true,
			setup: func(p *mocks.MockProvider, pm *mocks.MockPackageManager, rm *mocks.MockRepositoryManager) {
				p.EXPECT().GetID().Return("nvidia").AnyTimes()
			},
		},
		{
			name:      "ListAvailableFails",
			drivers:   []string{"nvidia:570.86.16"},
			expectErr: true,
			setup: func(p *mocks.MockProvider, pm *mocks.MockPackageManager, rm *mocks.MockRepositoryManager) {
				p.EXPECT().GetID().Return("nvidia").AnyTimes()
				p.EXPECT().GetName().Return("NVIDIA").AnyTimes()
				p.EXPECT().ListAvailable().Return(nil, fmt.Errorf("failed to list"))
			},
		},
		{
			name:      "DriverVersionNotAvailable",
			drivers:   []string{"nvidia:999.99.99"},
			expectErr: true,
			setup: func(p *mocks.MockProvider, pm *mocks.MockPackageManager, rm *mocks.MockRepositoryManager) {
				p.EXPECT().GetID().Return("nvidia").AnyTimes()
				p.EXPECT().GetName().Return("NVIDIA").AnyTimes()
				p.EXPECT().ListAvailable().Return([]api.DriverID{
					{ProviderID: "nvidia", Version: "570.86.16"},
				}, nil)
			},
		},
		{
			name:      "NoCompatibleHardware",
			drivers:   []string{"nvidia:570.86.16"},
			expectErr: true,
			setup: func(p *mocks.MockProvider, pm *mocks.MockPackageManager, rm *mocks.MockRepositoryManager) {
				p.EXPECT().GetID().Return("nvidia").AnyTimes()
				p.EXPECT().GetName().Return("NVIDIA").AnyTimes()
				p.EXPECT().ListAvailable().Return([]api.DriverID{
					{ProviderID: "nvidia", Version: "570.86.16"},
				}, nil)
				p.EXPECT().DetectHardware().Return(false, nil)
			},
		},
		{
			name:      "ForceSkipsHardwareCheck",
			drivers:   []string{"nvidia:570.86.16"},
			force:     true,
			expectErr: false,
			setup: func(p *mocks.MockProvider, pm *mocks.MockPackageManager, rm *mocks.MockRepositoryManager) {
				p.EXPECT().GetID().Return("nvidia").AnyTimes()
				p.EXPECT().GetName().Return("NVIDIA").AnyTimes()
				p.EXPECT().ListAvailable().Return([]api.DriverID{
					{ProviderID: "nvidia", Version: "570.86.16"},
				}, nil)
				rm.EXPECT().EnsureRepositoriesEnabled().Return(nil)
				p.EXPECT().Install([]api.DriverID{{ProviderID: "nvidia", Version: "570.86.16"}}).Return([]string{"nvidia-driver"}, nil)
				pm.EXPECT().Install([]string{"nvidia-driver"}, false, false).Return(nil)
			},
		},
		{
			name:      "SuccessfulInstall",
			drivers:   []string{"nvidia:570.86.16"},
			expectErr: false,
			setup: func(p *mocks.MockProvider, pm *mocks.MockPackageManager, rm *mocks.MockRepositoryManager) {
				p.EXPECT().GetID().Return("nvidia").AnyTimes()
				p.EXPECT().GetName().Return("NVIDIA").AnyTimes()
				p.EXPECT().ListAvailable().Return([]api.DriverID{
					{ProviderID: "nvidia", Version: "570.86.16"},
				}, nil)
				p.EXPECT().DetectHardware().Return(true, nil)
				rm.EXPECT().EnsureRepositoriesEnabled().Return(nil)
				p.EXPECT().Install([]api.DriverID{{ProviderID: "nvidia", Version: "570.86.16"}}).Return([]string{"nvidia-driver"}, nil)
				pm.EXPECT().Install([]string{"nvidia-driver"}, false, false).Return(nil)
			},
		},
		{
			name:      "RepositoryEnableFails",
			drivers:   []string{"nvidia:570.86.16"},
			expectErr: true,
			setup: func(p *mocks.MockProvider, pm *mocks.MockPackageManager, rm *mocks.MockRepositoryManager) {
				p.EXPECT().GetID().Return("nvidia").AnyTimes()
				p.EXPECT().GetName().Return("NVIDIA").AnyTimes()
				p.EXPECT().ListAvailable().Return([]api.DriverID{
					{ProviderID: "nvidia", Version: "570.86.16"},
				}, nil)
				p.EXPECT().DetectHardware().Return(true, nil)
				rm.EXPECT().EnsureRepositoriesEnabled().Return(fmt.Errorf("repo enable failed"))
			},
		},
		{
			name:      "ProviderInstallFails",
			drivers:   []string{"nvidia:570.86.16"},
			expectErr: true,
			setup: func(p *mocks.MockProvider, pm *mocks.MockPackageManager, rm *mocks.MockRepositoryManager) {
				p.EXPECT().GetID().Return("nvidia").AnyTimes()
				p.EXPECT().GetName().Return("NVIDIA").AnyTimes()
				p.EXPECT().ListAvailable().Return([]api.DriverID{
					{ProviderID: "nvidia", Version: "570.86.16"},
				}, nil)
				p.EXPECT().DetectHardware().Return(true, nil)
				rm.EXPECT().EnsureRepositoriesEnabled().Return(nil)
				p.EXPECT().Install([]api.DriverID{{ProviderID: "nvidia", Version: "570.86.16"}}).Return(nil, fmt.Errorf("install failed"))
			},
		},
		{
			name:      "PackageManagerInstallFails",
			drivers:   []string{"nvidia:570.86.16"},
			expectErr: true,
			setup: func(p *mocks.MockProvider, pm *mocks.MockPackageManager, rm *mocks.MockRepositoryManager) {
				p.EXPECT().GetID().Return("nvidia").AnyTimes()
				p.EXPECT().GetName().Return("NVIDIA").AnyTimes()
				p.EXPECT().ListAvailable().Return([]api.DriverID{
					{ProviderID: "nvidia", Version: "570.86.16"},
				}, nil)
				p.EXPECT().DetectHardware().Return(true, nil)
				rm.EXPECT().EnsureRepositoriesEnabled().Return(nil)
				p.EXPECT().Install([]api.DriverID{{ProviderID: "nvidia", Version: "570.86.16"}}).Return([]string{"nvidia-driver"}, nil)
				pm.EXPECT().Install([]string{"nvidia-driver"}, false, false).Return(fmt.Errorf("dnf failed"))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockProvider := mocks.NewMockProvider(ctrl)
			mockPM := mocks.NewMockPackageManager(ctrl)
			mockRM := mocks.NewMockRepositoryManager(ctrl)

			tt.setup(mockProvider, mockPM, mockRM)

			deps := api.CoreDeps{
				PackageManager:    mockPM,
				RepositoryManager: mockRM,
				Providers:         []api.Provider{mockProvider},
			}

			err := InstallSpecific(deps, tt.drivers, tt.batchMode, tt.dryRun, tt.force)
			if (err != nil) != tt.expectErr {
				t.Errorf("InstallSpecific() error = %v, expectErr %v", err, tt.expectErr)
			}
		})
	}
}

func TestInstallAutoDetect(t *testing.T) {
	tests := []struct {
		name      string
		setup     func(*mocks.MockProvider, *mocks.MockPackageManager, *mocks.MockRepositoryManager)
		expectErr bool
	}{
		{
			name:      "NoHardwareDetected",
			expectErr: true,
			setup: func(p *mocks.MockProvider, pm *mocks.MockPackageManager, rm *mocks.MockRepositoryManager) {
				p.EXPECT().GetName().Return("NVIDIA").AnyTimes()
				p.EXPECT().DetectHardware().Return(false, nil)
			},
		},
		{
			name:      "HardwareDetectedButNoDriversAvailable",
			expectErr: true,
			setup: func(p *mocks.MockProvider, pm *mocks.MockPackageManager, rm *mocks.MockRepositoryManager) {
				p.EXPECT().GetName().Return("NVIDIA").AnyTimes()
				p.EXPECT().DetectHardware().Return(true, nil)
				p.EXPECT().ListAvailable().Return([]api.DriverID{}, nil)
			},
		},
		{
			name:      "ListAvailableFails",
			expectErr: true,
			setup: func(p *mocks.MockProvider, pm *mocks.MockPackageManager, rm *mocks.MockRepositoryManager) {
				p.EXPECT().GetName().Return("NVIDIA").AnyTimes()
				p.EXPECT().DetectHardware().Return(true, nil)
				p.EXPECT().ListAvailable().Return(nil, fmt.Errorf("list failed"))
			},
		},
		{
			name:      "SuccessfulAutoDetectInstall",
			expectErr: false,
			setup: func(p *mocks.MockProvider, pm *mocks.MockPackageManager, rm *mocks.MockRepositoryManager) {
				p.EXPECT().GetID().Return("nvidia").AnyTimes()
				p.EXPECT().GetName().Return("NVIDIA").AnyTimes()
				p.EXPECT().DetectHardware().Return(true, nil)
				p.EXPECT().ListAvailable().Return([]api.DriverID{
					{ProviderID: "nvidia", Version: "570.86.16"},
				}, nil)
				rm.EXPECT().EnsureRepositoriesEnabled().Return(nil)
				p.EXPECT().Install([]api.DriverID{{ProviderID: "nvidia", Version: "570.86.16"}}).Return([]string{"nvidia-driver"}, nil)
				pm.EXPECT().Install([]string{"nvidia-driver"}, false, false).Return(nil)
			},
		},
		{
			name:      "RepositoryEnableFails",
			expectErr: true,
			setup: func(p *mocks.MockProvider, pm *mocks.MockPackageManager, rm *mocks.MockRepositoryManager) {
				p.EXPECT().GetID().Return("nvidia").AnyTimes()
				p.EXPECT().GetName().Return("NVIDIA").AnyTimes()
				p.EXPECT().DetectHardware().Return(true, nil)
				p.EXPECT().ListAvailable().Return([]api.DriverID{
					{ProviderID: "nvidia", Version: "570.86.16"},
				}, nil)
				rm.EXPECT().EnsureRepositoriesEnabled().Return(fmt.Errorf("repo error"))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockProvider := mocks.NewMockProvider(ctrl)
			mockPM := mocks.NewMockPackageManager(ctrl)
			mockRM := mocks.NewMockRepositoryManager(ctrl)

			tt.setup(mockProvider, mockPM, mockRM)

			deps := api.CoreDeps{
				PackageManager:    mockPM,
				RepositoryManager: mockRM,
				Providers:         []api.Provider{mockProvider},
			}

			err := InstallAutoDetect(deps, false, false)
			if (err != nil) != tt.expectErr {
				t.Errorf("InstallAutoDetect() error = %v, expectErr %v", err, tt.expectErr)
			}
		})
	}
}
