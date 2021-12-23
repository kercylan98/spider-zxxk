package spider

import (
	"context"
	"github.com/PuerkitoBio/goquery"
	"github.com/chromedp/chromedp"
	"strings"
)

type Stage struct {
	Name string
	LevelId string
}

// GetStages 获取学科网科目学段
func (slf *Spider) GetStages() ([]Stage, error) {
	var (
		document *goquery.Document
		html string
		err error
		sel = `#indexnva > div > div.tab`
	)

	err = slf.browser.Custom(context.Background(), func() []chromedp.Action {
		return []chromedp.Action{
			chromedp.Navigate(`https://www.zxxk.com`),
			chromedp.WaitVisible(sel),
			chromedp.OuterHTML(sel, &html),
		}
	})
	if err != nil {
		return nil, err
	}
	document, err = slf.browser.ToDocument(html)
	if err != nil {
		return nil, err
	}

	var stages []Stage

	document.Find(`.name`).Each(func(i int, selection *goquery.Selection) {
		stages = append(stages, Stage{
			Name:    strings.TrimSpace(selection.Text()),
			LevelId: selection.AttrOr("data-levelid", ""),
		})
	})

	return stages, nil
}