module github.com/makeworld-the-better-one/amfora

go 1.15

require (
	code.rocketnine.space/tslocum/cview v1.5.6-0.20210530175404-7e8817f20bdc
	github.com/atotto/clipboard v0.1.4
	github.com/dustin/go-humanize v1.0.0
	github.com/gdamore/tcell/v2 v2.3.3
	github.com/makeworld-the-better-one/go-gemini v0.11.0
	github.com/mitchellh/go-homedir v1.1.0
	github.com/mmcdole/gofeed v1.1.2
	github.com/muesli/termenv v0.9.0
	github.com/rkoesters/xdg v0.0.0-20181125232953-edd15b846f9b
	github.com/schollz/progressbar/v3 v3.8.0
	github.com/spf13/viper v1.9.0
	github.com/stretchr/testify v1.7.0
	golang.org/x/text v0.3.7
)

replace github.com/atotto/clipboard => github.com/awfulcooking/clipboard v0.1.5-0.20211201163140-3a50b14162df

replace github.com/gdamore/tcell/v2 => github.com/awfulcooking/tcell/v2 v2.4.1-0.20211128170204-5ebcb5571e5d

replace github.com/makeworld-the-better-one/go-gemini => github.com/awfulcooking/go-gemini v0.11.2-0.20211202044711-498169b7378e

replace golang.org/x/term => github.com/awfulcooking/term v0.0.0-20211128155416-2652f7c0d88b
