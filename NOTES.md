# Notes

## Issues
- URL for each tab should not be stored as a string - in the current code there's lots of reparsing the URL

## Upstream Bugs
- Bookmark keys aren't deleted, just set to `""`
  - Waiting on [this viper PR](https://github.com/spf13/viper/pull/519) to be merged
- [ANSI conversion is messed up](https://code.rocketnine.space/tslocum/cview/issues/48)
- [WordWrap is broken in some cases](https://code.rocketnine.space/tslocum/cview/issues/27) - close #156 if this is fixed
- [Prevent panic when reformatting](https://code.rocketnine.space/tslocum/cview/issues/50) - can't reliably reproduce or debug
- [Unicode bullet symbol mask causes issues with PasswordInput](https://code.rocketnine.space/tslocum/cview/issues/55)


## Upstream PRs
