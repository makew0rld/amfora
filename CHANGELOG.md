# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.8.0] - 2021-02-17
### Added
- **Media type handlers** - open non-text files in another application (#121, #134)
- Ability to set custom keybindings in config (#135)
- Added scrollbar, by default only appears on pages that go off-screen (#89, #107)
- More internal about pages, see `about:about` (#160, #187)

### Changed
- Update cview to `d776e728ef6d2a9990a5cd86a70b31f0678613e2` for large performance and feature updates (#107)
- Update to tcell v2 (dependency of cview)
- Display page even if mediatype params are malformed (#141)
- Sensitive input fields (status code 11) display with asterisks over the text (#106)

### Fixed
- Don't use cache when URL is typed in bottom bar (#159)
- Fix downloading of pages that are too large or timed out
- `about:` URLs can be typed into the bottom bar (#167)
- Bookmarks modal closes on ESC like the others (#173)
- Handle empty META string (#176)
- Whitespace around the URL entered in the bottom bar is stripped (#184)
- Don't break visiting IPv6 hosts when port 1965 is specified (#195)
- More reliable start, no more flash of unindented text, or text that stays unindented (#107)
- Pages with ANSI resets don't use the terminal's default text and background colors (#107)
- ANSI documents don't leak color into the left margin (#107)
- Rendering very long documents is now ~96% faster, excluding gemtext parsing (#26, #107)
- Due to that same change, less memory is used per-page (#26, #107)


## [1.7.2] - 2020-12-21
### Fixed
- Viewing subscriptions after subscribing to a certain user page won't crash Amfora (#157)


## [1.7.1] - 2020-12-21
### Fixed
- Fixed bug that caused Amfora to crash when subscribing to a page (#151)


## [1.7.0] - 2020-12-20
### Added
- **Subscriptions** to feeds and page changes (#61)
- Opening local files with `file://` URIs (#103, #117)
- `show_link` option added in config to optionally see the URL (#133)
- Support for Unicode in domain names (IDNs)
- Unnecessarily encoded characters in URLs will be decoded (#138)
- URLs are NFC-normalized before any processing (#138)
- Links to the wiki in the new tab
- Cache times out after 30 minutes by default (#110)
- `about:version` page (#126)

### Changed
- Updated [go-gemini](https://github.com/makeworld-the-better-one/go-gemini) to v0.11.0
  - Supports CN-only wildcard certs
  - Time out when header takes too long
- Preformatted text is now light yellow by default
- Downloading a file no longer uses a second request
- You can go back to the new tab page in history (#96)

### Fixed
- Single quotes are used in the default config for commands and paths so that Windows paths with backslashes will be parsed correctly
- Downloading now uses proxies when appropriate
- User-entered URLs with invalid characters will be percent-encoded (#138)
- Custom downloads dir is actually used (#148)
- Empty quote lines no longer disappear


## [1.6.0] - 2020-11-04
### Added
- **Support client certificates** through config (#112)
- `ansi` config setting, to disable ANSI colors in pages (#79, #86)
- Edit current URL with <kbd>e</kbd> (#87)
- If `emoji_favicons` is enabled, new bookmarks will have the domain's favicon prepended (#69, #90)
- The `BROWSER` env var is now also checked when opening web links on Unix (#93)
- More accurate error messages based on server response code

### Changed
- Disabling the `color` config setting also disables ANSI colors in pages (#79, #86)
- Updated [go-isemoji](https://github.com/makeworld-the-better-one/go-isemoji) to v1.1.0 to support Emoji 13.1 for favicons
- The web browser code doesn't check for Xorg anymore, just display variables (#93)
- Bookmarks can be made to non-gemini URLs (#94)
- Remove pointless directory fallbacks (#101)
- Don't load page from cache when redirected to it (#114)

### Fixed
- XDG user dir file is parsed instead of looking for XDG env vars (#97, #100)
- Support paths with spaces in HTTP browser config setting (#77)
- Clicking "Change" on an existing bookmark without changing the text no longer removes it (#91)
- Display HTTP Error if "Open In Portal" fails (#81)
- Support ANSI color codes again, but only in preformatted blocks (#59)
- Make the `..` command work lke it used to in v1.4.0


## [1.5.0] - 2020-09-01
### Added
- **Proxy support** - see the `[proxies]` section in the config (#66, #80)
- **Emoji favicons** can now be seen if `emoji_favicons` is enabled in the config (#62)
- `shift_numbers` key in the config was added, so that non US keyboard users can navigate tabs (#64)
- <kbd>F1</kbd> and <kbd>F2</kbd> keys for navigating to the previous and next tabs (#64)
- Resolving any relative path (starts with a `.`) in the bottom bar is supported, not just `..` (#71)
- You can now set external programs in the config to open other schemes, like `gopher://` or `magnet:` (#74)
- Auto-redirecting can be enabled - redirect within Gemini up to 5 times automatically (#75) 
- Help page now documents paging keys (#78)
- The new tab page can be customized by creating a gemtext file called `newtab.gmi` in the config directory (#67, #83)

### Changed
- Update to [go-gemini](https://github.com/makeworld-the-better-one/go-gemini) v0.8.4

### Fixed
- Two digit (and higher) link texts are now in line with one digit ones (#60)
- Race condition when reloading pages that could have caused the cache to still be used
- Prevent panic (crash) when the server sends an error with an empty meta string (#73)
- URLs with with colon-only schemes (like `mailto:`) are properly recognized
- You can no longer navigate through the history when the help page is open (#55, #78)


## [1.4.0] - 2020-07-28
### Added
- **Theming** - check out [default-config.toml](./default-config.toml) for details (#46)
- <kbd>Tab</kbd> now also enters link selecting mode, like <kbd>Enter</kbd> (#48)
- Number keys can be pressed to navigate to links 1 through 10 (#47)
- Permanent redirects are cached for the session (#22)
- `.ansi` is also supported for `text/x-ansi` files, as well as the already supported `.ans`

### Changed
- Documented <kbd>Ctrl-C</kbd> as "Hard quit"
- Updated [cview](https://gitlab.com/tslocum/cview/) to latest commit: `cc7796c4ca44e3908f80d93e92e73694562d936a`
- The bottom bar label now uses the same color as the tabs at the top
- Tab and blue link colors were changed very slightly to be part of the 256 Xterm colors, for better terminal support

### Fixed
- You can't change link selection while the page is loading
- Only one request is made for each URL - `v1.3.0` accidentally made two requests each time (#50)
- Using the `..` command doesn't keep the query string (#49)
- Any error that occurs when downloading a file will be displayed, and the partially downloaded file will be deleted
- Allow for opening a new tab while the current one is loading
- Pressing Escape after typing in the bottom bar no longer jumps you back to the top of the page
- Repeated redirects where the last one is cancelled by the user doesn't leave the `Loading...` text in the bottom bar (#53)


## [1.3.0] - 2020-07-10
### Added
- **Downloading content** (#38)
- Configurable page size limit - `page_max_size` in config (#30)
- Configurable page timeout - `page_max_time` in config
- Link and heading lines are wrapped just like regular text lines
- Wrapped list items are indented to stay behind the bullet (#35)
- Certificate expiry date is stored when the cert IDs match (#39)
- What link was selected is remembered as you browse through history
- Render ANSI codes in `text/x-ansi` pages, or text pages that end with `.ans` (#45)

### Changed
- Pages are rewrapped dynamically, whenever the terminal size changes (#33)
- TOFU warning message mentions how long the previous cert was still valid for (#34)

### Fixed
- Many potential network and display race conditions eliminated
- Whether a tab is loading stays indicated when you switch away from it and go back
- Plain text documents are displayed faithfully (there were some edge conditions)
- Opening files in portal.mozz.us uses the `http` setting in the config (#42)


## [1.2.0] - 2020-07-02
### Added
- Alt-Left and Alt-Right for history navigation (#23)
- You can type `..` in the bottom bar to go up a directory in the URL (#21)
- Error popup for when input string would result in a too long out-of-spec URL (#25)
- Paging, using <kbd>d</kbd> and <kbd>u</kbd>, as well as <kbd>Page Up</kbd> and <kbd>Page Down</kbd> (#19)
- <kbd>Esc</kbd> can exit link highlighting mode (#24)
- Selected link URL is displayed in the bottom bar (#24)
- Pressing <kbd>Ctrl-T</kbd> with a link selected opens it in a new tab (#27)
- Writing `new:N` in the bottom bar will open link number N in a new tab (#27)
- Quote lines are now in italics (#28)

### Changed
- Bottom bar now says `URL/Num./Search: ` when space is pressed
- Update to [go-gemini](https://github.com/makeworld-the-better-one/go-gemini) v0.6.0
- Help layout doesn't have borders anymore
- Pages with query strings are still cached (#29)
- URLs or searches typed in the bottom bar are not loaded from the cache (#29)

### Fixed
- Actual unicode bullet symbol is used for lists: U+2022
- Performance when loading very long cached pages improved (#26)
- Doesn't crash when wrapping certain complex lines (#20)
- Input fields are always in focus when they appear (#5)
- Reloading the new tab page doesn't cause an error popup
- Help table cells are hardwrapped so the text can still be read entirely on an 80-column terminal
- New tab text is wrapped to terminal width like other pages (#31)
- TOFU "continue anyway" popup has a question mark at the end


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
