package api

import (
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
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
			got := insertAt(tc.in, tc.idx, moved)
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

func TestInsertAtDoesNotMutateInput(t *testing.T) {
	base := ids(3)
	original := make([]uuid.UUID, len(base))
	copy(original, base)

	_ = insertAt(base, 1, uuid.New())

	for i := range base {
		if base[i] != original[i] {
			t.Fatalf("input slice was mutated at index %d", i)
		}
	}
}

func TestRemoveID(t *testing.T) {
	base := ids(3)
	got := removeID(base, base[1])

	if len(got) != 2 {
		t.Fatalf("length = %d, want 2", len(got))
	}
	if got[0] != base[0] || got[1] != base[2] {
		t.Fatalf("removeID kept the wrong elements: %v", got)
	}
}

func TestRemoveIDMissingIsNoOp(t *testing.T) {
	base := ids(2)
	got := removeID(base, uuid.New())
	if len(got) != 2 {
		t.Fatalf("length = %d, want 2 (absent id should change nothing)", len(got))
	}
}

func TestClampIndex(t *testing.T) {
	cases := map[int32]int32{-10: 0, 0: 0, 2: 2, 3: 3, 99: 3}
	for in, want := range cases {
		if got := clampIndex(in, 3); got != want {
			t.Errorf("clampIndex(%d, 3) = %d, want %d", in, got, want)
		}
	}
}

func TestValidateCardFields(t *testing.T) {
	t.Run("rejects blank title", func(t *testing.T) {
		if _, err := validateCardFields(CardRequest{Title: "   "}); err == nil {
			t.Fatal("expected an error for a whitespace-only title")
		}
	})

	t.Run("rejects unknown priority", func(t *testing.T) {
		_, err := validateCardFields(CardRequest{Title: "ok", Priority: "nuclear"})
		if err == nil {
			t.Fatal("expected an error for an unknown priority")
		}
	})

	t.Run("defaults priority and normalises blank text", func(t *testing.T) {
		blank := "   "
		f, err := validateCardFields(CardRequest{Title: "  hello  ", Description: &blank, Assignee: &blank})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if f.title != "hello" {
			t.Errorf("title = %q, want %q (should be trimmed)", f.title, "hello")
		}
		if f.priority != defaultPriority {
			t.Errorf("priority = %q, want %q", f.priority, defaultPriority)
		}
		if *f.description != "" || *f.assignee != "" {
			t.Errorf("whitespace should normalise to empty, got %q / %q", *f.description, *f.assignee)
		}
	})

	t.Run("nil optional text normalises to empty", func(t *testing.T) {
		f, err := validateCardFields(CardRequest{Title: "ok"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if f.description == nil || *f.description != "" {
			t.Errorf("description = %v, want empty string", f.description)
		}
	})
}

func TestDueDateRoundTrip(t *testing.T) {
	t.Run("preserves a calendar date exactly", func(t *testing.T) {
		in := "2026-10-15"
		stored, err := dueDateIn(&in)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		out := dueDateOut(stored)
		if out == nil || *out != in {
			t.Fatalf("round trip = %v, want %q", out, in)
		}
	})

	t.Run("rejects a malformed date", func(t *testing.T) {
		bad := "15-10-2026"
		if _, err := dueDateIn(&bad); err == nil {
			t.Fatal("expected an error for a non-ISO date")
		}
	})

	t.Run("nil and empty both mean no due date", func(t *testing.T) {
		for _, in := range []*string{nil, ptr("")} {
			stored, err := dueDateIn(in)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if out := dueDateOut(stored); out != nil {
				t.Errorf("expected null, got %q", *out)
			}
		}
	})
}

func ptr(s string) *string { return &s }

func TestMapNotFound(t *testing.T) {
	t.Run("pgx.ErrNoRows becomes a 404", func(t *testing.T) {
		err := mapNotFound(pgx.ErrNoRows, "board")
		var nErr NotFoundError
		if !errors.As(err, &nErr) {
			t.Fatalf("expected NotFoundError, got %T (%v)", err, err)
		}
		if nErr.Message != "board not found" {
			t.Errorf("message = %q, want %q", nErr.Message, "board not found")
		}
	})

	t.Run("other errors pass through untouched", func(t *testing.T) {
		sentinel := errors.New("connection reset")
		if got := mapNotFound(sentinel, "board"); !errors.Is(got, sentinel) {
			t.Fatalf("expected the original error, got %v", got)
		}
	})
}
