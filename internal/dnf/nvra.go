package dnf

import (
	"strings"
)

func parseNameFromNVRA(nvra string) string {
	last := max(0, strings.LastIndex(nvra, "-"))
	prev := max(0, strings.LastIndex(nvra[:last], "-"))
	return nvra[:prev]
}
