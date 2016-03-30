<img src="http://i.imgur.com/Cz0Ee3h.png" width="384" height="142" alt="buddha" />

buddha
=======

[![Build Status](https://travis-ci.org/pusher/buddha.svg?branch=master)](https://travis-ci.org/pusher/buddha)

Buddha executes a set of commands in lock step, while issuing health checks before and after executing the next command.

Buddha is designed to work with the [God](http://godrb.com/) process manager. Unlike God who watches over processes, Buddha helps guide processes through reincarnation.

Requirements:

  - Go 1.5+

[GoDoc](https://godoc.org/github.com/pusher/buddha)


Configuration
-------------

A buddha configuration file consists of an array of jobs. Each job consists of a set of commands and checks. There are three kinds of checks run in the following order:

  - "necessity checks": Check if the job is worth running; for example the most recent version may already running. The default is to skip the job if it is deemed unnecessary by *all* necessity checks
  - "before health checks": Check the state of the system is correct before running. These will retry if the check returns false. The default is to skip the job if all attempts return false for *any* health check
  - "after health checks": Checks performed after the command and act as validation. These will retry if the check returns false. The default is to terminate the buddha run if all attempts return false for *any* health check

Every health check is executed within a timed constraint, as noted below:

  - **Grace:** the period between executing a command and performaing health checks, to allow the application a window in which to initialise
  - **Timeout:** the period in which a health check has to execute, if a health check exceeds this it is deemed to have failed and will have its response ignored
  - **Interval:** the backoff period after a failed check before trying again up to the `failures` limit

Below is an example of starting a redis server, ensuring is comes up with a TCP health check, starting our demo application, and ensuring it responds healthily before terminating.

**Example:**

```json
[
  {
    "name": "my_app", // unique name for job
    "root": true,     // require that job runs as root user
    "commands": [
      {
        "path": "service",          // path to command (if not a path, $PATH environment will be searched)
        "args": ["redis", "start"], // arguments to pass to command

        // necessity check to see if we need to run the command
        // "exec" checks exit codes are assumed to have the meanings: 0 => true, 1 => false, 2 => error
        "necessity": [
          {"type": "exec", "name": "check_latest_version", "path": "is_latest_version" }
        ],

        // health checks to execute after command
        // failures will terminate the buddha run
        "after": [
          {"type": "tcp", "name": "redis", "addr": "127.0.0.1:6379"}
        ],

        "grace": "5s",    // grace period between commands
        "timeout": "1s",  // timeout for health check execution
        "interval": "2s", // interval between health checks
        "failures": 5     // maximum health check failures to tolerate
      },
      {
        "path": "service",
        "args": ["my_app", "restart"],

        // health checks to execute before command
        // failures will skip the current command
        "before": [
          {"type": "http", "name": "http_failed", "path": "http://127.0.0.1:8080/health_check", "expect": [500]}
        ],

        "after": [
          {"type": "http", "name": "http_success", "path": "http://127.0.0.1:8080/health_check", "expect": [200]}
        ],

        "grace": "5s",
        "timeout": "1s",
        "interval": "2s",
        "failures": 5
      }
    ]
  }
]

```


Usage
-----

### Runtime Options

  - `--on-before-fail`: determines the job behaviour on a before check failing.
    - `continue` continue with execution of the job
    - `skip` skip and start the next job (default)
    - `stop` end the current buddha run and exit
  - `--on-after-fail`: determines the run behavior on an after check failing.
    - `continue` continue with execution of next job
    - `stop` end the current buddha run and exit (default)

```
usage: buddha [flags] <jobs...>

flags:
  --config-dir=/etc/buddha.d    global job configuration directory
  --config=<file>               manually specify job configuration file
  --stdin                       accept job configuration from STDIN
  --lock-path=/tmp/buddha.lock  path to lock file
  --on-unnecessary=skip         job behaviour if necessity checks deem it unnecessary (continue|skip)
  --on-before-fail=skip         behaviour on before check failure (continue|skip|stop)
  --on-after-fail=stop          behaviour on after check failure (continue|stop)
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
    $ cat demo.json | buddha --stdin all
```
