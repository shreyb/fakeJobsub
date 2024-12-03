package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
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

	listCmd := flag.NewFlagSet("list", flag.ContinueOnError)
	listKeys := listCmd.String("keys", "", "Comma-separated list of keys to query")

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

	if args[1] == submitCmd.Name() {
		if err := checkSubmitForGroup(*submitGroup); err != nil {
			return errors.New("--group must be specified")
		}
	}

	// Remove later
	fmt.Printf("Flag set: %s\n", args[1])
	switch args[1] {
	case "submit":
		fmt.Printf("num = %d\n", *submitNum)
	case "list":
		fmt.Printf("keys = %s\n", *listKeys)
	}

	return nil
}

func parseArgs(args []string) error {
	return nil
}
