package plugins

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// writeFile writes the file in the directory, creating its parent directories, and returns its path
func writeFile(t *testing.T, dir, name, content string) string {
	path := filepath.Join(dir, name)
	require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o755))
	require.NoError(t, os.WriteFile(path, []byte(content), 0o644))
	return path
}

func TestUbuntuESM(t *testing.T) {
	origRepoFiles, origStatusFile := ubuntuESMRepoFiles, ubuntuProStatusFile
	t.Cleanup(func() { ubuntuESMRepoFiles, ubuntuProStatusFile = origRepoFiles, origStatusFile })
	dir := t.TempDir()
	ubuntuESMRepoFiles = []string{filepath.Join(dir, "ubuntu-esm-infra.sources"), filepath.Join(dir, "ubuntu-esm-infra.list")}
	ubuntuProStatusFile = filepath.Join(dir, "status.json")

	assert.Nil(t, ubuntuESM())

	writeFile(t, dir, "ubuntu-esm-infra.list", "deb https://esm.ubuntu.com/infra/ubuntu focal-infra-security main\n")
	assert.Equal(t, &extendedSupport{name: "ESM"}, ubuntuESM())

	writeFile(t, dir, "status.json", `{"attached": true, "expires": "2027-01-15T00:00:00+00:00", "services": []}`)
	support := ubuntuESM()
	require.NotNil(t, support)
	assert.Equal(t, time.Date(2027, 1, 15, 0, 0, 0, 0, time.UTC), support.expires.UTC())
	assert.Equal(t, "the Ubuntu Pro subscription", support.subscription)
}

func TestDebianELTS(t *testing.T) {
	origFile, origDir := aptSourcesFile, aptSourcesDir
	t.Cleanup(func() { aptSourcesFile, aptSourcesDir = origFile, origDir })
	dir := t.TempDir()
	aptSourcesFile = writeFile(t, dir, "sources.list", "deb http://deb.debian.org/debian bullseye main\n")
	aptSourcesDir = filepath.Join(dir, "sources.list.d")

	assert.Nil(t, debianELTS())

	writeFile(t, aptSourcesDir, "commented.list", "# deb http://deb.freexian.com/extended-lts bullseye-lts main\n")
	assert.Nil(t, debianELTS())

	writeFile(t, aptSourcesDir, "extended-lts.sources", "Types: deb\nURIs: http://deb.freexian.com/extended-lts\nSuites: bullseye-lts\n")
	assert.Equal(t, &extendedSupport{name: "ELTS"}, debianELTS())
}

func TestRHELELS(t *testing.T) {
	origFile := rhelReposFile
	t.Cleanup(func() { rhelReposFile = origFile })
	dir := t.TempDir()
	rhelReposFile = filepath.Join(dir, "redhat.repo")

	assert.Nil(t, rhelELS())

	writeFile(t, dir, "redhat.repo", `[rhel-7-server-rpms]
name = Red Hat Enterprise Linux 7 Server (RPMs)
enabled = 1

[rhel-7-server-els-rpms]
name = Red Hat Enterprise Linux 7 Server - Extended Life Cycle Support (RPMs)
enabled = 0
`)
	assert.Nil(t, rhelELS())

	writeFile(t, dir, "redhat.repo", `[rhel-7-server-els-rpms]
name = Red Hat Enterprise Linux 7 Server - Extended Life Cycle Support (RPMs)
enabled = 1
`)
	assert.Equal(t, &extendedSupport{name: "ELS"}, rhelELS())
}

func TestSLESLTSS(t *testing.T) {
	origFile := slesLTSSProductFile
	t.Cleanup(func() { slesLTSSProductFile = origFile })
	dir := t.TempDir()
	slesLTSSProductFile = filepath.Join(dir, "SLES-LTSS.prod")

	assert.Nil(t, slesLTSS())

	writeFile(t, dir, "SLES-LTSS.prod", "<product/>\n")
	assert.Equal(t, &extendedSupport{name: "LTSS"}, slesLTSS())
}
