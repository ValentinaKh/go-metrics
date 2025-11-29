package pool

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

type TestResetter struct {
	Value int
}

func (t *TestResetter) Reset() {
	t.Value = 0
}

func newTestResetter() *TestResetter {
	return &TestResetter{Value: 42}
}

func TestNew(t *testing.T) {
	p := New(func() *TestResetter {
		return &TestResetter{}
	})
	if p == nil {
		t.Fatal("Expected non-nil pool")
	}
}

func TestGet(t *testing.T) {
	p := New(newTestResetter)
	assert.Equal(t, 42, p.Get().Value)
}

func TestPut(t *testing.T) {
	p := New(newTestResetter)
	obj := p.Get()
	obj.Value = 999

	p.Put(obj)

	assert.Equal(t, 0, obj.Value)
}
