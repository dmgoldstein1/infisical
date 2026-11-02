package sources

import (
	"math"
	"strconv"
	"testing"
)

// CodeQL go/incorrect-integer-conversion: resource numbers flow into
// githubv4.Int (int32). Values outside (0, MaxInt32] must be rejected instead
// of wrapping on 64-bit platforms.
func TestParseResourceNumber(t *testing.T) {
	valid, err := parseResourceNumber("issue", "42")
	if err != nil || valid != 42 {
		t.Fatalf("expected 42, got %d, err=%v", valid, err)
	}
	for _, id := range []string{
		"0", "-1", "abc", "",
		strconv.FormatInt(int64(math.MaxInt32)+1, 10),
		"99999999999999999999",
	} {
		if _, err := parseResourceNumber("issue", id); err == nil {
			t.Errorf("expected error for %q, got nil", id)
		}
	}
	max, err := parseResourceNumber("pr", strconv.Itoa(math.MaxInt32))
	if err != nil || max != math.MaxInt32 {
		t.Fatalf("expected MaxInt32, got %d, err=%v", max, err)
	}
}
