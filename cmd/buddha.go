package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
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
	ConfigDir    = flag.String("config-dir", "/etc/buddha.d", "")
	ConfigFile   = flag.String("config", "", "")
	ConfigStdin  = flag.Bool("stdin", false, "")
	LockPath     = flag.String("lock-path", filepath.Join(os.TempDir(), "buddha.lock"), "")
	OnBeforeFail = flag.String("on-before-fail", "skip", "")
	OnAfterFail  = flag.String("on-after-fail", "stop", "")
	ShowVersion  = flag.Bool("version", false, "")
)

// --help usage page
func Usage() {
	fmt.Println(`usage: buddha [flags] <jobs...>

flags:
  --config-dir=/etc/buddha.d    global job configuration directory
  --config=<file>               manually specify job configuration file
  --stdin                       accept job configuration from STDIN
  --lock-path=/tmp/buddha.lock  path to lock file
  --on-before-fail=skip         job behaviour on before check failure (continue|skip|stop)
  --on-after-fail=stop          run behaviour on after check failure (continue|stop)
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
	fmt.Println("Build Version:", BuildVersion)
	fmt.Println("Build Revision:", BuildRevision)
}

func init() {
	flag.Usage = Usage
	flag.Parse()

	if *OnBeforeFail != "continue" &&
		*OnBeforeFail != "skip" &&
		*OnBeforeFail != "stop" {
		fmt.Println(*OnBeforeFail, "is not a valid value for --on-before-fail")
		os.Exit(2)
	}

	if *OnAfterFail != "continue" &&
		*OnAfterFail != "stop" {
		fmt.Println(*OnAfterFail, " is not a valid value for --on-after-fail")
		os.Exit(2)
	}

	if *ShowVersion {
		Version()
		os.Exit(0)

		return
	}
}

func main() {
	var jobs buddha.Jobs
	var err error

	if *ConfigFile != "" {
		// load manual job configuration file
		jobs, err = buddha.OpenFile(*ConfigFile)
		if err != nil {
			log.Println(log.LevelFail, "fatal: could not read config file %s", *ConfigFile)
		}
	} else if *ConfigStdin {
		// load job configuration from stdin
		jobs, err = buddha.Open(os.Stdin)
		if err != nil {
			log.Println(log.LevelFail, "fatal: could not read config from STDIN")
		}
	} else {
		jobs, err = buddha.OpenDir(*ConfigDir)
		if err != nil {
			log.Println(log.LevelFail, "fatal: could not read config directory %s", *ConfigDir)
		}
	}

	if err != nil {
		log.Println(log.LevelFail, "fatal: %s", err)

		os.Exit(2)
		return
	}

	// exit with status code of run
	os.Exit(run(jobs))

}

func run(jobs buddha.Jobs) int {
	lock, err := flock.Lock(*LockPath)
	if err != nil {
		if err == flock.ErrLocked {
			log.Println(log.LevelFail, "fatal: another instance of buddha is running")
			return 2
		}

		log.Println(log.LevelFail, "fatal: could not obtain exclusive lock at %s", *LockPath)
		log.Println(log.LevelFail, "fatal: %s", err)
		return 1
	}
	defer lock.Close()

	// sort jobs by name
	sort.Sort(jobs)

	jobsToRun := flag.Args()
	if len(jobsToRun) == 0 {
		log.Println(log.LevelFail, "please specify job names, or 'all' to run all")
		return 2
	}

	// if not running all jobs, filter job list
	if jobsToRun[0] != "all" {
		jobs = jobs.Filter(jobsToRun)
	}

	// perform sanity checks against jobs
	for i := 0; i < len(jobs); i++ {
		if jobs[i].Root && (os.Getuid() != 0) {
			log.Println(log.LevelFail, "fatal: job %s requires root privileges", jobs[i].Name)
			return 1
		}
	}

	// execute jobs
	for i := 0; i < len(jobs); i++ {
		err := runJob(jobs[i])
		if err != nil {
			log.Println(log.LevelFail, "fatal: job %s failed", jobs[i].Name)
			log.Println(log.LevelFail, "fatal: %s", err)
			return 1
		}
	}

	return 0
}

func runJob(job *buddha.Job) error {
	log.Println(log.LevelPrim, "Job: %s", job.Name)

	for _, cmd := range job.Commands {
		log.Println(log.LevelPrim, "Command: %s", cmd.Name)

		// execute before health checks
		// these will execute once and depending on --on-before-fail skip this job
		log.Println(log.LevelScnd, "Executing before checks")
		result, err := executeChecks(cmd, cmd.Before, 1)
		if err != nil {
			log.Println(log.LevelFail, "fatal: before checks failed, ending run")
			return err
		}
		if !result {
			if *OnBeforeFail == "stop" {
				log.Println(log.LevelFail, "fatal: before checks returned false, ending run")
				return nil
			} else if *OnBeforeFail == "continue" {
				log.Println(log.LevelFail, "warning: before checks returned false, continuing anyway")
			} else {
				log.Println(log.LevelInfo, "Before checks returned false, skipping job")
				continue
			}
		}

		// execute command
		log.Println(log.LevelScnd, "Executing Command: %s %s", cmd.Path, strings.Join(cmd.Args, " "))
		cmd.Stdout = execStdout
		err = cmd.Execute()
		if err != nil {
			log.Println(log.LevelFail, "fatal: %s", err)
			return err
		}

		// grace period between executing command and executing health checks/next command
		log.Println(log.LevelInfo, "Waiting %s grace...", cmd.Grace)
		time.Sleep(cmd.Grace.Duration())

		// execute after health checks
		log.Println(log.LevelScnd, "Executing after checks")
		result, err = executeChecks(cmd, cmd.After, cmd.Failures)
		if err != nil {
			log.Println(log.LevelFail, "fatal: after checks failed, ending run. err: %s", err)
			return err
		}
		if !result {
			if *OnAfterFail == "continue" {
				log.Println(log.LevelFail, "warning: after checks returned false, continuing anyway")
				continue
			}

			log.Println(log.LevelFail, "fatal: after checks returned false, ending run")
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
func executeChecks(cmd buddha.Command, checks buddha.Checks, failures int) (bool, error) {
	if len(checks) == 0 {
		return true, nil
	}

	wg := new(sync.WaitGroup)
	done := make(chan bool, len(checks))
	fail := make(chan error, 1)

	for _, check := range checks {
		wg.Add(1)

		go executeCheck(wg, cmd, check, done, fail, failures)
	}
	wg.Wait()
	close(done)

	select {
	case err := <-fail:
		return false, err
	default:
		// Only return true if all checks returned true
		aggregatedResult := true
		for result := range done {
			aggregatedResult = aggregatedResult && result
		}
		return aggregatedResult, nil
	}
}

// execute a check synchronously as defined by check settings as part of a worker waitgroup
func executeCheck(wg *sync.WaitGroup, cmd buddha.Command, check buddha.Check, done chan bool, fail chan error, failures int) {
	defer wg.Done()

	for i := 1; i <= failures; i++ {
		log.Println(log.LevelInfo, "Check %d/%d: %s: checking...", i, failures, check.String())
		result, err := check.Execute(cmd.Timeout.Duration())
		if err != nil {
			fail <- err
			return
		}
		if result {
			log.Println(log.LevelInfo, "Check %d/%d: %s success!", i, failures, check.String())
			done <- true
			return
		}
		log.Println(log.LevelInfo, "Check %d/%d: %s: returned false", i, failures, check.String())

		if i < failures {
			log.Println(log.LevelInfo, "Check %d/%d: %s: waiting %s...", i, failures, check.String(), cmd.Interval)
			time.Sleep(cmd.Interval.Duration())
		}
	}
	log.Println(log.LevelInfo, "All attempts returned false")
	done <- false
}
