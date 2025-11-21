package rhsm

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"

	"github.com/mizdebsk/rhel-drivers/internal/log"
)

func repoEnabled(path, repoID string) bool {
	f, err := os.Open(filepath.Clean(path))
	if err != nil {
		return false
	}
	defer func() {
		if err := f.Close(); err != nil {
			log.Warnf("failed to close file %s: %v", path, err)
		}
	}()

	sc := bufio.NewScanner(f)
	inSection := false

	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" {
			continue
		}

		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			section := strings.TrimSpace(line[1 : len(line)-1])
			inSection = section == repoID
			continue
		}
		if !inSection {
			continue
		}

		if strings.HasPrefix(strings.ToLower(line), "enabled") {
			parts := strings.SplitN(line, "=", 2)
			if len(parts) != 2 {
				return false
			}
			val := parts[1]
			val = strings.SplitN(val, "#", 2)[0]
			val = strings.ToLower(strings.TrimSpace(val))

			return val == "1" || val == "true" || val == "yes" || val == "on"
		}
	}

	return false
}
