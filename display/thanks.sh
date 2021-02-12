#!/usr/bin/env sh

echo "package display\n" > thanks.go
echo -n 'var thanks = []byte(`' >> thanks.go
cat ../THANKS.md |tr '`' "'" >> thanks.go
echo '`)' >> thanks.go

