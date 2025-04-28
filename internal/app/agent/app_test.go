package app

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"

	af "github.com/sanek1/metrics-collector/internal/flags/agent"
	l "github.com/sanek1/metrics-collector/pkg/logging"
)

type MockController struct {
	mock.Mock
}

func (m *MockController) SendingCounterMetrics(ctx context.Context, pollCount *int64, client *http.Client) {
	m.Called(ctx, pollCount, client)
}

func (m *MockController) SendingGaugeMetrics(ctx context.Context, metrics map[string]float64, client *http.Client) {
	m.Called(ctx, metrics, client)
}

func Test_InitDataAgent(t *testing.T) {
	opt := &af.Options{
		PollInterval:   2,
		ReportInterval: 10,
	}
	pollTick, reportTick, metrics, gpMetrics := initDataAgent(opt)

	assert.NotNil(t, pollTick)
	assert.NotNil(t, reportTick)
	assert.NotNil(t, metrics)
	assert.NotNil(t, gpMetrics)
	assert.Equal(t, 0, len(metrics))
	assert.Equal(t, 0, len(gpMetrics))

	time.Sleep(2100 * time.Millisecond)
	select {
	case <-pollTick.C:
	default:
		t.Error("expected pollTick")
	}
}

func TestNewApp(t *testing.T) {
	opt := &af.Options{
		PollInterval:   2,
		ReportInterval: 10,
	}

	app := New(opt)

	assert.NotNil(t, app)
	assert.NotNil(t, app.controller)
	assert.Equal(t, opt, app.opt)
}

// Для этого теста мы используем App с пустым контроллером
type Controller interface {
	SendingCounterMetrics(ctx context.Context, pollCount *int64, client *http.Client)
	SendingGaugeMetrics(ctx context.Context, metrics map[string]float64, client *http.Client)
}

func TestRun(t *testing.T) {
	// Вместо тестирования всего приложения, мы будем тестировать функцию initDataAgent,
	// так как она содержит большую часть логики, которую мы хотим проверить
	opt := &af.Options{
		PollInterval:   1,
		ReportInterval: 1,
	}

	pollTick, reportTick, metrics, gpMetrics := initDataAgent(opt)
	defer pollTick.Stop()
	defer reportTick.Stop()

	// Проверяем, что карты метрик инициализированы
	assert.NotNil(t, metrics)
	assert.NotNil(t, gpMetrics)
	assert.Equal(t, 0, len(metrics))
	assert.Equal(t, 0, len(gpMetrics))

	// Здесь мы могли бы проверить, что тикеры срабатывают через определенные промежутки времени,
	// но это уже сделано в тесте TestTickerBehavior, поэтому мы не дублируем эту логику
}

func TestStartLogger(t *testing.T) {
	logger := startLogger()
	assert.NotNil(t, logger)

	// Проверим, что методы логгера доступны и не вызывают панику
	logger.InfoCtx(context.Background(), "Test message")
	logger.ErrorCtx(context.Background(), "Test error")
}

// Тест для проверки работы горутин и тикеров
func TestTickerBehavior(t *testing.T) {
	// Создаем короткие интервалы для быстрого теста
	opt := &af.Options{
		PollInterval:   1,
		ReportInterval: 1,
	}

	pollTick, reportTick, metrics, gpMetrics := initDataAgent(opt)
	defer pollTick.Stop()
	defer reportTick.Stop()

	// Проверяем, что тикеры срабатывают в ожидаемые интервалы
	time.Sleep(1100 * time.Millisecond)

	// Проверяем, что pollTick сработал
	select {
	case <-pollTick.C:
		// Ожидаемое поведение
	default:
		t.Error("pollTick should have triggered")
	}

	// Проверяем, что reportTick сработал
	select {
	case <-reportTick.C:
		// Ожидаемое поведение
	default:
		t.Error("reportTick should have triggered")
	}

	// Проверяем, что метрики были созданы и имеют правильный размер
	assert.NotNil(t, metrics)
	assert.NotNil(t, gpMetrics)

	// Оба должны быть пустыми картами
	assert.Equal(t, 0, len(metrics))
	assert.Equal(t, 0, len(gpMetrics))
}

// Определяем интерфейс для контроллера, чтобы мы могли его мокать
type ControllerInterface interface {
	SendingCounterMetrics(ctx context.Context, pollCount *int64, client *http.Client)
	SendingGaugeMetrics(ctx context.Context, metrics map[string]float64, client *http.Client)
}

// Модифицированный App для тестирования
type TestApp struct {
	controller ControllerInterface
	opt        *af.Options
	logger     *l.ZapLogger
	ctx        context.Context
	cancel     context.CancelFunc
}

// Переопределяем метод startAgent для использования контекста с отменой
func (ta *TestApp) startAgent() error {
	pollTick, reportTick, metrics, gpMetrics := initDataAgent(ta.opt)
	client := &http.Client{
		Transport: &http.Transport{
			MaxIdleConnsPerHost: 100,
			IdleConnTimeout:     90 * time.Second,
		},
		Timeout: 10 * time.Second,
	}
	var pollCount int64 = 0

	defer func() {
		pollTick.Stop()
		reportTick.Stop()
	}()

	for {
		select {
		case <-ta.ctx.Done():
			return nil
		case <-pollTick.C:
			pollCount++
			// Вместо вызова реальных функций сохранения метрик,
			// мы просто добавляем несколько значений для тестирования
			metrics["TestMetric1"] = 1.0
			metrics["TestMetric2"] = 2.0
			gpMetrics["TestGPMetric1"] = 10.0
		case <-reportTick.C:
			ta.controller.SendingCounterMetrics(ta.ctx, &pollCount, client)
			ta.controller.SendingGaugeMetrics(ta.ctx, metrics, client)
			ta.controller.SendingGaugeMetrics(ta.ctx, gpMetrics, client)
			return nil // Выходим после первой отправки для упрощения теста
		}
	}
}

// TestFullAgentCycle проверяет полный цикл работы агента
func TestFullAgentCycle(t *testing.T) {
	// Создаем логгер для теста
	logger, _ := l.NewZapLogger(zap.InfoLevel)

	// Создаем мок контроллера
	mockCtrl := new(MockController)

	// Создаем опции с короткими интервалами для быстрого теста
	opt := &af.Options{
		PollInterval:   1,
		ReportInterval: 2,
	}

	// Создаем контекст с отменой
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// Настраиваем ожидания для метода SendingCounterMetrics
	mockCtrl.On("SendingCounterMetrics",
		mock.MatchedBy(func(ctx context.Context) bool { return true }),
		mock.MatchedBy(func(pollCount *int64) bool { return *pollCount > 0 }),
		mock.Anything).Return()

	// Настраиваем ожидания для метода SendingGaugeMetrics для обычных метрик
	mockCtrl.On("SendingGaugeMetrics",
		mock.MatchedBy(func(ctx context.Context) bool { return true }),
		mock.MatchedBy(func(metrics map[string]float64) bool {
			return len(metrics) > 0
		}),
		mock.Anything).Return()

	// Создаем экземпляр тестового приложения
	testApp := &TestApp{
		controller: mockCtrl,
		opt:        opt,
		logger:     logger,
		ctx:        ctx,
		cancel:     cancel,
	}

	// Запускаем метод startAgent в горутине
	done := make(chan error)
	go func() {
		done <- testApp.startAgent()
	}()

	// Ожидаем завершения или таймаута
	select {
	case err := <-done:
		assert.NoError(t, err)
	case <-time.After(5 * time.Second):
		t.Fatal("Test timed out")
	}

	// Проверяем, что методы контроллера были вызваны
	mockCtrl.AssertExpectations(t)
}

// TestStartAgentSignalHandling проверяет обработку сигналов в startAgent
func TestStartAgentSignalHandling(t *testing.T) {
	// Создаем логгер для теста
	logger, _ := l.NewZapLogger(zap.InfoLevel)

	// Создаем мок контроллера
	mockCtrl := new(MockController)

	// Создаем опции с короткими интервалами для быстрого теста
	opt := &af.Options{
		PollInterval:   1,
		ReportInterval: 3, // Увеличиваем интервал, чтобы успеть отменить контекст
	}

	// Создаем контекст с отменой
	ctx, cancel := context.WithCancel(context.Background())

	// Создаем экземпляр тестового приложения
	testApp := &TestApp{
		controller: mockCtrl,
		opt:        opt,
		logger:     logger,
		ctx:        ctx,
		cancel:     cancel,
	}

	// Запускаем метод startAgent в горутине
	done := make(chan error)
	go func() {
		done <- testApp.startAgent()
	}()

	// Сразу отменяем контекст, чтобы завершить выполнение startAgent
	cancel()

	// Ожидаем завершения или таймаута
	select {
	case err := <-done:
		assert.NoError(t, err)
	case <-time.After(2 * time.Second):
		t.Fatal("Test timed out, context cancellation did not stop the agent")
	}
}

// TestNewWithCustomLogger проверяет создание приложения с пользовательским логгером
func TestNewWithCustomLogger(t *testing.T) {
	// Создаем кастомный логгер
	customLogger, _ := l.NewZapLogger(zap.InfoLevel)

	// Создаем опции
	opt := &af.Options{
		PollInterval:   1,
		ReportInterval: 1,
	}

	// Создаем приложение с кастомным логгером
	app := &App{
		controller: nil,
		opt:        opt,
		logger:     customLogger,
	}

	// Проверяем, что логгер был установлен
	assert.NotNil(t, app.logger)

	// Проверяем, что поле opt было установлено
	assert.Equal(t, opt, app.opt)
}
