# filemap

Package filemap provides functions for manipulating maps stored in
files. The major delimiter separates key-value pairs and the minor
delimiter separates keys from corresponding values. For example,
using newline as the major delimiter and space as the minor
delimiter, the file containing the following:

key1 value1\nkey2 value2a,value2b,value2c\nkey3 \n"

is represented as the following map:
{
  "key1": "value1",
  "key2": "value2a,value2b,value2c",
  "key3": "",
}

NB: Keys and values must not contain strings used as delimiters.

# Contributing

See [here](https://github.com/ausocean/utils/src/master/README.md) under "Contributing"
for information on how to contribute.

# License

See [here](https://github.com/ausocean/utils/src/master/README.md) under "License"
for licensing.
