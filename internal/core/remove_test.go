package core

import (
	"fmt"
	"testing"

	"github.com/golang/mock/gomock"

	"github.com/mizdebsk/rhel-drivers/internal/api"
	"github.com/mizdebsk/rhel-drivers/internal/mocks"
)

func TestRemoveSpecific(t *testing.T) {
	tests := []struct {
		name      string
		drivers   []string
		setup     func(*mocks.MockProvider, *mocks.MockPackageManager)
		expectErr bool
	}{
		{
			name:      "EmptyDriversList",
			drivers:   []string{},
			expectErr: true,
			setup:     func(p *mocks.MockProvider, pm *mocks.MockPackageManager) {},
		},
		{
			name:      "InvalidDriverFormat",
			drivers:   []string{"invalid-format"},
			expectErr: true,
			setup:     func(p *mocks.MockProvider, pm *mocks.MockPackageManager) {},
		},
		{
			name:      "UnknownProvider",
			drivers:   []string{"unknown:1.0"},
			expectErr: true,
			setup: func(p *mocks.MockProvider, pm *mocks.MockPackageManager) {
				p.EXPECT().GetID().Return("nvidia").AnyTimes()
			},
		},
		{
			name:      "ListInstalledFails",
			drivers:   []string{"nvidia:570.86.16"},
			expectErr: true,
			setup: func(p *mocks.MockProvider, pm *mocks.MockPackageManager) {
				p.EXPECT().GetID().Return("nvidia").AnyTimes()
				p.EXPECT().GetName().Return("NVIDIA").AnyTimes()
				p.EXPECT().ListInstalled().Return(nil, fmt.Errorf("failed to list"))
			},
		},
		{
			name:      "DriverNotInstalled",
			drivers:   []string{"nvidia:570.86.16"},
			expectErr: true,
			setup: func(p *mocks.MockProvider, pm *mocks.MockPackageManager) {
				p.EXPECT().GetID().Return("nvidia").AnyTimes()
				p.EXPECT().GetName().Return("NVIDIA").AnyTimes()
				p.EXPECT().ListInstalled().Return([]api.DriverID{
					{ProviderID: "nvidia", Version: "560.35.03"},
				}, nil)
			},
		},
		{
			name:      "SuccessfulRemove",
			drivers:   []string{"nvidia:570.86.16"},
			expectErr: false,
			setup: func(p *mocks.MockProvider, pm *mocks.MockPackageManager) {
				p.EXPECT().GetID().Return("nvidia").AnyTimes()
				p.EXPECT().GetName().Return("NVIDIA").AnyTimes()
				p.EXPECT().ListInstalled().Return([]api.DriverID{
					{ProviderID: "nvidia", Version: "570.86.16"},
				}, nil)
				p.EXPECT().Remove([]api.DriverID{{ProviderID: "nvidia", Version: "570.86.16"}}).Return([]string{"nvidia-driver"}, nil)
				pm.EXPECT().Remove([]string{"nvidia-driver"}, false, false).Return(nil)
			},
		},
		{
			name:      "ProviderRemoveFails",
			drivers:   []string{"nvidia:570.86.16"},
			expectErr: true,
			setup: func(p *mocks.MockProvider, pm *mocks.MockPackageManager) {
				p.EXPECT().GetID().Return("nvidia").AnyTimes()
				p.EXPECT().GetName().Return("NVIDIA").AnyTimes()
				p.EXPECT().ListInstalled().Return([]api.DriverID{
					{ProviderID: "nvidia", Version: "570.86.16"},
				}, nil)
				p.EXPECT().Remove([]api.DriverID{{ProviderID: "nvidia", Version: "570.86.16"}}).Return(nil, fmt.Errorf("remove failed"))
			},
		},
		{
			name:      "PackageManagerRemoveFails",
			drivers:   []string{"nvidia:570.86.16"},
			expectErr: true,
			setup: func(p *mocks.MockProvider, pm *mocks.MockPackageManager) {
				p.EXPECT().GetID().Return("nvidia").AnyTimes()
				p.EXPECT().GetName().Return("NVIDIA").AnyTimes()
				p.EXPECT().ListInstalled().Return([]api.DriverID{
					{ProviderID: "nvidia", Version: "570.86.16"},
				}, nil)
				p.EXPECT().Remove([]api.DriverID{{ProviderID: "nvidia", Version: "570.86.16"}}).Return([]string{"nvidia-driver"}, nil)
				pm.EXPECT().Remove([]string{"nvidia-driver"}, false, false).Return(fmt.Errorf("dnf failed"))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockProvider := mocks.NewMockProvider(ctrl)
			mockPM := mocks.NewMockPackageManager(ctrl)

			tt.setup(mockProvider, mockPM)

			deps := api.CoreDeps{
				PackageManager: mockPM,
				Providers:      []api.Provider{mockProvider},
			}

			err := RemoveSpecific(deps, tt.drivers, false, false)
			if (err != nil) != tt.expectErr {
				t.Errorf("RemoveSpecific() error = %v, expectErr %v", err, tt.expectErr)
			}
		})
	}
}

func TestRemoveAll(t *testing.T) {
	tests := []struct {
		name      string
		setup     func(*mocks.MockProvider, *mocks.MockPackageManager)
		expectErr bool
	}{
		{
			name:      "NoInstalledDrivers",
			expectErr: true,
			setup: func(p *mocks.MockProvider, pm *mocks.MockPackageManager) {
				p.EXPECT().GetName().Return("NVIDIA").AnyTimes()
				p.EXPECT().ListInstalled().Return([]api.DriverID{}, nil)
			},
		},
		{
			name:      "ListInstalledFails",
			expectErr: true,
			setup: func(p *mocks.MockProvider, pm *mocks.MockPackageManager) {
				p.EXPECT().GetName().Return("NVIDIA").AnyTimes()
				p.EXPECT().ListInstalled().Return(nil, fmt.Errorf("list failed"))
			},
		},
		{
			name:      "SuccessfulRemoveAll",
			expectErr: false,
			setup: func(p *mocks.MockProvider, pm *mocks.MockPackageManager) {
				p.EXPECT().GetID().Return("nvidia").AnyTimes()
				p.EXPECT().GetName().Return("NVIDIA").AnyTimes()
				p.EXPECT().ListInstalled().Return([]api.DriverID{
					{ProviderID: "nvidia", Version: "570.86.16"},
					{ProviderID: "nvidia", Version: "560.35.03"},
				}, nil)
				p.EXPECT().Remove(gomock.Any()).Return([]string{"nvidia-driver-570", "nvidia-driver-560"}, nil)
				pm.EXPECT().Remove(gomock.Any(), false, false).Return(nil)
			},
		},
		{
			name:      "ProviderRemoveFails",
			expectErr: true,
			setup: func(p *mocks.MockProvider, pm *mocks.MockPackageManager) {
				p.EXPECT().GetID().Return("nvidia").AnyTimes()
				p.EXPECT().GetName().Return("NVIDIA").AnyTimes()
				p.EXPECT().ListInstalled().Return([]api.DriverID{
					{ProviderID: "nvidia", Version: "570.86.16"},
				}, nil)
				p.EXPECT().Remove(gomock.Any()).Return(nil, fmt.Errorf("remove failed"))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockProvider := mocks.NewMockProvider(ctrl)
			mockPM := mocks.NewMockPackageManager(ctrl)

			tt.setup(mockProvider, mockPM)

			deps := api.CoreDeps{
				PackageManager: mockPM,
				Providers:      []api.Provider{mockProvider},
			}

			err := RemoveAll(deps, false, false)
			if (err != nil) != tt.expectErr {
				t.Errorf("RemoveAll() error = %v, expectErr %v", err, tt.expectErr)
			}
		})
	}
}
