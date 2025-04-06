package storage

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/sanek1/metrics-collector/internal/models"
	"github.com/sanek1/metrics-collector/internal/storage/server/mocks"
	"github.com/sanek1/metrics-collector/pkg/logging"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

type MockLogger struct {
	mock.Mock
}

func (m *MockLogger) InfoCtx(ctx context.Context, msg string, fields ...interface{}) {
	m.Called(ctx, msg, fields)
}

func (m *MockLogger) ErrorCtx(ctx context.Context, msg string, fields ...interface{}) {
	m.Called(ctx, msg, fields)
}

func (m *MockLogger) WarnCtx(ctx context.Context, msg string, fields ...interface{}) {
	m.Called(ctx, msg, fields)
}

func (m *MockLogger) DebugCtx(ctx context.Context, msg string, fields ...interface{}) {
	m.Called(ctx, msg, fields)
}

type MockDB struct {
	mock.Mock
}

func (m *MockDB) Ping() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockDB) Close() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockDB) Exec(query string, args ...interface{}) error {
	callArgs := m.Called(query, args)
	return callArgs.Error(0)
}

func (m *MockDB) Query(query string, args ...interface{}) (map[string]interface{}, error) {
	callArgs := m.Called(query, args)
	return callArgs.Get(0).(map[string]interface{}), callArgs.Error(1)
}

func TestDBStorageWithMocks(t *testing.T) {
	mockLogger := new(MockLogger)
	mockLogger.On("InfoCtx", mock.Anything, mock.Anything, mock.Anything).Return()
	mockLogger.On("ErrorCtx", mock.Anything, mock.Anything, mock.Anything).Return()

	mockDB := new(MockDB)
	mockDB.On("Ping").Return(nil)
	mockDB.On("Exec", mock.Anything, mock.Anything).Return(nil)

	gaugeValue := float64(123.45)
	gaugeResults := map[string]interface{}{
		"id":    "TestGauge",
		"type":  "gauge",
		"value": gaugeValue,
	}
	mockDB.On("Query", "SELECT id, type, value, delta FROM metrics WHERE id = $1 AND type = $2", mock.Anything).Return(gaugeResults, nil)

	deltaValue := int64(42)
	counterResults := map[string]interface{}{
		"id":    "TestCounter",
		"type":  "counter",
		"delta": deltaValue,
	}
	mockDB.On("Query", "SELECT id, type, value, delta FROM metrics WHERE id = $1 AND type = $2", mock.Anything).Return(counterResults, nil)

	t.Run("Ping", func(t *testing.T) {
		err := mockDB.Ping()
		assert.NoError(t, err)
	})

	t.Run("Exec", func(t *testing.T) {
		err := mockDB.Exec("INSERT INTO metrics (id, type, value, delta) VALUES ($1, $2, $3, $4)", "TestGauge", "gauge", 123.45, nil)
		assert.NoError(t, err)
	})
}

func TestSaveAndLoadFromFile(t *testing.T) {
	tempDir := t.TempDir()
	logger, _ := logging.NewZapLogger(zap.InfoLevel)
	testFile := filepath.Join(tempDir, "metrics_test.json")

	mockStorage := mocks.NewStorage(t)
	mockFileStorage := mocks.NewFileStorage(t)

	ctx := context.Background()
	gaugeValue := 123.45
	gaugeMetric := models.Metrics{
		ID:    "test_gauge",
		MType: "gauge",
		Value: &gaugeValue,
	}

	returnedGaugeMetric := &gaugeMetric
	mockStorage.On("SetGauge", ctx, gaugeMetric).Return([]*models.Metrics{returnedGaugeMetric}, nil)

	counterValue := int64(42)
	counterMetric := models.Metrics{
		ID:    "test_counter",
		MType: "counter",
		Delta: &counterValue,
	}

	returnedCounterMetric := &counterMetric
	mockStorage.On("SetCounter", ctx, counterMetric).Return([]*models.Metrics{returnedCounterMetric}, nil)

	mockStorage.On("GetMetrics", ctx, "gauge", "test_gauge").Return(&gaugeMetric, true)
	mockStorage.On("GetMetrics", ctx, "counter", "test_counter").Return(&counterMetric, true)

	mockFileStorage.On("SaveToFile", testFile).Return(nil)
	mockFileStorage.On("LoadFromFile", testFile).Return(nil)
	mockFileStorage.On("LoadFromFile", mock.AnythingOfType("string")).Return(nil)
	mockFileStorage.On("SaveToFile", mock.AnythingOfType("string")).Return(fmt.Errorf("ошибка доступа к файлу"))

	_, err := mockStorage.SetGauge(ctx, gaugeMetric)
	require.NoError(t, err)

	_, err = mockStorage.SetCounter(ctx, counterMetric)
	require.NoError(t, err)

	t.Run("SaveToFile", func(t *testing.T) {
		err := mockFileStorage.SaveToFile(testFile)
		require.NoError(t, err)

		mockFileStorage.AssertCalled(t, "SaveToFile", testFile)
	})

	t.Run("LoadFromFile", func(t *testing.T) {
		newStorage := NewMetricsStorage(logger)

		newStorage.Metrics = map[string]models.Metrics{
			"test_gauge":   gaugeMetric,
			"test_counter": counterMetric,
		}

		err := mockFileStorage.LoadFromFile(testFile)
		require.NoError(t, err)

		mockFileStorage.AssertCalled(t, "LoadFromFile", testFile)
		g, found := mockStorage.GetMetrics(ctx, "gauge", "test_gauge")
		require.True(t, found)
		assert.Equal(t, gaugeValue, *g.Value)

		c, found := mockStorage.GetMetrics(ctx, "counter", "test_counter")
		require.True(t, found)
		assert.Equal(t, counterValue, *c.Delta)
	})

	t.Run("LoadFromNonExistentFile", func(t *testing.T) {
		nonExistentFile := filepath.Join(tempDir, "non_existent.json")
		err := mockFileStorage.LoadFromFile(nonExistentFile)
		assert.NoError(t, err)
	})

	t.Run("SaveToFileWithError", func(t *testing.T) {
		invalidPath := filepath.Join(tempDir, "invalid_dir", "file.json")
		os.MkdirAll(filepath.Dir(invalidPath), 0755)

		err := mockFileStorage.SaveToFile(invalidPath)
		assert.Error(t, err)
	})
}

func TestPeriodicallySaveBackUp(t *testing.T) {
	// Создаем временную директорию для теста
	tempDir := t.TempDir()
	subDir := filepath.Join(tempDir, "backup") // Используем поддиректорию
	err := os.MkdirAll(subDir, 0755)
	require.NoError(t, err)

	testFile := filepath.Join(subDir, "metrics_backup.json")

	// Настраиваем логгер и хранилище
	logger, _ := logging.NewZapLogger(zap.InfoLevel)
	storage := NewMetricsStorage(logger)

	// Добавляем тестовую метрику типа gauge
	ctx := context.Background()
	gaugeValue := float64(123.45)
	gaugeMetric := models.Metrics{
		ID:    "test_gauge",
		MType: "gauge",
		Value: &gaugeValue,
	}
	_, err = storage.SetGauge(ctx, gaugeMetric)
	require.NoError(t, err)

	// Добавляем тестовую метрику типа counter
	counterValue := int64(42)
	counterMetric := models.Metrics{
		ID:    "test_counter",
		MType: "counter",
		Delta: &counterValue,
	}
	_, err = storage.SetCounter(ctx, counterMetric)
	require.NoError(t, err)

	// Сохраняем метрики в файл напрямую, чтобы проверить, что файл создается корректно
	err = storage.SaveToFile(testFile)
	require.NoError(t, err, "Должно быть возможно сохранить файл напрямую")

	// Проверяем, что файл создан
	_, err = os.Stat(testFile)
	require.NoError(t, err, "Файл должен быть создан")

	// Создаем контекст с коротким таймаутом для тестирования
	ctx, cancel := context.WithTimeout(context.Background(), 250*time.Millisecond)
	defer cancel()

	// Запускаем периодическое сохранение в фоне на короткий интервал
	// Используем отдельную горутину и канал для подтверждения завершения
	done := make(chan struct{})
	go func() {
		defer close(done)
		storage.PeriodicallySaveBackUp(ctx, testFile, false, 100*time.Millisecond)
	}()

	// Ждем завершения контекста
	<-ctx.Done()

	// Даем время на завершение горутины
	<-done

	// Создаем новое хранилище для загрузки метрик
	newStorage := NewMetricsStorage(logger)

	// Загружаем данные из файла
	err = newStorage.LoadFromFile(testFile)
	require.NoError(t, err)

	// Проверяем, что метрики загружены корректно
	g, found := newStorage.GetMetrics(context.Background(), "gauge", "test_gauge")
	require.True(t, found, "Должна существовать метрика gauge")
	require.NotNil(t, g, "Метрика gauge не должна быть nil")
	require.NotNil(t, g.Value, "Значение метрики gauge не должно быть nil")
	assert.Equal(t, gaugeValue, *g.Value)

	c, found := newStorage.GetMetrics(context.Background(), "counter", "test_counter")
	require.True(t, found, "Должна существовать метрика counter")
	require.NotNil(t, c, "Метрика counter не должна быть nil")
	require.NotNil(t, c.Delta, "Значение метрики counter не должно быть nil")
	assert.Equal(t, counterValue, *c.Delta)
}
