package main

import (
	"flag"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/pusher/buddha"
)

var (
	BuildVersion  string
	BuildRevision string
)

var (
	ConfigDir   = flag.String("config-dir", "/etc/buddha.d", "")
	ConfigFile  = flag.String("config", "", "")
	ConfigStdin = flag.Bool("stdin", false, "")
	ShowVersion = flag.Bool("version", false, "")
	ConfirmAll  = flag.Bool("y", false, "")
)

const USAGE = `usage: buddha [flags] job_file jobs...

flags:
  --config-dir=/etc/buddha.d  global job configuration directory
  --config=<file>             manually specify job configuration file
  --stdin                     accept job configuration from stdin
  --version                   display version information
  -y                          confirm run all
  -h, --help                  shows this help

examples:
  to invoke api_server from /etc/buddha.d:
    $ buddha api_server
  to invoke all jobs from /etc/buddha.d:
    $ buddha -y all
  to invoke server from /my/app:
    $ buddha --config-dir=/my/app server
  to invoke demo.json file:
    $ buddha --config=demo.json all
  to invoke jobs from stdin:
    $ cat demo.json | buddha --stdin all
`

// --help usage page
func Usage() {
	fmt.Fprint(os.Stderr, USAGE)
}

// --version
func Version() {
	fmt.Println("Build Version:", BuildVersion)
	fmt.Println("Build Revision:", BuildRevision)
}

func main() {
	flag.Usage = Usage
	flag.Parse()
	if *ShowVersion {
		Version()
		return
	}

	var jobs *buddha.Jobs
	if *ConfigFile != "" {
		// load manual job configuration file
		j, err := buddha.OpenFile(*ConfigFile)
		if err != nil {
			log.Fail(2, "config:", err)
			return
		}

		jobs = j
	} else if *ConfigStdin {
		// load job configuration from stdin
		j, err := buddha.Open(os.Stdin)
		if err != nil {
			log.Fail(2, "config:", err)
			return
		}

		jobs = j
	} else {
		j, err := buddha.OpenDir(*ConfigDir)
		if err != nil {
			log.Fail(2, "config:", err)
			return
		}

		jobs = j
	}

	jobsToRun := flag.Args()
	if len(jobsToRun) == 0 {
		log.Fail(2, "please specify job name 'all' to run all jobs")
		return
	}

	if jobsToRun[0] == "all" {
		if !*ConfirmAll {
			log.Warn("all will execute in the order of the source config file/directory.")
			log.Fail(2, "Please re-run with the -y flag to acknowledge.")
		}

		for _, job := range *jobs {
			err := runJob(job)
			if err != nil {
				return
			}
		}
		return
	}

	for _, jobName := range jobsToRun {
		for _, job := range *jobs {
			if job.Name == jobName {
				err := runJob(job)
				if err != nil {
					return
				}
			}
		}
	}
}

func runJob(job *buddha.Job) error {
	log.Info("job:", job.Name)

	for _, cmd := range job.Commands {
		log.Info("command:", cmd.Name)

		// execute before health checks
		// these will only skip command, not terminate the run
		err := executeChecks(cmd, cmd.Before)
		if err != nil {
			log.Warn("checks: before:", err)
			continue
		}

		// execute command
		log.Info("exec:", cmd.Path, cmd.Args)
		cmd.Stdout = execStdout
		err = cmd.Execute()
		if err != nil {
			log.Fail(3, "exec:", err)
			return err
		}

		// grace period between executing command and executing health checks/next command
		log.Info("grace: waiting", cmd.Grace)
		time.Sleep(cmd.Grace.Duration())

		// execute after health checks
		err = executeChecks(cmd, cmd.After)
		if err != nil {
			log.Fail(4, "checks: after:", err)
			return err
		}
	}

	return nil
}

// pipe exec stdout to log
func execStdout(line string) {
	log.Info("stdout:", line)
}

// execute independent checks in worker goroutines
func executeChecks(cmd buddha.Command, checks buddha.Checks) error {
	if len(checks) == 0 {
		return nil
	}

	wg := new(sync.WaitGroup)
	fail := make(chan error, len(checks))

	for _, check := range checks {
		wg.Add(1)

		go executeCheck(wg, cmd, check, fail)
	}
	wg.Wait()

	// use select to pop from fail channel non-blocking
	select {
	case err := <-fail:
		return err
	default:
		// no-op
	}

	return nil
}

// execute a check synchronously as defined by check settings as part of a worker waitgroup
func executeCheck(wg *sync.WaitGroup, cmd buddha.Command, check buddha.Check, fail chan error) {
	defer wg.Done()

	var err error
	for i := 1; i <= cmd.Failures; i++ {
		log.Infof("check %d/%d: %s: checking", i, cmd.Failures, check.String())
		err = check.Execute(cmd.Timeout.Duration())
		if err == nil {
			log.Infof("check %d/%d: %s: success", i, cmd.Failures, check.String())
			break
		}
		log.Warnf("check %d/%d: %s: %s", i, cmd.Failures, check.String(), err)

		log.Infof("check %d/%d: %s: waiting interval %s", i, cmd.Failures, check.String(), cmd.Interval)
		time.Sleep(cmd.Interval.Duration())
	}

	if err != nil {
		fail <- err
	}
}

// return true if string is found in array of strings
func inArray(a string, b []string) bool {
	for _, s := range b {
		if s == a {
			return true
		}
	}

	return false
}
