package trace

import (
	"net/http"
	"sync"
	"time"
)

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

type Wrapper interface {
	Setup(name string)
	Teardown()
	Child(name string) Wrapper
	Event(e Event)
}

type Span struct {
	Context Context
	impl    []Wrapper
	event   TimespanEvent
}

type Trace struct {
	Span
	Name    string
	Message string
	Start   time.Time
	sync.Mutex
	stats map[string]int
}

type TraceOption func(*Trace)

func New(name string, impl []Wrapper, opts ...TraceOption) *Trace {
	t := &Trace{Span: Span{impl: impl, event: newSpanEvent(name)}}
	for _, opt := range opts {
		opt(t)
	}
	for _, i := range t.impl {
		i.Setup(name)
	}
	return t
}

func DefaultContext(ctx Context) TraceOption {
	return func(t *Trace) {
		t.Context = ctx
	}
}

func FromRequest(req *http.Request) TraceOption {
	return func(t *Trace) {
		t.event = newServerEvent(t.event.Message(), req)
	}
}
func (s *Span) BeginSpan(msg string, opts ...EventOption) *Span {
	impl := make([]Wrapper, 0, len(s.impl))
	for _, i := range s.impl {
		impl = append(impl, i.Child(msg))
	}
	return &Span{Context: s.Context, event: newSpanEvent(msg, opts...), impl: impl}
}

func (s *Span) End(opts ...EventOption) {
	s.event.setDuration(time.Since(s.event.Start()))
	s.Event(s.event, opts...)
	for _, i := range s.impl {
		i.Teardown()
	}
}

func (s *Span) Event(e Event, opts ...EventOption) {
	e.opts(opts)
	e.Context().add(s.Context)
	for _, i := range s.impl {
		i.Event(e)
	}
}

/*func (s *Span) Duration() time.Duration {
	return s.event.duration
}*/

func (s *Span) Crit(msg string, opts ...EventOption) {
	s.Event(newLogEvent(Crit, msg, opts...))
}

func (s *Span) Error(msg string, opts ...EventOption) {
	s.Event(newLogEvent(Err, msg, opts...))
}

func (s *Span) Warn(msg string, opts ...EventOption) {
	s.Event(newLogEvent(Warn, msg, opts...))
}

func (s *Span) Info(msg string, opts ...EventOption) {
	s.Event(newLogEvent(Info, msg, opts...))
}
func (s *Span) Debug(msg string, opts ...EventOption) {
	s.Event(newLogEvent(Debug, msg, opts...))
}
