package plugins

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// extendedSupport is the paid support after the standard one ends, which endoflife.date calls "eoes".
type extendedSupport struct {
	name string
	// when the subscription ends, zero if unknown
	expires      time.Time
	subscription string
}

// extendedSupportChecks return the extended support enabled on this machine, nil if none, by endoflife.date product.
var extendedSupportChecks = map[string]func() *extendedSupport{
	"debian": debianELTS,
	"rhel":   rhelELS,
	"sles":   slesLTSS,
	"ubuntu": ubuntuESM,
}

// for mocking in tests
var ubuntuESMRepoFiles = []string{
	"/etc/apt/sources.list.d/ubuntu-esm-infra.sources",
	"/etc/apt/sources.list.d/ubuntu-esm-infra.list",
}
var ubuntuProStatusFile = "/var/lib/ubuntu-advantage/status.json"
var aptSourcesFile = "/etc/apt/sources.list"
var aptSourcesDir = "/etc/apt/sources.list.d"
var rhelReposFile = "/etc/yum.repos.d/redhat.repo"
var slesLTSSProductFile = "/etc/products.d/SLES-LTSS.prod"

// ubuntuESM checks if Ubuntu Pro's ESM for the main repository is enabled, as the Pro client
// adds its apt source then, and when the Pro subscription ends, from the status cached by the Pro client.
func ubuntuESM() *extendedSupport {
	enabled := false
	for _, path := range ubuntuESMRepoFiles {
		if _, err := os.Stat(path); err == nil {
			enabled = true
		}
	}
	if !enabled {
		return nil
	}

	support := &extendedSupport{name: "ESM"}
	var status struct {
		Expires string `json:"expires"`
	}
	if data, err := os.ReadFile(ubuntuProStatusFile); err == nil && json.Unmarshal(data, &status) == nil {
		if expires, err := time.Parse(time.RFC3339, status.Expires); err == nil {
			support.expires, support.subscription = expires, "the Ubuntu Pro subscription"
		}
	}
	return support
}

// debianELTS checks if there is an apt source of Freexian's Extended LTS.
func debianELTS() *extendedSupport {
	paths := []string{aptSourcesFile}
	for _, pattern := range []string{"*.list", "*.sources"} {
		matches, _ := filepath.Glob(filepath.Join(aptSourcesDir, pattern))
		paths = append(paths, matches...)
	}
	for _, path := range paths {
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		for _, line := range strings.Split(string(data), "\n") {
			line = strings.TrimSpace(line)
			if !strings.HasPrefix(line, "#") && strings.Contains(line, "deb.freexian.com/extended-lts") {
				return &extendedSupport{name: "ELTS"}
			}
		}
	}
	return nil
}

// rhelELS checks if any Extended Life Cycle Support repository, f.e. rhel-7-server-els-rpms,
// is enabled in the repositories file managed by subscription-manager.
func rhelELS() *extendedSupport {
	data, err := os.ReadFile(rhelReposFile)
	if err != nil {
		return nil
	}
	section := ""
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			section = line[1 : len(line)-1]
			continue
		}
		key, value, ok := strings.Cut(line, "=")
		if ok && strings.Contains(section, "-els-") && strings.TrimSpace(key) == "enabled" && strings.TrimSpace(value) == "1" {
			return &extendedSupport{name: "ELS"}
		}
	}
	return nil
}

// slesLTSS checks if the Long Term Service Pack Support product is registered, as SUSEConnect installs its file then.
func slesLTSS() *extendedSupport {
	if _, err := os.Stat(slesLTSSProductFile); err != nil {
		return nil
	}
	return &extendedSupport{name: "LTSS"}
}
