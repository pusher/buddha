buddha
=======

Buddha executes a set of commands in lock stop, while issuing health checks before executing the next command.

Buddha is designed to work with the [God](http://godrb.com/) process manager. Unlike God who watches over processed, Buddha helps guide processes through reincarnation.

[GoDoc](https://godoc.org/github.com/pusher/buddha)


Example Config
--------------

Below is an example of starting a redis server, ensuring is comes up with a TCP health check, starting our demo application, and ensuring it responds healthily before terminating.

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


Todo
----

   - [ ] differentiate stdout/stderr from command in log output
   - [ ] output (optional) result data from checks
