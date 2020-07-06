package main

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/chromedp/cdproto/emulation"
	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
)

func capture(drawioURL string, data ConvertData) ([]byte, error) {
	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	var attr = map[string]string{} // attributed rendered diagram
	var res []byte                 // captured screenshot data
	evalStatement := fmt.Sprintf(`window.render(%s).enabled`, data.ToRenderParams())

	// Run
	logf("Start: ")
	if err := chromedp.Run(ctx,
		// 1. Access to drawio page, and wait for "networkIdle" event. (refs: https://github.com/chromedp/chromedp/issues/431#issuecomment-592950397
		enableLifeCycleEvents(),
		navigateAndWaitFor(drawioURL, "networkIdle"),
		// 2. Render dialog
		renderDiagram(evalStatement, &attr),
		// 3. Take screenshot
		takeScreenshot(&attr, &res),
	); err != nil {
		return nil, err
	}
	return res, nil
}

func renderDiagram(statement string, attr *map[string]string) chromedp.Tasks {
	logf("  2. Render a diagram on drawio.")
	debugf("EvalStatement: %s", statement)
	var dummy interface{} // Dummy data that never be used
	return chromedp.Tasks{
		// NOTE: If evaluating `window.render({...})`, then it returns error ("Object reference chain is too long (-32000)") because of deepness of return value. For avoiding this issue, the following code get `enabled` field in return value, but the value never be used
		chromedp.Evaluate(statement, &dummy),
		chromedp.Attributes("#LoadingComplete", attr),
	}
}

type Bounds struct {
	X      float64 `json:"x"`
	Y      float64 `json:"y"`
	Width  int64   `json:"width"`
	Height int64   `json:"height"`
}

func takeScreenshot(attr *map[string]string, res *[]byte) chromedp.Tasks {
	return chromedp.Tasks{
		chromedp.ActionFunc(func(ctx context.Context) error {
			logf("  3. Capture screenshot.")
			// TODO: Might need to adjust https://github.com/jgraph/draw-image-export2/blob/26f8c493648e74010c5051a1c5984c019f3ea07e/export.js#L541
			s := (*attr)["bounds"]
			var bounds Bounds
			if err := json.Unmarshal([]byte(s), &bounds); err != nil {
				return err
			}
			// force viewport emulation
			if err := emulation.SetDeviceMetricsOverride(bounds.Width, bounds.Height, 1, false).WithScreenOrientation(&emulation.ScreenOrientation{
				Type:  emulation.OrientationTypePortraitPrimary,
				Angle: 0,
			}).Do(ctx); err != nil {
				return err
			}
			// capture screenshot
			scale, err := strconv.ParseFloat((*attr)["scale"], 64)
			if err != nil {
				scale = 1
			}

			debugf("Bounds for viewport: %#v", bounds)
			debugf("Scale: %v", scale)
			if *res, err = page.CaptureScreenshot().WithClip(&page.Viewport{
				X:      bounds.X,
				Y:      bounds.Y,
				Width:  float64(bounds.Width),
				Height: float64(bounds.Height),
				Scale:  scale,
			}).Do(ctx); err != nil {
				return err
			}
			return nil
		}),
	}
}

func enableLifeCycleEvents() chromedp.ActionFunc {
	return func(ctx context.Context) error {
		err := page.Enable().Do(ctx)
		if err != nil {
			return err
		}
		err = page.SetLifecycleEventsEnabled(true).Do(ctx)
		if err != nil {
			return err
		}
		return nil
	}
}

func navigateAndWaitFor(url string, eventName string) chromedp.ActionFunc {
	logf("  1. Wait for opening drawio: %s", url)
	return func(ctx context.Context) error {
		_, _, _, err := page.Navigate(url).Do(ctx)
		if err != nil {
			return err
		}

		return waitFor(ctx, eventName)
	}
}

// waitFor blocks until eventName is received.
// Examples of events you can wait for:
//     init, DOMContentLoaded, firstPaint,
//     firstContentfulPaint, firstImagePaint,
//     firstMeaningfulPaintCandidate,
//     load, networkAlmostIdle, firstMeaningfulPaint, networkIdle
//
// This is not super reliable, I've already found incidental cases where
// networkIdle was sent before load. It's probably smart to see how
// puppeteer implements this exactly.
func waitFor(ctx context.Context, eventName string) error {
	// TODO: timeout setting
	ch := make(chan struct{})
	cctx, cancel := context.WithCancel(ctx)
	debugf(`Wait for "%s"`, eventName)
	chromedp.ListenTarget(cctx, func(ev interface{}) {
		switch e := ev.(type) {
		case *page.EventLifecycleEvent:
			debugf("Received lifecycle event: %s", e.Name)
			if e.Name == eventName {
				cancel()
				close(ch)
			}
		}
	})
	select {
	case <-ch:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}

}
