package run

import "testing"

func TestRedactLine(t *testing.T) {
	cases := []struct {
		in, want string
	}{
		{"API_KEY=supersecret", "API_KEY=***"},
		{"token: mytoken123", "token: ***"},
		{"Bearer ghp_abc123xyz", "***"},
		{"plain log line", "plain log line"},
	}
	for _, tc := range cases {
		got := RedactLine(tc.in)
		if got != tc.want {
			t.Errorf("RedactLine(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}
