package condor

import (
	"os"
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

func TestSubmit(t *testing.T) {}

func TestList(t *testing.T) {}

func TestGetFilename(t *testing.T) {
	temp := t.TempDir()
	s := Schedd{Name: "example"}
	expected := temp + "/" + "fakeJobsubSchedd_example"

	fn := s.getFilename(temp)
	if fn != expected {
		t.Errorf("Got wrong filename.  Wanted %s, got %s", expected, fn)
	}
}
