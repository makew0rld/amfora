#!/usr/bin/env sh

echo "package display\n" > display/license.go
echo -n 'var license = []byte(`' >> display/license.go
cat LICENSE |tr '`' "'" >> display/license.go
echo '`)' >> display/license.go

