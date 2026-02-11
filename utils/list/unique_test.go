package list

import (
	"reflect"
	"testing"
)

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

func TestNonZeroUniques_Int(t *testing.T) {
	input := []int{0, 1, 2, 2, 0, 3, 1}
	expected := []int{1, 2, 3}

	result := NonZeroUniques(input)

	if !reflect.DeepEqual(result, expected) {
		t.Fatalf("got %v, want %v", result, expected)
	}
}

func TestNonZeroUniques_String(t *testing.T) {
	input := []string{"", "a", "b", "a", "", "c"}
	expected := []string{"a", "b", "c"}

	result := NonZeroUniques(input)

	if !reflect.DeepEqual(result, expected) {
		t.Fatalf("got %v, want %v", result, expected)
	}
}

func TestNonZeroUniques_AllZero(t *testing.T) {
	input := []int{0, 0, 0}
	expected := []int{}

	result := NonZeroUniques(input)

	if !reflect.DeepEqual(result, expected) {
		t.Fatalf("got %v, want %v", result, expected)
	}
}

func TestNonZeroUniques_AlreadyUnique(t *testing.T) {
	input := []int{1, 2, 3}
	expected := []int{1, 2, 3}

	result := NonZeroUniques(input)

	if !reflect.DeepEqual(result, expected) {
		t.Fatalf("got %v, want %v", result, expected)
	}
}

func TestNonZeroUniques_Struct(t *testing.T) {
	type S struct {
		A int
	}

	input := []S{{}, {1}, {2}, {1}, {}}
	expected := []S{{1}, {2}}

	result := NonZeroUniques(input)

	if !reflect.DeepEqual(result, expected) {
		t.Fatalf("got %v, want %v", result, expected)
	}
}

func TestNonZeroUniques_Empty(t *testing.T) {
	input := []int{}
	expected := []int{}

	result := NonZeroUniques(input)

	if !reflect.DeepEqual(result, expected) {
		t.Fatalf("got %v, want %v", result, expected)
	}
}
