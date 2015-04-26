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

The full list of flags supported are:

Short | Long       | Description 
---   | ---        | ---         
-c     | --config     | Path to the configuration file.
-v     | --verbose    | Verbose logging, can be specified multiple times
-q     | --quiet      | Quiet mode, disables all logging 
-d     | --dir        | Directory to execute the commands in (by default, the current directory).
-m     | --match      | File patterns to match (by default, "*")
-x     | --excludeDir | Directories to exclude from watching (by default, "*.git", "*.hg")
-p     | --proxy      | List of proxy ports to set up. See [Proxy](#proxies) for more information.


## Proxies

Proxy ports can be used to avoid connections failing while the server is being reloaded. Proxy ports will attempt to try connect to the target port for a minute before giving up. There are two types of proxy ports: TCP and HTTP.

TCP proxy ports do not modify the request in any way, and can be used for any TCP protocol.

HTTP proxy ports act as a HTTP proxy, and will do things like use a correct `Host` header, and allow for custom path prefixes. E.g. if you set up a proxy using `-p http:9090:8080/server`, then a request made to `localhost:9090/test` will actually be forwarded to `localhost:8080/server/test`.

## Configuration file
A YAML configuration file can be used using the `--config` (or `-c` for short) flag. When a configuration file is specified, configuration flags are ignored.

Sample configuration files can be seen in the [test](test) directory.

The YAML file 

For more complex configurations, it may make sense to set up a YAML config.

TODO: document yaml config here.
