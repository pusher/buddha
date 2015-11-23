package main

import (
	"flag"
	"fmt"
	"os"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/pusher/buddha"
	"github.com/pusher/buddha/flock"
	"github.com/pusher/buddha/log"
)

var (
	BuildVersion  string = "development"
	BuildRevision string = "development"
)

var (
	ConfigDir   = flag.String("config-dir", "/etc/buddha.d", "")
	ConfigFile  = flag.String("config", "", "")
	ConfigStdin = flag.Bool("stdin", false, "")
	LockPath    = flag.String("lock-path", "/tmp/buddha.lock", "")
	ShowVersion = flag.Bool("version", false, "")
)

// --help usage page
func Usage() {
	fmt.Println(`usage: buddha [flags] <jobs...>

flags:
  --config-dir=/etc/buddha.d    global job configuration directory
  --config=<file>               manually specify job configuration file
  --stdin                       accept job configuration from STDIN
  --lock-path=/tmp/buddha.lock  path to lock file
  --version                     display version information

examples:
  to invoke api_server from /etc/buddha.d:
    $ buddha api_server
  to invoke all jobs from /etc/buddha.d:
    $ buddha all
  to invoke server from /my/app:
    $ buddha --config-dir=/my/app server
  to invoke demo.json file:
    $ buddha --config=demo.json all
  to invoke jobs from stdin:
    $ cat demo.json | buddha --stdin all`)
}

// --version
func Version() {
	fmt.Printf("Build Version: %s\r\n", BuildVersion)
	fmt.Printf("Build Revision: %s\r\n", BuildRevision)
}

func init() {
	flag.Usage = Usage
	flag.Parse()

	if *ShowVersion {
		Version()
		os.Exit(0)

		return
	}
}

func main() {
	lock, err := flock.Lock(*LockPath)
	if err != nil {
		if err == flock.ErrLocked {
			log.Println(log.LevelFail, "fatal: another instance of buddha is running")
			os.Exit(2)

			return
		}

		log.Println(log.LevelFail, "fatal: could not obtain exclusive lock at %s", *LockPath)
		log.Println(log.LevelFail, "fatal: %s", err)
		os.Exit(1)

		return
	}
	defer lock.Close()

	var jobs buddha.Jobs
	if *ConfigFile != "" {
		// load manual job configuration file
		jobs, err = buddha.OpenFile(*ConfigFile)
		if err != nil {
			log.Println(log.LevelFail, "fatal: could not read config file %s", *ConfigFile)
			log.Println(log.LevelFail, "fatal: %s", err)
			os.Exit(2)

			return
		}
	} else if *ConfigStdin {
		// load job configuration from stdin
		jobs, err = buddha.Open(os.Stdin)
		if err != nil {
			log.Println(log.LevelFail, "fatal: could not read config from STDIN")
			log.Println(log.LevelFail, "fatal: %s", err)
			os.Exit(2)

			return
		}
	} else {
		jobs, err = buddha.OpenDir(*ConfigDir)
		if err != nil {
			log.Println(log.LevelFail, "fatal: could not read config directory %s", *ConfigDir)
			log.Println(log.LevelFail, "fatal: %s", err)
			os.Exit(2)

			return
		}
	}

	// sort jobs by name
	sort.Sort(jobs)

	jobsToRun := flag.Args()
	if len(jobsToRun) == 0 {
		log.Println(log.LevelFail, "please specify job names, or 'all' to run all")
		os.Exit(2)

		return
	}

	// if not running all jobs, filter job list
	if jobsToRun[0] != "all" {
		jobs = jobs.Filter(jobsToRun)
	}

	// perform sanity checks against jobs
	for i := 0; i < len(jobs); i++ {
		if jobs[i].Root && (os.Getuid() != 0) {
			log.Println(log.LevelFail, "fatal: job %s requires root privileges", jobs[i].Name)
			os.Exit(1)
		}
	}

	// execute jobs
	for i := 0; i < len(jobs); i++ {
		err := runJob(jobs[i])
		if err != nil {
			log.Println(log.LevelFail, "fatal: job %s failed", jobs[i].Name)
			log.Println(log.LevelFail, "fatal: %s", err)
			os.Exit(1)

			return
		}
	}
}

func runJob(job *buddha.Job) error {
	log.Println(log.LevelPrim, "Job: %s", job.Name)

	for _, cmd := range job.Commands {
		log.Println(log.LevelPrim, "Command: %s", job.Name)

		// execute before health checks
		// these will only skip command, not terminate the run
		log.Println(log.LevelScnd, "Executing health checks")
		err := executeChecks(cmd, cmd.Before)
		if err != nil {
			log.Println(log.LevelFail, "error: before checks failed, skipping run")
			continue
		}

		// execute command
		log.Println(log.LevelScnd, "Executing Command: %s %s", cmd.Path, strings.Join(cmd.Args, " "))
		cmd.Stdout = execStdout
		err = cmd.Execute()
		if err != nil {
			log.Println(log.LevelFail, "fatal: command exited with non-zero status")
			log.Println(log.LevelFail, "fatal: %s", err)

			return err
		}

		// grace period between executing command and executing health checks/next command
		log.Println(log.LevelInfo, "Waiting %s grace...", cmd.Grace)
		time.Sleep(cmd.Grace.Duration())

		// execute after health checks
		log.Println(log.LevelScnd, "Executing health checks")
		err = executeChecks(cmd, cmd.After)
		if err != nil {
			log.Println(log.LevelFail, "fatal: after checks failed")
			log.Println(log.LevelFail, "fatal: %s", err)

			return err
		}
	}

	return nil
}

// pipe exec stdout to log
func execStdout(line string) {
	log.Println(log.LevelInfo, line)
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
		log.Println(log.LevelInfo, "Check %d/%d: %s: checking...", i, cmd.Failures, check.String())
		err = check.Execute(cmd.Timeout.Duration())
		if err == nil {
			log.Println(log.LevelInfo, "Check %d/%d: %s success!", i, cmd.Failures, check.String())
			break
		}
		log.Println(log.LevelInfo, "Check %d/%d: %s: %s", i, cmd.Failures, check.String(), err)

		log.Println(log.LevelInfo, "Check %d/%d: %s: waiting %s...", i, cmd.Failures, check.String(), cmd.Interval)
		time.Sleep(cmd.Interval.Duration())
	}

	if err != nil {
		fail <- err
	}
}
