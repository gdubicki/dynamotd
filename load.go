package main

import (
	"fmt"
	linuxproc "github.com/c9s/goprocinfo/linux"
	"runtime"
)

func load() Row {
	var warningThreshold = 0.9
	var criticalThreshold = 1.1

	var cores = float64(runtime.NumCPU())

	var load1, load5, load15 float64

	if runtime.GOOS == "linux" {

		stat, err := linuxproc.ReadLoadAvg("/proc/stat")
		if err != nil {
			return Row{
				singleColorLabel("Load"),
				singleColorValue("Can't read /proc/stat"),
			}
		}

		load1 = stat.Last1Min
		load5 = stat.Last5Min
		load15 = stat.Last15Min

	} else { // # macos

		//cmd := exec.Command("uptime")
		//out, _ := cmd.CombinedOutput()
		//err := cmd.Run()
		//if err != nil {
		//	log.Fatalf("Running 'uptime' failed with %s\n", err)
		//}
		//
		//re := regexp.MustCompile(`load average: (.*)`)
		//loadAverageString := re.FindSubmatch(out)
		//loadAverageStringSplitted := strings.Split(string(loadAverageString[0]), ", ")
		//
		//var err1, err2, err3 error
		//load1, err1 = strconv.ParseFloat(loadAverageStringSplitted[0], 64)
		//load5, err2 = strconv.ParseFloat(loadAverageStringSplitted[1], 64)
		//load15, err3 = strconv.ParseFloat(loadAverageStringSplitted[2], 64)
		//if err1 != nil || err2 != nil || err3 != nil {
		//	log.Fatalf("Interpreting 'uptime' load values (%s, %s, %s) failed\n",
		//		loadAverageStringSplitted[0], loadAverageStringSplitted[1], loadAverageStringSplitted[2])
		//}
		load1 = 20.0
		load5 = 12.1
		load15 = 3.0
	}

	load1color := valueOkColor
	load5color := valueOkColor
	load15color := valueOkColor

	if load1 >= warningThreshold*cores {
		load1color = valueWarningColor
	}
	if load5 >= warningThreshold*cores {
		load5color = valueWarningColor
	}
	if load15 >= warningThreshold*cores {
		load15color = valueWarningColor
	}

	if load1 >= criticalThreshold*cores {
		load1color = valueCriticalColor
	}
	if load5 >= criticalThreshold*cores {
		load5color = valueCriticalColor
	}
	if load15 >= criticalThreshold*cores {
		load15color = valueCriticalColor
	}

	// text format like: 0.12, 0.4, 0.5 (1 / 5 / 15), with colors
	return Row{
		singleColorLabel("Load"),
		toColorText(
			ColorString{load1color, fmt.Sprintf("%.2f", load1)},
			valueDescription(" / "),
			ColorString{load5color, fmt.Sprintf("%.2f", load5)},
			valueDescription(" / "),
			ColorString{load15color, fmt.Sprintf("%.2f", load15)},
			valueDescription(" (1 / 5 / 15)"),
		),
	}
}
