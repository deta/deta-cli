package cmd

import (
	"reflect"
	"testing"
)

func TestParseArgs(t *testing.T) {
	testCases := []struct {
		args     []string
		action   string
		progArgs map[string]interface{}
	}{
		{
			[]string{"default", "--name", "jimmy", "--age", "33", "-active"},
			"default",
			map[string]interface{}{
				"name":   "jimmy",
				"age":    "33",
				"active": true,
			},
		},
		{
			[]string{"------name", "jimmy", "", "--address", "some street", "--dank", "--age"},
			"",
			map[string]interface{}{
				"name":    "jimmy",
				"age":     "",
				"dank":    "",
				"address": "some street",
			},
		},
		{
			[]string{"--name", "jimmy", "--", "--name", "jonah"},
			"",
			map[string]interface{}{
				"name": []string{"jimmy", "jonah"},
			},
		},
	}

	for _, tc := range testCases {
		action, progArgs := parseArgs(tc.args)
		if action != tc.action {
			t.Errorf("Error in parsing action, got: '%s' expected: '%s' args:%v", action, tc.action, tc.args)
		}

		if !reflect.DeepEqual(progArgs, tc.progArgs) {
			t.Errorf("Error in parsing prog args, got: %v expected: %v args:%v", progArgs, tc.progArgs, tc.args)
		}
	}
}
