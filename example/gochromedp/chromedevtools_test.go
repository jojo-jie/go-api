package gochromedp

import (
	"context"
	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
	"io/ioutil"
	"testing"
)

func TestWebPDF(t *testing.T) {
	// 创建 context
	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	// 生成pdf
	var buf []byte
	if err := chromedp.Run(ctx, printToPDF(`https://www.baidu.com/`, &buf)); err != nil {
		t.Fatal(err)
	}
	if err := ioutil.WriteFile("colobu.pdf", buf, 0644); err != nil {
		t.Fatal(err)
	}
}

func printToPDF(urlstr string, res *[]byte) chromedp.Tasks {
	return chromedp.Tasks{
		chromedp.Navigate(urlstr),
		chromedp.ActionFunc(func(ctx context.Context) error {
			buf, _, err := page.PrintToPDF().WithPrintBackground(false).Do(ctx)
			if err != nil {
				return err
			}
			*res = buf
			return nil
		}),
	}
}
