package db

import "testing"

func TestPrepareAnyRowAndPointerSlice(t *testing.T) {
	l := 5
	s1, s2 := prepareAnyRowAndPointerSlice(l)

	if len(s1) != len(s2) {
		t.Errorf("Expected slices to have same length.  len(s1) = %d, len(s2) = %d", len(s1), len(s2))
	}

	for idx := range s1 {
		if &s1[idx] != s2[idx] {
			t.Errorf("s2 value should be address of corresponding s1 value. &s1 = %p, &s2 = %p", &s1, s2)
		}
	}

}

// There should be other tests to ensure that the database is opened or created properly, that the various db-changing/retrieving methods work correctly, etc.
