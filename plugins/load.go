package plugins

import (
	"fmt"
	linuxproc "github.com/c9s/goprocinfo/linux"
	. "github.com/gdubicki/dynamotd/dynamotd"
	"log"
	"os/exec"
	"regexp"
	"runtime"
	"strconv"
)

func Load() Row {
	// TODO: make below configurable
	var warningThreshold = 0.9
	var criticalThreshold = 1.1

	var cores = float64(runtime.NumCPU())

	var load1, load5, load15 float64

	if runtime.GOOS == "linux" {

		stat, err := linuxproc.ReadLoadAvg("/proc/stat")
		if err != nil {
			return Row{
				Label: SingleColorLabel("Load"),
				Value: SingleColorValue("Can't read /proc/stat"),
			}
		}

		load1 = stat.Last1Min
		load5 = stat.Last5Min
		load15 = stat.Last15Min

	} else { // # macos

		cmd := exec.Command("uptime")
		out, err := cmd.CombinedOutput()
		if err != nil {
			log.Fatalf("Running 'uptime' failed with %s\n", err)
		}

		loads := getParams(`load average: (?P<Load1>\d+\.\d+), (?P<Load5>\d+\.\d+), (?P<Load15>\d+\.\d+)`, string(out))

		var err1, err2, err3 error
		load1, err1 = strconv.ParseFloat(loads["Load1"], 64)
		load5, err2 = strconv.ParseFloat(loads["Load5"], 64)
		load15, err3 = strconv.ParseFloat(loads["Load15"], 64)
		if err1 != nil || err2 != nil || err3 != nil {
			log.Fatalf("Interpreting 'uptime' load values (%s, %s, %s) failed\n",
				loads["Load1"], loads["Load5"], loads["Load15"])
		}
	}

	load1color := ValueOkColor
	load5color := ValueOkColor
	load15color := ValueOkColor

	if load1 >= warningThreshold*cores {
		load1color = ValueWarningColor
	}
	if load5 >= warningThreshold*cores {
		load5color = ValueWarningColor
	}
	if load15 >= warningThreshold*cores {
		load15color = ValueWarningColor
	}

	if load1 >= criticalThreshold*cores {
		load1color = ValueCriticalColor
	}
	if load5 >= criticalThreshold*cores {
		load5color = ValueCriticalColor
	}
	if load15 >= criticalThreshold*cores {
		load15color = ValueCriticalColor
	}

	// text format like: 0.12, 0.4, 0.5 (1 / 5 / 15), with colors
	return Row{
		Label: SingleColorLabel("Load"),
		Value: ToColorText(
			ColorString{Color: load1color, Text: fmt.Sprintf("%.2f", load1)},
			ValueDescription(" / "),
			ColorString{Color: load5color, Text: fmt.Sprintf("%.2f", load5)},
			ValueDescription(" / "),
			ColorString{Color: load15color, Text: fmt.Sprintf("%.2f", load15)},
			ValueDescription(" (1 / 5 / 15)"),
		),
	}
}

/**
 * Parses input string with the given regular expression and returns the
 * group values defined in the expression.
 *
 * Copied from https://stackoverflow.com/a/39635221/2693875
 */
func getParams(regEx, input string) (paramsMap map[string]string) {

	var compRegEx = regexp.MustCompile(regEx)
	match := compRegEx.FindStringSubmatch(input)

	paramsMap = make(map[string]string)
	for i, name := range compRegEx.SubexpNames() {
		if i > 0 && i <= len(match) {
			paramsMap[name] = match[i]
		}
	}
	return
}
