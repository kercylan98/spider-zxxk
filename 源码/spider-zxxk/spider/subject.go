package spider

import (
	"context"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/chromedp/chromedp"
	"strings"
)

type Subject struct {
	Name string
	Url string
}

// GetSubjects 特定年级的获取所有学科
func (slf *Spider) GetSubjects(stageLevelId string) ([]Subject, error) {
	var (
		document *goquery.Document
		html string
		err error
		sel = `#indexnva > div > div.tab-wrap`
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

	var subjects []Subject
	// 特定学段
	document.Find(fmt.Sprintf(`ul[data-levelid="%s"]`, stageLevelId)).
		// 获取学科url
		Find("li").Each(func(i int, selection *goquery.Selection) {
		var name string
		var url string
		if _, exist := selection.Attr(`data-subjectid`); exist {
			aTag := selection.Find(`.subject-name`)
			name = strings.TrimSpace(aTag.Find(`.name`).Text())
			url = strings.TrimSpace(aTag.AttrOr("href", ""))

			slice := strings.Split(url, "/books")
			if stageLevelId != "4" {
				subjects = append(subjects, Subject{
					Name: name,
					Url:  strings.ReplaceAll(slice[0] + "/books", "//books", "/books"),
				})
			}else {
				subjects = append(subjects, Subject{
					Name: name,
					Url:  strings.ReplaceAll(strings.ReplaceAll(slice[0] + "/books", "//books", "/books"), ".com/books", ".com/h/books/"),
				})
			}
		}else {
			selection.Find(`.subject-name`).Find(`a`).Each(func(i int, selection *goquery.Selection) {
				name = strings.TrimSpace(selection.Text())
				url = strings.TrimSpace(selection.AttrOr("href", ""))

				slice := strings.Split(url, "/books")

				if stageLevelId != "4" {
					subjects = append(subjects, Subject{
						Name: name,
						Url:  strings.ReplaceAll(slice[0] + "/books", "//books", "/books"),
					})
				}else {
					subjects = append(subjects, Subject{
						Name: name,
						Url: strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(slice[0] + "/books", "//books", "/books"), ".com/books", ".com/h/books/"), ".com/books", ".com/h/books/") ,
					})
				}
			})
		}


	})

	return subjects, nil
}