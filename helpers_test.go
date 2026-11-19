package main

import "testing"

func TestEnvFileValue(t *testing.T) {
	cases := map[string]string{
		"plain":                    "plain",
		"say \"hi\"":               "say \\\"hi\\\"",
		"line 1\nline 2":           "line 1\\nline 2",
		"windows\r\nline\r\n":      "windows\\nline\\n",
		"## RATIONALE\n\n- item\n": "## RATIONALE\\n\\n- item\\n",
	}
	for in, want := range cases {
		if got := envFileValue(in); got != want {
			t.Errorf("envFileValue(%q) = %q, want %q", in, got, want)
		}
	}
}
