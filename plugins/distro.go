package plugins

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/gdubicki/dynamotd/dynamotd"
)

func init() {
	dynamotd.Register(distroLabel, distroSettings{Warning: 90, Critical: 0}, Distro, "distro")
}

const distroLabel = "Distro"

// distroSettings are the thresholds in days of support left until the EOL.
type distroSettings struct {
	Warning  int `yaml:"warning"`
	Critical int `yaml:"critical"`
}

func (settings distroSettings) Validate() error {
	if settings.Warning < settings.Critical {
		return fmt.Errorf("warning threshold (%d days) is lower than the critical one (%d days)", settings.Warning, settings.Critical)
	}
	return nil
}

// eolProducts maps the os-release IDs to the endoflife.date products whose release names are
// the os-release VERSION_ID or its prefix, f.e. "9" for RHEL's "9.4".
var eolProducts = map[string]string{
	"almalinux":     "almalinux",
	"alpine":        "alpine-linux",
	"amzn":          "amazon-linux",
	"centos":        "centos",
	"debian":        "debian",
	"devuan":        "devuan",
	"fedora":        "fedora",
	"linuxmint":     "linuxmint",
	"nixos":         "nixos",
	"ol":            "oracle-linux",
	"opensuse-leap": "opensuse",
	"pop":           "pop-os",
	"rhel":          "rhel",
	"rocky":         "rocky-linux",
	"sles":          "sles",
	"ubuntu":        "ubuntu",
}

// EOL dates rarely change, but new releases appear, so it's enough to refresh them daily
const eolCacheTTL = 24 * time.Hour

const maxEOLResponseSize = 1024 * 1024

// for mocking in tests
var osReleasePaths = []string{"/etc/os-release", "/usr/lib/os-release"}
var eolAPIURL = "https://endoflife.date/api/v1/products/"
var eolCacheDir = defaultEOLCacheDir
var fetchEOLReleases = fetchEOLReleasesReal

type eolRelease struct {
	Name    string `json:"name"`
	IsEOL   bool   `json:"isEol"`
	EOLFrom string `json:"eolFrom"`
	// the end of the paid extended support, f.e. Ubuntu ESM
	EOESFrom string `json:"eoesFrom"`
}

type eolCache struct {
	FetchedAt time.Time    `json:"fetchedAt"`
	Releases  []eolRelease `json:"releases"`
}

// Distro shows the distro name and version, and when its support ends (EOL), using the data from endoflife.date.
func Distro(ctx context.Context, settings distroSettings) []dynamotd.Row {
	osRelease, err := readOSRelease()
	if errors.Is(err, fs.ErrNotExist) {
		// not Linux
		return nil
	}
	if err != nil {
		return []dynamotd.Row{dynamotd.ErrorRow(distroLabel, err)}
	}
	return []dynamotd.Row{{
		Label: dynamotd.SingleColorLabel(distroLabel),
		Value: dynamotd.ToColorText(distroValue(ctx, osRelease, settings)...),
	}}
}

func distroValue(ctx context.Context, osRelease map[string]string, settings distroSettings) []dynamotd.ColorString {
	name := distroName(osRelease)
	product := eolProduct(osRelease)
	versionID := osRelease["VERSION_ID"]
	// f.e. Arch Linux or Debian testing, which don't have EOL dates
	if product == "" || versionID == "" {
		return []dynamotd.ColorString{dynamotd.ValueNeutral(name)}
	}

	releases, err := getEOLReleases(ctx, product)
	if err != nil {
		return []dynamotd.ColorString{dynamotd.ValueNeutral(name), dynamotd.ValueDescription(fmt.Sprintf(" (can't check EOL: %s)", err))}
	}
	release, found := matchEOLRelease(releases, versionID)
	if !found {
		return []dynamotd.ColorString{dynamotd.ValueNeutral(name)}
	}
	var support *extendedSupport
	if check, ok := extendedSupportChecks[product]; ok {
		support = check()
	}
	return eolValue(name, release, support, settings, now())
}

// supportEnd is when the support ends, with the reason if it's not the standard support
type supportEnd struct {
	date string
	// f.e. "ESM", empty for the standard support
	extendedSupport string
	// f.e. "the Ubuntu Pro subscription", if it ends earlier than the extended support
	subscription string
}

func getSupportEnd(release eolRelease, support *extendedSupport) supportEnd {
	end := supportEnd{date: release.EOLFrom}
	if support == nil || release.EOESFrom == "" {
		return end
	}
	extended := supportEnd{date: release.EOESFrom, extendedSupport: support.name}
	if !support.expires.IsZero() {
		if expires := support.expires.UTC().Format(time.DateOnly); expires < extended.date {
			extended.date, extended.subscription = expires, support.subscription
		}
	}
	// the dates have the same format, so they can be compared as strings
	if extended.date > end.date {
		return extended
	}
	return end
}

func eolValue(name string, release eolRelease, support *extendedSupport, settings distroSettings, today time.Time) []dynamotd.ColorString {
	end := getSupportEnd(release, support)
	eol, err := time.Parse(time.DateOnly, end.date)
	if err != nil {
		if release.IsEOL {
			return []dynamotd.ColorString{dynamotd.ValueCritical(name), dynamotd.ValueDescription(" - "), dynamotd.ValueCritical("EOL")}
		}
		// the EOL date is not announced yet
		return []dynamotd.ColorString{dynamotd.ValueNeutral(name)}
	}

	daysLeft := int(eol.Sub(time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, time.UTC)).Hours() / 24)

	var eolColor color.Attribute
	switch {
	case daysLeft <= settings.Critical:
		eolColor = dynamotd.ValueCriticalColor
	case daysLeft <= settings.Warning:
		eolColor = dynamotd.ValueWarningColor
	default:
		supported := " (supported until " + end.date
		if end.extendedSupport != "" {
			supported = " (supported with " + end.extendedSupport + " until " + end.date
		}
		if end.subscription != "" {
			supported += ", end of " + end.subscription
		}
		return []dynamotd.ColorString{dynamotd.ValueNeutral(name), dynamotd.ValueDescription(supported + ")")}
	}

	var status string
	var details []string
	switch {
	case daysLeft <= 0:
		status = "EOL since " + end.date
	case daysLeft == 1:
		status = "EOL tomorrow"
	default:
		status = fmt.Sprintf("EOL in %d days", daysLeft)
		details = append(details, "on "+end.date)
	}
	switch {
	case end.subscription != "":
		details = append(details, "end of "+end.subscription)
	case end.extendedSupport != "":
		details = append(details, "end of "+end.extendedSupport)
	}

	value := []dynamotd.ColorString{
		{Color: eolColor, Text: name},
		dynamotd.ValueDescription(" - "),
		{Color: eolColor, Text: status},
	}
	if len(details) > 0 {
		value = append(value, dynamotd.ValueDescription(" ("+strings.Join(details, ", ")+")"))
	}
	return value
}

func distroName(osRelease map[string]string) string {
	if name := osRelease["PRETTY_NAME"]; name != "" {
		return name
	}
	if name := strings.TrimSpace(osRelease["NAME"] + " " + osRelease["VERSION"]); name != "" {
		return name
	}
	return "Linux"
}

func eolProduct(osRelease map[string]string) string {
	// CentOS Stream has the same ID as CentOS Linux, but different releases
	if osRelease["ID"] == "centos" && strings.Contains(osRelease["NAME"], "Stream") {
		return "centos-stream"
	}
	return eolProducts[osRelease["ID"]]
}

// matchEOLRelease finds the release with the name being the version or its longest prefix ending before a dot.
func matchEOLRelease(releases []eolRelease, version string) (eolRelease, bool) {
	for {
		for _, release := range releases {
			if release.Name == version {
				return release, true
			}
		}
		dot := strings.LastIndex(version, ".")
		if dot < 0 {
			return eolRelease{}, false
		}
		version = version[:dot]
	}
}

func readOSRelease() (map[string]string, error) {
	for _, path := range osReleasePaths {
		data, err := os.ReadFile(path)
		if errors.Is(err, fs.ErrNotExist) {
			continue
		}
		if err != nil {
			return nil, err
		}
		return parseOSRelease(string(data)), nil
	}
	return nil, fs.ErrNotExist
}

// parseOSRelease parses the os-release format: https://www.freedesktop.org/software/systemd/man/latest/os-release.html
func parseOSRelease(data string) map[string]string {
	values := map[string]string{}
	for _, line := range strings.Split(data, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		if len(value) >= 2 && (value[0] == '"' || value[0] == '\'') && value[len(value)-1] == value[0] {
			if value[0] == '"' {
				value = osReleaseEscapes.Replace(value[1 : len(value)-1])
			} else {
				value = value[1 : len(value)-1]
			}
		}
		values[key] = value
	}
	return values
}

var osReleaseEscapes = strings.NewReplacer(`\"`, `"`, `\\`, `\`, `\$`, `$`, "\\`", "`")

// getEOLReleases returns the releases of the product, cached to not slow down every login,
// and to show them even if endoflife.date is temporarily unreachable.
func getEOLReleases(ctx context.Context, product string) ([]eolRelease, error) {
	cachePath := ""
	if dir, err := eolCacheDir(); err == nil {
		cachePath = filepath.Join(dir, "eol-"+product+".json")
	}

	cached, cacheErr := readEOLCache(cachePath)
	if cacheErr == nil && now().Sub(cached.FetchedAt) < eolCacheTTL {
		return cached.Releases, nil
	}

	releases, err := fetchEOLReleases(ctx, product)
	if err != nil {
		if cacheErr == nil {
			return cached.Releases, nil
		}
		return nil, err
	}
	// it's just a cache, so ignore the errors
	_ = writeEOLCache(cachePath, eolCache{FetchedAt: now(), Releases: releases})
	return releases, nil
}

func defaultEOLCacheDir() (string, error) {
	if os.Geteuid() == 0 {
		return "/var/cache/dynamotd", nil
	}
	dir, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "dynamotd"), nil
}

func readEOLCache(path string) (eolCache, error) {
	var cache eolCache
	if path == "" {
		return cache, fs.ErrNotExist
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return cache, err
	}
	err = json.Unmarshal(data, &cache)
	return cache, err
}

// writeEOLCache writes the cache atomically, as dynamotd may run for multiple logins at the same time.
func writeEOLCache(path string, cache eolCache) error {
	if path == "" {
		return fs.ErrNotExist
	}
	data, err := json.Marshal(cache)
	if err != nil {
		return err
	}
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	file, err := os.CreateTemp(dir, ".eol-*.json")
	if err != nil {
		return err
	}
	defer func() { _ = os.Remove(file.Name()) }()
	if _, err := file.Write(data); err != nil {
		_ = file.Close()
		return err
	}
	if err := file.Close(); err != nil {
		return err
	}
	return os.Rename(file.Name(), path)
}

func fetchEOLReleasesReal(ctx context.Context, product string) ([]eolRelease, error) {
	// the overall timeout is in the context
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, eolAPIURL+product, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "dynamotd")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, describeHTTPError(err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected HTTP status %s", resp.Status)
	}

	var body struct {
		Result struct {
			Releases []eolRelease `json:"releases"`
		} `json:"result"`
	}
	if err := json.NewDecoder(io.LimitReader(resp.Body, maxEOLResponseSize)).Decode(&body); err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			return nil, fmt.Errorf("timed out")
		}
		return nil, fmt.Errorf("unexpected response")
	}
	if len(body.Result.Releases) == 0 {
		return nil, fmt.Errorf("unexpected response")
	}
	return body.Result.Releases, nil
}
