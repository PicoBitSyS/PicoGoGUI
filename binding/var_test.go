package binding

import (
	"sync"
	"testing"
)

func TestVarNotify(t *testing.T) {
	v := New(1)
	got := 0
	v.Subscribe(func(n int) { got = n })
	v.Set(42)
	if v.Get() != 42 || got != 42 {
		t.Fatalf("get=%d got=%d", v.Get(), got)
	}
}

func TestSetSilent(t *testing.T) {
	v := New("a")
	called := false
	v.Subscribe(func(string) { called = true })
	v.SetSilent("b")
	if v.Get() != "b" || called {
		t.Fatal("SetSilent should not notify")
	}
}

func TestUnsubscribe(t *testing.T) {
	v := New(0)
	n := 0
	unsub := v.Subscribe(func(int) { n++ })
	v.Set(1)
	unsub()
	v.Set(2)
	if n != 1 {
		t.Fatalf("n=%d", n)
	}
}

func TestConcurrentSet(t *testing.T) {
	v := New(0)
	var wg sync.WaitGroup
	for i := 0; i < 32; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			v.Set(i)
			_ = v.Get()
		}(i)
	}
	wg.Wait()
}

func TestSetFromSkipsOnlyMatchingOrigin(t *testing.T) {
	v := New("initial")
	a := NewOrigin()
	b := NewOrigin()
	var gotA, gotB, gotPlain string
	v.SubscribeFrom(a, func(value string) { gotA = value })
	v.SubscribeFrom(b, func(value string) { gotB = value })
	v.Subscribe(func(value string) { gotPlain = value })

	v.SetFrom(a, "changed")

	if gotA != "" {
		t.Fatalf("origin subscriber received its own update: %q", gotA)
	}
	if gotB != "changed" || gotPlain != "changed" {
		t.Fatalf("other subscribers not notified: b=%q plain=%q", gotB, gotPlain)
	}
}
