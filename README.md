# dio-exporter

`dio-exporter` is a CLI tool for exporting .drawio or .dio file as image file.

## Usage

Find drawio files (`.drawio` or `.dio`) from `./data` directory recursivelly, convert files to `.png` or `.svg`, and write converted files to `./dist`
```
$ dio-exporter-vX.X.X-${GOOS}-${GOARCH} \
  -in     ./data/ \
  -out    ./dist/ \
  -format png
```

Run drawio server locally

```
$ dio-exporter-vX.X.X-${GOOS}-${GOARCH} -debug-server
```

## Requirement

`dio-exporter` is using [chromedp](https://github.com/chromedp/chromedp) library for capturing screen shot, so you must install `google-chrome` before runnint `dio-exporter`.

https://www.google.com/intl/ja_jp/chrome/

## Build
`make dist`

## Test
`make test`


<!-- This feature doesn't used for now, becuase of problems about fonts
`Node.js` is needed for running tests, because tests are using [pixelmatch](https://github.com/mapbox/pixelmatch) for comparing exported images and oracle images.
-->