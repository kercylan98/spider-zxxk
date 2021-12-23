package browser

import (
	"context"
	"github.com/chromedp/chromedp"
	"testing"
)

func TestBrowser_Custom(t *testing.T) {
	b := New()
	var content string

	err := b.Custom(context.Background(), func() []chromedp.Action {
		return []chromedp.Action{
			chromedp.Navigate(`https://www.zxxk.com`),
			chromedp.WaitVisible(`#indexnva > div`),
			chromedp.OuterHTML(`#indexnva > div`, &content),

		}
	})

	if err != nil {
		t.Fatal(err)
	}
	t.Log(content)
}
