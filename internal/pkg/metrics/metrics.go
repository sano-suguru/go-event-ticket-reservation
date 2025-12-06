package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

// Metrics はアプリケーションのメトリクスを管理する
type Metrics struct {
	// HTTPリクエストの総数（method, path, status_code）
	HTTPRequestsTotal *prometheus.CounterVec

	// HTTPリクエストのレイテンシ（method, path）
	HTTPRequestDuration *prometheus.HistogramVec

	// 予約の総数（status: success, conflict, lock_failed, error）
	ReservationsTotal *prometheus.CounterVec

	// 分散ロックの操作時間（operation: acquire/release, status: success/failed）
	DistributedLockDuration *prometheus.HistogramVec

	// アクティブな予約数（status: pending, confirmed）
	ActiveReservations *prometheus.GaugeVec
}

// New は新しいMetricsインスタンスを作成し、デフォルトレジストリに登録する
func New() *Metrics {
	return NewWithRegistry(prometheus.DefaultRegisterer)
}

// NewWithRegistry は指定したレジストリにメトリクスを登録する
func NewWithRegistry(reg prometheus.Registerer) *Metrics {
	m := &Metrics{
		HTTPRequestsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "http_requests_total",
				Help: "Total number of HTTP requests",
			},
			[]string{"method", "path", "status_code"},
		),
		HTTPRequestDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "http_request_duration_seconds",
				Help:    "HTTP request latency in seconds",
				Buckets: []float64{.005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10},
			},
			[]string{"method", "path"},
		),
		ReservationsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "reservations_total",
				Help: "Total number of reservation attempts",
			},
			[]string{"status"},
		),
		DistributedLockDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "distributed_lock_duration_seconds",
				Help:    "Time spent on distributed lock operations",
				Buckets: []float64{.001, .005, .01, .025, .05, .1, .25, .5, 1},
			},
			[]string{"operation", "status"},
		),
		ActiveReservations: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "active_reservations",
				Help: "Current number of active reservations",
			},
			[]string{"status"},
		),
	}

	// レジストリに登録
	reg.MustRegister(
		m.HTTPRequestsTotal,
		m.HTTPRequestDuration,
		m.ReservationsTotal,
		m.DistributedLockDuration,
		m.ActiveReservations,
	)

	return m
}

// デフォルトのメトリクスインスタンス
var defaultMetrics *Metrics

// Init はデフォルトのメトリクスインスタンスを初期化する
func Init() *Metrics {
	defaultMetrics = New()
	return defaultMetrics
}

// Get はデフォルトのメトリクスインスタンスを返す
func Get() *Metrics {
	return defaultMetrics
}
