package plugins

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/gdubicki/dynamotd/dynamotd"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockReboot sets up a fake /var/run and /boot, with the given kernels installed in that order.
func mockReboot(t *testing.T, runningKernel string, rebootPackages *string, installedKernels ...string) {
	origFile, origBoot, origKernel := rebootRequiredFile, bootDir, getKernelVersion
	t.Cleanup(func() { rebootRequiredFile, bootDir, getKernelVersion = origFile, origBoot, origKernel })

	dir := t.TempDir()
	rebootRequiredFile = filepath.Join(dir, "reboot-required")
	if rebootPackages != nil {
		require.NoError(t, os.WriteFile(rebootRequiredFile, []byte("*** System restart required ***\n"), 0o644))
		if *rebootPackages != "" {
			require.NoError(t, os.WriteFile(rebootRequiredFile+".pkgs", []byte(*rebootPackages), 0o644))
		}
	}

	bootDir = filepath.Join(dir, "boot")
	require.NoError(t, os.Mkdir(bootDir, 0o755))
	installTime := time.Now().Add(-time.Hour)
	for _, kernel := range installedKernels {
		path := filepath.Join(bootDir, "vmlinuz-"+kernel)
		require.NoError(t, os.WriteFile(path, nil, 0o644))
		require.NoError(t, os.Chtimes(path, installTime, installTime))
		installTime = installTime.Add(time.Minute)
	}

	getKernelVersion = func(context.Context) (string, error) { return runningKernel, nil }
}

func rebootValue(t *testing.T) string {
	rows := RebootRequired(context.Background(), dynamotd.NoSettings{})
	if len(rows) == 0 {
		return ""
	}
	require.Len(t, rows, 1)
	return rows[0].Value.String()
}

func ptr(s string) *string { return &s }

func TestRebootRequired(t *testing.T) {
	tests := []struct {
		name     string
		running  string
		packages *string
		kernels  []string
		expected string
	}{
		{"nothing installed", "6.8.0-45-generic", nil, nil, ""},
		{"running the newest kernel", "6.8.0-50-generic", nil,
			[]string{"6.8.0-45-generic", "6.8.0-50-generic"}, ""},
		{"newer kernel installed", "6.8.0-45-generic", nil,
			[]string{"6.8.0-45-generic", "6.8.0-50-generic"},
			"required - most recently installed kernel: 6.8.0-50-generic, running: 6.8.0-45-generic"},
		{"rescue and unversioned kernels are ignored", "5.14.0-427.el9.x86_64", nil,
			[]string{"5.14.0-427.el9.x86_64", "0-rescue-abc123", "linux"}, ""},
		{"Debian/Ubuntu flag file without packages", "6.8.0-50-generic", ptr(""), nil, "required"},
		{"Debian/Ubuntu flag file with packages", "6.8.0-50-generic", ptr("libc6\nlinux-image-6.8.0-51-generic\nlibc6\n"), nil,
			"required - for libc6, linux-image-6.8.0-51-generic"},
		{"Debian/Ubuntu flag file with many packages", "6.8.0-50-generic", ptr("a\nb\nc\nd\n"), nil,
			"required - for a, b, c and more"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockReboot(t, tt.running, tt.packages, tt.kernels...)
			assert.Equal(t, tt.expected, rebootValue(t))
		})
	}
}
