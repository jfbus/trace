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

// Event is the generic event interface
type Event interface {
	Message() string
	Metric() string
	Context() Context
	opts([]EventOption)
}

// TimespanEvent is an event that has a start and a duration
type TimespanEvent interface {
	Event
	Start() time.Time
	Duration() time.Duration
	setDuration(time.Duration)
}

// LevelEvent is an event that has a criticity level
type LevelEvent interface {
	Event
	Level() int
}

// HttpServerEvent is a event for a http server request
type HttpServerEvent interface {
	TimespanEvent
	ServerRequest() *http.Request
}

type baseEvent struct {
	msg    string
	metric string
	ctx    Context
	lvl    int
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

func (e *baseEvent) Level() int {
	return e.lvl
}

func (e *baseEvent) opts(opts []EventOption) {
	for _, opt := range opts {
		opt(e)
	}
}

// Metric adds a metric to an event. Basic events will be measured with a counter/meter.
// Span events will be measured with a counter/meter and a timer.
func Metric(m string) EventOption {
	return func(e *baseEvent) {
		e.metric = m
	}
}

// AddContext adds some context to the event
func AddContext(ctx Context) EventOption {
	return func(e *baseEvent) {
		e.ctx.add(ctx)
	}
}

// Level sets a cricitity level to an event
func Level(lvl int) EventOption {
	return func(e *baseEvent) {
		e.lvl = lvl
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

// LogEvent creates a basic logging event, with a level and a message.
func LogEvent(lvl int, msg string, args ...interface{}) Event {
	return &logEvent{baseEvent: baseEvent{msg: msg, ctx: Context{}, lvl: lvl}, args: args}
}

func newLogEvent(lvl int, msg string, opts ...EventOption) *logEvent {
	e := &logEvent{baseEvent: baseEvent{msg: msg, ctx: Context{}, lvl: lvl}}
	e.opts(opts)
	return e
}

func newSpanEvent(msg string, opts ...EventOption) *spanEvent {
	e := &spanEvent{baseEvent: baseEvent{msg: msg, ctx: Context{}}, start: time.Now()}
	e.opts(opts)
	return e
}

func newHttpServerEvent(msg string, req *http.Request, opts ...EventOption) *serverEvent {
	e := &serverEvent{spanEvent: spanEvent{baseEvent: baseEvent{msg: msg, ctx: Context{}}, start: time.Now()}, req: req}
	e.opts(opts)
	return e
}
