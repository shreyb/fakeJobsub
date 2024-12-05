package main

import (
	"errors"
	"fmt"
	"sync"

	"fakeJobsub/condor"
)

func checkSubmitForGroup(group string) error {
	if group == "" {
		return errors.New("failed submit group")
	}
	return nil
}

// listJobsFromSchedds concurrently queries all elements in schedds and returns
// their rows in the order given by schedds.  If there is an error querying one or
// more of the schedds, a non-nil error is returned indicating which schedds
// had errors, and what those errors were
func listJobsFromSchedds(schedds []*condor.Schedd, keys ...string) ([]string, error) {
	// Where all our rows will get stored by schedd
	scheddMap := make(map[string][]string, 0)
	for _, schedd := range schedds {
		// Initialize the slices that are the values in this map
		scheddMap[schedd.Name] = make([]string, 0)
	}

	// Listener for aggregator chan that collects all the rows.  Note that this
	// is simply an example to demonstrate channels.  In reality, this would
	// more clearly/easily be accomplished with a mutex, similar to errorList
	// below
	type entryForAgg struct {
		scheddName string
		row        string
	}
	aggregator := make(chan entryForAgg, len(schedds)) // Second argument is the buffer size of the channel
	aggDone := make(chan bool)                         // Channel to close when aggregation is done
	go func() {
		for entry := range aggregator {
			scheddMap[entry.scheddName] = append(scheddMap[entry.scheddName], entry.row)
		}
		close(aggDone)
	}()

	// Compile all the errors into errList
	type errorList struct {
		errs []error
		mux  sync.Mutex
	}
	errList := errorList{
		errs: make([]error, 0), // Initialize the error slice
		mux:  sync.Mutex{},
	}

	// Go query all the schedds
	var wg sync.WaitGroup // Calling wg.Wait() will block execution until all of the wg threads are done
	for _, schedd := range schedds {
		wg.Add(1) // Add a "Lock" the waitgroup
		go func(schedd *condor.Schedd) {
			defer wg.Done() // "Release" one "lock" from the waitgroup
			rows, err := schedd.List(0, keys...)
			if err != nil {
				// Add the error to our errList
				errList.mux.Lock()
				err2 := fmt.Errorf("%s: %w", schedd.Name, err)
				errList.errs = append(errList.errs, err2)
				errList.mux.Unlock()
				return
			}
			// All is well - send the rows to the aggregator
			for _, r := range rows {
				e := entryForAgg{
					scheddName: schedd.Name,
					row:        r,
				}
				aggregator <- e
			}
		}(schedd)
	}
	wg.Wait()         // Don't let execution proceed until all waitgroup "locks" are released
	close(aggregator) // Let the aggregator listener know that it can shut down
	<-aggDone         // Don't proceed until aggregation is done

	// Check for any errors
	if len(errList.errs) > 0 {
		var errCombined string
		for _, err := range errList.errs {
			errCombined += fmt.Sprintf(": %s", err.Error())
		}
		return nil, fmt.Errorf("Could not get list of jobs from schedds%s", errCombined)
	}

	// Compile the rows in order
	s := make([]string, 0)
	for _, schedd := range schedds {
		s = append(s, schedd.Name) // So we can display by schedd
		for _, row := range scheddMap[schedd.Name] {
			s = append(s, row)
		}
		s = append(s, "") // Empty row between schedds
	}

	return s, nil
}
