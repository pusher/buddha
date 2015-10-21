package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
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
	ShowVersion = flag.Bool("version", false, "display version information")
)

// --help usage page
func Usage() {
	fmt.Print("usage: buddha [flags] job_file\r\n\r\n")

	fmt.Print("flags:\r\n")
	fmt.Print("  --config-dir=/etc/buddha.d  global job configuration directory\r\n")
	fmt.Print("  --config=<file>             manually specify job configuration file\r\n")
	fmt.Print("  --version                   display version information\r\n\r\n")

	fmt.Print("examples:\r\n")
	fmt.Print("  to invoke /etc/buddha.d/api_server.json:\r\n")
	fmt.Print("    $ buddha api_server\r\n")
	fmt.Print("  to invoke /my/app/server.json:\r\n")
	fmt.Print("    $ buddha --config-dir=/my/app server\r\n")
	fmt.Print("  to invoke demo.json file:\r\n")
	fmt.Print("    $ buddha --config=demo.json\r\n")
	fmt.Print("  to invoke jobs from stdin:\r\n")
	fmt.Print("    $ cat demo.json | buddha -\r\n")
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
		}

		jobs = j
	} else {
		// load named job configuration file from global
		if flag.Arg(0) == "" {
			log.Fail(2, "config: please specify a job name or config file")
			return
		}

		path := filepath.Join(*ConfigDir, flag.Arg(0)+".json")
		j, err := buddha.OpenFile(path)
		if err != nil {
			log.Fail(1, "fail: config:", err)
			return
		}

		jobs = j
	}

	for _, job := range *jobs {
		runJob(job)
	}
}

func runJob(job *buddha.Job) {
	log.Info("job:", job.Name)

	for _, cmd := range job.Commands {
		log.Info("exec:", cmd.Path, cmd.Args)

		cmd.Stdout = execStdout
		err := cmd.Execute()
		if err != nil {
			log.Fail(3, "exec:", err)
			return
		}

		checks := cmd.All()
		if len(checks) == 0 {
			log.Info("check: skipping checks")
			continue
		}

		// grace period between executing command and executing health checks
		log.Info("grace: waiting", cmd.Grace)
		time.Sleep(cmd.Grace.Duration())

		err = executeChecks(cmd, checks)
		if err != nil {
			log.Fail(4, "checks:", err)
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
	wg := new(sync.WaitGroup)
	fail := make(chan error, 1)

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
