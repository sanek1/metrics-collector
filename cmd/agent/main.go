package main

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"runtime"
	"time"

	"github.com/sanek1/metrics-collector/internal/validation"
)

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

	var pollCount int64 = 0
	metrics := make(map[string]float64, countMetrics)
	for {
		select {
		case <-pollTick.C:
			pollCount++
			metrics = make(map[string]float64, countMetrics)
			pollMetrics(metrics)

		case <-reportTick.C:
			reportMetrics(metrics, &pollCount, client, logger)
		}
	}
}

func reportMetrics(metrics map[string]float64, pollCount *int64, client *http.Client, logger *log.Logger) {
	fmt.Fprintf(os.Stdout, "--------- start response ---------\n\n")
	addr := Options.flagRunAddr
	metricURL := fmt.Sprint("http://", addr, "/update/counter/PollCount/", *pollCount)
	metrics2 := validation.Metrics{
		ID:    "PollCount",
		MType: "counter",
		Delta: pollCount,
	}
	if err := reportClient(client, metricURL, metrics2, logger); err != nil {
		logger.Printf("Error reporting metrics: %v", err)
	}

	for name, v := range metrics {
		metricURL = fmt.Sprint("http://", addr, "/update/gauge/", name, "/", v)

		metrics3 := validation.Metrics{
			ID:    name,
			MType: "gauge",
			Value: &v,
		}

		if err := reportClient(client, metricURL, metrics3, logger); err != nil {
			logger.Printf("Error reporting metrics: %v", err)
		}
	}
	fmt.Fprintf(os.Stdout, "--------- end response ---------\n\n")
	fmt.Fprintf(os.Stdout, "--------- NEW ITERATION %d ---------> \n\n", *pollCount)
}

func reportClient(client *http.Client, url string, m validation.Metrics, logger *log.Logger) error {
	ctx := context.Background()
	reqBody, err := buildMetrickBody(m)
	if err != nil {
		logger.Printf("Error building request body: %v", err)
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(reqBody))
	if err != nil {
		logger.Printf("Error creating request: %v", err)
		return err
	}

	req.Header.Set("Content-Type", "application/json")
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

	if err := json.Unmarshal(body, &m); err != nil {
		logger.Printf("Error unmarshaling response body: %v", err)
		return err
	}

	defer resp.Body.Close()

	if _, err := io.Copy(io.Discard, resp.Body); err != nil {
		logger.Printf("Error discarding response body: %v", err)
	}

	return nil
}

func pollMetrics(metrics map[string]float64) {
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

func buildMetrickBody(m validation.Metrics) ([]byte, error) {
	body, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}
	return body, nil
}
