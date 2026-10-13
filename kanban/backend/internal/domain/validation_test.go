package domain

import "testing"

func ptr(s string) *string { return &s }

// The card slices delegate every field check to RequireName, RequirePriority
// and NormaliseText, so these tests are what stands between a blank title and
// a 500 from the Postgres CHECK constraint.

func TestRequireName(t *testing.T) {
	t.Run("rejects blank title", func(t *testing.T) {
		if _, err := RequireName("   ", "title"); err == nil {
			t.Fatal("expected an error for a whitespace-only title")
		}
	})

	t.Run("trims surrounding whitespace", func(t *testing.T) {
		got, err := RequireName("  hello  ", "title")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != "hello" {
			t.Errorf("title = %q, want %q (should be trimmed)", got, "hello")
		}
	})

	t.Run("rejects an over-long name", func(t *testing.T) {
		tooLong := make([]byte, MaxNameLength+1)
		for i := range tooLong {
			tooLong[i] = 'a'
		}
		if _, err := RequireName(string(tooLong), "title"); err == nil {
			t.Fatal("expected an error for a name past the length limit")
		}
	})
}

func TestRequirePriority(t *testing.T) {
	t.Run("rejects unknown priority", func(t *testing.T) {
		if _, err := RequirePriority("nuclear"); err == nil {
			t.Fatal("expected an error for an unknown priority")
		}
	})

	t.Run("defaults when omitted", func(t *testing.T) {
		got, err := RequirePriority("")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != DefaultPriority {
			t.Errorf("priority = %q, want %q", got, DefaultPriority)
		}
	})

	t.Run("accepts every allowed priority", func(t *testing.T) {
		for _, p := range Priorities {
			if _, err := RequirePriority(p); err != nil {
				t.Errorf("RequirePriority(%q) returned %v", p, err)
			}
		}
	})
}

func TestNormaliseText(t *testing.T) {
	t.Run("nil optional text normalises to empty", func(t *testing.T) {
		got := NormaliseText(nil)
		if got == nil || *got != "" {
			t.Errorf("NormaliseText(nil) = %v, want empty string", got)
		}
	})

	t.Run("whitespace normalises to empty", func(t *testing.T) {
		blank := "   "
		for _, got := range []*string{NormaliseText(&blank), NormaliseText(ptr(""))} {
			if *got != "" {
				t.Errorf("whitespace should normalise to empty, got %q", *got)
			}
		}
	})

	t.Run("real text survives", func(t *testing.T) {
		if got := NormaliseText(ptr(" Priya ")); *got != "Priya" {
			t.Errorf("NormaliseText = %q, want %q", *got, "Priya")
		}
	})
}

func TestDueDateRoundTrip(t *testing.T) {
	t.Run("preserves a calendar date exactly", func(t *testing.T) {
		in := "2026-10-15"
		stored, err := DueDateIn(&in)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		out := DueDateOut(stored)
		if out == nil || *out != in {
			t.Fatalf("round trip = %v, want %q", out, in)
		}
	})

	t.Run("rejects a malformed date", func(t *testing.T) {
		bad := "15-10-2026"
		if _, err := DueDateIn(&bad); err == nil {
			t.Fatal("expected an error for a non-ISO date")
		}
	})

	t.Run("nil and empty both mean no due date", func(t *testing.T) {
		for _, in := range []*string{nil, ptr("")} {
			stored, err := DueDateIn(in)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if out := DueDateOut(stored); out != nil {
				t.Errorf("expected null, got %q", *out)
			}
		}
	})
}
