# trace

Mailjet tracing library

It wraps :

* logging,
* metrics,
* tracing.

We are currently using log15, go-metrics and appdash, but those libs can be swapped with other libs, provided you build wrappers for the lib.

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