#!/usr/bin/env bash

head -n 1 default.go | tee default.go > /dev/null
echo -n 'var defaultConf = []byte(`' >> default.go
cat ../default-config.toml >> default.go
echo '`)' >> default.go