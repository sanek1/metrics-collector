// package storage
package storage

import (
	"crypto/rand"
	"encoding/binary"
	"log"
	"runtime"

	"github.com/shirou/gopsutil/v3/mem"
)

//go:nocover
func InitPoolMetrics(metrics map[string]float64) {
	var num uint32
	err := binary.Read(rand.Reader, binary.LittleEndian, &num)
	if err != nil {
		log.Printf("Error reading random number: %v", err)
	}
	var rtm runtime.MemStats
	runtime.ReadMemStats(&rtm)
	metrics["Alloc"] = float64(rtm.Alloc)
	metrics["BuckHashSys"] = float64(rtm.BuckHashSys)
	metrics["Frees"] = float64(rtm.Frees)
	metrics["GCCPUFraction"] = float64(rtm.GCCPUFraction)
	metrics["GCSys"] = float64(rtm.GCSys)
	metrics["HeapAlloc"] = float64(rtm.HeapAlloc)
	metrics["HeapIdle"] = float64(rtm.HeapIdle)
	metrics["HeapInuse"] = float64(rtm.HeapInuse)
	metrics["HeapObjects"] = float64(rtm.HeapObjects)
	metrics["HeapReleased"] = float64(rtm.HeapReleased)
	metrics["HeapSys"] = float64(rtm.HeapSys)
	metrics["LastGC"] = float64(rtm.LastGC)
	metrics["Lookups"] = float64(rtm.Lookups)
	metrics["MCacheInuse"] = float64(rtm.MCacheInuse)
	metrics["MCacheSys"] = float64(rtm.MCacheSys)
	metrics["MSpanInuse"] = float64(rtm.MSpanInuse)
	metrics["MSpanSys"] = float64(rtm.MSpanSys)
	metrics["Mallocs"] = float64(rtm.Mallocs)
	metrics["NextGC"] = float64(rtm.NextGC)
	metrics["NumForcedGC"] = float64(rtm.NumForcedGC)
	metrics["NumGC"] = float64(rtm.NumGC)
	metrics["OtherSys"] = float64(rtm.OtherSys)
	metrics["PauseTotalNs"] = float64(rtm.PauseTotalNs)
	metrics["StackInuse"] = float64(rtm.StackInuse)
	metrics["StackSys"] = float64(rtm.StackSys)
	metrics["Sys"] = float64(rtm.Sys)
	metrics["TotalAlloc"] = float64(rtm.TotalAlloc)
	metrics["RandomValue"] = float64(num)
}

//go:nocover
func GetGopsuiteMetrics(gpMetrics map[string]float64) map[string]float64 {
	v, err := mem.VirtualMemory()
	if err != nil {
		log.Printf("Error getting virtual memory: %v", err)
		return gpMetrics
	}
	gpMetrics["TotalMemory"] = float64(v.Total)
	gpMetrics["FreeMemory"] = float64(v.Available)
	gpMetrics["CPUutilization1"] = float64(v.Used)
	return gpMetrics
}
