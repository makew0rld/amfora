#!/usr/bin/env sh

cat > thanks.go <<-EOF
//nolint
package display

//go:generate ./thanks.sh
EOF
echo -n 'var thanks = []byte(`' >> thanks.go
cat ../THANKS.md | tr '`' "'" >> thanks.go
echo '`)' >> thanks.go

