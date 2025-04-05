// Package storage представляет собой библиотеку хранилища метрик
package storage

import (
	"context"
	"time"

	m "github.com/sanek1/metrics-collector/internal/models"
)

// Storage определяет основной интерфейс для хранения и доступа к метрикам.
// Предоставляет методы для установки и получения метрик разных типов.
type Storage interface {
	// SetGauge устанавливает одну или несколько метрик типа gauge.
	// Принимает контекст выполнения и вариативный список метрик.
	// Возвращает slice обновленных метрик и ошибку, если она возникла.
	SetGauge(ctx context.Context, models ...m.Metrics) ([]*m.Metrics, error)

	// SetCounter устанавливает одну или несколько метрик типа counter.
	// Для counter-метрик новое значение добавляется к текущему значению счетчика.
	// Принимает контекст выполнения и вариативный список метрик.
	// Возвращает slice обновленных метрик и ошибку, если она возникла.
	SetCounter(ctx context.Context, models ...m.Metrics) ([]*m.Metrics, error)

	// GetAllMetrics возвращает список всех доступных метрик в хранилище.
	// Возвращает slice строк с именами метрик.
	GetAllMetrics() []string

	// GetMetrics получает метрику по её типу и имени.
	// Принимает контекст выполнения, тип метрики и имя метрики.
	// Возвращает указатель на метрику и boolean-флаг, указывающий существует ли метрика.
	GetMetrics(ctx context.Context, metricType, metricName string) (*m.Metrics, bool)
}

// DatabaseStorage определяет интерфейс для хранилища метрик, использующего базу данных.
// Предоставляет методы для проверки соединения с БД и управления схемой данных.
type DatabaseStorage interface {
	// PingIsOk проверяет доступность базы данных.
	// Возвращает true, если соединение с БД установлено и работает корректно.
	PingIsOk() bool

	// EnsureMetricsTableExists проверяет существование необходимой таблицы метрик
	// и создает её, если она отсутствует.
	// Принимает контекст выполнения.
	// Возвращает ошибку, если не удалось создать таблицу.
	EnsureMetricsTableExists(ctx context.Context) error
}

// FileStorage определяет интерфейс для хранилища метрик с поддержкой
// сохранения и загрузки из файла.
type FileStorage interface {
	// SaveToFile сохраняет все метрики в файл с указанным именем.
	// Принимает имя файла.
	// Возвращает ошибку, если не удалось сохранить метрики в файл.
	SaveToFile(fname string) error

	// LoadFromFile загружает метрики из файла с указанным именем.
	// Принимает имя файла.
	// Возвращает ошибку, если не удалось загрузить метрики из файла.
	LoadFromFile(fname string) error

	// PeriodicallySaveBackUp запускает периодическое сохранение метрик в файл.
	// Параметры:
	//   - ctx: контекст для возможности отмены операции
	//   - filename: имя файла для сохранения
	//   - restore: флаг, указывающий, нужно ли восстанавливать данные при запуске
	//   - interval: интервал между сохранениями
	PeriodicallySaveBackUp(ctx context.Context, filename string, restore bool, interval time.Duration)
}

type StorageHelper interface {
	FilterBatchesBeforeSaving(metrics []m.Metrics) []m.Metrics
	SortingBatchData(existingMetrics []*m.Metrics, metrics []m.Metrics) (updatingBatch, insertingBatch []m.Metrics)
	CollectorQuery(ctx context.Context, metrics []m.Metrics) (query string, mTypes []string, args []interface{})
	UpdateMetrics(ctx context.Context, models []m.Metrics) error
	InsertMetric(ctx context.Context, models []m.Metrics) error
	GetMetricsOnDBs(ctx context.Context, models ...m.Metrics) ([]*m.Metrics, error)
	SetMetrics(ctx context.Context, models []m.Metrics) ([]*m.Metrics, error)
	EnsureMetricsTableExists(ctx context.Context) error
}
