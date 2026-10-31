//go:build screenshots

package plugins

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/fatih/color"
	"github.com/gdubicki/dynamotd/dynamotd"
	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/disk"
	"github.com/shirou/gopsutil/v4/host"
	"github.com/shirou/gopsutil/v4/mem"
	"github.com/stretchr/testify/require"
)

// Generates the example outputs for the README screenshots, using the real rows with example data:
// tools/screenshots/generate.sh

var screenshotNow = time.Date(2026, 10, 27, 22, 30, 0, 0, time.Local)

type scenario struct {
	uptime        time.Duration
	load          [3]float64
	cpuBefore     cpu.TimesStat
	cpuAfter      cpu.TimesStat
	memory        mem.VirtualMemoryStat
	swap          mem.SwapMemoryStat
	oomKills      int
	lastOOMKill   string
	processes     []processSample
	partitions    []disk.PartitionStat
	usages        map[string]disk.UsageStat
	failedUnits   []string
	rebootPkgs    *string
	clockInSync   bool
	users         []string
	externalIPErr error
}

const cores = 4

func (s scenario) apply(t *testing.T) {
	dir := t.TempDir()

	sampleCPUTimes = func(context.Context) (cpu.TimesStat, cpu.TimesStat, error) { return s.cpuBefore, s.cpuAfter, nil }
	showIOWaitAndSteal = true

	getVirtualMemory = func(context.Context) (*mem.VirtualMemoryStat, error) { return &s.memory, nil }
	getSwap = func(context.Context) (*mem.SwapMemoryStat, error) { return &s.swap, nil }

	vmstatPath = dir + "/vmstat"
	require.NoError(t, os.WriteFile(vmstatPath, []byte(fmt.Sprintf("oom_kill %d\n", s.oomKills)), 0o644))
	bootTime := screenshotNow.Add(-s.uptime)
	getBootTime = func() (uint64, error) { return uint64(bootTime.Unix()), nil }
	now = func() time.Time { return screenshotNow }
	readKernelLog = func() ([]string, error) {
		if s.lastOOMKill == "" {
			return []string{"6,1,1000000,-;Linux version 6.8.0-45-generic"}, nil
		}
		killTime := screenshotNow.Add(-2*time.Hour - 16*time.Minute).Sub(bootTime).Microseconds()
		return []string{fmt.Sprintf("3,500,%d,-;Out of memory: Killed process 2412 (%s) total-vm:12000000kB", killTime, s.lastOOMKill)}, nil
	}

	getLoadPerCore = func(context.Context) (float64, error) { return s.load[0] / cores, nil }
	getMemoryUsedPercent = func(context.Context) (float64, error) {
		return float64(s.memory.Total-s.memory.Available) / float64(s.memory.Total) * 100, nil
	}
	sampleProcesses = func(context.Context, bool) ([]processSample, error) { return s.processes, nil }

	getPartitions = func(context.Context, bool) ([]disk.PartitionStat, error) { return s.partitions, nil }
	getUsage = func(_ context.Context, path string) (*disk.UsageStat, error) {
		usage := s.usages[path]
		usage.Path = path
		return &usage, nil
	}

	systemdRunning = func() bool { return true }
	listFailedUnits = func(context.Context) ([]string, error) { return s.failedUnits, nil }

	rebootRequiredFile = dir + "/reboot-required"
	if s.rebootPkgs != nil {
		require.NoError(t, os.WriteFile(rebootRequiredFile, nil, 0o644))
		require.NoError(t, os.WriteFile(rebootRequiredFile+".pkgs", []byte(*s.rebootPkgs), 0o644))
	}
	bootDir = dir

	clockSynchronized = func() (bool, bool) { return s.clockInSync, true }

	getUsers = func(context.Context) ([]host.UserStat, error) {
		var users []host.UserStat
		for _, user := range s.users {
			users = append(users, host.UserStat{User: user})
		}
		return users, nil
	}
}

// rows in the default order, using the real row types where they can be fed with example data,
// and the same formatting as the real ones for the others (timestamp, FQDN, IP, uptime, load)
func (s scenario) rows(t *testing.T) []dynamotd.Row {
	ctx := context.Background()
	single := func(label string, value ...dynamotd.ColorString) []dynamotd.Row {
		return []dynamotd.Row{{Label: dynamotd.SingleColorLabel(label), Value: dynamotd.ToColorText(value...)}}
	}

	externalIP := ipColorString("203.0.113.45", s.externalIPErr)
	uptimeSeconds := uint64(s.uptime.Seconds())
	uptimeThresholds := dynamotd.Thresholds{Warning: 90, Critical: 365}
	loadThresholds := dynamotd.Thresholds{Warning: 0.9, Critical: 1.1}
	load := func(value float64) dynamotd.ColorString {
		return dynamotd.ColorString{Color: loadThresholds.Color(value / cores), Text: fmt.Sprintf("%.2f", value)}
	}

	var rows []dynamotd.Row
	for _, rowRows := range [][]dynamotd.Row{
		single(timestampLabel, dynamotd.ValueNeutral(screenshotNow.Format("2006-01-02 15:04:05 MST"))),
		Clock(ctx, dynamotd.NoSettings{}),
		{{}},
		single(fqdnLabel, dynamotd.ValueNeutral("web01.example.com")),
		single(ipLabel, ipColorString("10.0.1.23", nil), dynamotd.ValueDescription(" (external: "), externalIP, dynamotd.ValueDescription(")")),
		single(uptimeLabel, dynamotd.ColorString{Color: GetUptimeColor(uptimeSeconds, uptimeThresholds), Text: GetUptimeString(uptimeSeconds)}),
		RebootRequired(ctx, dynamotd.NoSettings{}),
		{{}},
		Users(ctx, dynamotd.NoSettings{}),
		{{}},
		single(loadLabel, load(s.load[0]), dynamotd.ValueDescription(" / "), load(s.load[1]), dynamotd.ValueDescription(" / "), load(s.load[2]),
			dynamotd.ValueDescription(" with "), dynamotd.ValueNeutral(fmt.Sprintf("%d", cores)), dynamotd.ValueDescription(" cores at "), dynamotd.ValueNeutral("3500 MHz")),
		CPUUsage(ctx, defaultCPUUsageSettings),
		TopProcesses(ctx, topProcessesSettings{Load: 0.9, Memory: 90, Count: 3}),
		Memory(ctx, memorySettings{Thresholds: dynamotd.Thresholds{Warning: 90, Critical: 95}, Swap: dynamotd.Thresholds{Warning: 50, Critical: 80}}),
		OOMKills(ctx, dynamotd.NoSettings{}),
		DiskSpace(ctx, diskSpaceSettings{Thresholds: dynamotd.Thresholds{Warning: 75, Critical: 90},
			Inodes: dynamotd.Thresholds{Warning: 75, Critical: 90}, Always: []string{"/"}, Others: true}),
		{{}},
		FailedUnits(ctx, dynamotd.NoSettings{}),
	} {
		rows = append(rows, rowRows...)
	}
	return rows
}

func printTo(t *testing.T, path string, rows []dynamotd.Row) {
	file, err := os.Create(path)
	require.NoError(t, err)
	defer file.Close()

	stdout := os.Stdout
	os.Stdout = file
	defer func() { os.Stdout = stdout }()
	dynamotd.Print(rows)
}

func TestScreenshots(t *testing.T) {
	color.NoColor = false
	const gib = 1024 * 1024 * 1024
	day := 24 * time.Hour

	allGood := scenario{
		uptime:      12*day + 7*time.Hour + 41*time.Minute,
		load:        [3]float64{0.42, 0.38, 0.35},
		cpuBefore:   cpu.TimesStat{},
		cpuAfter:    cpu.TimesStat{User: 6, System: 2, Iowait: 0.4, Idle: 91.6},
		memory:      mem.VirtualMemoryStat{Total: 16 * gib, Available: 11*gib + gib/3},
		swap:        mem.SwapMemoryStat{Total: 4 * gib, Used: 0, UsedPercent: 0},
		partitions:  []disk.PartitionStat{{Device: "/dev/sda1", Mountpoint: "/", Fstype: "ext4"}},
		usages:      map[string]disk.UsageStat{"/": {Total: 100 * gib, Used: 38*gib + gib/5, UsedPercent: 38.2, InodesTotal: 1000, InodesUsedPercent: 12}},
		clockInSync: true,
	}

	rebootPkgs := "linux-image-6.8.0-51-generic\nlibc6\nopenssl\nsystemd\n"
	allProblems := scenario{
		uptime:    400*day + 3*time.Hour + 12*time.Minute,
		load:      [3]float64{7.92, 6.41, 4.13},
		cpuBefore: cpu.TimesStat{},
		cpuAfter:  cpu.TimesStat{User: 28, System: 9, Iowait: 41, Steal: 7, Idle: 15},
		memory:    mem.VirtualMemoryStat{Total: 16 * gib, Available: gib / 2},
		swap:      mem.SwapMemoryStat{Total: 4 * gib, Used: 3*gib + gib/2, UsedPercent: 87.5},
		oomKills:  3, lastOOMKill: "java",
		processes: []processSample{
			{pid: 2412, name: "java", cpuPercent: 187, rss: 9*gib + gib/2},
			{pid: 1890, name: "postgres", cpuPercent: 64, rss: 3*gib + gib/8},
			{pid: 30121, name: "rsync", cpuPercent: 22, rss: gib / 10},
			{pid: 811, name: "nginx", cpuPercent: 4, rss: gib / 20},
		},
		partitions: []disk.PartitionStat{
			{Device: "/dev/sda1", Mountpoint: "/", Fstype: "ext4"},
			{Device: "/dev/sdb1", Mountpoint: "/var", Fstype: "xfs"},
			{Device: "/dev/sdc1", Mountpoint: "/srv/data", Fstype: "ext4"},
		},
		usages: map[string]disk.UsageStat{
			"/":         {Total: 50 * gib, Used: 41 * gib, UsedPercent: 82, InodesTotal: 1000, InodesUsedPercent: 31},
			"/var":      {Total: 200 * gib, Used: 194 * gib, UsedPercent: 97, InodesTotal: 1000, InodesUsedPercent: 20},
			"/srv/data": {Total: 500 * gib, Used: 61 * gib, UsedPercent: 12.2, InodesTotal: 1000, InodesUsedPercent: 96},
		},
		failedUnits:   []string{"nginx.service", "backup.timer"},
		rebootPkgs:    &rebootPkgs,
		clockInSync:   false,
		users:         []string{"alice", "bob", "bob"},
		externalIPErr: fmt.Errorf("timed out"),
	}

	outDir := os.Getenv("SCREENSHOTS_DIR")
	require.NotEmpty(t, outDir)

	allGood.apply(t)
	printTo(t, outDir+"/all-good.ansi", allGood.rows(t))

	allProblems.apply(t)
	printTo(t, outDir+"/all-problems.ansi", allProblems.rows(t))
}
