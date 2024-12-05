This is a little fake example CLI tool that mimicks some of the basic functionality of https://github.com/fermitools/jobsub_lite. I wrote it to demonstrate some of the basic concepts in writing a Go CLI.

# Concepts covered, in no particular order

* Go code CLI structure
* Some best practices, using guidelines from https://go.dev/doc/effective_go
* `package`s
* `func`s
* types
* `func main` as the entrypoint
* `func init` and its purpose in some packages
* `nil`
* `var`, `:=`
* Error handling
* slices/range over slice
* interface
* maps
* channel/goroutines/concurrency
* Tests

# Functionality of this "tool"

`fakeJobsub` pretends to submit jobs to a batch system (at Fermilab, we use [HTCondor](https://htcondor.org/) for high-throughput grid computing), and can fetch the job queue from that "batch system".  The "batch system" here is simply a sqlite database that is written to/read from $TMPDIR on disk.  The CLI format and arguments are derived from [jobsub_lite](https://github.com/fermitools/jobsub_lite).  Keep in mind - we're basically just writing to and reading from a sqlite database - no actual jobs are being submitted in this mock, and thus no jobs will be run.

## Install the tool

To use this, you must have [Go](https://go.dev) and [sqlite3](https://www.sqlite.org/) installed.  Download this repository:

```
$ git clone https://github.com/shreyb/fakeJobsub
```

Then build the `fakeJobsub` executable:

```
$ cd fakeJobsub
$ go build 
```

This should create an executable, `fakeJobsub`, in the repository directory.  You can use the tool from there, or adjust your `PATH` so that you don't need to type `./` before every `fakeJobsub` invocation.  For the rest of the README, the assumption will be that `fakeJobsub` is running from that repository directory, with no `PATH` adjustments.


## Using `fakeJobsub`

To pretend to submit a job, you can do something like:

```
$ ./fakeJobsub submit --group myexperiment --num 5
```

The tool will write an entry into the backing sqlite database, and sleep for a few seconds to simulate network latency and batch system activity.  

You can list jobs in the queue by running:

```
$ ./fakeJobsub list
```

## Multiple "Access Points"
This tool has multiple simulated scheduler machines (schedds/Access Points) hardcoded (schedd1, schedd2).  By default, the `submit` subcommand will randomly pick one "Access Point" to submit jobs to (meaning the corresponding backing DB will be written to).  The `list` subcommand will return results from all "Access Points" by default (all backing DBs will be queried).  To target one "Access Point", use the `--schedd` flag to either subcommand:

```
$ ./fakeJobsub submit --group myexperiment --schedd schedd1
```

and

```
$ ./fakeJobsub list --schedd schedd1
```


## More list functions 

The `list` subcommand allows you to query only certain (valid) keys.  As of this writing, the valid keys are "clusterid, group, num".  Pass these in as a comma-separated list with `--keys` flag to `list`.

One can also query a specific clusterid on an "Access Point" by using the `--clusterid` flag with the `list` subcommand.  In that case, `--schedd` must be specified.  For example:

```
$ ./fakeJobsub list --schedd schedd1 --clusterid 2 --keys "clusterid,num"
```
