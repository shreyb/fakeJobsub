package condor

// Submit submits a certain number of jobs based on the config
func (s *Schedd) Submit(numJobs int) {}

// List returns a list of the jobs in the queue.  If keys are given, it will only return the values for those keys
func (s *Schedd) List(keys ...string) {}

// Schedd is a condor Schedd
type Schedd struct {
}
