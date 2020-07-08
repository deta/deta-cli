package cmd

import (
	"errors"
	"testing"
)

func TestGetCronTypeFromExpr(t *testing.T) {
	testCases := []struct {
		expr     string
		cronType string
		err      error
	}{
		{"1 minute", "rate", nil},
		{"2 hours", "rate", nil},
		{"0 10 * * ? *", "cron", nil},
		{"0/15 * * * ? *", "cron", nil},
		{"0/5 8-17 ? * MON-FRI *", "cron", nil},
		{"akdlkf", "", errInvalidExp},
		{"0 10 * * ?", "", errInvalidExp},
		{"a b c d e f g h", "", errInvalidExp},
	}

	for _, tc := range testCases {
		cronType, err := getCronTypeFromExpr(tc.expr)
		if cronType != tc.cronType {
			t.Errorf("got unexpected cron type: expected %s got %s for expression %s", cronType, tc.cronType, tc.expr)
		}
		if !errors.Is(err, tc.err) {
			t.Errorf("got unexpected error: expected %s got %s for expression %s", err, tc.err, tc.expr)
		}
	}
}
