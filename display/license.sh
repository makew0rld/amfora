#!/usr/bin/env sh

cat > license.go <<-EOF
package display

//go:generate ./license.sh
EOF
echo -n 'var license = []byte(`' >> license.go
cat ../LICENSE | tr '`' "'" >> license.go
echo '`)' >> license.go

