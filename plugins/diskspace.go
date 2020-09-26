package plugins

import (
	"fmt"
	humanize "github.com/dustin/go-humanize"
	. "github.com/gdubicki/dynamotd/dynamotd"
	"github.com/minio/minio/pkg/disk"
)

func DiskSpace(path string) Row {
	var warningThreshold = 75.0
	var criticalThreshold = 90.0

	var color Color

	di, err := disk.GetInfo(path)
	if err != nil {
		panic(fmt.Errorf("error getting disk info"))
	}

	percentageUsed := (float64(di.Total-di.Free)/float64(di.Total))*100

	if percentageUsed >= warningThreshold {
		color = ValueCriticalColor
	} else if percentageUsed >= criticalThreshold {
		color = ValueWarningColor
	} else {
		color = ValueOkColor
	}

	return Row{
		Label: SingleColorLabel("Disk space (/)"),
		Value: ToColorText(
			ColorString{Color: color, Text: fmt.Sprintf("%s", humanize.Bytes(di.Total-di.Free))},
			ValueDescription(" of "),
			ValueNeutral(fmt.Sprintf("%s", humanize.Bytes(di.Total))),
			ValueDescription(" disk space used ("),
			ColorString{Color: color, Text: fmt.Sprintf("%0.2f%%", percentageUsed)},
			ValueDescription(")"),
		),
	}
}
