package appdash

import (
	"sync"
	"time"

	"github.com/jfbus/trace"
	"sourcegraph.com/sourcegraph/appdash"
	"sourcegraph.com/sourcegraph/appdash/httptrace"
)

type spanEvent struct {
	Span       string
	ClientSend time.Time
	ClientRecv time.Time
}

func (spanEvent) Schema() string      { return "mj_spanEvent" }
func (spanEvent) Important() []string { return []string{"Name"} }
func (e spanEvent) Start() time.Time  { return e.ClientSend }
func (e spanEvent) End() time.Time    { return e.ClientRecv }

func init() {
	appdash.RegisterEvent(spanEvent{})
}

// Wrap generates a Wrapper around an appdash collector
// 	memStore := appdash.NewMemoryStore()
// 	store := &appdash.RecentStore{
// 		MinEvictAge: 60 * time.Second,
// 		DeleteStore: memStore,
// 	}
// 	coll := appdash.NewLocalCollector(store)
// 	server.RegisterTraceCollector(coll)
// 	tr := trace.New("name", []trace.Wrapper{appdash_trace.Wrap(appdashColl)})
func Wrap(coll appdash.Collector) trace.Wrapper {
	return &appdashWrapper{coll: coll}
}

type appdashWrapper struct {
	sync.Mutex
	coll  appdash.Collector
	rec   *appdash.Recorder
	child bool
}

func (w *appdashWrapper) Setup(name string) {
	w.rec = appdash.NewRecorder(appdash.NewRootSpanID(), w.coll)
	w.rec.Name(name)
}

func (w *appdashWrapper) Teardown() {}

func (w *appdashWrapper) Child(name string) trace.Wrapper {
	nw := &appdashWrapper{
		coll:  w.coll,
		rec:   w.rec.Child(),
		child: true,
	}
	nw.rec.Name(name)
	return nw
}

func (w *appdashWrapper) Event(e trace.Event) {
	if se, ok := e.(trace.HttpServerEvent); ok {
		if se.Duration() > 0 {
			sse := httptrace.NewServerEvent(se.ServerRequest())
			sse.ServerRecv = se.Start()
			sse.ServerSend = se.Start().Add(se.Duration())
			w.rec.Event(sse)
		}
	} else if de, ok := e.(trace.TimespanEvent); ok {
		if de.Duration() > 0 {
			w.rec.Event(&spanEvent{Span: de.Message(), ClientSend: de.Start(), ClientRecv: de.Start().Add(de.Duration())})
		}
	} else {
		w.rec.Msg(e.Message())
	}
}
