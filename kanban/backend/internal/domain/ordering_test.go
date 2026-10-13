package domain

import (
	"testing"

	"github.com/google/uuid"
)

func ids(n int) []uuid.UUID {
	out := make([]uuid.UUID, 0, n)
	for i := 0; i < n; i++ {
		out = append(out, uuid.New())
	}
	return out
}

func TestInsertAt(t *testing.T) {
	base := ids(3)
	a, b, c := base[0], base[1], base[2]
	moved := uuid.New()

	tests := []struct {
		name string
		in   []uuid.UUID
		idx  int32
		want []uuid.UUID
	}{
		{"front", base, 0, []uuid.UUID{moved, a, b, c}},
		{"middle", base, 1, []uuid.UUID{a, moved, b, c}},
		{"end", base, 2, []uuid.UUID{a, b, moved, c}},
		{"past end clamps to end", base, 99, []uuid.UUID{a, b, c, moved}},
		{"negative clamps to front", base, -4, []uuid.UUID{moved, a, b, c}},
		{"empty list", nil, 0, []uuid.UUID{moved}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := InsertAt(tc.in, tc.idx, moved)
			if len(got) != len(tc.want) {
				t.Fatalf("length = %d, want %d", len(got), len(tc.want))
			}
			for i := range got {
				if got[i] != tc.want[i] {
					t.Fatalf("index %d = %v, want %v", i, got[i], tc.want[i])
				}
			}
		})
	}
}

// InsertAt must build a new slice. The reindex loops in the move slices hold
// onto the original ordering to compare against, so an in-place insert would
// silently corrupt the caller.
func TestInsertAtDoesNotMutateInput(t *testing.T) {
	base := ids(3)
	original := make([]uuid.UUID, len(base))
	copy(original, base)

	_ = InsertAt(base, 1, uuid.New())

	for i := range base {
		if base[i] != original[i] {
			t.Fatalf("input slice was mutated at index %d", i)
		}
	}
}

func TestRemoveID(t *testing.T) {
	base := ids(3)
	got := RemoveID(base, base[1])

	if len(got) != 2 {
		t.Fatalf("length = %d, want 2", len(got))
	}
	if got[0] != base[0] || got[1] != base[2] {
		t.Fatalf("RemoveID kept the wrong elements: %v", got)
	}
}

func TestRemoveIDMissingIsNoOp(t *testing.T) {
	base := ids(2)
	got := RemoveID(base, uuid.New())
	if len(got) != 2 {
		t.Fatalf("length = %d, want 2 (absent id should change nothing)", len(got))
	}
}

func TestClampIndex(t *testing.T) {
	cases := map[int32]int32{-10: 0, 0: 0, 2: 2, 3: 3, 99: 3}
	for in, want := range cases {
		if got := ClampIndex(in, 3); got != want {
			t.Errorf("ClampIndex(%d, 3) = %d, want %d", in, got, want)
		}
	}
}

func TestOrdinals(t *testing.T) {
	got := Ordinals(4)
	want := []int32{0, 1, 2, 3}

	if len(got) != len(want) {
		t.Fatalf("length = %d, want %d", len(got), len(want))
	}
	for i := range got {
		if got[i] != want[i] {
			t.Fatalf("index %d = %d, want %d", i, got[i], want[i])
		}
	}
}

// IsGapFree is the invariant the smoke tests assert after every move, so it
// needs to actually reject the shapes a broken reindex would produce.
func TestIsGapFree(t *testing.T) {
	tests := []struct {
		name string
		in   []int32
		want bool
	}{
		{"empty", nil, true},
		{"contiguous", []int32{0, 1, 2}, true},
		{"gap", []int32{0, 1, 3}, false},
		{"duplicate", []int32{0, 1, 1}, false},
		{"starts above zero", []int32{1, 2, 3}, false},
		{"negative", []int32{-1, 0, 1}, false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := IsGapFree(tc.in); got != tc.want {
				t.Errorf("IsGapFree(%v) = %v, want %v", tc.in, got, tc.want)
			}
		})
	}
}
