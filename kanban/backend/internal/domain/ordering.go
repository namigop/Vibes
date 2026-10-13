package domain

import (
	"github.com/google/uuid"
)

// Ordering invariant shared by every slice that reorders things.
//
// `position` is a gap-free 0..n-1 ordinal within its parent — cards inside a
// column, columns inside a board. A move reindexes both the source and the
// destination, so a stale client can never corrupt the order and there is no
// fractional-position drift to clean up.
//
// The pure helpers below are used both to compute the new ordering and to
// validate it in tests.

// ClampIndex keeps a destination index inside [0, n] so a client that is a
// little out of date — or malicious — cannot corrupt the ordering. It is
// clamped rather than rejected so a stale drag does not fail.
func ClampIndex(index int32, n int32) int32 {
	if index < 0 {
		return 0
	}
	if index > n {
		return n
	}
	return index
}

// RemoveID returns ids without the given id. The input is not mutated.
func RemoveID(ids []uuid.UUID, id uuid.UUID) []uuid.UUID {
	out := make([]uuid.UUID, 0, len(ids))
	for _, existing := range ids {
		if existing != id {
			out = append(out, existing)
		}
	}
	return out
}

// InsertAt returns ids with id inserted at index, appending when the index is
// at or past the end.
func InsertAt(ids []uuid.UUID, index int32, id uuid.UUID) []uuid.UUID {
	index = ClampIndex(index, int32(len(ids)))
	out := make([]uuid.UUID, 0, len(ids)+1)
	out = append(out, ids[:index]...)
	out = append(out, id)
	out = append(out, ids[index:]...)
	return out
}

// Ordinals returns the positions 0..n-1 as int32s, ready to be written back.
func Ordinals(n int) []int32 {
	out := make([]int32, 0, n)
	for i := 0; i < n; i++ {
		out = append(out, int32(i))
	}
	return out
}

// IsGapFree reports whether positions are exactly 0..n-1 with no gaps and no
// duplicates. It is the invariant the smoke tests assert after every move.
func IsGapFree(positions []int32) bool {
	if len(positions) == 0 {
		return true
	}
	seen := make(map[int32]bool, len(positions))
	for _, p := range positions {
		if p < 0 || seen[p] {
			return false
		}
		seen[p] = true
	}
	for i := int32(0); i < int32(len(positions)); i++ {
		if !seen[i] {
			return false
		}
	}
	return true
}
