# Notes

- All the maps and stuff could be replaced with a `tab` struct
- And then just one single map of tab number to `tab`

## Bugs
- Wrapping is messed up on CHAZ post, but nothing else
  - Filed [issue 23](https://gitlab.com/tslocum/cview/-/issues/23)
- Error modal doesn't show the title
  - Filed [issue 24](https://gitlab.com/tslocum/cview/-/issues/24)
- Text background not reset on ANSI pages
  - Filed [issue 25](https://gitlab.com/tslocum/cview/-/issues/25)
- Inputfield isn't repeatedly in focus
  - Tried multiple focus options with App and Form funcs, but nothing worked

## Small todos
- Look at other todos in code
- Add "Why the name amfora" thing to README
- Add GIF to README
- Pass `gemini://egsam.pitr.ca/` test
  - Timeout for server not closing connection?
