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

// possible behaviours if checks return false
const (
	ContinueBehaviour = "continue"
	SkipBehaviour     = "skip"
	StopBehaviour     = "stop"
)

var (
	BuildVersion  string = "development"
	BuildRevision string = "development"
)

var (
	ConfigDir     = flag.String("config-dir", "/etc/buddha.d", "")
	ConfigFile    = flag.String("config", "", "")
	ConfigStdin   = flag.Bool("stdin", false, "")
	LockPath      = flag.String("lock-path", filepath.Join(os.TempDir(), "buddha.lock"), "")
	OnUnnecessary = flag.String("on-unnecessary", "skip", "")
	OnBeforeFail  = flag.String("on-before-fail", "skip", "")
	OnAfterFail   = flag.String("on-after-fail", "stop", "")
	ShowVersion   = flag.Bool("version", false, "")
)

// --help usage page
func Usage() {
	fmt.Println(`usage: buddha [flags] <jobs...>

flags:
  --config-dir=/etc/buddha.d    global job configuration directory
  --config=<file>               manually specify job configuration file
  --stdin                       accept job configuration from STDIN
  --lock-path=/tmp/buddha.lock  path to lock file
  --on-unnecessary=skip         job behaviour if necessity checks deem it unnecessary (continue|skip)
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

	if *OnUnnecessary != ContinueBehaviour &&
		*OnUnnecessary != SkipBehaviour {
		fmt.Println(*OnUnnecessary, "is not a valid value for --on-unnecessary")
		os.Exit(2)
	}

	if *OnBeforeFail != ContinueBehaviour &&
		*OnBeforeFail != SkipBehaviour &&
		*OnBeforeFail != StopBehaviour {
		fmt.Println(*OnBeforeFail, "is not a valid value for --on-before-fail")
		os.Exit(2)
	}

	if *OnAfterFail != ContinueBehaviour &&
		*OnAfterFail != StopBehaviour {
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
			log.Println(log.LevelFail, "fatal: job %s failed with unexpected error: %s", jobs[i].Name, err)
			return 1
		}
	}

	return 0
}

func runJob(job *buddha.Job) error {
	log.Println(log.LevelPrim, "Job: %s", job.Name)

	for _, cmd := range job.Commands {
		log.Println(log.LevelPrim, "Command: %s", cmd.Name)

		log.Println(log.LevelScnd, "Executing necessity checks")
		isNecessaryResults, err := executeChecks(cmd, cmd.Necessity, executeNecessityCheck)
		if err != nil {
			log.Println(log.LevelFail, "fatal: unexpected error from necessity check, ending run")
			return err
		}
		if allFalse(isNecessaryResults) {
			switch *OnUnnecessary {
			case ContinueBehaviour:
				log.Println(log.LevelFail, "warning: job unnecessary, continuing anyway")
			default:
				log.Println(log.LevelInfo, "Job deemed unnecessary, skipping")
				continue
			}
		}

		// execute before health checks
		// these will execute once and depending on --on-before-fail skip this job
		log.Println(log.LevelScnd, "Executing before checks")
		checksResults, err := executeChecks(cmd, cmd.Before, executeHealthCheck)
		if err != nil {
			log.Println(log.LevelFail, "fatal: unexpected error from before check, ending run")
			return err
		}
		if anyFalse(checksResults) {
			switch *OnBeforeFail {
			case StopBehaviour:
				log.Println(log.LevelFail, "fatal: before returned false, ending run")
				return nil
			case ContinueBehaviour:
				log.Println(log.LevelFail, "warning: before returned false, continuing anyway")
			default:
				log.Println(log.LevelFail, "warning: before returned false, skipping job")
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
		checksResults, err = executeChecks(cmd, cmd.After, executeHealthCheck)
		if err != nil {
			log.Println(log.LevelFail, "fatal: unexpected error from after check, ending run. err: %s", err)
			return err
		}
		if anyFalse(checksResults) {
			if *OnAfterFail == ContinueBehaviour {
				log.Println(log.LevelFail, "warning: after checks failed, continuing anyway")
				continue
			}

			log.Println(log.LevelFail, "fatal: after checks failed, ending run")
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
func executeChecks(cmd buddha.Command, checks buddha.Checks, executeCheck ExecuteCheck) ([]bool, error) {
	if len(checks) == 0 {
		return nil, nil
	}

	wg := new(sync.WaitGroup)
	done := make(chan bool, len(checks))
	fail := make(chan error, 1)

	for _, check := range checks {
		wg.Add(1)

		go executeCheck(wg, cmd, check, done, fail)
	}
	wg.Wait()
	close(done)

	select {
	case err := <-fail:
		return nil, err
	default:
		results := make([]bool, len(checks))
		i := 0
		for result := range done {
			results[i] = result
			i++
		}
		return results, nil
	}
}

type ExecuteCheck func(*sync.WaitGroup, buddha.Command, buddha.Check, chan bool, chan error)

func executeNecessityCheck(wg *sync.WaitGroup, cmd buddha.Command, check buddha.Check, done chan bool, fail chan error) {
	defer wg.Done()

	log.Println(log.LevelInfo, "Check %s: checking...", check.String())
	err := check.Execute(cmd.Timeout.Duration())
	if err != nil {
		switch e := err.(type) {
		case buddha.CheckFalse:
			log.Println(log.LevelInfo, "Check %s: deemed job unnecessary: %s", check.String(), e)
			done <- false
		default:
			// unexpected failure
			fail <- err
		}
	} else {
		log.Println(log.LevelInfo, "Check %s: deemed job necessary", check.String())
		done <- true
	}
}

// execute a check synchronously as defined by check settings as part of a worker waitgroup
func executeHealthCheck(wg *sync.WaitGroup, cmd buddha.Command, check buddha.Check, done chan bool, fail chan error) {
	defer wg.Done()

	for i := 1; i <= cmd.Failures; i++ {
		log.Println(log.LevelInfo, "Check %d/%d: %s: checking...", i, cmd.Failures, check.String())
		err := check.Execute(cmd.Timeout.Duration())
		if err != nil {
			switch e := err.(type) {
			case buddha.CheckFalse:
				log.Println(log.LevelInfo, "Check %d/%d: %s: returned false: %s", i, cmd.Failures, check.String(), e)
				if i < cmd.Failures {
					log.Println(log.LevelInfo, "Check %d/%d: %s: waiting %s...", i, cmd.Failures, check.String(), cmd.Interval)
					time.Sleep(cmd.Interval.Duration())
				}
			default:
				// unexpected failure, do not retry
				fail <- err
				return
			}
		} else {
			log.Println(log.LevelInfo, "Check %d/%d: %s success!", i, cmd.Failures, check.String())
			done <- true
			return
		}
	}
	done <- false
}

func allFalse(arr []bool) bool {
	if arr == nil {
		return false
	}
	aggregate := true
	for _, v := range arr {
		aggregate = aggregate && !v
	}
	return aggregate
}

func anyFalse(arr []bool) bool {
	for _, v := range arr {
		if v == false {
			return true
		}
	}
	return false
}
