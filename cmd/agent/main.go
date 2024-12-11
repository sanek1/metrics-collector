package main

import (
	"fmt"
	"io"
	"log"
	"math/rand/v2"
	"net/http"
	"os"
	"runtime"
	"time"
)

type gauge float64
type counter int64

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

	PollCount := counter(0)
	metrics := make(map[string]gauge, 30)
	for {
		select {
		case <-pollTick.C:
			PollCount++
			pollMetrics(metrics)
		case <-reportTick.C:
			reportMetrics(metrics, PollCount, client, logger)
		}
	}
}

func reportMetrics(metrics map[string]gauge, PollCount counter, client *http.Client, logger *log.Logger) {
	addr := Options.flagRunAddr

	metricURL := fmt.Sprint("http://", addr, "/update/counter/PollCount/", PollCount)
	if err := reportClient(client, metricURL, logger); err != nil {
		logger.Printf("Error reporting metrics: %v", err)
	}

	for k, v := range metrics {
		metricURL = fmt.Sprint("http://", addr, "/update/gauge/", k, "/", v)
		if err := reportClient(client, metricURL, logger); err != nil {
			logger.Printf("Error reporting metrics: %v", err)
		}
	}

	fmt.Fprintf(os.Stdout, " Iteration ---------> %d\n\n", PollCount)
}

func reportClient(client *http.Client, url string, logger *log.Logger) error {
	req, err := http.NewRequest(http.MethodPost, url, nil)
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
	metrics["RandomValue"] = gauge(rand.Float64())
}
