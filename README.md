# Tunnelify
Tunnelify is a deployable proxy server and tunnel written in go

[Installing](#installing) | [Quickstart](#quickstart) | [Configuration](#configuration)


## Installing
For now, you can only install tunnelify using `go get`. To install, run:
```sh
$ go get github.com/kofoworola/tunnelify
```

## Quickstart
After installing tunnelify, run this to start up the proxy:
```sh
$ tunnelify start <PATH TO CONFIG FILE>
```

Now the proxy is listening on whatever value is set in your config's `server.host` value and is proxying every request sent through it.


## Configuration
Recommended configuration format is json, but tunnelify also supports toml and yaml. 
Config values can also be set via Environment variables. For example, to set the value of `server.host` via 
Environments, update the value of the `SERVER_HOST`; essentially replace all `.` in the key with `_` and 
change to upper case.

### Available config values
| Name     | Type | Description           | Default |
|----------|------|-----------------------| ----- |
| `server.host`| string | Adress the proxy's server will listen on | null|
| `server.auth`| []string| Array of allowed [Basic](https://tools.ietf.org/html/rfc7617) authorization strings in the form `user-id:password`| [] |
| `server.timeout` | duration| Amount of time the proxy will attempt to establish an outbound connection for | 30s |
| `hideIP` | boolean | Hide the IP of the source of the request | false |
| `logging` | []string | An array of file or URL paths to write logging to (logs are written to `stderr` regardless | [] |

