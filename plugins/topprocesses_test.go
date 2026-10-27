package plugins

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func mockTopProcesses(t *testing.T, loadPerCore, memoryUsedPercent float64) *bool {
	origLoad, origMemory, origSample := getLoadPerCore, getMemoryUsedPercent, sampleProcesses
	t.Cleanup(func() { getLoadPerCore, getMemoryUsedPercent, sampleProcesses = origLoad, origMemory, origSample })

	measuredCPU := new(bool)
	getLoadPerCore = func(context.Context) (float64, error) { return loadPerCore, nil }
	getMemoryUsedPercent = func(context.Context) (float64, error) { return memoryUsedPercent, nil }
	sampleProcesses = func(_ context.Context, measureCPU bool) ([]processSample, error) {
		*measuredCPU = measureCPU
		return []processSample{
			{pid: 1, name: "init", cpuPercent: 0, rss: 10 << 20},
			{pid: 100, name: "java", cpuPercent: 180, rss: 4 << 30},
			{pid: 200, name: "postgres", cpuPercent: 25, rss: 6 << 30},
			{pid: 300, name: "nginx", cpuPercent: 5, rss: 50 << 20},
		}, nil
	}
	return measuredCPU
}

var defaultTopProcessesSettings = topProcessesSettings{Load: 0.9, Memory: 90, Count: 3}

func TestTopProcessesNotShownWhenAllIsWell(t *testing.T) {
	mockTopProcesses(t, 0.5, 50)
	assert.Empty(t, TopProcesses(context.Background(), defaultTopProcessesSettings))
}

func TestTopProcessesHighLoad(t *testing.T) {
	measuredCPU := mockTopProcesses(t, 1.5, 50)

	rows := TopProcesses(context.Background(), defaultTopProcessesSettings)

	require.Len(t, rows, 1)
	assert.True(t, *measuredCPU)
	assert.Equal(t, "Top CPU", rows[0].Label.String())
	assert.Equal(t, "java (100) 180%, postgres (200) 25%, nginx (300) 5%", rows[0].Value.String())
}

func TestTopProcessesHighMemory(t *testing.T) {
	measuredCPU := mockTopProcesses(t, 0.1, 95)

	settings := defaultTopProcessesSettings
	settings.Count = 2
	rows := TopProcesses(context.Background(), settings)

	require.Len(t, rows, 1)
	// no need to wait for measuring the CPU usage
	assert.False(t, *measuredCPU)
	assert.Equal(t, "Top RAM", rows[0].Label.String())
	assert.Equal(t, "postgres (200) 6 GiB, java (100) 4 GiB", rows[0].Value.String())
}

func TestTopProcessesAlways(t *testing.T) {
	mockTopProcesses(t, 0.1, 10)

	settings := defaultTopProcessesSettings
	settings.Always = true
	rows := TopProcesses(context.Background(), settings)

	assert.Equal(t, []string{"Top CPU", "Top RAM"}, labels(rows))
}

func TestTopProcessesSettingsValidation(t *testing.T) {
	assert.ErrorContains(t, topProcessesSettings{Count: 0}.Validate(), "count must be at least 1")
}
