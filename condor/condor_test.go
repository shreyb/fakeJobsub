package condor

import (
	"fakeJobsub/db"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

// Test where we start with a new DB file, submit, retrieve list

func TestInit(t *testing.T) {
	if DefaultSchedd == nil {
		t.Error("DefaultSchedd should be non-nil")
	}
	if DefaultSchedd.Name != "DefaultSchedd" {
		t.Errorf("DefaultSchedd name should be 'DefaultSchedd'.  Got %s", DefaultSchedd.Name)
	}
	if DefaultSchedd.db == nil {
		t.Error("Failed to intitialize DefaultSchedd database")
	}

}

func TestGetSchedd(t *testing.T) {
	t.Run("default", func(t *testing.T) {
		name := ""
		s, err := GetSchedd(name)
		if err != nil {
			t.Errorf("Should have gotten nil error.  Got %v instead", err)
		}
		if s != DefaultSchedd {
			t.Errorf("Should have gotten DefaultSchedd.  Got %v instead", s)
		}
	})

	t.Run("named", func(t *testing.T) {
		name := "test0"
		s, err := GetSchedd(name)
		if err != nil {
			t.Errorf("Should have gotten nil error.  Got %v instead", err)
		}
		if s.Name != name {
			t.Errorf("Schedd.Name should have been %s.  Got %s instead", name, s.Name)
		}
		if s.db == nil {
			t.Error("Should not have gotten nil db for test schedd")
		}
	})

	t.Run("DB error", func(t *testing.T) {
		name := "test1"
		t.Setenv("TMPDIR", os.DevNull)
		_, err := GetSchedd(name)
		if err == nil || !strings.Contains(err.Error(), "stat database") {
			t.Errorf("Should have gotten error indicating that database could not be opened.  Got %v instead", err)
		}
	})
}

func TestSubmit(t *testing.T) {
	name := "name"
	group := "testgroup"
	numJobs := 42

	// Setup DB and schedd
	s := &Schedd{Name: name}
	s.Name = name
	d, err := db.CreateOrOpenDB(s.getFilename(t.TempDir()))
	if err != nil {
		t.Errorf("Could not create test db: %s", err.Error())
	}
	s.db = d

	if err := s.Submit(group, numJobs); err != nil {
		t.Errorf("Failed to submit test jobs: %s", err.Error())
	}

	expectedHeader := "clusterid\tgroup\tnum"
	expectedRow := fmt.Sprintf("1\t%s\t%d", group, numJobs)
	expectedResult := []string{expectedHeader, expectedRow}
	rows, err := s.db.RetrieveJobsFromDB(1)
	if err != nil {
		t.Errorf("Should have gotten nil error.  Got %v instead", err)
	}

	if !slices.Equal(rows, expectedResult) {
		t.Errorf("Got wrong result.  Expected %v, got %v", expectedResult, rows)
	}

}

func TestList(t *testing.T) {
	// Setup DB
	name := "test1"
	s := &Schedd{Name: name}
	s.Name = name
	d, err := db.CreateOrOpenDB(s.getFilename(t.TempDir()))
	if err != nil {
		t.Errorf("Could not create test db: %s", err.Error())
	}
	s.db = d

	if err := s.db.InsertJobIntoDB(42, "testgroup", 17); err != nil {
		t.Errorf("Could not create row in test db: %s", err.Error())
	}
	if err := s.db.InsertJobIntoDB(43, "testgroup", 17); err != nil {
		t.Errorf("Could not create row in test db: %s", err.Error())
	}

	// Now retrieve the value but only some columns, and one of the clusterids
	t.Run("Valid result", func(t *testing.T) {
		expectedHeader := ("clusterid\tgroup")
		expectedRow := ("42\ttestgroup")
		expectedResult := []string{expectedHeader, expectedRow}
		result, err := s.List(42, "clusterid", "group")
		if err != nil {
			t.Errorf("Should have gotten nil error.  Got %v instead", err)
		}
		if !slices.Equal(expectedResult, result) {
			t.Errorf("Got wrong result.  Expected %v, got %v", expectedResult, result)
		}
	})

	// Try to get an invalid row
	t.Run("Invalid result", func(t *testing.T) {
		_, err = s.List(22)
		if err == nil || !strings.Contains(err.Error(), "could not list jobs") {
			t.Errorf("Got unexpected error. Expected error that indicated that jobs could not be listed; got %v", err)
		}
	})

}

func TestSubmitAndList(t *testing.T) {
	// Submit a single job to a particular schedd, then list it and make sure we get the right thing
	// Setup DB
	name := "test1"
	s := &Schedd{Name: name}
	s.Name = name
	d, err := db.CreateOrOpenDB(s.getFilename(t.TempDir()))
	if err != nil {
		t.Errorf("Could not create test db: %s", err.Error())
	}
	s.db = d

	// Submit a job
	group := "testgroup"
	numJobs := 42
	if err := s.Submit(group, numJobs); err != nil {
		t.Errorf("Failed to submit test jobs: %s", err.Error())
	}

	// List that cluster
	t.Run("Valid result", func(t *testing.T) {
		expectedHeader := ("clusterid\tgroup\tnum")
		expectedRow := ("1\ttestgroup\t42")
		expectedResult := []string{expectedHeader, expectedRow}
		result, err := s.List(1, "clusterid", "group", "num")
		if err != nil {
			t.Errorf("Should have gotten nil error.  Got %v instead", err)
		}
		if !slices.Equal(expectedResult, result) {
			t.Errorf("Got wrong result.  Expected %v, got %v", expectedResult, result)
		}
	})
}

func TestGetFilename(t *testing.T) {
	temp := t.TempDir()
	s := Schedd{Name: "example"}
	expected := filepath.Join(temp, "fakeJobsubSchedd_example.db")

	fn := s.getFilename(temp)
	if fn != expected {
		t.Errorf("Got wrong filename.  Wanted %s, got %s", expected, fn)
	}
}
