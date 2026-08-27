package core

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/golang/mock/gomock"

	"github.com/mizdebsk/radii/internal/api"
	"github.com/mizdebsk/radii/internal/mocks"
)

func TestList(t *testing.T) {
	tests := []struct {
		name           string
		listInst       bool
		listAvail      bool
		hwdetect       bool
		compatibleOnly bool
		setup          func(*mocks.MockProvider, *mocks.MockPackageManager, *mocks.MockRepositoryManager)
		expectErr      bool
		expectLen      int
		checkFunc      func([]api.DriverStatus) error
	}{
		{
			name:           "ListInstalledOnly",
			listInst:       true,
			listAvail:      false,
			hwdetect:       false,
			compatibleOnly: false,
			setup: func(p *mocks.MockProvider, pm *mocks.MockPackageManager, rm *mocks.MockRepositoryManager) {
				p.EXPECT().GetID().Return("nvidia").AnyTimes()
				p.EXPECT().GetName().Return("NVIDIA").AnyTimes()
				p.EXPECT().ListInstalled().Return([]api.DriverID{
					{ProviderID: "nvidia", Version: "570.86.16"},
				}, nil)
			},
			expectErr: false,
			expectLen: 1,
			checkFunc: func(result []api.DriverStatus) error {
				if !result[0].Installed || result[0].Available {
					return fmt.Errorf("expected installed=true, available=false")
				}
				return nil
			},
		},
		{
			name:           "ListAvailableOnly",
			listInst:       false,
			listAvail:      true,
			hwdetect:       false,
			compatibleOnly: false,
			setup: func(p *mocks.MockProvider, pm *mocks.MockPackageManager, rm *mocks.MockRepositoryManager) {
				p.EXPECT().GetID().Return("nvidia").AnyTimes()
				p.EXPECT().GetName().Return("NVIDIA").AnyTimes()
				rm.EXPECT().GetRepoIDs(true).Return([]string{"rhel-10-for-x86_64-baseos-rpms"}, nil)
				pm.EXPECT().SetEnableRepos(gomock.Any())
				p.EXPECT().ListAvailable().Return([]api.DriverID{
					{ProviderID: "nvidia", Version: "570.86.16"},
				}, nil)
			},
			expectErr: false,
			expectLen: 1,
			checkFunc: func(result []api.DriverStatus) error {
				if result[0].Installed || !result[0].Available {
					return fmt.Errorf("expected installed=false, available=true")
				}
				return nil
			},
		},
		{
			name:           "ListBothInstalledAndAvailable",
			listInst:       true,
			listAvail:      true,
			hwdetect:       false,
			compatibleOnly: false,
			setup: func(p *mocks.MockProvider, pm *mocks.MockPackageManager, rm *mocks.MockRepositoryManager) {
				p.EXPECT().GetID().Return("nvidia").AnyTimes()
				p.EXPECT().GetName().Return("NVIDIA").AnyTimes()
				rm.EXPECT().GetRepoIDs(true).Return([]string{"rhel-10-for-x86_64-baseos-rpms"}, nil)
				pm.EXPECT().SetEnableRepos(gomock.Any())
				p.EXPECT().ListInstalled().Return([]api.DriverID{
					{ProviderID: "nvidia", Version: "570.86.16"},
				}, nil)
				p.EXPECT().ListAvailable().Return([]api.DriverID{
					{ProviderID: "nvidia", Version: "570.86.16"},
					{ProviderID: "nvidia", Version: "560.35.03"},
				}, nil)
			},
			expectErr: false,
			expectLen: 2,
			checkFunc: func(result []api.DriverStatus) error {
				for _, r := range result {
					if r.ID.Version == "570.86.16" && (!r.Installed || !r.Available) {
						return fmt.Errorf("570.86.16 should be both installed and available")
					}
					if r.ID.Version == "560.35.03" && (r.Installed || !r.Available) {
						return fmt.Errorf("560.35.03 should be available but not installed")
					}
				}
				return nil
			},
		},
		{
			name:           "ListWithHardwareDetection",
			listInst:       true,
			listAvail:      true,
			hwdetect:       true,
			compatibleOnly: false,
			setup: func(p *mocks.MockProvider, pm *mocks.MockPackageManager, rm *mocks.MockRepositoryManager) {
				p.EXPECT().GetID().Return("nvidia").AnyTimes()
				p.EXPECT().GetName().Return("NVIDIA").AnyTimes()
				p.EXPECT().DetectHardware().Return(true, nil)
				rm.EXPECT().GetRepoIDs(true).Return([]string{"rhel-10-for-x86_64-baseos-rpms"}, nil)
				pm.EXPECT().SetEnableRepos(gomock.Any())
				p.EXPECT().ListInstalled().Return([]api.DriverID{}, nil)
				p.EXPECT().ListAvailable().Return([]api.DriverID{
					{ProviderID: "nvidia", Version: "570.86.16"},
				}, nil)
			},
			expectErr: false,
			expectLen: 1,
			checkFunc: func(result []api.DriverStatus) error {
				if !result[0].Compatible {
					return fmt.Errorf("expected driver to be marked as compatible")
				}
				return nil
			},
		},
		{
			name:           "GetRepoIDsFails",
			listInst:       false,
			listAvail:      true,
			hwdetect:       false,
			compatibleOnly: false,
			setup: func(p *mocks.MockProvider, pm *mocks.MockPackageManager, rm *mocks.MockRepositoryManager) {
				rm.EXPECT().GetRepoIDs(true).Return(nil, fmt.Errorf("repo error"))
			},
			expectErr: true,
			expectLen: 0,
		},
		{
			name:           "GetRepoIDsReturnsEmpty",
			listInst:       false,
			listAvail:      true,
			hwdetect:       false,
			compatibleOnly: false,
			setup: func(p *mocks.MockProvider, pm *mocks.MockPackageManager, rm *mocks.MockRepositoryManager) {
				p.EXPECT().GetID().Return("nvidia").AnyTimes()
				p.EXPECT().GetName().Return("NVIDIA").AnyTimes()
				rm.EXPECT().GetRepoIDs(true).Return(nil, nil)
				p.EXPECT().ListAvailable().Return([]api.DriverID{}, nil)
			},
			expectErr: false,
			expectLen: 0,
		},
		{
			name:           "ListInstalledFails",
			listInst:       true,
			listAvail:      false,
			hwdetect:       false,
			compatibleOnly: false,
			setup: func(p *mocks.MockProvider, pm *mocks.MockPackageManager, rm *mocks.MockRepositoryManager) {
				p.EXPECT().GetID().Return("nvidia").AnyTimes()
				p.EXPECT().GetName().Return("NVIDIA").AnyTimes()
				p.EXPECT().ListInstalled().Return(nil, fmt.Errorf("list failed"))
			},
			expectErr: true,
			expectLen: 0,
		},
		{
			name:           "ListAvailableFails",
			listInst:       true,
			listAvail:      true,
			hwdetect:       false,
			compatibleOnly: false,
			setup: func(p *mocks.MockProvider, pm *mocks.MockPackageManager, rm *mocks.MockRepositoryManager) {
				p.EXPECT().GetID().Return("nvidia").AnyTimes()
				p.EXPECT().GetName().Return("NVIDIA").AnyTimes()
				rm.EXPECT().GetRepoIDs(true).Return([]string{"rhel-10-for-x86_64-baseos-rpms"}, nil)
				pm.EXPECT().SetEnableRepos(gomock.Any())
				p.EXPECT().ListInstalled().Return([]api.DriverID{}, nil)
				p.EXPECT().ListAvailable().Return(nil, fmt.Errorf("list failed"))
			},
			expectErr: true,
			expectLen: 0,
		},
		{
			name:           "EmptyResults",
			listInst:       true,
			listAvail:      true,
			hwdetect:       false,
			compatibleOnly: false,
			setup: func(p *mocks.MockProvider, pm *mocks.MockPackageManager, rm *mocks.MockRepositoryManager) {
				p.EXPECT().GetID().Return("nvidia").AnyTimes()
				p.EXPECT().GetName().Return("NVIDIA").AnyTimes()
				rm.EXPECT().GetRepoIDs(true).Return([]string{"rhel-10-for-x86_64-baseos-rpms"}, nil)
				pm.EXPECT().SetEnableRepos(gomock.Any())
				p.EXPECT().ListInstalled().Return([]api.DriverID{}, nil)
				p.EXPECT().ListAvailable().Return([]api.DriverID{}, nil)
			},
			expectErr: false,
			expectLen: 0,
		},
		{
			name:           "ListCompatibleOnlyFiltersToCompatible",
			listInst:       true,
			listAvail:      true,
			hwdetect:       true,
			compatibleOnly: true,
			setup: func(p *mocks.MockProvider, pm *mocks.MockPackageManager, rm *mocks.MockRepositoryManager) {
				p.EXPECT().GetID().Return("nvidia").AnyTimes()
				p.EXPECT().GetName().Return("NVIDIA").AnyTimes()
				p.EXPECT().DetectHardware().Return(true, nil)
				rm.EXPECT().GetRepoIDs(true).Return([]string{"rhel-10-for-x86_64-baseos-rpms"}, nil)
				pm.EXPECT().SetEnableRepos(gomock.Any())
				p.EXPECT().ListInstalled().Return([]api.DriverID{}, nil)
				p.EXPECT().ListAvailable().Return([]api.DriverID{
					{ProviderID: "nvidia", Version: "570"},
					{ProviderID: "nvidia", Version: "560"},
				}, nil)
			},
			expectErr: false,
			expectLen: 2,
			checkFunc: func(result []api.DriverStatus) error {
				for _, r := range result {
					if !r.Compatible {
						return fmt.Errorf("expected only compatible drivers, got Compatible=false for %s", r.ID.Version)
					}
				}
				return nil
			},
		},
		{
			name:           "ListCompatibleOnlyExcludesIncompatible",
			listInst:       true,
			listAvail:      true,
			hwdetect:       true,
			compatibleOnly: true,
			setup: func(p *mocks.MockProvider, pm *mocks.MockPackageManager, rm *mocks.MockRepositoryManager) {
				p.EXPECT().GetID().Return("nvidia").AnyTimes()
				p.EXPECT().GetName().Return("NVIDIA").AnyTimes()
				p.EXPECT().DetectHardware().Return(false, nil)
				rm.EXPECT().GetRepoIDs(true).Return([]string{"rhel-10-for-x86_64-baseos-rpms"}, nil)
				pm.EXPECT().SetEnableRepos(gomock.Any())
				p.EXPECT().ListInstalled().Return([]api.DriverID{}, nil)
				p.EXPECT().ListAvailable().Return([]api.DriverID{
					{ProviderID: "nvidia", Version: "570"},
				}, nil)
			},
			expectErr: false,
			expectLen: 0,
			checkFunc: func(result []api.DriverStatus) error {
				if len(result) != 0 {
					return fmt.Errorf("expected no drivers when hardware not compatible, got %d", len(result))
				}
				return nil
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

			result, err := List(deps, tt.listInst, tt.listAvail, tt.hwdetect, tt.compatibleOnly)
			if (err != nil) != tt.expectErr {
				t.Errorf("List() error = %v, expectErr %v", err, tt.expectErr)
				return
			}

			if len(result) != tt.expectLen {
				t.Errorf("List() returned %d results, expected %d", len(result), tt.expectLen)
				return
			}

			if tt.checkFunc != nil {
				if err := tt.checkFunc(result); err != nil {
					t.Errorf("List() check failed: %v", err)
				}
			}
		})
	}
}

func TestFilterCompatible(t *testing.T) {
	tests := []struct {
		name string
		in   []api.DriverStatus
		want []api.DriverStatus
	}{
		{
			name: "Empty",
			in:   nil,
			want: nil,
		},
		{
			name: "AllCompatible",
			in: []api.DriverStatus{
				{ID: api.DriverID{ProviderID: "nvidia", Version: "570"}, Compatible: true},
				{ID: api.DriverID{ProviderID: "nvidia", Version: "560"}, Compatible: true},
			},
			want: []api.DriverStatus{
				{ID: api.DriverID{ProviderID: "nvidia", Version: "570"}, Compatible: true},
				{ID: api.DriverID{ProviderID: "nvidia", Version: "560"}, Compatible: true},
			},
		},
		{
			name: "NoneCompatible",
			in: []api.DriverStatus{
				{ID: api.DriverID{ProviderID: "nvidia", Version: "570"}, Compatible: false},
			},
			want: nil,
		},
		{
			name: "MixedFiltersToCompatibleOnly",
			in: []api.DriverStatus{
				{ID: api.DriverID{ProviderID: "nvidia", Version: "570"}, Compatible: false},
				{ID: api.DriverID{ProviderID: "nvidia", Version: "560"}, Compatible: true},
				{ID: api.DriverID{ProviderID: "amd", Version: "1.0"}, Compatible: false},
				{ID: api.DriverID{ProviderID: "amd", Version: "2.0"}, Compatible: true},
			},
			want: []api.DriverStatus{
				{ID: api.DriverID{ProviderID: "nvidia", Version: "560"}, Compatible: true},
				{ID: api.DriverID{ProviderID: "amd", Version: "2.0"}, Compatible: true},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := filterCompatible(tt.in)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("filterCompatible() = %v, want %v", got, tt.want)
			}
		})
	}
}
