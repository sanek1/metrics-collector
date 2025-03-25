package models

import (
	"context"
)

// MetricsStorage определяет интерфейс для хранения и управления метриками.
// Этот интерфейс предоставляет методы для установки, получения и управления метриками,
// а также для проверки состояния подключения к хранилищу.
type MetricsStorage interface {
	// SetGauge устанавливает значения для метрик типа gauge.
	// Принимает контекст и одну или несколько метрик.
	// Возвращает обновленные метрики и ошибку, если она возникла.
	SetGauge(ctx context.Context, metrics ...Metrics) ([]*Metrics, error)

	// SetCounter устанавливает значения для метрик типа counter.
	// Принимает контекст и одну или несколько метрик.
	// Возвращает обновленные метрики и ошибку, если она возникла.
	SetCounter(ctx context.Context, metrics ...Metrics) ([]*Metrics, error)

	// GetMetrics возвращает метрику по ее типу и имени.
	// Возвращает указатель на метрику и флаг существования метрики.
	GetMetrics(ctx context.Context, metricType, metricName string) (*Metrics, bool)

	// GetAllMetrics возвращает список всех доступных метрик.
	GetAllMetrics() []string

	// Ping проверяет доступность хранилища метрик.
	// Возвращает ошибку, если хранилище недоступно.
	Ping(ctx context.Context) error
}
