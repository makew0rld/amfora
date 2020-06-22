# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]
### Added
- Support over 55 charsets (#3)
- Add titles to error modals
- Store ports in TOFU database (#7)
- Search from bottom bar
- Wrapping based on terminal width (#1)
- `left_margin` config option
- Right margin for text (#1)
- Desktop entry file
- Option to continue anyway when cert doesn't match TOFU database
- Display all `text/*` documents, not just gemini and plain (#12)

### Changed
- Connection timeout is 15 seconds (was 5s)
- Hash `SubjectPublicKeyInfo` for TOFU instead (#7)
- `wrap_width` config option became `max_width`

### Removed
- Opening multiple URLs from the command line (threading issues)

### Fixed
- Reset bottom bar on error / invalid URL
- Side scrolling doesn't cut off text on the left side (#1)
- Mark status code 21 as invalid


## [1.0.0] - 2020-06-18
Initial release.

### Added
- Tabbed browsing
- TOFU
- Styled content
- Basic history for each tab
- Input
