# Notes

- All the maps and stuff could be replaced with a `tab` struct
- And then just one single map of tab number to `tab`

## Bugs
- Wrapping is messed up on CHAZ post, but nothing else
  - Filed [issue 23](https://gitlab.com/tslocum/cview/-/issues/23)
- Text background not reset on ANSI pages
  - Filed [issue 25](https://gitlab.com/tslocum/cview/-/issues/25)
- Modal styling messed up when wrapped - example occurence is the error modal for a long unsupported scheme URL
  - Filed [issue 26](https://gitlab.com/tslocum/cview/-/issues/26)
  - Add some bold back into modal text after this is fixed