package plugins

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/fatih/color"
	"github.com/gdubicki/dynamotd/dynamotd"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var distroTestNow = time.Date(2026, 11, 8, 22, 30, 0, 0, time.Local)

const ubuntuOSRelease = `PRETTY_NAME="Ubuntu 24.04.1 LTS"
NAME="Ubuntu"
VERSION_ID="24.04"
VERSION="24.04.1 LTS (Noble Numbat)"
ID=ubuntu
ID_LIKE=debian
`

// mockDistro makes Distro use the given os-release file content (none if empty), the given releases
// fetched from endoflife.date (or the error), and a temporary cache directory, which it returns
func mockDistro(t *testing.T, osRelease string, releases []eolRelease, fetchErr error) (cacheDir string, fetches *int) {
	origPaths, origCacheDir, origFetch, origNow, origChecks := osReleasePaths, eolCacheDir, fetchEOLReleases, now, extendedSupportChecks
	t.Cleanup(func() {
		osReleasePaths, eolCacheDir, fetchEOLReleases, now, extendedSupportChecks = origPaths, origCacheDir, origFetch, origNow, origChecks
	})
	// to not depend on the extended support of the machine running the tests
	extendedSupportChecks = map[string]func() *extendedSupport{}

	dir := t.TempDir()
	osReleasePaths = []string{filepath.Join(dir, "os-release")}
	if osRelease != "" {
		require.NoError(t, os.WriteFile(osReleasePaths[0], []byte(osRelease), 0o644))
	}
	cacheDir = filepath.Join(dir, "cache")
	eolCacheDir = func() (string, error) { return cacheDir, nil }
	fetches = new(int)
	fetchEOLReleases = func(context.Context, string) ([]eolRelease, error) {
		*fetches++
		return releases, fetchErr
	}
	now = func() time.Time { return distroTestNow }
	return cacheDir, fetches
}

var defaultDistroSettings = distroSettings{Warning: 90, Critical: 0}

func TestDistro(t *testing.T) {
	tests := []struct {
		name      string
		osRelease string
		releases  []eolRelease
		fetchErr  error
		expected  string
	}{
		{"supported", ubuntuOSRelease, []eolRelease{{Name: "24.04", EOLFrom: "2029-05-31"}}, nil,
			"Ubuntu 24.04.1 LTS (supported until 2029-05-31)"},
		{"EOL soon", ubuntuOSRelease, []eolRelease{{Name: "24.04", EOLFrom: "2026-12-31"}}, nil,
			"Ubuntu 24.04.1 LTS - EOL in 53 days (on 2026-12-31)"},
		{"EOL", ubuntuOSRelease, []eolRelease{{Name: "24.04", IsEOL: true, EOLFrom: "2026-05-31"}}, nil,
			"Ubuntu 24.04.1 LTS - EOL since 2026-05-31"},
		{"release not found", ubuntuOSRelease, []eolRelease{{Name: "26.04", EOLFrom: "2031-05-29"}}, nil,
			"Ubuntu 24.04.1 LTS"},
		{"can't fetch", ubuntuOSRelease, nil, errors.New("timed out"),
			"Ubuntu 24.04.1 LTS (can't check EOL: timed out)"},
		{"no EOL data for the distro", "PRETTY_NAME=\"Arch Linux\"\nID=arch\n", nil, errors.New("not called"),
			"Arch Linux"},
		{"no version", "PRETTY_NAME=\"Debian GNU/Linux trixie/sid\"\nID=debian\n", nil, errors.New("not called"),
			"Debian GNU/Linux trixie/sid"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDistro(t, tt.osRelease, tt.releases, tt.fetchErr)
			rows := Distro(context.Background(), defaultDistroSettings)
			require.Len(t, rows, 1)
			assert.Equal(t, distroLabel, rows[0].Label.String())
			assert.Equal(t, tt.expected, rows[0].Value.String())
		})
	}
}

func TestDistroWithExtendedSupport(t *testing.T) {
	mockDistro(t, ubuntuOSRelease, []eolRelease{{Name: "24.04", EOLFrom: "2026-05-31", EOESFrom: "2031-04-30"}}, nil)
	extendedSupportChecks["ubuntu"] = func() *extendedSupport { return &extendedSupport{name: "ESM"} }
	rows := Distro(context.Background(), defaultDistroSettings)
	require.Len(t, rows, 1)
	assert.Equal(t, "Ubuntu 24.04.1 LTS (supported with ESM until 2031-04-30)", rows[0].Value.String())
}

func TestDistroWithoutOSRelease(t *testing.T) {
	mockDistro(t, "", nil, nil)
	assert.Empty(t, Distro(context.Background(), defaultDistroSettings))
}

func TestEOLValue(t *testing.T) {
	tests := []struct {
		name          string
		release       eolRelease
		settings      distroSettings
		expectedText  string
		expectedColor color.Attribute
	}{
		{"supported", eolRelease{EOLFrom: "2027-02-07"}, defaultDistroSettings,
			"Ubuntu (supported until 2027-02-07)", dynamotd.ValueNeutralColor},
		{"warning threshold", eolRelease{EOLFrom: "2027-02-06"}, defaultDistroSettings,
			"Ubuntu - EOL in 90 days (on 2027-02-06)", dynamotd.ValueWarningColor},
		{"tomorrow", eolRelease{EOLFrom: "2026-11-09"}, defaultDistroSettings,
			"Ubuntu - EOL tomorrow", dynamotd.ValueWarningColor},
		{"today", eolRelease{EOLFrom: "2026-11-08"}, defaultDistroSettings,
			"Ubuntu - EOL since 2026-11-08", dynamotd.ValueCriticalColor},
		{"custom critical threshold", eolRelease{EOLFrom: "2026-11-20"}, distroSettings{Warning: 90, Critical: 30},
			"Ubuntu - EOL in 12 days (on 2026-11-20)", dynamotd.ValueCriticalColor},
		{"EOL without a date", eolRelease{IsEOL: true}, defaultDistroSettings,
			"Ubuntu - EOL", dynamotd.ValueCriticalColor},
		{"EOL not announced", eolRelease{}, defaultDistroSettings,
			"Ubuntu", dynamotd.ValueNeutralColor},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			value := dynamotd.ToColorText(eolValue("Ubuntu", tt.release, nil, tt.settings, distroTestNow)...)
			assert.Equal(t, tt.expectedText, value.String())
			assert.Equal(t, tt.expectedColor, value.Parts()[0].Color)
		})
	}
}

func TestEOLValueWithExtendedSupport(t *testing.T) {
	esm := &extendedSupport{name: "ESM"}
	expiresOn := func(date string) *extendedSupport {
		expires, err := time.Parse(time.RFC3339, date+"T00:00:00+00:00")
		require.NoError(t, err)
		return &extendedSupport{name: "ESM", expires: expires, subscription: "the Ubuntu Pro subscription"}
	}
	tests := []struct {
		name          string
		release       eolRelease
		support       *extendedSupport
		expectedText  string
		expectedColor color.Attribute
	}{
		{"supported", eolRelease{EOLFrom: "2025-05-31", EOESFrom: "2030-05-31"}, esm,
			"Ubuntu (supported with ESM until 2030-05-31)", dynamotd.ValueNeutralColor},
		{"ends soon", eolRelease{EOLFrom: "2021-04-30", EOESFrom: "2026-12-31"}, esm,
			"Ubuntu - EOL in 53 days (on 2026-12-31, end of ESM)", dynamotd.ValueWarningColor},
		{"ended", eolRelease{EOLFrom: "2019-04-30", EOESFrom: "2024-04-30"}, esm,
			"Ubuntu - EOL since 2024-04-30 (end of ESM)", dynamotd.ValueCriticalColor},
		{"no extended support date", eolRelease{EOLFrom: "2025-05-31"}, esm,
			"Ubuntu - EOL since 2025-05-31", dynamotd.ValueCriticalColor},
		{"subscription ends first", eolRelease{EOLFrom: "2025-05-31", EOESFrom: "2030-05-31"}, expiresOn("2027-06-15"),
			"Ubuntu (supported with ESM until 2027-06-15, end of the Ubuntu Pro subscription)", dynamotd.ValueNeutralColor},
		{"subscription ended", eolRelease{EOLFrom: "2025-05-31", EOESFrom: "2030-05-31"}, expiresOn("2026-10-01"),
			"Ubuntu - EOL since 2026-10-01 (end of the Ubuntu Pro subscription)", dynamotd.ValueCriticalColor},
		{"subscription ends after the extended support", eolRelease{EOLFrom: "2025-05-31", EOESFrom: "2030-05-31"}, expiresOn("2099-01-01"),
			"Ubuntu (supported with ESM until 2030-05-31)", dynamotd.ValueNeutralColor},
		{"subscription ends before the standard support", eolRelease{EOLFrom: "2029-05-31", EOESFrom: "2034-04-30"}, expiresOn("2026-10-01"),
			"Ubuntu (supported until 2029-05-31)", dynamotd.ValueNeutralColor},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			value := dynamotd.ToColorText(eolValue("Ubuntu", tt.release, tt.support, defaultDistroSettings, distroTestNow)...)
			assert.Equal(t, tt.expectedText, value.String())
			assert.Equal(t, tt.expectedColor, value.Parts()[0].Color)
		})
	}
}

func TestMatchEOLRelease(t *testing.T) {
	releases := []eolRelease{{Name: "3.20"}, {Name: "15.6"}, {Name: "9"}, {Name: "2023"}}
	tests := []struct {
		version  string
		expected string
	}{
		{"3.20.3", "3.20"},
		{"15.6", "15.6"},
		{"9.4", "9"},
		{"2023", "2023"},
		{"3.2", ""},
		{"10.1", ""},
	}
	for _, tt := range tests {
		release, found := matchEOLRelease(releases, tt.version)
		assert.Equal(t, tt.expected != "", found, "version %s", tt.version)
		assert.Equal(t, tt.expected, release.Name, "version %s", tt.version)
	}
}

func TestEOLProduct(t *testing.T) {
	assert.Equal(t, "rhel", eolProduct(map[string]string{"ID": "rhel"}))
	assert.Equal(t, "centos", eolProduct(map[string]string{"ID": "centos", "NAME": "CentOS Linux"}))
	assert.Equal(t, "centos-stream", eolProduct(map[string]string{"ID": "centos", "NAME": "CentOS Stream"}))
	assert.Empty(t, eolProduct(map[string]string{"ID": "arch"}))
}

func TestParseOSRelease(t *testing.T) {
	values := parseOSRelease("# comment\n" +
		"NAME=\"Red Hat Enterprise Linux\"\n" +
		"VERSION_ID='9.4'\n" +
		"ID=rhel\n" +
		`PRETTY_NAME="Say \"hi\" for \$5 \\ \` + "` \"\n" +
		"INVALID\n")
	assert.Equal(t, map[string]string{
		"NAME":        "Red Hat Enterprise Linux",
		"VERSION_ID":  "9.4",
		"ID":          "rhel",
		"PRETTY_NAME": "Say \"hi\" for $5 \\ ` ",
	}, values)
}

func TestDistroName(t *testing.T) {
	assert.Equal(t, "Ubuntu 24.04.1 LTS", distroName(map[string]string{"PRETTY_NAME": "Ubuntu 24.04.1 LTS", "NAME": "Ubuntu"}))
	assert.Equal(t, "Ubuntu 24.04.1 LTS", distroName(map[string]string{"NAME": "Ubuntu", "VERSION": "24.04.1 LTS"}))
	assert.Equal(t, "Linux", distroName(map[string]string{}))
}

func TestGetEOLReleasesCache(t *testing.T) {
	ctx := context.Background()
	fetched := []eolRelease{{Name: "24.04", EOLFrom: "2029-05-31"}}
	cacheDir, fetches := mockDistro(t, "", fetched, nil)

	// fetched and cached
	releases, err := getEOLReleases(ctx, "ubuntu")
	require.NoError(t, err)
	assert.Equal(t, fetched, releases)
	assert.Equal(t, 1, *fetches)
	assert.FileExists(t, filepath.Join(cacheDir, "eol-ubuntu.json"))

	// from the cache
	releases, err = getEOLReleases(ctx, "ubuntu")
	require.NoError(t, err)
	assert.Equal(t, fetched, releases)
	assert.Equal(t, 1, *fetches)

	// the cache expired, but the fetch fails, so the stale cache is used
	now = func() time.Time { return distroTestNow.Add(eolCacheTTL) }
	fetchEOLReleases = func(context.Context, string) ([]eolRelease, error) {
		*fetches++
		return nil, errors.New("timed out")
	}
	releases, err = getEOLReleases(ctx, "ubuntu")
	require.NoError(t, err)
	assert.Equal(t, fetched, releases)
	assert.Equal(t, 2, *fetches)

	// without the cache, the error is returned
	_, err = getEOLReleases(ctx, "debian")
	assert.EqualError(t, err, "timed out")
}

func TestFetchEOLReleases(t *testing.T) {
	origURL := eolAPIURL
	t.Cleanup(func() { eolAPIURL = origURL })

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/ubuntu":
			_, _ = w.Write([]byte(`{"result": {"name": "ubuntu", "releases": [
				{"name": "26.04", "isEol": false, "eolFrom": "2031-05-29", "eoesFrom": "2036-04-23"},
				{"name": "6.06", "isEol": true, "eolFrom": null}
			]}}`))
		case "/garbage":
			_, _ = w.Write([]byte(`<html>`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()
	eolAPIURL = server.URL + "/"

	releases, err := fetchEOLReleasesReal(context.Background(), "ubuntu")
	require.NoError(t, err)
	assert.Equal(t, []eolRelease{{Name: "26.04", EOLFrom: "2031-05-29", EOESFrom: "2036-04-23"}, {Name: "6.06", IsEOL: true}}, releases)

	_, err = fetchEOLReleasesReal(context.Background(), "garbage")
	assert.EqualError(t, err, "unexpected response")

	_, err = fetchEOLReleasesReal(context.Background(), "nope")
	assert.EqualError(t, err, "unexpected HTTP status 404 Not Found")
}

func TestDistroSettingsValidate(t *testing.T) {
	_, err := dynamotd.Lookup([]dynamotd.RowConfig{{Name: "distro", Settings: settingsNodeFor(t, "warning: 10\ncritical: 30")}})
	assert.ErrorContains(t, err, "warning threshold (10 days) is lower than the critical one (30 days)")
}
