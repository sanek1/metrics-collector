package main

import (
	"context"
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"runtime"
	"time"
)

type gauge float64
type counter int64

const (
	countMetrics = 30
)

func main() {
	ParseFlags()
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	logger := log.New(os.Stdout, "Agent\t", log.Ldate|log.Ltime)
	logger.Println("started")

	pollTick := time.NewTicker(time.Duration(Options.pollInterval) * time.Second)
	reportTick := time.NewTicker(time.Duration(Options.reportInterval) * time.Second)
	defer func() {
		pollTick.Stop()
		reportTick.Stop()
	}()
	client := &http.Client{}

	pollCount := counter(0)
	metrics := make(map[string]gauge, countMetrics)
	for {
		select {
		case <-pollTick.C:
			pollCount++
			pollMetrics(metrics)
		case <-reportTick.C:
			reportMetrics(metrics, pollCount, client, logger)
		}
	}
}

func reportMetrics(metrics map[string]gauge, pollCount counter, client *http.Client, logger *log.Logger) {
	addr := Options.flagRunAddr

	metricURL := fmt.Sprint("http://", addr, "/update/counter/PollCount/", pollCount)
	if err := reportClient(client, metricURL, logger); err != nil {
		logger.Printf("Error reporting metrics: %v", err)
	}

	for k, v := range metrics {
		metricURL = fmt.Sprint("http://", addr, "/update/gauge/", k, "/", v)
		if err := reportClient(client, metricURL, logger); err != nil {
			logger.Printf("Error reporting metrics: %v", err)
		}
	}

	fmt.Fprintf(os.Stdout, " Iteration ---------> %d\n\n", pollCount)
}

func reportClient(client *http.Client, url string, logger *log.Logger) error {
	ctx := context.Background()
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, http.NoBody)
	if err != nil {
		logger.Printf("Error creating request: %v", err)
		return err
	}
	req.Header.Set("Content-Type", "text/plain")
	cookie := &http.Cookie{
		Name:   "Token",
		Value:  "TEST_TOKEN",
		MaxAge: 360,
	}
	req.AddCookie(cookie)

	resp, err := client.Do(req)
	if err != nil {
		logger.Printf("Error sending request: %v", err)
		return err
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Printf("Error reading response body: %v", err)
		return err
	}
	os.Stdout.Write(body)
	fmt.Fprintf(os.Stdout, " Url: %s status: %d\n", url, resp.StatusCode)

	defer resp.Body.Close()

	if _, err := io.Copy(io.Discard, resp.Body); err != nil {
		logger.Printf("Error discarding response body: %v", err)
	}

	return nil
}

func pollMetrics(metrics map[string]gauge) {
	var num uint32
	err := binary.Read(rand.Reader, binary.LittleEndian, &num)
	if err != nil {
		log.Printf("Error reading random number: %v", err)
	}
	var rtm runtime.MemStats
	runtime.ReadMemStats(&rtm)
	metrics["Alloc"] = gauge(rtm.Alloc)
	metrics["BuckHashSys"] = gauge(rtm.BuckHashSys)
	metrics["Frees"] = gauge(rtm.Frees)
	metrics["GCCPUFraction"] = gauge(rtm.GCCPUFraction)
	metrics["GCSys"] = gauge(rtm.GCSys)
	metrics["HeapAlloc"] = gauge(rtm.HeapAlloc)
	metrics["HeapIdle"] = gauge(rtm.HeapIdle)
	metrics["HeapInuse"] = gauge(rtm.HeapInuse)
	metrics["HeapObjects"] = gauge(rtm.HeapObjects)
	metrics["HeapReleased"] = gauge(rtm.HeapReleased)
	metrics["HeapSys"] = gauge(rtm.HeapSys)
	metrics["LastGC"] = gauge(rtm.LastGC)
	metrics["Lookups"] = gauge(rtm.Lookups)
	metrics["MCacheInuse"] = gauge(rtm.MCacheInuse)
	metrics["MCacheSys"] = gauge(rtm.MCacheSys)
	metrics["MSpanInuse"] = gauge(rtm.MSpanInuse)
	metrics["MSpanSys"] = gauge(rtm.MSpanSys)
	metrics["Mallocs"] = gauge(rtm.Mallocs)
	metrics["NextGC"] = gauge(rtm.NextGC)
	metrics["NumForcedGC"] = gauge(rtm.NumForcedGC)
	metrics["NumGC"] = gauge(rtm.NumGC)
	metrics["OtherSys"] = gauge(rtm.OtherSys)
	metrics["PauseTotalNs"] = gauge(rtm.PauseTotalNs)
	metrics["StackInuse"] = gauge(rtm.StackInuse)
	metrics["StackSys"] = gauge(rtm.StackSys)
	metrics["Sys"] = gauge(rtm.Sys)
	metrics["TotalAlloc"] = gauge(rtm.TotalAlloc)
	metrics["RandomValue"] = gauge(num)
}
