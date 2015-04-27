# autobld
autobld is a utility to automatically recompile your server application when changes are made to your source code.

autobld can also set up a proxy for your server's listening ports that will block while the server reloads. This avoids you constantly refreshing till the server is up.

## Quick Start

If all you want is to automatically restart the server for you when any change is made to any file in the current directory, then do:

```
autobld python test.py
```

### Limit watched files

However, you may want to be more selective, and only restart the server when code changes (e.g. any python files). You can specify file patterns to watch by using the `--match` or `-m` flag.
```
autobld -m *.py python test.py
```

### Proxy ports

autobld can set up a proxy which blocks while the server is reloading. You can set up a TCP port proxy by using `--proxy` (`-p` for short).

For example, if you have a server that listens on port 8080, you can set up a proxy port on 9090 that forwards to 8080:
```
autobld -m *.py -p 9090:8080 python test.py
```

You can also use a HTTP proxy instead, which also allows forwarding to a custom path:
```
autobld -m *.py -p http:9090:8080/server python test.py
```

This sets up a HTTP proxy which will forward `:9090/[URL]` to `localhost:8080/test/[URL]`. For more information, see the [Proxies](#proxies) section.

## Configuration Flags

Short | Long       | Description
---   | ---        | ---
-c     | --config     | Path to the configuration file.
-v     | --verbose    | Verbose logging, can be specified multiple times
-q     | --quiet      | Quiet mode, disables all logging
-d     | --dir        | Directory to execute the commands in (by default, the current directory).
-m     | --match      | File patterns to match (by default, `*`)
-x     | --excludeDir | Directories to exclude from watching (by default, `*.git`, `*.hg`)
-p     | --proxy      | List of proxy ports to set up. See [Proxy](#proxies) for more information.

### Timeouts
Timeouts described in [Timeouts](#timeouts) can be controlled using the following flags:

Flag | Description
--- | ---
--changeTimeout | Change timeout
--killTimeout | Kill timeout

## Proxies

Proxy ports can be used to avoid connections failing while the server is being reloaded. Proxy ports will attempt to try connect to the target port for a minute before giving up. There are two types of proxy ports: TCP and HTTP.

TCP proxy ports do not modify the request in any way, and can be used for any TCP protocol.

HTTP proxy ports act as a HTTP proxy, and will do things like use a correct `Host` header, and allow for custom path prefixes. E.g. if you set up a proxy using `-p http:9090:8080/server`, then a request made to `localhost:9090/test` will actually be forwarded to `localhost:8080/server/test`.

## Timeouts
There are also some more advanced timeout configurations:

**Change timeout**: The amount of time to wait after a change is detected before restarting the task. This is used to avoid restarting the task multiple times of there are many files saved within a short period. The default change timeout is 1 second.

**Kill timeout**: The amount of time to wait after sending Ctrl-C before using a Kill signal to kill a task. The default kill timeout is 1 second.

Timeouts are specified in the format used by [ParseDuration](http://golang.org/pkg/time/#ParseDuration), which supports values such as `1s` for 1 second, or `250ms` for 250 milliseconds.


## Configuration file
A YAML configuration file can be used using the `--config` (or `-c` for short) flag. When a configuration file is specified, configuration flags are ignored.

`autobld -c autobld.yaml`

Sample configuration files can be seen in the [test](test) directory. Below is an explanation of the different YAML options:
```yaml
# The base directory used as the working directory for executing commands.
# This directory can be absolute, or relative to the location of the configuration file.
# If no directory is given, it defaults to the directory of the configuration file.
baseDir: src

# The action and arguments to run. This will be rerun if any changes are detected.
action: ["go", "run", "main.go"]

# Proxy is the configuration for the proxies to set up. Multiple proxies can be specified.
proxy:
# If no type is given, the default proxy is a TCP proxy.
- port: 9090
  forwardTo: 8080
# HTTP proxies can forward the request with a specified prefix.
- port: 9091
  forwardTo: 8080
  type: http
  httpPath: /server

# Matchers specify the directories and file patterns within the directory to watch for changes.
# Multiple matchers can be specified, allowing different patterns to be watched in different directories.
# If no matchers are specified, all files in baseDir are watched.
matchers:
# If no directories are specified for a matcher, it defaults to baseDir.
- patterns: ["*.go", "*.sh"]
```

Here is an example with more complex matchers:
```yaml
baseDir: src
action: ["go", "run", "main.go"]
matchers:
# Reload on any changes to *.go files or *.py files in the test/server and test/library directory.
# Ignore any directory named "frontend" within these directories.
- dirs: ["server", "library"]
  patterns: ["*.go", "*.py"]
  excludeDirs: ["frontend", "client"]
# Reload on any changes to the yaml files in the config folder
- dirs: ["config"]
  patterns: ["*.yaml"]
```

### Timeouts

[Timeouts](#timeouts) can be also specified in the configuration file:
```yaml
action: ["go", "run", "main.go"]
changeTimeout: 3s
killTimeout: 5s
```
