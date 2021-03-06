package gometrics

import (
	"time"

	"github.com/jfbus/trace"
	"github.com/rcrowley/go-metrics"
)

type MetricsOption func(*gometricsWrapper)

// Wrap creates a wrapper around a go-metrics registry.
// All metrics will be extracted from the registry.
// When an event is received with a Metric(name) option,
// a metric with the same name will be searched.
// The following metric types are supported :
// Meter/Counter (all events) and Timer (span event only)
//	t := trace.New("name", []trace.Wrapper{gometrics.Wrap(metrics.DefaultRegistry)})
func Wrap(r metrics.Registry) trace.Wrapper {
	w := &gometricsWrapper{metrics: map[string]interface{}{}}
	r.Each(func(name string, metric interface{}) {
		w.metrics[name] = metric
	})
	return w
}

type gometricsWrapper struct {
	metrics map[string]interface{}
}

func (w *gometricsWrapper) Setup(name string)               {}
func (w *gometricsWrapper) Teardown()                       {}
func (w *gometricsWrapper) Child(name string) trace.Wrapper { return w }

func (w *gometricsWrapper) Event(e trace.Event) {
	if m := e.Metric(); m != "" {
		if de, ok := e.(trace.TimespanEvent); ok {
			w.time(m, de.Duration())
		} else {
			w.count(m)
		}
	}
}

func (w *gometricsWrapper) count(name string) {
	m := w.metrics[name]
	if mm, ok := m.(metrics.Meter); ok {
		mm.Mark(1)
	} else if mc, ok := m.(metrics.Counter); ok {
		mc.Inc(1)
	}
}

func (w *gometricsWrapper) time(name string, dur time.Duration) {
	m := w.metrics[name]
	if mt, ok := m.(metrics.Timer); ok {
		mt.Update(dur)
	} else if mm, ok := m.(metrics.Meter); ok {
		mm.Mark(1)
	} else if mc, ok := m.(metrics.Counter); ok {
		mc.Inc(1)
	}
}
