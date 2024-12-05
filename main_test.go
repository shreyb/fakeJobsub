package main

import (
	"errors"
	"strings"
	"testing"
)

func TestRun(t *testing.T) {
	var args []string

	t.Run("Test 0:  Nothing given at all", func(t *testing.T) {
		args = []string{}
		if err := run(args); !errors.Is(err, errUsage) {
			t.Errorf("Should have gotten errUsage.  Got %v instead", err)
		}
	},
	)

	t.Run("Test 1:  No subcommand", func(t *testing.T) {
		args = []string{"fakeJobsub"}
		if err := run(args); !errors.Is(err, errUsage) {
			t.Errorf("Should have gotten errUsage.  Got %v instead", err)
		}
	},
	)

	t.Run("Test 2: Invalid subcommand", func(t *testing.T) {
		args = []string{"fakeJobsub", "badcommand"}
		if err := run(args); err.Error() != "invalid subcommand" {
			t.Errorf("Should have gotten an error indicating \"invalid subcommand\".  Got %s", err)
		}
	},
	)

	t.Run("Test 3: Bad args: submit", func(t *testing.T) {
		args = []string{"fakeJobsub", "submit", "--group", "fermilab", "--oopsbadarg"}
		if err := run(args); !errors.Is(err, errParseFlags) {
			t.Errorf("Should have gotten errParseFlags.  Got %v instead", err)
		}
	},
	)

	t.Run("Test 4: Bad args: list", func(t *testing.T) {
		args = []string{"fakeJobsub", "list", "--oopsbadarg"}
		if err := run(args); !errors.Is(err, errParseFlags) {
			t.Errorf("Should have gotten errParseFlags.  Got %v instead", err)
		}
	},
	)

	t.Run("Test 5: submit with no --group", func(t *testing.T) {
		args = []string{"fakeJobsub", "submit", "--num", "5"}
		if err := run(args); err.Error() != "--group must be specified" {
			t.Errorf("Should have gotten error indicating that group should be specified.  Got %v instead", err)
		}
	},
	)

	t.Run("Test 6:  Valid submit, no num", func(t *testing.T) {
		args = []string{"fakeJobsub", "submit", "--group", "fermilab"}
		if err := run(args); err != nil {
			t.Errorf("Should have gotten nil error. Got %v instead", err)
		}
	},
	)

	t.Run("Test 7: Valid submit, num", func(t *testing.T) {
		args = []string{"fakeJobsub", "submit", "--group", "fermilab", "--num", "5"}
		if err := run(args); err != nil {
			t.Errorf("Should have gotten nil error. Got %v instead", err)
		}
	},
	)

	t.Run("Test 7: Valid list, no keys", func(t *testing.T) {
		args = []string{"fakeJobsub", "list"}
		if err := run(args); err != nil {
			t.Errorf("Should have gotten nil error. Got %v instead", err)
		}
	},
	)

	// Note:  for tests like these where we're dereferencing the error, we should really FIRST do a check to see if the error is nil, and fail at that point.  I didn't want to complicate the tests though.
	t.Run("Test 8: Valid list, bad keys", func(t *testing.T) {
		args = []string{"fakeJobsub", "list", "--keys", "foo,bar,baz"}
		if err := run(args); !strings.Contains(err.Error(), "invalid column: foo") {
			t.Errorf("Should have gotten non-nil error. Got %v instead", err)
		}
	},
	)

	t.Run("Test 9: Valid list, good keys", func(t *testing.T) {
		args = []string{"fakeJobsub", "list", "--keys", "clusterid, group"}
		if err := run(args); err != nil {
			t.Errorf("Should have gotten non-nil error. Got %v instead", err)
		}
	},
	)

	t.Run("Test 10: submit to a specific schedd", func(t *testing.T) {
		args = []string{"fakeJobsub", "submit", "--group", "fermilab", "--schedd", "schedd1"}
		if err := run(args); err != nil {
			t.Errorf("Should have gotten nil error. Got %v instead", err)
		}
	},
	)

	t.Run("Test 11: submit to a specific invalid schedd", func(t *testing.T) {
		args = []string{"fakeJobsub", "submit", "--group", "fermilab", "--schedd", "schedd42"}
		if err := run(args); err != nil {
			t.Errorf("Should have gotten nil error. Got %v instead", err)
		}
	},
	)

	t.Run("Test 12: list from a specific invalid schedd", func(t *testing.T) {
		args = []string{"fakeJobsub", "list", "--schedd", "schedd42"}
		if err := run(args); !strings.Contains(err.Error(), "invalid schedd") {
			t.Errorf("Should have gotten error indicating that the schedd was invalid. Got %v instead", err)
		}
	},
	)

	t.Run("Test 13: list from a specific valid schedd", func(t *testing.T) {
		args = []string{"fakeJobsub", "list", "--schedd", "schedd1"}
		if err := run(args); err != nil {
			t.Errorf("Should have gotten nil error. Got %v instead", err)
		}
	},
	)

	t.Run("Test 14: list for clusterid, but no valid schedd", func(t *testing.T) {
		args = []string{"fakeJobsub", "list", "--clusterid", "1"}
		if err := run(args); !strings.Contains(err.Error(), "must set --schedd flag") {
			t.Errorf("Should have gotten error indicating that --schedd flag needs to be set. Got %v instead", err)
		}
	},
	)
}
