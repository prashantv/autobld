# autobld
autobld is a utility to automatically recompile your server application when changes are made to your source code.

autobld can also set up a proxy for your server's listening ports that will block while the server reloads. This avoids you constantly refreshing till the server is up.

## Simple use case

If all you want is to automatically restart the server for you when any change is made to any file in the current directory, then do:

```
autobld python test.py
```

However, you may want to be more selective, and only restart the server when code changes (e.g. any python files). You can specify file patterns to watch by using the --match or -m flag.
```
autobld -m *.py python test.py
```

## More cmplex use cases
For more complex configurations, it may make sense to set up a YAML config.

TODO: document yaml config here.
