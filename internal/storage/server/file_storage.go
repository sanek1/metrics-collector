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

// func (ms *MetricsStorage) PeriodicallySaveBackUp(ctx context.Context, filename string, restore bool, interval time.Duration) {
// 	ticker := time.NewTicker(interval)
// 	go func() {
// 		for range ticker.C {
// 			if err := ms.DumpToFile(filename); err != nil {
// 				log.Printf("Ошибка сохранения метрик: %v", err)
// 			}
// 		}
// 	}()
// }

// func (ms *MetricsStorage) DumpToFile(filename string) error {
// 	// Создаем временный файл для атомарной записи
// 	tmpFile := filename + ".tmp"
// 	f, err := os.Create(tmpFile)
// 	if err != nil {
// 		return err
// 	}

// 	// Сериализуем все метрики за одну блокировку
// 	metrics := ms.GetAllMetrics()

// 	encoder := json.NewEncoder(f)
// 	if err := encoder.Encode(metrics); err != nil {
// 		f.Close()
// 		os.Remove(tmpFile)
// 		return err
// 	}

// 	if err := f.Sync(); err != nil {
// 		f.Close()
// 		return err
// 	}

// 	if err := f.Close(); err != nil {
// 		return err
// 	}

// 	// Атомарная замена файла
// 	return os.Rename(tmpFile, filename)
// }

func (ms *MetricsStorage) PeriodicallySaveBackUp(ctx context.Context, filename string, restore bool, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	if restore {
		err := ms.LoadFromFile(filename)
		if err != nil {
			ms.Logger.ErrorCtx(ctx, "Error loading metrics from file")
		}
	}

	for range ticker.C {
		err := ms.SaveToFile(filename)
		ms.Logger.InfoCtx(ctx, "saving to file was successful")
		if err != nil {
			ms.Logger.ErrorCtx(ctx, "Error saving metrics to file"+err.Error())
		}
	}

	for {
		select {
		case <-ticker.C:
			err := ms.SaveToFile(filename)
			ms.Logger.InfoCtx(ctx, "saving to file was successful")
			if err != nil {
				ms.Logger.ErrorCtx(ctx, "Error saving metrics to file"+err.Error())
			}

		case <-ctx.Done():
			ms.Logger.InfoCtx(ctx, "Backup process stopped.")
			return
		}
	}
}
