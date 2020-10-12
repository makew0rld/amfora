# Notes

## Issues
- URL for each tab should not be stored as a string - in the current code there's lots of reparsing the URL


## Regressions


## Upstream Bugs
- Wrapping messes up on brackets
  - Filed [issue 23](https://gitlab.com/tslocum/cview/-/issues/23)
- Wrapping panics on strings with brackets and Asian characters
  - Filed cview [issue 27](https://gitlab.com/tslocum/cview/-/issues/27)
  - The panicking was reported and fixed in Amfora [issue 20](https://github.com/makeworld-the-better-one/amfora/issues/20), but the lines are now just not wrapped
- Text background not reset on ANSI pages
  - Filed [issue 25](https://gitlab.com/tslocum/cview/-/issues/25)
- Modal styling messed up when wrapped - example occurence is the error modal for a long unsupported scheme URL
  - Filed [issue 26](https://gitlab.com/tslocum/cview/-/issues/26)
  - Add some bold back into modal text after this is fixed
- Bookmark keys aren't deleted, just set to `""`
  - Waiting on [this viper PR](https://github.com/spf13/viper/pull/519) to be merged
- Help table cells aren't dynamically wrapped
  - Filed [issue 29](https://gitlab.com/tslocum/cview/-/issues/29)
