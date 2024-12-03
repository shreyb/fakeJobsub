package main

import "testing"

func TestCheckSubmitForGroup(t *testing.T) {
	if err := checkSubmitForGroup(""); err == nil {
		t.Error("Should have gotten non-nil error for checkSubmitForGroup when no group given")
	}
	if err := checkSubmitForGroup("fermilab"); err != nil {
		t.Error("Should have gotten non-nil error for checkSubmitForGroup when no group given")
	}
}
