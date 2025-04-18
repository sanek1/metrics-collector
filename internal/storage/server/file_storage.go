// Package storage представляет собой библиотеку хранилища метрик
package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"
)

func (ms *MetricsStorage) SaveToFile(fname string) error {
	// serialize to json
	data, err := json.MarshalIndent(ms.Metrics, "", "   ")
	if err != nil {
		return err
	}
	// save to file
	err = os.WriteFile(fname, data, fileMode)
	if err != nil {
		return err
	}
	fmt.Printf("Data saved to file: %s\n", fname)
	return nil
}

func (ms *MetricsStorage) LoadFromFile(filename string) error {
	content, err := os.ReadFile(filename)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Println("Data file not found. Let's start with empty values.")
			return nil
		}
		return fmt.Errorf("file read error: %v", err)
	}

	err = json.Unmarshal(content, &ms.Metrics)
	if err != nil {
		return fmt.Errorf("data unmarshalling error: %v", err)
	}

	fmt.Println("Previous metric values have been loaded.")
	return nil
}

func (ms *MetricsStorage) PeriodicallySaveBackUp(ctx context.Context, filename string, restore bool, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	if restore {
		err := ms.LoadFromFile(filename)
		if err != nil {
			ms.Logger.ErrorCtx(ctx, "Error loading metrics from file")
		}
	}
	err := ms.SaveToFile(filename)
	if err != nil {
		ms.Logger.ErrorCtx(ctx, "Error saving metrics to file: "+err.Error())
	} else {
		ms.Logger.InfoCtx(ctx, "saving to file was successful")
	}

	for {
		select {
		case <-ticker.C:
			err := ms.SaveToFile(filename)
			if err != nil {
				ms.Logger.ErrorCtx(ctx, "Error saving metrics to file: "+err.Error())
			} else {
				ms.Logger.InfoCtx(ctx, "saving to file was successful")
			}
		case <-ctx.Done():
			ms.Logger.InfoCtx(ctx, "Backup process stopped.")
			return
		}
	}
}
