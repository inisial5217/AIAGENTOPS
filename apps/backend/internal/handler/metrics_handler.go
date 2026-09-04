package handler

import (
	"strconv"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Metrics holds prometheus collectors
type Metrics struct {
	httpRequestsTotal       *prometheus.CounterVec
	httpRequestDuration     *prometheus.HistogramVec
	activeWebsocketConns    prometheus.Gauge
	dbPoolActiveConnections prometheus.Gauge
	dbPoolIdleConnections   prometheus.Gauge
	db                      *pgxpool.Pool
}

// NewMetrics initializes metrics
func NewMetrics(db *pgxpool.Pool) *Metrics {
	m := &Metrics{
		httpRequestsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "http_requests_total",
				Help: "Total number of HTTP requests processed",
			},
			[]string{"method", "path", "status_code"},
		),
		httpRequestDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "http_request_duration_seconds",
				Help:    "Histogram of response duration for HTTP requests",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"method", "path"},
		),
		activeWebsocketConns: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name: "active_websocket_connections",
				Help: "Current active WebSocket connections",
			},
		),
		dbPoolActiveConnections: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name: "db_pool_active_connections",
				Help: "Current active database connections in pool",
			},
		),
		dbPoolIdleConnections: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name: "db_pool_idle_connections",
				Help: "Current idle database connections in pool",
			},
		),
		db: db,
	}

	// register prometheus metrics
	prometheus.MustRegister(
		m.httpRequestsTotal,
		m.httpRequestDuration,
		m.activeWebsocketConns,
		m.dbPoolActiveConnections,
		m.dbPoolIdleConnections,
	)

	return m
}

// Handler returns promhttp handler
func (m *Metrics) Handler() echo.HandlerFunc {
	h := promhttp.Handler()
	return func(c echo.Context) error {
		// update db pool stats
		if m.db != nil {
			stat := m.db.Stat()
			m.dbPoolActiveConnections.Set(float64(stat.AcquiredConns()))
			m.dbPoolIdleConnections.Set(float64(stat.IdleConns()))
		}
		h.ServeHTTP(c.Response(), c.Request())
		return nil
	}
}

// Middleware records http metrics
func (m *Metrics) Middleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			start := time.Now()

			err := next(c)
			if err != nil {
				c.Error(err)
			}

			duration := time.Since(start).Seconds()
			status := strconv.Itoa(c.Response().Status)
			path := c.Path()
			if path == "" {
				path = c.Request().URL.Path
			}
			method := c.Request().Method

			m.httpRequestsTotal.WithLabelValues(method, path, status).Inc()
			m.httpRequestDuration.WithLabelValues(method, path).Observe(duration)

			return nil
		}
	}
}
