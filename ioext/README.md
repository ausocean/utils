# ioext

Package ioext provides input/output extension functionality, namely, a
"MultiWriteCloser" structure that implements io.WriteCloser and duplicates its
writes for one or more passed io.WriteClosers, similar to the Unix tee(1)
command. Close of the MultiWriteCloser is passed on to all
of the provided io.WriteClosers.

# Contributing

See [here](https://github.com/ausocean/utils/src/master/README.md) under "Contributing"
for information on how to contribute.

# License

See [here](https://github.com/ausocean/utils/src/master/README.md) under "License"
for licensing.
