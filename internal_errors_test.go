package newrelic

import (
	"runtime"
	"testing"

	"github.com/newrelic/go-agent/internal"
)

type myError struct{}

func (e myError) Error() string { return "my msg" }

func TestNoticeErrorBackground(t *testing.T) {
	app := testApp(nil, nil, t)
	txn := app.StartTransaction("hello", nil, nil)
	err := txn.NoticeError(myError{})
	if nil != err {
		t.Error(err)
	}
	txn.End()
	app.ExpectErrors(t, []internal.WantError{{
		TxnName: "OtherTransaction/Go/hello",
		Msg:     "my msg",
		Klass:   "newrelic.myError",
		Caller:  "go-agent.TestNoticeErrorBackground",
		URL:     "",
	}})
	app.ExpectErrorEvents(t, []internal.WantErrorEvent{{
		TxnName: "OtherTransaction/Go/hello",
		Msg:     "my msg",
		Klass:   "newrelic.myError",
	}})
	app.ExpectMetrics(t, backgroundErrorMetrics)
}

func TestNoticeErrorWeb(t *testing.T) {
	app := testApp(nil, nil, t)
	txn := app.StartTransaction("hello", nil, helloRequest)
	err := txn.NoticeError(myError{})
	if nil != err {
		t.Error(err)
	}
	txn.End()
	app.ExpectErrors(t, []internal.WantError{{
		TxnName: "WebTransaction/Go/hello",
		Msg:     "my msg",
		Klass:   "newrelic.myError",
		Caller:  "go-agent.TestNoticeErrorWeb",
		URL:     "/hello",
	}})
	app.ExpectErrorEvents(t, []internal.WantErrorEvent{{
		TxnName: "WebTransaction/Go/hello",
		Msg:     "my msg",
		Klass:   "newrelic.myError",
	}})
	app.ExpectMetrics(t, webErrorMetrics)
}

func TestNoticeErrorTxnEnded(t *testing.T) {
	app := testApp(nil, nil, t)
	txn := app.StartTransaction("hello", nil, nil)
	txn.End()
	err := txn.NoticeError(myError{})
	if err != errAlreadyEnded {
		t.Error(err)
	}
	txn.End()
	app.ExpectErrors(t, []internal.WantError{})
	app.ExpectErrorEvents(t, []internal.WantErrorEvent{})
	app.ExpectMetrics(t, backgroundMetrics)
}

func TestNoticeErrorHighSecurity(t *testing.T) {
	cfgFn := func(cfg *Config) { cfg.HighSecurity = true }
	app := testApp(nil, cfgFn, t)
	txn := app.StartTransaction("hello", nil, nil)
	err := txn.NoticeError(myError{})
	if nil != err {
		t.Error(err)
	}
	txn.End()
	app.ExpectErrors(t, []internal.WantError{{
		TxnName: "OtherTransaction/Go/hello",
		Msg:     highSecurityErrorMsg,
		Klass:   "newrelic.myError",
		Caller:  "go-agent.TestNoticeErrorHighSecurity",
	}})
	app.ExpectErrorEvents(t, []internal.WantErrorEvent{{
		TxnName: "OtherTransaction/Go/hello",
		Msg:     highSecurityErrorMsg,
		Klass:   "newrelic.myError",
	}})
	app.ExpectMetrics(t, backgroundErrorMetrics)
}

func TestNoticeErrorLocallyDisabled(t *testing.T) {
	cfgFn := func(cfg *Config) { cfg.ErrorCollector.Enabled = false }
	app := testApp(nil, cfgFn, t)
	txn := app.StartTransaction("hello", nil, nil)
	err := txn.NoticeError(myError{})
	if errorsLocallyDisabled != err {
		t.Error(err)
	}
	txn.End()
	app.ExpectErrors(t, []internal.WantError{})
	app.ExpectErrorEvents(t, []internal.WantErrorEvent{})
	app.ExpectMetrics(t, backgroundMetrics)
}

func TestNoticeErrorRemotelyDisabled(t *testing.T) {
	replyfn := func(reply *internal.ConnectReply) { reply.CollectErrors = false }
	app := testApp(replyfn, nil, t)
	txn := app.StartTransaction("hello", nil, nil)
	err := txn.NoticeError(myError{})
	if errorsRemotelyDisabled != err {
		t.Error(err)
	}
	txn.End()
	app.ExpectErrors(t, []internal.WantError{})
	app.ExpectErrorEvents(t, []internal.WantErrorEvent{})
	app.ExpectMetrics(t, backgroundMetrics)
}

func TestNoticeErrorNil(t *testing.T) {
	app := testApp(nil, nil, t)
	txn := app.StartTransaction("hello", nil, nil)
	err := txn.NoticeError(nil)
	if errNilError != err {
		t.Error(err)
	}
	txn.End()
	app.ExpectErrors(t, []internal.WantError{})
	app.ExpectErrorEvents(t, []internal.WantErrorEvent{})
	app.ExpectMetrics(t, backgroundMetrics)
}

func TestNoticeErrorEventsLocallyDisabled(t *testing.T) {
	cfgFn := func(cfg *Config) { cfg.ErrorCollector.CaptureEvents = false }
	app := testApp(nil, cfgFn, t)
	txn := app.StartTransaction("hello", nil, nil)
	err := txn.NoticeError(myError{})
	if nil != err {
		t.Error(err)
	}
	txn.End()
	app.ExpectErrors(t, []internal.WantError{{
		TxnName: "OtherTransaction/Go/hello",
		Msg:     "my msg",
		Klass:   "newrelic.myError",
		Caller:  "go-agent.TestNoticeErrorEventsLocallyDisabled",
	}})
	app.ExpectErrorEvents(t, []internal.WantErrorEvent{})
	app.ExpectMetrics(t, backgroundErrorMetrics)
}

func TestNoticeErrorEventsRemotelyDisabled(t *testing.T) {
	replyfn := func(reply *internal.ConnectReply) { reply.CollectErrorEvents = false }
	app := testApp(replyfn, nil, t)
	txn := app.StartTransaction("hello", nil, nil)
	err := txn.NoticeError(myError{})
	if nil != err {
		t.Error(err)
	}
	txn.End()
	app.ExpectErrors(t, []internal.WantError{{
		TxnName: "OtherTransaction/Go/hello",
		Msg:     "my msg",
		Klass:   "newrelic.myError",
		Caller:  "go-agent.TestNoticeErrorEventsRemotelyDisabled",
	}})
	app.ExpectErrorEvents(t, []internal.WantErrorEvent{})
	app.ExpectMetrics(t, backgroundErrorMetrics)
}

type errorWithClass struct{ class string }

func (e errorWithClass) Error() string      { return "my msg" }
func (e errorWithClass) ErrorClass() string { return e.class }

func TestErrorWithClasser(t *testing.T) {
	app := testApp(nil, nil, t)
	txn := app.StartTransaction("hello", nil, nil)
	err := txn.NoticeError(errorWithClass{class: "zap"})
	if nil != err {
		t.Error(err)
	}
	txn.End()
	app.ExpectErrors(t, []internal.WantError{{
		TxnName: "OtherTransaction/Go/hello",
		Msg:     "my msg",
		Klass:   "zap",
		Caller:  "go-agent.TestErrorWithClasser",
		URL:     "",
	}})
	app.ExpectErrorEvents(t, []internal.WantErrorEvent{{
		TxnName: "OtherTransaction/Go/hello",
		Msg:     "my msg",
		Klass:   "zap",
	}})
	app.ExpectMetrics(t, backgroundErrorMetrics)
}

func TestErrorWithClasserReturnsEmpty(t *testing.T) {
	app := testApp(nil, nil, t)
	txn := app.StartTransaction("hello", nil, nil)
	err := txn.NoticeError(errorWithClass{class: ""})
	if nil != err {
		t.Error(err)
	}
	txn.End()
	app.ExpectErrors(t, []internal.WantError{{
		TxnName: "OtherTransaction/Go/hello",
		Msg:     "my msg",
		Klass:   "newrelic.errorWithClass",
		Caller:  "go-agent.TestErrorWithClasserReturnsEmpty",
		URL:     "",
	}})
	app.ExpectErrorEvents(t, []internal.WantErrorEvent{{
		TxnName: "OtherTransaction/Go/hello",
		Msg:     "my msg",
		Klass:   "newrelic.errorWithClass",
	}})
	app.ExpectMetrics(t, backgroundErrorMetrics)
}

type withStackTrace struct{ trace []uintptr }

func makeErrorWithStackTrace() error {
	callers := make([]uintptr, 20)
	written := runtime.Callers(1, callers)
	return withStackTrace{
		trace: callers[0:written],
	}
}

func (e withStackTrace) Error() string         { return "my msg" }
func (e withStackTrace) StackTrace() []uintptr { return e.trace }

func TestErrorWithStackTrace(t *testing.T) {
	app := testApp(nil, nil, t)
	txn := app.StartTransaction("hello", nil, nil)
	e := makeErrorWithStackTrace()
	err := txn.NoticeError(e)
	if nil != err {
		t.Error(err)
	}
	txn.End()
	app.ExpectErrors(t, []internal.WantError{{
		TxnName: "OtherTransaction/Go/hello",
		Msg:     "my msg",
		Klass:   "newrelic.withStackTrace",
		Caller:  "go-agent.makeErrorWithStackTrace",
		URL:     "",
	}})
	app.ExpectErrorEvents(t, []internal.WantErrorEvent{{
		TxnName: "OtherTransaction/Go/hello",
		Msg:     "my msg",
		Klass:   "newrelic.withStackTrace",
	}})
	app.ExpectMetrics(t, backgroundErrorMetrics)
}

func TestErrorWithStackTraceReturnsNil(t *testing.T) {
	app := testApp(nil, nil, t)
	txn := app.StartTransaction("hello", nil, nil)
	e := withStackTrace{trace: nil}
	err := txn.NoticeError(e)
	if nil != err {
		t.Error(err)
	}
	txn.End()
	app.ExpectErrors(t, []internal.WantError{{
		TxnName: "OtherTransaction/Go/hello",
		Msg:     "my msg",
		Klass:   "newrelic.withStackTrace",
		Caller:  "go-agent.TestErrorWithStackTraceReturnsNil",
		URL:     "",
	}})
	app.ExpectErrorEvents(t, []internal.WantErrorEvent{{
		TxnName: "OtherTransaction/Go/hello",
		Msg:     "my msg",
		Klass:   "newrelic.withStackTrace",
	}})
	app.ExpectMetrics(t, backgroundErrorMetrics)
}