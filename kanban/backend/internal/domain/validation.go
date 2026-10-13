package domain

import (
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

// dateLayout is the wire format for cards.due_date.
//
// A due date is a calendar day, not an instant. Postgres hands us a pgtype.Date
// at UTC midnight, so formatting in UTC yields exactly the day that was stored
// for every timezone. The frontend formats locally for display and must never
// round-trip this through a timestamp.
const dateLayout = "2006-01-02"

// DefaultColumnNames are the stages a newly created board starts with.
var DefaultColumnNames = []string{"To Do", "In Progress", "Review", "Done"}

// Priorities is the closed set the database CHECK constraint enforces. It is
// repeated here so the API can answer 400 with a useful message instead of
// leaking a Postgres constraint violation as a 500.
var Priorities = []string{"low", "medium", "high", "urgent"}

// DefaultPriority applies when a client omits the field.
const DefaultPriority = "medium"

// MaxNameLength bounds board and column names.
const MaxNameLength = 120

// RequireName trims and validates a human-supplied name.
func RequireName(raw, field string) (string, error) {
	name := strings.TrimSpace(raw)
	if name == "" {
		return "", Invalid("%s must not be empty", field)
	}
	if len(name) > MaxNameLength {
		return "", Invalid("%s must be %d characters or fewer", field, MaxNameLength)
	}
	return name, nil
}

// NormaliseText turns nil and whitespace-only text into an empty string, so
// the database never stores "   " as a meaningful value.
func NormaliseText(s *string) *string {
	empty := ""
	if s == nil {
		return &empty
	}
	trimmed := strings.TrimSpace(*s)
	if trimmed == "" {
		return &empty
	}
	return &trimmed
}

// RequirePriority validates a priority, defaulting when empty.
func RequirePriority(raw string) (string, error) {
	if raw == "" {
		return DefaultPriority, nil
	}
	for _, p := range Priorities {
		if p == raw {
			return p, nil
		}
	}
	return "", Invalid("priority must be one of low, medium, high, urgent")
}

// DueDateIn converts a nullable "YYYY-MM-DD" into a pgtype.Date. An absent or
// empty value means "no due date".
func DueDateIn(s *string) (pgtype.Date, error) {
	if s == nil || *s == "" {
		return pgtype.Date{}, nil
	}
	t, err := time.Parse(dateLayout, *s)
	if err != nil {
		return pgtype.Date{}, Invalid("due_date must be formatted as YYYY-MM-DD")
	}
	return pgtype.Date{Time: t, Valid: true}, nil
}

// DueDateOut converts a pgtype.Date into a nullable "YYYY-MM-DD".
func DueDateOut(d pgtype.Date) *string {
	if !d.Valid {
		return nil
	}
	s := d.Time.UTC().Format(dateLayout)
	return &s
}
