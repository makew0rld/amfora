#!/usr/bin/env sh

echo "package display\n" > display/thanks.go
echo -n 'var thanks = []byte(`' >> display/thanks.go
cat THANKS.md |tr '`' "'" >> display/thanks.go
echo '`)' >> display/thanks.go

