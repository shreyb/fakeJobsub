package condor

import (
	"fmt"
	"os"
	"path"
	"time"

	"fakeJobsub/db"
)

// DefaultSchedd is the default schedd whose backend is in the default db location
var DefaultSchedd *Schedd

// Schedd is a condor Schedd
type Schedd struct {
	Name string
	db   scheddDB
}

func init() {
	DefaultSchedd = &Schedd{Name: "DefaultSchedd"}
	d, err := db.CreateOrOpenDB(DefaultSchedd.getFilename())
	if err != nil {
		panic(err)
	}
	DefaultSchedd.db = d
}

// GetSchedd opens the underlying db.FakeJobsubDB for further operations
func GetSchedd(name string) (*Schedd, error) {
	if name == "" {
		return DefaultSchedd, nil
	}

	s := &Schedd{Name: name}
	s.Name = name

	d, err := db.CreateOrOpenDB(s.getFilename())
	if err != nil {
		return nil, err
	}

	s.db = d
	return s, nil
}

// Submit submits a certain number of jobs based on the config
func (s *Schedd) Submit(group string, numJobs int) error {
	cid, err := s.db.GetNextClusterID()
	if err != nil {
		return fmt.Errorf("could not submit job: %w", err)
	}

	if err = s.db.InsertJobIntoDB(cid, group, numJobs); err != nil {
		return fmt.Errorf("could not submit job: %w", err)
	}

	// Fake some CPU-intensive activity
	fmt.Printf("Submitting....\n\n")
	time.Sleep(3 * time.Second)

	fmt.Printf("Submitted %d jobs to cluster %d for group %s\n", numJobs, cid, group)

	return nil
}

// List prints a list of the jobs in the queue.  If keys are given, it will only return the values for those keys.  It only allows filtering based on clusterID for simplicity in this demo
func (s *Schedd) List(clusterID int, keys ...string) ([]string, error) {
	rows, err := s.db.RetrieveJobsFromDB(clusterID, keys...)
	if err != nil {
		return nil, fmt.Errorf("could not list jobs: %w", err)
	}

	// Mock some processing time
	time.Sleep(2 * time.Second)

	return rows, nil
}

func (s *Schedd) getFilename() string {
	return path.Join(os.TempDir(), fmt.Sprintf("fakeJobsubSchedd_%s", s.Name))
}

// scheddDB contains the methods needed to interact with a jobs database for job submission and jobs listing purposes
type scheddDB interface {
	InsertJobIntoDB(int, string, int) error
	RetrieveJobsFromDB(int, ...string) ([]string, error)
	GetNextClusterID() (int, error)
}
