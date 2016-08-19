package test

import (
	"testing"
	"bytes"
	"github.com/lkiversonlk/OnlineChat/trace"
)

func TestNew(t *testing.T) {
	var buf bytes.Buffer
	tracer := trace.New(&buf)
	if tracer == nil {
		t.Error("Return from trace.New should not be nil")
	} else {
		tracer.Trace("Hello trace package.")
		if buf.String() != "Hello trace package.\n" {
			t.Errorf("Trace should not write '%s'.", buf.String())
		}
	}
}

func TestOff(t *testing.T) {
	var silentTracer = trace.NilTracer()
	silentTracer.Trace("something")
}