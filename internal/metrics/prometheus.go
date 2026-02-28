package metrics

import (
	"context"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// MAURecorder implements auth.ActiveMonthRecorder. It writes the MAU row to DynamoDB with a
// conditional put (only if not already present this month); only when the put succeeds does it
// increment the Prometheus counter. No in-memory set — DynamoDB is the source of truth.
type MAURecorder struct {
	store   *MAUStore
	counter prometheus.Counter
}

// NewMAURecorder creates a recorder that uses the store for persistence and registers its counter on reg.
func NewMAURecorder(store *MAUStore, reg prometheus.Registerer) (*MAURecorder, error) {
	counter := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "mau_unique_users_seen_total",
		Help: "Unique users this process has recorded this month (login or refresh). One increment per first-seen user per month. Recorded on every new session issue (code exchange and token refresh).",
	})
	if err := reg.Register(counter); err != nil {
		return nil, err
	}
	return &MAURecorder{store: store, counter: counter}, nil
}

// RecordActiveMonth records that the user was active. Inserts a row in DynamoDB only if not already present this month; increments the counter only on insert.
func (r *MAURecorder) RecordActiveMonth(ctx context.Context, userID string) error {
	inserted, err := r.store.RecordActiveMonthIfNew(ctx, userID)
	if err != nil {
		return err
	}
	if inserted {
		r.counter.Inc()
	}
	return nil
}

// Handler returns an http.Handler that serves the default Prometheus registry (GET /metrics).
func Handler() http.Handler {
	return promhttp.Handler()
}

// HandlerForRegistry returns an http.Handler that serves the given registry. Use in tests so each test server has its own registry.
func HandlerForRegistry(reg *prometheus.Registry) http.Handler {
	return promhttp.HandlerFor(reg, promhttp.HandlerOpts{})
}
