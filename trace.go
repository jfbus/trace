/*
Trace a package that wraps logging/metrics/zipkin-like application.

Usage :

	// start a trace
	t := trace.New(req.RequestURI, []trace.Wrapper{gometrics.Wrap(metrics.DefaultRegistry), log15.Wrap(log.DefaultLogger), appdash_trace.Wrap(appdashCollector)}, trace.FromHttpRequest(req))

	// start a span
	span := t.BeginSpan("DB transaction", trace.Metric("db.transaction"))
	...
	// send a log event
	span.Event(trace.LogEvent(trace.LvlErr, "Unable to do something : %s", err))
	...
	span.End()

	// update the "request" timer, log the request with the duration and the status code
	t.End(trace.Metric("request"), trace.AddContext(trace.Context{"status": statusCode}), trace.Level(trace.LvlInfo))

Wrappers for log15, go-metrics and appdash are included, other libs can easily be used by writing the adequate wrapper.
*/
package trace

import (
	"net/http"
	"time"
)

// Context defines the context of an event
type Context map[string]interface{}

func (c Context) add(ns ...Context) {
	for _, n := range ns {
		if n == nil {
			continue
		}
		for k, v := range n {
			c[k] = v
		}
	}
}

// Wrapper wraps a service managing events (logger, tracing system, metrics collection, ...)
type Wrapper interface {
	Setup(name string)
	Teardown()
	Child(name string) Wrapper
	Event(e Event)
}

// A Trace is a basic unit of work, having a start time, an event time, and having
// child spans or events.
type Trace struct {
	Context Context
	impl    []Wrapper
	event   TimespanEvent
}

type TraceOption func(*Trace)

// New creates a new trace using a list of wrappers and some options.
func New(name string, impl []Wrapper, opts ...TraceOption) *Trace {
	t := &Trace{impl: impl, event: newSpanEvent(name)}
	for _, opt := range opts {
		opt(t)
	}
	for _, i := range t.impl {
		i.Setup(name)
	}
	return t
}

// DefaultContext is a TraceOption that defines a default context that will be used for
// all spans and events.
func DefaultContext(ctx Context) TraceOption {
	return func(t *Trace) {
		t.Context = ctx
	}
}

// FromHttpRequest is a TraceOption that marks the trace as a server trace.
func FromHttpRequest(req *http.Request) TraceOption {
	return func(t *Trace) {
		t.event = newHttpServerEvent(t.event.Message(), req)
	}
}

// BeginSpan starts a new child span. Some additional EventOptions can be set.
// They will be added to all events in this span.
func (t *Trace) BeginSpan(msg string, opts ...EventOption) *Trace {
	impl := make([]Wrapper, 0, len(t.impl))
	for _, i := range t.impl {
		impl = append(impl, i.Child(msg))
	}
	return &Trace{Context: t.Context, event: newSpanEvent(msg, opts...), impl: impl}
}

// End ends a span. It will trigger a span event, with the specified options.
func (t *Trace) End(opts ...EventOption) {
	t.event.setDuration(time.Since(t.event.Start()))
	t.Event(t.event, opts...)
	for _, i := range t.impl {
		i.Teardown()
	}
}

// Event records an event in a span.
func (t *Trace) Event(e Event, opts ...EventOption) {
	e.opts(opts)
	e.Context().add(t.Context)
	for _, i := range t.impl {
		i.Event(e)
	}
}

// Crit sends a critical log event
func (t *Trace) Crit(msg string, opts ...EventOption) {
	t.Event(newLogEvent(LvlCrit, msg, opts...))
}

// Crit sends an error log event
func (t *Trace) Error(msg string, opts ...EventOption) {
	t.Event(newLogEvent(LvlErr, msg, opts...))
}

// Crit sends a warning log event
func (t *Trace) Warn(msg string, opts ...EventOption) {
	t.Event(newLogEvent(LvlWarn, msg, opts...))
}

// Crit sends an info log event
func (t *Trace) Info(msg string, opts ...EventOption) {
	t.Event(newLogEvent(LvlInfo, msg, opts...))
}

// Crit sends a debug log event
func (t *Trace) Debug(msg string, opts ...EventOption) {
	t.Event(newLogEvent(LvlDebug, msg, opts...))
}
