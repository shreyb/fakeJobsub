package main

import (
	"errors"
	"strings"
	"testing"
)

func TestRun(t *testing.T) {
	var args []string

	// Test 0:  Nothing given at all
	args = []string{}
	if err := run(args); !errors.Is(err, errUsage) {
		t.Errorf("Should have gotten errUsage.  Got %v instead", err)
	}

	// Test 1:  No subcommand
	args = []string{"fakeJobsub"}
	if err := run(args); !errors.Is(err, errUsage) {
		t.Errorf("Should have gotten errUsage.  Got %v instead", err)
	}

	// Test 2: Invalid subcommand
	args = []string{"fakeJobsub", "badcommand"}
	if err := run(args); err.Error() != "invalid subcommand" {
		t.Errorf("Should have gotten an error indicating \"invalid subcommand\".  Got %s", err)
	}

	// Test 3: Bad args: submit
	args = []string{"fakeJobsub", "submit", "--group", "fermilab", "--oopsbadarg"}
	if err := run(args); !errors.Is(err, errParseFlags) {
		t.Errorf("Should have gotten errParseFlags.  Got %v instead", err)
	}

	// Test 4: Bad args: list
	args = []string{"fakeJobsub", "list", "--oopsbadarg"}
	if err := run(args); !errors.Is(err, errParseFlags) {
		t.Errorf("Should have gotten errParseFlags.  Got %v instead", err)
	}

	// Test 5: submit with no --group
	args = []string{"fakeJobsub", "submit", "--num", "5"}
	if err := run(args); err.Error() != "--group must be specified" {
		t.Errorf("Should have gotten error indicating that group should be specified.  Got %v instead", err)
	}

	// Test 6:  Valid submit, no num
	args = []string{"fakeJobsub", "submit", "--group", "fermilab"}
	if err := run(args); err != nil {
		t.Errorf("Should have gotten nil error. Got %v instead", err)
	}

	// Test 7: Valid submit, num
	args = []string{"fakeJobsub", "submit", "--group", "fermilab", "--num", "5"}
	if err := run(args); err != nil {
		t.Errorf("Should have gotten nil error. Got %v instead", err)
	}

	// Test 7: Valid list, no keys
	args = []string{"fakeJobsub", "list"}
	if err := run(args); err != nil {
		t.Errorf("Should have gotten nil error. Got %v instead", err)
	}

	// Test 8: Valid list, bad keys
	args = []string{"fakeJobsub", "list", "--keys", "foo,bar,baz"}
	if err := run(args); !strings.Contains(err.Error(), "invalid column: foo") {
		t.Errorf("Should have gotten non-nil error. Got %v instead", err)
	}

	// Test 9: Valid list, good keys
	args = []string{"fakeJobsub", "list", "--keys", "clusterid, group"}
	if err := run(args); err != nil {
		t.Errorf("Should have gotten non-nil error. Got %v instead", err)
	}

}
