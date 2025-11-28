package core

import (
	"testing"

	"github.com/golang/mock/gomock"

	"github.com/mizdebsk/rhel-drivers/internal/api"
	"github.com/mizdebsk/rhel-drivers/internal/mocks"
)

func TestParseDriverID(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expected    api.DriverID
		expectError bool
	}{
		{
			name:  "ValidInput",
			input: "nvidia:450.80.02",
			expected: api.DriverID{
				ProviderID: "nvidia",
				Version:    "450.80.02",
			},
		},
		{
			name:        "MissingVersion",
			input:       "nvidia:",
			expected:    api.DriverID{},
			expectError: true,
		},
		{
			name:        "MissingVendor",
			input:       ":450.80.02",
			expected:    api.DriverID{},
			expectError: true,
		},
		{
			name:        "EmptyInput",
			input:       "",
			expected:    api.DriverID{},
			expectError: true,
		},
		{
			name:        "NoSeparator",
			input:       "nvidia450.80.02",
			expected:    api.DriverID{},
			expectError: true,
		},
		{
			name:  "ValidInputWithSpaces",
			input: "  nvidia  :  450.80.02  ",
			expected: api.DriverID{
				ProviderID: "  nvidia  ",
				Version:    "  450.80.02  ",
			},
		},
		{
			name:        "ColonOnly",
			input:       ":",
			expected:    api.DriverID{},
			expectError: true,
		},
		{
			name:        "MultipleColons",
			input:       "nvidia:450.80.02:extra",
			expected:    api.DriverID{},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseDriverID(tt.input)
			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got nil for input %q", tt.input)
				}
			} else {
				if err != nil {
					t.Errorf("did not expect error but got %v for input %q", err, tt.input)
				}
				if result != tt.expected {
					t.Errorf("expected %v but got %v for input %q", tt.expected, result, tt.input)
				}
			}
		})
	}
}

func TestLookupProvider(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	prov1 := mocks.NewMockProvider(ctrl)
	prov2 := mocks.NewMockProvider(ctrl)
	prov1.EXPECT().GetID().Return("Ann").AnyTimes()
	prov2.EXPECT().GetID().Return("Ben").AnyTimes()

	deps := api.CoreDeps{
		Providers: []api.Provider{prov1, prov2},
	}

	tests := []struct {
		name       string
		prov       string
		expectErr  bool
		expectProv api.Provider
	}{
		{
			name:       "Found",
			prov:       "Ann",
			expectProv: prov1,
		},
		{
			name:      "NotFound",
			prov:      "Bob",
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			driver := api.DriverID{ProviderID: tt.prov}
			prov, err := lookupProvider(deps, driver)

			if tt.expectErr {
				if err == nil {
					t.Fatalf("lookupProvider() error = nil, want non-nil")
				}
				if prov != nil {
					t.Fatalf("lookupProvider() provider = %v, want nil", prov)
				}
			} else {
				if err != nil {
					t.Fatalf("lookupProvider() unexpected error = %v", err)
				}
				if prov != tt.expectProv {
					t.Fatalf("lookupProvider() = %v, want %v", prov, tt.expectProv)
				}
			}
		})
	}
}

func TestResolveDriver(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	prov1 := mocks.NewMockProvider(ctrl)
	prov2 := mocks.NewMockProvider(ctrl)
	prov1.EXPECT().GetID().Return("Alpha").AnyTimes()
	prov2.EXPECT().GetID().Return("Beta").AnyTimes()

	deps := api.CoreDeps{
		Providers: []api.Provider{prov1, prov2},
	}

	tests := []struct {
		name       string
		driver     string
		expectErr  bool
		expectProv api.Provider
		expectVer  string
	}{
		{
			name:       "Found",
			driver:     "Alpha:1.0",
			expectProv: prov1,
			expectVer:  "1.0",
		},
		{
			name:      "InvalidFormat",
			driver:    "Gamma",
			expectErr: true,
		},
		{
			name:      "ProviderNotFound",
			driver:    "Sigma:42",
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			driver, prov, err := resolveDriver(deps, tt.driver)

			if tt.expectErr {
				if err == nil {
					t.Fatalf("lookupProvider() error = nil, want non-nil")
				}
				if prov != nil {
					t.Fatalf("lookupProvider() provider = %v, want nil", prov)
				}
			} else {
				if err != nil {
					t.Fatalf("lookupProvider() unexpected error = %v", err)
				}
				if prov != tt.expectProv {
					t.Fatalf("lookupProvider() provider = %v, want %v", prov, tt.expectProv)
				}
				expectedDriver := api.DriverID{ProviderID: tt.expectProv.GetID(), Version: tt.expectVer}
				if driver != expectedDriver {
					t.Fatalf("lookupProvider() driver = %v, want %v", driver, expectedDriver)
				}
			}
		})
	}
}
