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
	ConfigDir   = flag.String("config-dir", "/etc/buddha.d", "global job configuration directory")
	ConfigFile  = flag.String("config", "", "manually specify job coniguration file")
	ConfigStdin = flag.Bool("stdin", false, "accept config from stdin")
	ShowVersion = flag.Bool("version", false, "display version information")
)

// --help usage page
func Usage() {
	fmt.Print("usage: buddha [flags] job_file\r\n\r\n")

	fmt.Print("flags:\r\n")
	fmt.Print("  --config-dir=/etc/buddha.d  global job configuration directory\r\n")
	fmt.Print("  --config=<file>             manually specify job configuration file\r\n")
	fmt.Print("  --stdin                     accept job configiguration from STDIN\r\n")
	fmt.Print("  --version                   display version information\r\n\r\n")

	fmt.Print("examples:\r\n")
	fmt.Print("  to invoke api_server from /etc/buddha.d:\r\n")
	fmt.Print("    $ buddha api_server\r\n")
	fmt.Print("  to invoke all jobs from /etc/buddha.d:\r\n")
	fmt.Print("    $ buddha all\r\n")
	fmt.Print("  to invoke server from /my/app:\r\n")
	fmt.Print("    $ buddha --config-dir=/my/app server\r\n")
	fmt.Print("  to invoke demo.json file:\r\n")
	fmt.Print("    $ buddha --config=demo.json all\r\n")
	fmt.Print("  to invoke jobs from stdin:\r\n")
	fmt.Print("    $ cat demo.json | buddha --stdin all\r\n")
}

// --version
func Version() {
	fmt.Printf("Build Version: %s\r\n", BuildVersion)
	fmt.Printf("Build Revision: %s\r\n", BuildRevision)
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
	} else if flag.Arg(0) == "-" {
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

	for _, job := range *jobs {
		if jobsToRun[0] == "all" || inArray(job.Name, jobsToRun) {
			runJob(job)
		}
	}
}

func runJob(job *buddha.Job) {
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
			return
		}

		// grace period between executing command and executing health checks/next command
		log.Info("grace: waiting", cmd.Grace)
		time.Sleep(cmd.Grace.Duration())

		// execute after health checks
		err = executeChecks(cmd, cmd.After)
		if err != nil {
			log.Fail(4, "checks: after:", err)
			return
		}
	}
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
		log.Infof("check %d/%d: %s: checking\r\n", i, cmd.Failures, check.String())
		err = check.Execute(cmd.Timeout.Duration())
		if err == nil {
			log.Infof("check %d/%d: %s: success", i, cmd.Failures, check.String())
			break
		}
		log.Warnf("check %d/%d: %s: %s\r\n", i, cmd.Failures, check.String(), err)

		log.Infof("check %d/%d: %s: waiting interval %s\r\n", i, cmd.Failures, check.String(), cmd.Interval)
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
