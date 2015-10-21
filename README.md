buddha
=======

Buddha executes a set of commands in lock stop, while issuing health checks before executing the next command.

Buddha is designed to work with the [God](http://godrb.com/) process manager. Unlike God who watches over processed, Buddha helps guide processes through reincarnation.

Requirements:

  - Go 1.5+

[GoDoc](https://godoc.org/github.com/pusher/buddha)


Configuration
-------------

A buddha configuration file consists of an array of jobs. Each job executes a set of commands and performs health checks after every command and before continuing to the next one. Currently implemented are HTTP and TCP health checks.

Every health check is executed within a timed constraint, as noted below:

  - **Grace:** the period between executing a command and performaing health checks, to allow the application a window in which to initialise
  - **Timeout:** the period in which a health check has to execute, if a health check exceeds this it is deemed to have failed and will have its response ignored
  - **Interval:** the backoff period after a failed check before trying again up to the `failures` limit

Below is an example of starting a redis server, ensuring is comes up with a TCP health check, starting our demo application, and ensuring it responds healthily before terminating.

**Example:**

```json
[
  {
    "name": "app",
    "commands": [
      {
      	// execute command `service redis start`
      	// expect a zero-exit from command
        "path": "service",
        "args": ["redis", "start"],

        "tcp": [
          // name the check redis in logs, check against 127.0.0.1:6379
          {"name": "redis", "addr": "127.0.0.1:6379"}
        ],

        "grace": "2s",    // allow redis 2s to start before health checking
        "timeout": "1s",  // execution timeout for health check
        "interval": "2s", // timeout between health checks
        "failures": 5     // maximum health check failures before terminating
      },
      {
      	// execute command `service my_app start`
      	// expect a zero-exit from command
        "path": "service",
        "args": ["my_app", "start"],

        "http": [
          // name the check app in logs, check the URL http://127.0.0.1:8080/health_check
          // and expected an HTTP 200 status code before continuing
          {"name": "app", "path": "http://127.0.0.1:8080/health_check", "expect": [200]}
        ],

        "grace": "5s",    // allow app 5s to start before health checking
        "timeout": "1s",  // execution timeout for health check
        "interval": "5s", // timeout between health checks
        "failures": 5     // maximum health check failures before terminating
      }
    ]
  }
]
```
