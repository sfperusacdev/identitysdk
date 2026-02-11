package staticstore

import "testing"

func TestSetAndGet(t *testing.T) {
	s := New[string, int]()

	s.Set("a", 10)

	v, ok := s.Get("a")
	if !ok {
		t.Fatalf("expected key to exist")
	}

	if v != 10 {
		t.Fatalf("expected 10, got %d", v)
	}
}

func TestGetMissing(t *testing.T) {
	s := New[string, int]()

	_, ok := s.Get("missing")
	if ok {
		t.Fatalf("expected key to not exist")
	}
}

func TestDelete(t *testing.T) {
	s := New[string, int]()

	s.Set("a", 1)
	s.Delete("a")

	_, ok := s.Get("a")
	if ok {
		t.Fatalf("expected key to be deleted")
	}
}

func TestLen(t *testing.T) {
	s := New[string, int]()

	if s.Len() != 0 {
		t.Fatalf("expected len 0")
	}

	s.Set("a", 1)
	s.Set("b", 2)

	if s.Len() != 2 {
		t.Fatalf("expected len 2, got %d", s.Len())
	}
}

func TestOverwrite(t *testing.T) {
	s := New[string, int]()

	s.Set("a", 1)
	s.Set("a", 2)

	v, _ := s.Get("a")

	if v != 2 {
		t.Fatalf("expected 2, got %d", v)
	}
}

func TestConcurrentAccess(t *testing.T) {
	s := New[int, int]()

	done := make(chan struct{})

	for i := range 100 {
		go func(v int) {
			s.Set(v, v)
			done <- struct{}{}
		}(i)
	}

	for range 100 {
		<-done
	}

	for i := range 100 {
		if _, ok := s.Get(i); !ok {
			t.Fatalf("missing key %d", i)
		}
	}
}
