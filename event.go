package trace

import (
	"fmt"
	"net/http"
	"time"
)

const (
	LvlDebug = iota
	LvlInfo
	LvlWarn
	LvlErr
	LvlCrit
)

type EventOption func(*baseEvent)

type Event interface {
	Message() string
	Metric() string
	Context() Context
	opts([]EventOption)
}

type TimespanEvent interface {
	Event
	Start() time.Time
	Duration() time.Duration
	setDuration(time.Duration)
}

type CriticityEvent interface {
	Event
	Crit() int
}

type ServerEvent interface {
	TimespanEvent
	ServerRequest() *http.Request
}

type baseEvent struct {
	msg    string
	metric string
	ctx    Context
	crit   int
}

func (e *baseEvent) Message() string {
	return e.msg
}

func (e *baseEvent) Metric() string {
	return e.metric
}

func (e *baseEvent) Context() Context {
	return e.ctx
}

func (e *baseEvent) Crit() int {
	return e.crit
}

func (e *baseEvent) opts(opts []EventOption) {
	for _, opt := range opts {
		opt(e)
	}
}

func Metric(m string) EventOption {
	return func(e *baseEvent) {
		e.metric = m
	}
}

func AddContext(ctx Context) EventOption {
	return func(e *baseEvent) {
		e.ctx.add(ctx)
	}
}

func Crit(crit int) EventOption {
	return func(e *baseEvent) {
		e.crit = crit
	}
}

type logEvent struct {
	baseEvent
	args []interface{}
}

func (e *logEvent) Message() string {
	if e.args != nil {
		return fmt.Sprintf(e.msg, e.args...)
	}
	return e.msg
}

type spanEvent struct {
	baseEvent
	start    time.Time
	duration time.Duration
}

func (e *spanEvent) Context() Context {
	if e.duration != 0 {
		e.ctx["duration"] = e.duration
	}
	return e.ctx
}

func (e *spanEvent) Start() time.Time {
	return e.start
}

func (e *spanEvent) Duration() time.Duration {
	return e.duration
}

func (e *spanEvent) setDuration(d time.Duration) {
	e.duration = d
}

type serverEvent struct {
	spanEvent
	req *http.Request
}

func (e *serverEvent) ServerRequest() *http.Request {
	return e.req
}

func LogEvent(crit int, msg string, args ...interface{}) Event {
	return &logEvent{baseEvent: baseEvent{msg: msg, ctx: Context{}, crit: crit}, args: args}
}

func newLogEvent(crit int, msg string, opts ...EventOption) *logEvent {
	e := &logEvent{baseEvent: baseEvent{msg: msg, ctx: Context{}, crit: crit}}
	e.opts(opts)
	return e
}

func newSpanEvent(msg string, opts ...EventOption) *spanEvent {
	e := &spanEvent{baseEvent: baseEvent{msg: msg, ctx: Context{}}, start: time.Now()}
	e.opts(opts)
	return e
}

func newServerEvent(msg string, req *http.Request, opts ...EventOption) *serverEvent {
	e := &serverEvent{spanEvent: spanEvent{baseEvent: baseEvent{msg: msg, ctx: Context{}}, start: time.Now()}, req: req}
	e.opts(opts)
	return e
}
