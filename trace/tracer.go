package trace

import (
	"io"
	"fmt"
)

type Tracer interface {
	Trace(...interface{})
}

type nilTracer struct {}

type tracer struct {
	out io.Writer
}

func (t *nilTracer) Trace(a ...interface{}){}

func (t *tracer) Trace(a ...interface{}) {
	t.out.Write([]byte(fmt.Sprint(a...)))
	t.out.Write([]byte("\n"))
}

func New(w io.Writer) Tracer{
	return &tracer{out: w}
}

func NilTracer() Tracer{
	return &nilTracer{}
}