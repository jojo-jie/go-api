package gochromedp

import (
	"context"
	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"
	"io/ioutil"
	"math/rand"
	"os"
	"path/filepath"
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

func TestBarPDF(t *testing.T) {
	generateEcharts()
	ctx, cancel := chromedp.NewContext(context.Background(), chromedp.WithDebugf(t.Logf))
	defer cancel()
	// 定义tasks
	elementScreenshot := func(urlstr, sel string, res *[]byte) chromedp.Tasks {
		return chromedp.Tasks{
			chromedp.Navigate(urlstr),
			chromedp.Screenshot(sel, res, chromedp.NodeVisible),
		}
	}

	// 生成截图
	var buf []byte
	barFile, err := filepath.Abs("./bar.html")
	if err != nil {
		t.Fatal(err)
	}
	buf, _ = os.ReadFile(barFile)
	if err := chromedp.Run(ctx, elementScreenshot(`file://`+barFile, `canvas`, &buf)); err != nil {
		t.Fatal(err)
	}

	// 将截图写到文件中
	if err := ioutil.WriteFile("bar.png", buf, 0o644); err != nil {
		t.Fatal(err)
	}
}

func generateEcharts() {
	bar := charts.NewBar()
	// set some global options like Title/Legend/ToolTip or anything else
	bar.SetGlobalOptions(charts.WithTitleOpts(opts.Title{
		Title:    "sexy bar",
		Subtitle: "Ghost",
	}))

	// Put data into instance
	bar.SetXAxis([]string{"Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun"}).
		AddSeries("Category A", generateBarItems()).
		AddSeries("Category B", generateBarItems())

	// Where the magic happens
	f, _ := os.Create("bar.html")
	bar.Render(f)
}
func generateBarItems() []opts.BarData {
	items := make([]opts.BarData, 0)
	for i := 0; i < 7; i++ {
		items = append(items, opts.BarData{
			Value: rand.Intn(300),
		})
	}
	return items
}
