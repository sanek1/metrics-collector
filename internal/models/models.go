package models

const (
	TypeGauge   = "gauge"
	TypeCounter = "counter"
)

type Metrics struct {
	ID    string   `json:"id" db:"id"`                 // Name of the metric
	MType string   `json:"type" db:"type"`             // Type of the metric
	Delta *int64   `json:"delta,omitempty" db:"delta"` // Count of the metric
	Value *float64 `json:"value,omitempty" db:"value"` // Gauge value
}

func NewMetricCounter(id string, delta *int64) *Metrics {
	return &Metrics{
		ID:    id,
		MType: TypeCounter,
		Delta: delta,
	}
}

func NewMetricGauge(id string, value *float64) *Metrics {
	return &Metrics{
		ID:    id,
		MType: TypeGauge,
		Value: value,
	}
}

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
