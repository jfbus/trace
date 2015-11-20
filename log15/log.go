package log15

import (
	"github.com/jfbus/trace"
	log "gopkg.in/inconshreveable/log15.v2"
)

// Wrap creates a wrapper for a log15 logger.
// All events will be logged. Span event will be logged with their duration.
// If a level has been set, it will be used.
//	t := trace.New("name", []trace.Wrapper{log15.Wrap(log.DefaultLogger)})
func Wrap(l log.Logger) trace.Wrapper {
	return &log15Wrapper{log: l}
}

type log15Wrapper struct {
	log log.Logger
}

func (w *log15Wrapper) Setup(name string)               {}
func (w *log15Wrapper) Teardown()                       {}
func (w *log15Wrapper) Child(name string) trace.Wrapper { return w }

func (w *log15Wrapper) Event(e trace.Event) {
	lvl := trace.LvlDebug
	if ce, ok := e.(trace.LevelEvent); ok {
		lvl = ce.Level()
	}

	switch lvl {
	case trace.LvlCrit:
		w.log.Crit(e.Message(), log.Ctx(e.Context()))
	case trace.LvlErr:
		w.log.Error(e.Message(), log.Ctx(e.Context()))
	case trace.LvlWarn:
		w.log.Warn(e.Message(), log.Ctx(e.Context()))
	case trace.LvlInfo:
		w.log.Info(e.Message(), log.Ctx(e.Context()))
	default:
		w.log.Debug(e.Message(), log.Ctx(e.Context()))
	}
}
