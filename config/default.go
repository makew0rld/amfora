package config

//go:generate ./default.sh
var defaultConf = []byte(`# This is the default config file.
# It also shows all the default values, if you don't create the file.

# All URL values may omit the scheme and/or port.

[a-general]
home = "gemini://gemini.circumlunar.space"
# What command to run to open a HTTP URL. Set to "default" to try to guess the browser,
# or set to "off" to not open HTTP URLs.
# If a command is set, than the URL will be added (in quotes) to the end of the command.
# A space will be prepended if necessary.
http = "default"

search = "gemini://gus.guru/search"  # Any URL that will accept a query string can be put here
color = true  # Whether colors will be used in the terminal
bullets = true  # Whether to replace list asterisks with unicode bullets
wrap_width = 100  # How many columns to wrap a page's text to. Preformatted blocks are not wrapped.

[bookmarks]
# Make sure to quote the key names if you edit this part yourself
# Example:
# "CAPCOM" = "gemini://gemini.circumlunar.space/capcom/"

# Options for page cache - which is only for text/gemini pages
# Increase the cache size to speed up browsing at the expense of memory
[cache]
# Zero values mean there is no limit
max_size = 0  # Size in bytes
max_pages = 30 # The maximum number of pages the cache can store
`)
