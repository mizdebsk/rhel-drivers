package core

import (
	"fmt"
	"strings"

	"github.com/mizdebsk/rhel-drivers/internal/api"
)

func parseDriverID(input string) (api.DriverID, error) {
	parts := strings.SplitN(input, ":", 2)
	if len(parts) != 2 || parts[0] == "" {
		return api.DriverID{}, fmt.Errorf("invalid driver ID format: %q (expected 'vendor:version')", input)
	}
	return api.DriverID{
		ProviderID: parts[0],
		Version:    parts[1],
	}, nil
}
