#!/usr/bin/env sh

echo "package display\n" > license.go
echo -n 'var license = []byte(`' >> license.go
cat ../LICENSE |tr '`' "'" >> license.go
echo '`)' >> license.go

