package plugins

import (
	"fmt"
	"github.com/Tonyfilla/go-humanize"
	"github.com/fatih/color"
	. "github.com/gdubicki/dynamotd/dynamotd"
	"github.com/minio/minio/pkg/disk"
)

func DiskSpace(path string) Row {
	var warningThreshold = 75.0
	var criticalThreshold = 90.0

	var color color.Attribute

	di, err := disk.GetInfo(path)
	if err != nil {
		panic(fmt.Errorf("error getting disk info"))
	}

	percentageUsed := (float64(di.Total-di.Free) / float64(di.Total)) * 100

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
			ColorString{Color: color, Text: fmt.Sprintf("%s", humanize.IBytesCustomCeil(di.Total-di.Free, 2))},
			ValueDescription(" of "),
			ValueNeutral(fmt.Sprintf("%s", humanize.IBytesCustomCeil(di.Total, 0))),
			ValueDescription(" disk space used ("),
			ColorString{Color: color, Text: fmt.Sprintf("%0.2f%%", percentageUsed)},
			ValueDescription(")"),
		),
	}
}
