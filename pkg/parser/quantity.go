package parser

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// ParseCPUToMillicores converts Kubernetes CPU quantity to millicores
// Examples: "100m" -> 100, "0.5" -> 500, "1" -> 1000, "2.5" -> 2500
func ParseCPUToMillicores(value string) (int64, error) {
	if value == "" {
		return 0, nil
	}

	value = strings.TrimSpace(value)

	// Handle millicore format (e.g., "100m")
	if strings.HasSuffix(value, "m") {
		millis := strings.TrimSuffix(value, "m")
		return strconv.ParseInt(millis, 10, 64)
	}

	// Handle decimal/integer format (e.g., "0.5", "1", "2.5")
	cores, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid CPU quantity: %s", value)
	}

	// Convert cores to millicores
	return int64(cores * 1000), nil
}

// ParseMemoryToMiB converts Kubernetes memory quantity to MiB
// Examples: "128Mi" -> 128, "1Gi" -> 1024, "512Mi" -> 512, "1024Ki" -> 1
func ParseMemoryToMiB(value string) (int64, error) {
	if value == "" {
		return 0, nil
	}

	value = strings.TrimSpace(value)

	// Regex to extract number and unit
	re := regexp.MustCompile(`^(\d+(?:\.\d+)?)\s*([A-Za-z]*)$`)
	matches := re.FindStringSubmatch(value)

	if len(matches) != 3 {
		return 0, fmt.Errorf("invalid memory quantity: %s", value)
	}

	numStr := matches[1]
	unit := matches[2]

	num, err := strconv.ParseFloat(numStr, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid memory quantity: %s", value)
	}

	// Convert to MiB based on unit
	var mib float64
	switch unit {
	case "Ki":
		mib = num / 1024 // KiB to MiB
	case "Mi", "":
		mib = num // Already in MiB
	case "Gi":
		mib = num * 1024 // GiB to MiB
	case "Ti":
		mib = num * 1024 * 1024 // TiB to MiB
	case "Pi":
		mib = num * 1024 * 1024 * 1024 // PiB to MiB
	case "Ei":
		mib = num * 1024 * 1024 * 1024 * 1024 // EiB to MiB
	// Decimal units (less common, but supported)
	case "k", "K":
		mib = num * 1000 / 1024 / 1024 // KB to MiB
	case "M":
		mib = num * 1000 * 1000 / 1024 / 1024 // MB to MiB
	case "G":
		mib = num * 1000 * 1000 * 1000 / 1024 / 1024 // GB to MiB
	case "T":
		mib = num * 1000 * 1000 * 1000 * 1000 / 1024 / 1024 // TB to MiB
	default:
		// Assume bytes if no unit
		mib = num / 1024 / 1024
	}

	return int64(mib), nil
}
