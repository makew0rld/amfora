# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]
### Added
- Alt-Left and Alt-Right for history navigation (#23)
- You can type `..` in the bottom bar to go up a directory in the URL (#21)
- Error popup for when input string would result in a too long out-of-spec URL (#25)
- Paging, using <kbd>d</kbd> and <kbd>u</kbd>, as well as <kbd>Page Up</kbd> and <kbd>Page Down</kbd> (#19)
- <kbd>Esc</kbd> can exit link highlighting mode (#24)
- Selected link URL is displayed in the bottom bar (#24)
- Pressing <kbd>Ctrl-T</kbd> with a link selected opens it in a new tab (#27)
- Writing `new:N` in the bottom bar will open link number N in a new tab (#27)

### Changed
- Bottom bar now says `URL/Num./Search: ` when space is pressed
- Update to [go-gemini](https://github.com/makeworld-the-better-one/go-gemini) v0.6.0
- Help layout doesn't have borders anymore

### Fixed
- Actual unicode bullet symbol is used for lists: U+2022
- Performance when loading very long cached pages improved (#26)
- Doesn't crash when wrapping certain complex lines (#20)
- Input fields are always in focus when they appear (#5)
- Reloading the new tab page doesn't cause an error popup
- Help table cells are hardwrapped so the text can still be read entirely on an 80-column terminal

## [1.1.0] - 2020-06-24
### Added
- **Bookmarks** (#10)
- **Support over 55 charsets** (#3)
- **Search using the bottom bar**
- Add titles to all modals
- Store ports in TOFU database (#7)
- Search from bottom bar
- Wrapping based on terminal width (#1)
- `left_margin` config option (#1)
- Right margin for text (#1)
- Desktop entry file
- Option to continue anyway when cert doesn't match TOFU database
- Display all `text/*` documents, not just gemini and plain (#12)
- Prefer XDG environment variables if they're set, to specify config dir, etc (#11)
- Version and help commands - `-v`, `--version`, `--help`, `-h` (#14)

### Changed
- Connection timeout is 15 seconds (was 5s)
- Hash `SubjectPublicKeyInfo` for TOFU instead (#7)
- `wrap_width` config option became `max_width` (#1)
- Make the help table look better

### Removed
- Opening multiple URLs from the command line

### Fixed
- Reset bottom bar on error / invalid URL
- Side scrolling doesn't cut off text on the left side (#1)
- Mark status code 21 as invalid
- Bottom bar is not in focus after clicking Enter
- Badly formed links on pages can no longer crash the browser
- Disabling color in config affects UI elements (#16)
- Keep bold for headings even with color disabled
- Don't make whole link text bold when color is disabled
- Get domain from URL for TOFU, not from certificate


## [1.0.0] - 2020-06-18
Initial release.

### Added
- Tabbed browsing
- TOFU
- Styled content
- Basic history for each tab
- Input
