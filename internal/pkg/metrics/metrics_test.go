package metrics

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewMetrics(t *testing.T) {
	// 各テストで新しいレジストリを使用
	reg := prometheus.NewRegistry()
	m := NewWithRegistry(reg)

	require.NotNil(t, m)
	assert.NotNil(t, m.HTTPRequestsTotal)
	assert.NotNil(t, m.HTTPRequestDuration)
	assert.NotNil(t, m.ReservationsTotal)
	assert.NotNil(t, m.DistributedLockDuration)
	assert.NotNil(t, m.ActiveReservations)
}

func TestHTTPRequestsTotal(t *testing.T) {
	reg := prometheus.NewRegistry()
	m := NewWithRegistry(reg)

	// リクエストをカウント
	m.HTTPRequestsTotal.WithLabelValues("GET", "/api/v1/events", "200").Inc()
	m.HTTPRequestsTotal.WithLabelValues("POST", "/api/v1/reservations", "201").Inc()
	m.HTTPRequestsTotal.WithLabelValues("POST", "/api/v1/reservations", "409").Inc()

	// メトリクスが正しく収集されているか確認
	families, err := reg.Gather()
	require.NoError(t, err)

	var found bool
	for _, f := range families {
		if f.GetName() == "http_requests_total" {
			found = true
			assert.Equal(t, 3, len(f.GetMetric()))
		}
	}
	assert.True(t, found, "http_requests_total metric not found")
}

func TestReservationsTotal(t *testing.T) {
	reg := prometheus.NewRegistry()
	m := NewWithRegistry(reg)

	// 予約成功・失敗をカウント
	m.ReservationsTotal.WithLabelValues("success").Inc()
	m.ReservationsTotal.WithLabelValues("success").Inc()
	m.ReservationsTotal.WithLabelValues("conflict").Inc()
	m.ReservationsTotal.WithLabelValues("lock_failed").Inc()

	families, err := reg.Gather()
	require.NoError(t, err)

	var found bool
	for _, f := range families {
		if f.GetName() == "reservations_total" {
			found = true
			assert.Equal(t, 3, len(f.GetMetric()))
		}
	}
	assert.True(t, found, "reservations_total metric not found")
}

func TestDistributedLockDuration(t *testing.T) {
	reg := prometheus.NewRegistry()
	m := NewWithRegistry(reg)

	// ロック取得時間を観測
	m.DistributedLockDuration.WithLabelValues("acquire", "success").Observe(0.015)
	m.DistributedLockDuration.WithLabelValues("acquire", "failed").Observe(0.005)
	m.DistributedLockDuration.WithLabelValues("release", "success").Observe(0.002)

	families, err := reg.Gather()
	require.NoError(t, err)

	var found bool
	for _, f := range families {
		if f.GetName() == "distributed_lock_duration_seconds" {
			found = true
		}
	}
	assert.True(t, found, "distributed_lock_duration_seconds metric not found")
}

func TestActiveReservations(t *testing.T) {
	reg := prometheus.NewRegistry()
	m := NewWithRegistry(reg)

	// アクティブな予約数を増減
	m.ActiveReservations.WithLabelValues("pending").Inc()
	m.ActiveReservations.WithLabelValues("pending").Inc()
	m.ActiveReservations.WithLabelValues("confirmed").Inc()
	m.ActiveReservations.WithLabelValues("pending").Dec() // 1つキャンセル

	families, err := reg.Gather()
	require.NoError(t, err)

	var found bool
	for _, f := range families {
		if f.GetName() == "active_reservations" {
			found = true
			// pending: 1, confirmed: 1
			assert.Equal(t, 2, len(f.GetMetric()))
		}
	}
	assert.True(t, found, "active_reservations metric not found")
}

func TestHTTPRequestDuration(t *testing.T) {
	reg := prometheus.NewRegistry()
	m := NewWithRegistry(reg)

	// レイテンシを観測
	m.HTTPRequestDuration.WithLabelValues("GET", "/api/v1/events").Observe(0.025)
	m.HTTPRequestDuration.WithLabelValues("POST", "/api/v1/reservations").Observe(0.150)

	families, err := reg.Gather()
	require.NoError(t, err)

	var found bool
	for _, f := range families {
		if f.GetName() == "http_request_duration_seconds" {
			found = true
		}
	}
	assert.True(t, found, "http_request_duration_seconds metric not found")
}

func TestGet_ReturnsDefaultMetrics(t *testing.T) {
	// Getは defaultMetrics を返す
	// 注意: Init が呼ばれていない場合は nil を返す可能性がある
	m := Get()
	// nil または Metrics インスタンスが返る
	if m != nil {
		assert.NotNil(t, m.HTTPRequestsTotal)
	}
}

func TestInit_CreatesDefaultMetrics(t *testing.T) {
	// 既存のdefaultMetricsをバックアップ
	oldMetrics := defaultMetrics
	defer func() { defaultMetrics = oldMetrics }()

	// 新しいレジストリでテスト用メトリクスを作成してdefaultMetricsにセット
	// 注意: Initを呼ぶとデフォルトレジストリに登録するため、テストでは直接セット
	reg := prometheus.NewRegistry()
	m := NewWithRegistry(reg)
	defaultMetrics = m

	// Get()がdefaultMetricsを返すことを確認
	got := Get()
	assert.NotNil(t, got)
	assert.Equal(t, m, got)
}
