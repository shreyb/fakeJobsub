package main

import (
	"errors"
	"flag"
	"fmt"
	"math/rand"
	"os"
	"slices"
	"strings"

	"fakeJobsub/condor"
)

var (
	errParseFlags = errors.New("could not parse flags")
	errUsage      = errors.New("usage called")
)

var schedds = []string{"schedd1", "schedd2"}

func main() {
	if err := run(os.Args); err != nil {
		if errors.Is(err, errParseFlags) {
			os.Exit(2)
		}
		fmt.Printf("Error running fakeJobsub: %s\n", err)
		os.Exit(1)
	}
}

// Note - by making run depend on args, I now can TEST it!
// This is the main reason folks sometimes split out a "run" function from the main function - since main isn't really that testable as is.
func run(args []string) error {
	// Flags
	submitCmd := flag.NewFlagSet("submit", flag.ContinueOnError)
	submitNum := submitCmd.Int("num", 1, "Number of jobs to submit")
	submitGroup := submitCmd.String("group", "", "Group/Experiment")
	submitSchedd := submitCmd.String("schedd", "", "schedd to submit to.  If blank, one will be randomly chosen")
	submitVerbose := submitCmd.Bool("verbose", false, "Verbose mode")

	listCmd := flag.NewFlagSet("list", flag.ContinueOnError)
	listKeys := listCmd.String("keys", "", "Comma-separated list of keys to query")
	listClusterID := listCmd.Int("clusterid", 0, "ClusterID to query. Must also specify --schedd.")
	listSchedd := listCmd.String("schedd", "", "schedd to query from.  If blank, will query all configured schedds")
	listVerbose := listCmd.Bool("verbose", false, "Verbose mode")

	// Map of our flagsets to their names.  Very contrived.  Gives us something like {"submit": submitCmd, "list": listCmd}
	flagSetMap := make(map[string]*flag.FlagSet, 0)
	flagSetMap[submitCmd.Name()] = submitCmd
	flagSetMap[listCmd.Name()] = listCmd

	// Parse args
	if len(args) < 2 {
		fmt.Fprintf(os.Stderr, "fakeJobsub must be run with the \"submit\" or \"list\" subcommand\n\n")
		submitCmd.Usage()
		listCmd.Usage()
		return errUsage
	}

	subcommand := args[1]

	flSet, ok := flagSetMap[subcommand]
	if !ok {
		fmt.Println("Invalid subcommand.  Must run fakeJobsub with the \"submit\" or \"list\" subcommand.")
		submitCmd.Usage()
		listCmd.Usage()
		return errors.New("invalid subcommand")
	}

	if err := flSet.Parse(args[2:]); err != nil {
		return errParseFlags
	}

	// Subcommand logic
	switch subcommand {
	case submitCmd.Name():
		if err := checkSubmitForGroup(*submitGroup); err != nil {
			return errors.New("--group must be specified")
		}

		if *submitVerbose {
			fmt.Printf("num = %d\n", *submitNum)
			fmt.Printf("group = %s\n", *submitGroup)
			fmt.Printf("schedd = %s\n", *submitSchedd)
		}

		// Pick a schedd based on --schedd
		scheddName := *submitSchedd
		if !slices.Contains(schedds, *submitSchedd) {
			// randomly pick one
			fmt.Printf("Given schedd %s is not in the list of configured schedds: %v.  Picking one randomly.\n", *submitSchedd, schedds)
			scheddName = schedds[rand.Intn(len(schedds))]
		}
		if scheddName == "" {
			// randomly pick one
			scheddName = schedds[rand.Intn(len(schedds))]
		}

		schedd, err := condor.GetSchedd(scheddName)
		if err != nil {
			return fmt.Errorf("could not get schedd: %w", err)
		}

		if err := schedd.Submit(*submitGroup, *submitNum); err != nil {
			return fmt.Errorf("could not submit job: %w", err)
		}
		fmt.Println("Submitted job(s) successfully")
		return nil

	case listCmd.Name():
		if *listVerbose {
			fmt.Printf("keys = %s\n", *listKeys)
			fmt.Printf("clusterID = %d\n", *listClusterID)
			fmt.Printf("schedd = %s\n", *listSchedd)
		}

		// Stop and return an error if we specified --clusterid but not --schedd
		if *listClusterID != 0 && *listSchedd == "" {
			return errors.New("must set --schedd flag if --clusterid is specified")
		}

		keys := make([]string, 0)
		if *listKeys != "" {
			keysRaw := strings.Split(*listKeys, ",")
			keys = make([]string, 0, len(keysRaw))
			for _, key := range keysRaw {
				keys = append(keys, strings.TrimSpace(key))
			}
		}

		// We're running query on one schedd
		if *listSchedd != "" {
			if !slices.Contains(schedds, *listSchedd) {
				return fmt.Errorf("invalid schedd: %s.  Please choose from valid schedds %v or do not set the --schedd flag", *listSchedd, schedds)
			}

			schedd, err := condor.GetSchedd(*listSchedd)
			if err != nil {
				return fmt.Errorf("could not get schedd: %w", err)
			}

			rows, err := schedd.List(*listClusterID, keys...)
			if err != nil {
				return fmt.Errorf("could not list jobs: %w", err)
			}

			// Print our rows
			for _, row := range rows {
				fmt.Println(row)
			}
			return nil
		}

		// Don't have specific schedd - query them all!
		scheddObjs := make([]*condor.Schedd, 0, len(schedds))
		for _, s := range schedds {
			schedd, err := condor.GetSchedd(s)
			if err != nil {
				return fmt.Errorf("could not get schedd: %w", err)
			}
			scheddObjs = append(scheddObjs, schedd)
		}
		rows, err := listJobsFromSchedds(scheddObjs, keys...)
		if err != nil {
			return fmt.Errorf("could not list jobs from all schedds: %w", err)
		}

		// Print the rows!
		for _, row := range rows {
			fmt.Println(row)
		}
		return nil
	}
	return nil
}
