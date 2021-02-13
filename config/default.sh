#!/usr/bin/env sh

cat > default.go <<-EOF
package config

//go:generate ./default.sh
EOF
echo -n 'var defaultConf = []byte(`' >> default.go
cat ../default-config.toml >> default.go
echo '`)' >> default.go
