package list

import "testing"

func TestUnique(t *testing.T) {
	in := []int{1, 2, 2, 3, 1, 4, 4, 5}
	out := Unique(in)

	expected := []int{1, 2, 3, 4, 5}

	if len(out) != len(expected) {
		t.Fatalf("len mismatch: got %v, want %v", out, expected)
	}

	for i := range expected {
		if out[i] != expected[i] {
			t.Fatalf("value mismatch: got %v, want %v", out, expected)
		}
	}
}

func TestUniqueString(t *testing.T) {
	in := []string{"a", "b", "a", "c", "b", "d"}
	out := Unique(in)

	expected := []string{"a", "b", "c", "d"}

	if len(out) != len(expected) {
		t.Fatalf("len mismatch: got %v, want %v", out, expected)
	}

	for i := range expected {
		if out[i] != expected[i] {
			t.Fatalf("value mismatch: got %v, want %v", out, expected)
		}
	}
}
