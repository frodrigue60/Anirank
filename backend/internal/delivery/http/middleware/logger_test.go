package middleware

import "testing"

func TestShouldLogRequest(t *testing.T) {
	cases := []struct {
		status int
		level  string
		want   bool
	}{
		{200, "info", true},
		{404, "info", true},
		{500, "info", true},
		{200, "warn", false},
		{404, "warn", true},
		{500, "warn", true},
		{200, "error", false},
		{404, "error", false},
		{499, "error", false},
		{500, "error", true},
		{503, "error", true},
		{200, "", false}, // default = error
	}

	for _, tc := range cases {
		got := shouldLogRequest(tc.status, tc.level)
		if got != tc.want {
			t.Errorf("shouldLogRequest(%d, %q) = %v, want %v", tc.status, tc.level, got, tc.want)
		}
	}
}
