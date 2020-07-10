# dio-exporter

`dio-exporter` is CLI tools for exporting .drawio or .dio file as image file.

## Usage

## Requirement

`dio-exporter` is using [chromedp](https://github.com/chromedp/chromedp) library for capturing screen shot, so you must install `google-chrome` before runnint `dio-exporter`.

https://www.google.com/intl/ja_jp/chrome/

## Build
`make dist`

## Test

`make test`

`Node.js` is needed for running tests, because tests are using [pixelmatch](https://github.com/mapbox/pixelmatch) for comparing exported images and oracle images.