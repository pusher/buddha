<img src="http://i.imgur.com/Cz0Ee3h.png" width="384" height="142" alt="buddha" />

buddha
=======

Buddha executes a set of commands in lock stop, while issuing health checks before executing the next command.

Buddha is designed to work with the [God](http://godrb.com/) process manager. Unlike God who watches over processes, Buddha helps guide processes through reincarnation.

Requirements:

  - Go 1.5+

[GoDoc](https://godoc.org/github.com/pusher/buddha)


Configuration
-------------

A buddha configuration file consists of an array of jobs. Each job executes a set of commands and performs health checks after every command and before continuing to the next one. Currently implemented are HTTP, TCP and script health checks.

Every health check is executed within a timed constraint, as noted below:

  - **Grace:** the period between executing a command and performaing health checks, to allow the application a window in which to initialise
  - **Timeout:** the period in which a health check has to execute, if a health check exceeds this it is deemed to have failed and will have its response ignored
  - **Interval:** the backoff period after a failed check before trying again up to the `failures` limit

Below is an example of starting a redis server, ensuring is comes up with a TCP health check, starting our demo application, and ensuring it responds healthily before terminating.

**Example:**

```json
[
  {
    "name": "my_app",
    "commands": [
      {
        "path": "service",
        "args": ["redis", "start"],

        "after": [
          {"type": "tcp", "name": "redis", "addr": "127.0.0.1:6379"}
        ],

        "grace": "5s",
        "timeout": "1s",
        "interval": "2s",
        "failures": 5
      },
      {
        "path": "service",
        "args": ["my_app", "restart"],

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

```
usage: buddha [flags] job_file

flags:
  --config-dir=/etc/buddha.d  global job configuration directory
  --config=<file>             manually specify job configuration file
  --stdin                     accept job configiguration from STDIN
  --version                   display version information

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
