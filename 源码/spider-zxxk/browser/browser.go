package browser

import (
	"context"
	"github.com/PuerkitoBio/goquery"
	"github.com/chromedp/chromedp"
	"log"
	"strings"
	"time"
)

func New() *Browser {
	var (
		browser = new(Browser)
	)
	browser.Timeout = 15 * time.Second

	return browser
}

// Browser 浏览器
type Browser struct {
	Timeout  		time.Duration
}

// Custom 自定义操作流程
func (slf *Browser) Custom(c context.Context, actions func() []chromedp.Action ) error {
	ac, acCancel := chromedp.NewExecAllocator(c, append(chromedp.DefaultExecAllocatorOptions[:], []chromedp.ExecAllocatorOption {
		chromedp.Flag("headless", true), // debug使用
		chromedp.Flag("blink-settings", "imagesEnabled=false"),
		chromedp.UserAgent(`Mozilla/5.0 (Windows NT 6.3; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/73.0.3683.103 Safari/537.36`),
	}...)...)
	defer acCancel()
	bc, cbCancel := chromedp.NewContext(ac, chromedp.WithLogf(log.Printf))
	defer cbCancel()
	ctx, cancel := context.WithTimeout(bc, slf.Timeout)
	defer cancel()
	if err := chromedp.Run(ctx, actions()...); err != nil {
		return err
	}
	return nil
}

// ToDocument 转换为Document
func (slf *Browser) ToDocument(html string) (*goquery.Document, error) {
	return goquery.NewDocumentFromReader(strings.NewReader(html))
}