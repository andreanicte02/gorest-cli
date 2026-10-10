package main

import (
	"bytes"
	"strings"
	"testing"
)

func TestRunValidation(t *testing.T) {
	t.Setenv("GOREST_TOKEN", "")
	tests := []struct {
		args    []string
		wantErr string
	}{
		{[]string{"get"}, "missing user id"},
		{[]string{"get", "abc"}, `invalid id "abc"`},
		{[]string{"create", "--name", "Ada"}, "a token is required"},
		{[]string{"delete", "5"}, "a token is required"},
		{[]string{"frobnicate"}, "unknown command"},
	}
	for _, tt := range tests {
		t.Run(strings.Join(tt.args, " "), func(t *testing.T) {
			err := run(tt.args, &bytes.Buffer{})
			if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("err = %v, want it to contain %q", err, tt.wantErr)
			}
		})
	}
}

func TestRunHelp(t *testing.T) {
	var out bytes.Buffer
	if err := run(nil, &out); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "Usage:") {
		t.Fatalf("help output missing usage:\n%s", out.String())
	}
}
