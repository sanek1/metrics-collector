package storage

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInitPoolMetrics(t *testing.T) {
	metrics := make(map[string]float64)
	InitPoolMetrics(metrics)

	expectedMetrics := []string{
		"Alloc",
		"BuckHashSys",
		"Frees",
		"GCCPUFraction",
		"GCSys",
		"HeapAlloc",
		"HeapIdle",
		"HeapInuse",
		"HeapObjects",
		"HeapReleased",
		"HeapSys",
		"LastGC",
		"Lookups",
		"MCacheInuse",
		"MCacheSys",
		"MSpanInuse",
		"MSpanSys",
		"Mallocs",
		"NextGC",
		"NumForcedGC",
		"NumGC",
		"OtherSys",
		"PauseTotalNs",
		"StackInuse",
		"StackSys",
		"Sys",
		"TotalAlloc",
		"RandomValue",
	}

	for _, metricName := range expectedMetrics {
		value, exists := metrics[metricName]
		assert.True(t, exists, "metric %s should be present", metricName)

		if metricName != "GCCPUFraction" && metricName != "LastGC" {
			assert.GreaterOrEqual(t, value, float64(0), "metric %s should be not negative", metricName)
		}
	}

	assert.NotZero(t, metrics["RandomValue"], "RandomValue should be not zero")
}

func TestGetGopsuiteMetrics(t *testing.T) {
	metrics := make(map[string]float64)
	result := GetGopsuiteMetrics(metrics)

	assert.Equal(t, &metrics, &result, "function should return the same map")
	expectedMetrics := []string{
		"TotalMemory",
		"FreeMemory",
		"CPUutilization1",
	}

	for _, metricName := range expectedMetrics {
		value, exists := metrics[metricName]
		assert.True(t, exists, "metric %s should be present", metricName)
		assert.GreaterOrEqual(t, value, float64(0), "metric %s should be not negative", metricName)
	}

	assert.GreaterOrEqual(t, metrics["TotalMemory"], metrics["FreeMemory"], "")
}

func TestMetricsIntegration(t *testing.T) {
	metrics := make(map[string]float64)

	InitPoolMetrics(metrics)
	result := GetGopsuiteMetrics(metrics)
	assert.Equal(t, &metrics, &result, "function should return the same map")
	assert.Contains(t, metrics, "Alloc", "metric Alloc should be present")
	assert.Contains(t, metrics, "TotalMemory", "metreic TotalMemory should be present")
	assert.GreaterOrEqual(t, len(metrics), 28+3, "general metrics count")
}

func TestGetGopsuiteMetricsWithPrefilled(t *testing.T) {
	metrics := map[string]float64{
		"ExistingMetric": 123.45,
	}
	result := GetGopsuiteMetrics(metrics)
	assert.Equal(t, &metrics, &result, "function should return the same map")
	assert.Equal(t, 123.45, metrics["ExistingMetric"], "existing metric should not be changed")

	assert.Contains(t, metrics, "TotalMemory")
	assert.Contains(t, metrics, "FreeMemory")
	assert.Contains(t, metrics, "CPUutilization1")
}
