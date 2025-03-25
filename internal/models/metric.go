package models

// Типы метрик
const (
	// TypeGauge представляет тип метрики со значением с плавающей точкой
	TypeGauge = "gauge"
	// TypeCounter представляет тип метрики с целочисленным счетчиком
	TypeCounter = "counter"
)

// Metrics представляет собой структуру метрики, используемую для сбора и хранения
// различных типов метрик в системе мониторинга.
// Поддерживает два типа метрик: gauge (значение с плавающей точкой) и counter (целочисленный счетчик).
type Metrics struct {
	ID    string   `json:"id" db:"id"`                 // Name of the metric
	MType string   `json:"type" db:"type"`             // Type of the metric
	Delta *int64   `json:"delta,omitempty" db:"delta"` // Count of the metric
	Value *float64 `json:"value,omitempty" db:"value"` // Gauge value
}

// NewMetricCounter создает новую метрику типа counter с заданным ID и значением.
// Параметры:
//   - id: уникальный идентификатор метрики
//   - delta: указатель на значение счетчика
//
// Возвращает:
//   - указатель на созданную метрику
func NewMetricCounter(id string, delta *int64) *Metrics {
	return &Metrics{
		ID:    id,
		MType: TypeCounter,
		Delta: delta,
	}
}

// NewMetricGauge создает новую метрику типа gauge с заданным ID и значением.
// Параметры:
//   - id: уникальный идентификатор метрики
//   - value: указатель на значение с плавающей точкой
//
// Возвращает:
//   - указатель на созданную метрику
func NewMetricGauge(id string, value *float64) *Metrics {
	return &Metrics{
		ID:    id,
		MType: TypeGauge,
		Value: value,
	}
}

// NewArrMetricGauge создает массив метрик типа gauge из карты значений.
// Параметры:
//   - metrics: карта, где ключ - идентификатор метрики, значение - значение метрики
//
// Возвращает:
//   - срез метрик типа gauge
func NewArrMetricGauge(metrics map[string]float64) []Metrics {
	result := make([]Metrics, 0, len(metrics))
	for id, value := range metrics {
		result = append(result, Metrics{
			ID:    id,
			MType: TypeGauge,
			Value: &value,
		})
	}
	return result
}
