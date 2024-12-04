package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"

	"fakeJobsub/condor"
)

var errParseFlags = errors.New("could not parse flags")
var errUsage = errors.New("usage called")

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
	submitVerbose := submitCmd.Bool("verbose", false, "Verbose mode")

	listCmd := flag.NewFlagSet("list", flag.ContinueOnError)
	listKeys := listCmd.String("keys", "", "Comma-separated list of keys to query")
	listClusterID := listCmd.Int("clusterid", 0, "ClusterID to query")
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

	flSet, ok := flagSetMap[args[1]]
	if !ok {
		fmt.Println("Invalid subcommand.  Must run fakeJobsub with the \"submit\" or \"list\" subcommand.")
		submitCmd.Usage()
		listCmd.Usage()
		return errors.New("invalid subcommand")
	}

	if err := flSet.Parse(args[2:]); err != nil {
		return errParseFlags
	}

	// Get our schedd
	schedd, err := condor.GetSchedd("")
	if err != nil {
		return fmt.Errorf("could not get schedd: %w", err)
	}

	switch args[1] {
	case submitCmd.Name():
		if err := checkSubmitForGroup(*submitGroup); err != nil {
			return errors.New("--group must be specified")
		}

		if *submitVerbose {
			fmt.Printf("num = %d\n", *submitNum)
			fmt.Printf("group = %s\n", *submitGroup)
		}

		if err := schedd.Submit(*submitGroup, *submitNum); err != nil {
			return fmt.Errorf("could not submit job: %w", err)
		}
		fmt.Println("Submitted job(s) successfully")

	case listCmd.Name():
		if *listVerbose {
			fmt.Printf("keys = %s\n", *listKeys)
			fmt.Printf("clusterID = %d\n", *listClusterID)
		}

		keys := make([]string, 0)
		if *listKeys != "" {
			keysRaw := strings.Split(*listKeys, ",")
			keys := make([]string, 0, len(keysRaw))
			for _, key := range keysRaw {
				keys = append(keys, strings.TrimSpace(key))
			}
		}

		if err := schedd.List(*listClusterID, keys...); err != nil {
			return fmt.Errorf("could not list jobs: %w", err)
		}
	}

	return nil
}
