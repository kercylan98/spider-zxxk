package spider

import (
	"context"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/chromedp/chromedp"
	"strings"
)

// Resource 学科资源
type Resource struct {
	Domain string
	Name   string
	Url 	string
}

func (slf *Spider) GetResourceChildren(resource Resource) ([]Resource, error) {
	var (
		document *goquery.Document
		html string
		err error
		sel = `.body-tree`
	)

	//fmt.Println(resource.Name, "https://" + resource.Domain + resource.Url)
	err = slf.browser.Custom(context.Background(), func() []chromedp.Action {
		return []chromedp.Action{
			chromedp.Navigate("https://" + resource.Domain + resource.Url),
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

	var resources []Resource
	document.Find(fmt.Sprintf(`*[title='%s']`, resource.Name)).Each(func(i int, selection *goquery.Selection) {

		selection.Parent().Parent().Parent().Find(`ul`).Find("li").Each(func(i int, selection *goquery.Selection) {
			if selection.Parent().Prev().Find(fmt.Sprintf(`*[title='%s']`, resource.Name)).Size() == 0 {
				return
			}

			el := selection.Find(`a`).First()
			name := strings.TrimSpace(el.Text())
			url  := strings.TrimSpace(el.AttrOr("href", ""))
			if url != "" && url != resource.Url {
				resources = append(resources, Resource{
					Name:   name,
					Url:    url,
					Domain: resource.Domain,
				})
			}
		})
	})


	return resources, nil
}

func (slf *Spider) GetResource(subjectUrl string) ([]Resource, error) {
	var (
		document *goquery.Document
		html string
		err error
		sel = `body > div.sx_main > div.list-body.clearfix > div.body-tree > div > ul`
	)

	err = slf.browser.Custom(context.Background(), func() []chromedp.Action {
		return []chromedp.Action{
			chromedp.Navigate(subjectUrl),
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

	var resources []Resource
	document.Find(`li`).Each(func(i int, selection *goquery.Selection) {
		el := selection.Find(`div`).Find(`span`).Find(`a`)
		name := strings.TrimSpace(el.Text())
		url  := strings.TrimSpace(el.AttrOr("href", ""))
		if url != "" {
			replace := strings.ReplaceAll(subjectUrl, "https://", "")
			resources = append(resources, Resource{
				Name:   name,
				Url:    url,
				Domain: strings.Split(replace, "/")[0],
			})
		}
	})

	return resources, nil
}