# trace

Mailjet tracing library

It wraps :

* logging,
* metrics,
* tracing.

We are currently using log15, go-metrics and appdash, but those libs can be swapped with other libs, provided you build wrappers for the lib.

## Usage

```
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
```

## Supporting new libraries

wrappers must implement the following interface :

```
type Wrapper interface {
	Setup(name string)
	Teardown()
	Child(name string) Wrapper
	Event(e Event)
}
```

## License

MIT - see LICENSE