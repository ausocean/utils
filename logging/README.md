# logging

Package logging provides a "Logger" interface with methods for setting log level
and logging at each level. Two implementations of this interface are provided:
the "JSONLogger" which effectively wraps the zap.SugaredLogger and the "TestLogger"
which is a derivation of type testing.T with additional methods for wrapping the
testing.T.Log method.

# Contributing

See [here](https://github.com/ausocean/utils/src/master/README.md) under "Contributing"
for information on how to contribute.

# License

See [here](https://github.com/ausocean/utils/src/master/README.md) under "License"
for licensing.
