package log15

import (
	"github.com/mailjet/trace"
	log "gopkg.in/inconshreveable/log15.v2"
)

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
	crit := trace.LvlDebug
	if ce, ok := e.(trace.CriticityEvent); ok {
		crit = ce.Crit()
	}

	switch crit {
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
