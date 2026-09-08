package amd

import (
	"reflect"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/mizdebsk/radii/internal/api"
	"github.com/mizdebsk/radii/internal/mocks"
)

func TestInstallAndRemoveLatest(t *testing.T) {
	provider := NewProvider(nil).(*prov)
	drivers := []api.DriverID{{ProviderID: "amdgpu", Version: variantLatest}}
	want := []string{pkgKmodAmdgpu, pkgRocm}

	installed, err := provider.Install(drivers)
	if err != nil {
		t.Fatalf("Install() error = %v", err)
	}
	if !reflect.DeepEqual(installed, want) {
		t.Fatalf("Install() = %v, want %v", installed, want)
	}

	removed, err := provider.Remove(drivers)
	if err != nil {
		t.Fatalf("Remove() error = %v", err)
	}
	if !reflect.DeepEqual(removed, want) {
		t.Fatalf("Remove() = %v, want %v", removed, want)
	}
}

func TestInstallAndRemoveRejectUnknownVariant(t *testing.T) {
	provider := NewProvider(nil).(*prov)
	drivers := []api.DriverID{{ProviderID: "amdgpu", Version: "unknown"}}

	if _, err := provider.Install(drivers); err == nil {
		t.Fatal("Install() error = nil, want non-nil")
	}
	if _, err := provider.Remove(drivers); err == nil {
		t.Fatal("Remove() error = nil, want non-nil")
	}
}

func TestListInstalledPackageStates(t *testing.T) {
	tests := []struct {
		name      string
		installed []api.PackageInfo
		want      []api.DriverID
	}{
		{name: "neither package", installed: nil, want: []api.DriverID{}},
		{
			name:      "ROCm only",
			installed: []api.PackageInfo{{Name: pkgRocm}},
			want:      []api.DriverID{{ProviderID: "amdgpu", Version: variantLatest}},
		},
		{
			name:      "kernel module only",
			installed: []api.PackageInfo{{Name: pkgKmodAmdgpu}},
			want:      []api.DriverID{{ProviderID: "amdgpu", Version: variantLatest}},
		},
		{
			name: "complete stack",
			installed: []api.PackageInfo{
				{Name: pkgKmodAmdgpu},
				{Name: pkgRocm},
			},
			want: []api.DriverID{{ProviderID: "amdgpu", Version: variantLatest}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			packageManager := mocks.NewMockPackageManager(ctrl)
			packageManager.EXPECT().ListInstalledPackages().Return(tt.installed, nil)
			provider := NewProvider(packageManager).(*prov)

			got, err := provider.ListInstalled()
			if err != nil {
				t.Fatalf("ListInstalled() error = %v", err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("ListInstalled() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestListAvailablePackageStates(t *testing.T) {
	tests := []struct {
		name      string
		available []api.PackageInfo
		want      []api.DriverID
	}{
		{name: "neither package", available: nil, want: []api.DriverID{}},
		{name: "ROCm only", available: []api.PackageInfo{{Name: pkgRocm}}, want: []api.DriverID{}},
		{name: "kernel module only", available: []api.PackageInfo{{Name: pkgKmodAmdgpu}}, want: []api.DriverID{}},
		{
			name: "complete stack",
			available: []api.PackageInfo{
				{Name: pkgKmodAmdgpu},
				{Name: pkgRocm},
			},
			want: []api.DriverID{{ProviderID: "amdgpu", Version: variantLatest}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			packageManager := mocks.NewMockPackageManager(ctrl)
			packageManager.EXPECT().ListAvailablePackages().Return(tt.available, nil)
			provider := NewProvider(packageManager).(*prov)

			got, err := provider.ListAvailable()
			if err != nil {
				t.Fatalf("ListAvailable() error = %v", err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("ListAvailable() = %v, want %v", got, tt.want)
			}
		})
	}
}
