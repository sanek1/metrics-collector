package validation

const (
	TypeGauge   = "gauge"
	TypeCounter = "counter"
)

type Metrics struct {
	ID    string   `json:"id"`              // Name of the metric
	MType string   `json:"type"`            // Type of the metric
	Delta *int64   `json:"delta,omitempty"` // Count of the metric
	Value *float64 `json:"value,omitempty"` // Gauge value
}
