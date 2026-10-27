package plugins

import (
	"context"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"
	"unicode"

	"github.com/gdubicki/dynamotd/dynamotd"
	"github.com/shirou/gopsutil/v4/host"
)

func init() {
	dynamotd.Register(rebootLabel, dynamotd.NoSettings{}, RebootRequired, "reboot")
}

const rebootLabel = "Reboot"

// show at most this many package names, to keep the row short
const maxRebootPackagesShown = 3

// for mocking in tests
var rebootRequiredFile = "/var/run/reboot-required"
var bootDir = "/boot"
var getKernelVersion = host.KernelVersionWithContext

// RebootRequired shows if a reboot is required, f.e. to run the newest installed kernel.
func RebootRequired(ctx context.Context, _ dynamotd.NoSettings) []dynamotd.Row {
	reason, required := rebootReason(ctx)
	if !required {
		return nil
	}

	value := []dynamotd.ColorString{dynamotd.ValueWarning("required")}
	if reason != "" {
		value = append(value, dynamotd.ValueDescription(" - "+reason))
	}
	return []dynamotd.Row{{
		Label: dynamotd.SingleColorLabel(rebootLabel),
		Value: dynamotd.ToColorText(value...),
	}}
}

func rebootReason(ctx context.Context) (string, bool) {
	// Debian and Ubuntu packages create this file when they need a reboot,
	// with the list of the packages in the .pkgs file
	if _, err := os.Stat(rebootRequiredFile); err == nil {
		packages := readRebootPackages(rebootRequiredFile + ".pkgs")
		if len(packages) == 0 {
			return "", true
		}
		shown := packages[:min(len(packages), maxRebootPackagesShown)]
		reason := "for " + strings.Join(shown, ", ")
		if len(packages) > len(shown) {
			reason += " and more"
		}
		return reason, true
	}

	// on the other distros, check at least if there is a newer kernel installed
	running, err := getKernelVersion(ctx)
	if err != nil {
		return "", false
	}
	newest := newestInstalledKernel()
	if newest != "" && newest != running {
		return "most recently installed kernel: " + newest + ", running: " + running, true
	}
	return "", false
}

func readRebootPackages(path string) []string {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	var packages []string
	for _, line := range strings.Split(string(data), "\n") {
		if line = strings.TrimSpace(line); line != "" && !slices.Contains(packages, line) {
			packages = append(packages, line)
		}
	}
	return packages
}

// newestInstalledKernel returns the version of the most recently installed kernel in /boot,
// or an empty string if it can't be determined.
func newestInstalledKernel() string {
	paths, err := filepath.Glob(filepath.Join(bootDir, "vmlinuz-*"))
	if err != nil {
		return ""
	}

	var newest string
	var newestTime time.Time
	for _, path := range paths {
		version := strings.TrimPrefix(filepath.Base(path), "vmlinuz-")
		// skip f.e. RHEL's "vmlinuz-0-rescue-..." and Arch's "vmlinuz-linux" without the version
		if strings.Contains(version, "rescue") || version == "" || !unicode.IsDigit(rune(version[0])) {
			continue
		}
		info, err := os.Stat(path)
		if err != nil {
			continue
		}
		if info.ModTime().After(newestTime) {
			newest, newestTime = version, info.ModTime()
		}
	}
	return newest
}
