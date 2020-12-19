package trace

import (
	"bytes"
	"fmt"
	"testing"
)

func TestNew(t *testing.T) {
	var buf bytes.Buffer
	tracer := New(&buf)
	if tracer == nil {
		t.Errorf("Return from New should not be nil")
	} else {
		m := "Hello trace package."
		tracer.Trace(m)
		if buf.String() != fmt.Sprintf("%s\n", m) {
			t.Errorf("Trace should not write '%s'.", buf.String())
		}
	}
}
