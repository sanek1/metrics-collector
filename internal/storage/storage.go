package storage

type MemStorage struct {
	GaugeBuff   map[string]float64
	CounterBuff []int64
}
type IMetricStorage interface {
	//GetGauge(key string) float64
	//SetGauge(key string, value float64)
	//SetCounter(key string, value int64)
	//GetCounter(key int64) int64
}

func NewMemStorage() *MemStorage {
	return &MemStorage{
		GaugeBuff:   make(map[string]float64),
		CounterBuff: make([]int64, 0),
	}
}

func (m *MemStorage) GetGauge(key string) float64 {
	return m.GaugeBuff[key]
}
func (m *MemStorage) GetCounter(key int64) int64 {
	return m.CounterBuff[key]
}

func (m *MemStorage) SetCounter(key int64) {
	m.CounterBuff = append(m.CounterBuff, key)
}

func (m *MemStorage) SetGauge(key string, value float64) {
	m.GaugeBuff[key] = value
}
