# Amfora

<center>  <!-- I know, that's not how you usually do it :) -->
<img src="logo.png" alt="amphora logo" width="30%">
<h6>Modified from: amphora by Alvaro Cabrera from the Noun Project</h6>
</center>


[![go reportcard](https://goreportcard.com/badge/github.com/makeworld-the-better-one/amfora)](https://goreportcard.com/report/github.com/makeworld-the-better-one/amfora)
[![license GPLv3](https://img.shields.io/github/license/makeworld-the-better-one/amfora)](https://www.gnu.org/licenses/gpl-3.0.en.html)

<center>  <!-- I know, that's not how you usually do it :) -->
<a href="https://raw.githubusercontent.com/makeworld-the-better-one/amfora/master/demo-large.gif">
<img src="demo-large.gif" alt="Demo GIF" width="80%">
</a>
</center>

Amfora aims to be the best looking [Gemini](https://gemini.circumlunar.space/) client with the most features... all in the terminal. It does not support Gopher or other non-Web protocols - check out [Bombadillo](http://bombadillo.colorfield.space/) for that.

It also aims to be completely cross platform, with full Windows support. If you're on Windows, I would not recommend using the default terminal software. Maybe use [Cmder](https://cmder.net/) instead?

It fully passes Sean Conman's client torture test. It mostly passes the Egsam test.

## Installation

Download a binary from the [releases](https://github.com/makeworld-the-better-one/amfora/releases) page. On POSIX systems you might have to make the binary executable with `chmod +x <filename>`. On Windows, make sure you click "Advanced > Run anyway" after double-clicking, or something like that.

## Usage

Just call `amfora` or `amfora <url> <other-url>` on the terminal. On Windows it might be `amfora.exe` instead.

The project keeps many standard terminal keybindings and is intuitive. Press <kbd>?</kbd> inside the application to pull up the help menu with a list of all the keybindings, and <kbd>Esc</kbd> to leave it. If you have used Bombadillo you will find it similar.

It is designed with large or fullscreen terminals in mind. For optimal usage, make your terminal fullscreen. It was also designed with a dark background terminal in mind, but please file an issue if the colour choices look bad on your terminal setup. 

## Features / Roadmap
Features in italics are in the master branch, but not in the latest release.

- [x] URL browsing with TOFU and error handling
- [x] Tabbed browsing
- [x] Support ANSI color codes on pages, even for Windows
- [x] Styled page content (headings, links)
- [x] Basic forward/backward history, for each tab
- [x] Input (Status Code 10 & 11)
- [x] *Multiple charset support (over 55)*
- [ ] Built-in search using GUS
- [ ] Bookmarks
- [ ] Search in pages with <kbd>Ctrl-F</kbd>
- [ ] Download pages and arbitrary data
- [ ] Full mouse support
- [ ] Emoji favicons
  - See `gemini://mozz.us/files/rfc_gemini_favicon.gmi` for details
- [ ] Table of contents for pages
- [ ] ~~Collapsing of gemini site sections (as determined by headers)~~
- [ ] Full client certificate UX within the client
  - *I will be waiting for some spec changes to happen before implementing this*
  - Create transient and permanent certs within the client, per domain
  - Manage and browse them
  - https://lists.orbitalfox.eu/archives/gemini/2020/001400.html
- [ ] Subscribe to RSS and Atom feeds and display them
- [ ] Support Markdown rendering

## Configuration
The config file is written in the intuitive [TOML](https://github.com/toml-lang/toml) file format. See [default-config.toml](./default-config.toml) for details. By default this file is available at `~/.config/amfora/config.toml`.

On Windows, the file is in `%APPDATA%\amfora\config.toml`, which usually expands to `C:\Users\<username>\AppData\Roaming\amfora\config.toml`.

## Libraries
Amfora ❤️ open source!

- [cview](https://gitlab.com/tslocum/cview/) for the TUI
  - It's a fork of [tview](https://github.com/rivo/tview) with PRs merged and active support
  - It uses [tcell](https://github.com/gdamore/tcell) for low level terminal operations
- [Viper](https://github.com/spf13/viper) for configuration and TOFU storing
- [go-gemini](https://github.com/makeworld-the-better-one/go-gemini), my forked and updated Gemini client/server library

## License
This project is licensed under the GPL v3.0. See the [LICENSE](./LICENSE) file for details.
