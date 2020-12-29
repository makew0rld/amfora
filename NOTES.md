# Notes

## Issues
- URL for each tab should not be stored as a string - in the current code there's lots of reparsing the URL

## Upstream Bugs
- Bookmark keys aren't deleted, just set to `""`
  - Waiting on [this viper PR](https://github.com/spf13/viper/pull/519) to be merged
- Help table cells aren't dynamically wrapped
  - Filed [issue 29](https://gitlab.com/tslocum/cview/-/issues/29)
