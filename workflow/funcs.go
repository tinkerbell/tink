package workflow

import (
	"fmt"
	"strings"
)

// templateFuncs defines the custom functions available to workflow templates.
var templateFuncs = map[string]interface{}{
	"contains":        strings.Contains,
	"hasPrefix":       strings.HasPrefix,
	"hasSuffix":       strings.HasSuffix,
	"formatPartition": formatPartition,
}

// formatPartition formats a device path with partition for the device type. If it receives an
// unidentifiable device path it returns the dev.
//
// Examples
// 		formatPartition("/dev/nvme0n1", 0) -> /dev/nvme0n1p1
// 		formatPartition("/dev/sda", 1) -> /dev/sda1
func formatPartition(dev string, partition int) string {
	switch {
	case strings.HasPrefix(dev, "/dev/nvme"):
		return fmt.Sprintf("%vp%v", dev, partition)
	case strings.HasPrefix(dev, "/dev/sd"):
		return fmt.Sprintf("%v%v", dev, partition)
	}
	return dev
}
