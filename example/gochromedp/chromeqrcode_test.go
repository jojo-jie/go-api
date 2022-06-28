package gochromedp

import (
	"bytes"
	"context"
	"fmt"
	"github.com/chromedp/chromedp"
	"github.com/makiuchi-d/gozxing"
	"github.com/makiuchi-d/gozxing/qrcode"
	goQrcode "github.com/skip2/go-qrcode"
	"image"
	"io/ioutil"
	"testing"
	"time"
)

func TestQrcodeLogin(t *testing.T) {
	ctx, _ := chromedp.NewExecAllocator(
		context.Background(),
		append(
			chromedp.DefaultExecAllocatorOptions[:],
			chromedp.Flag("headless", false),
		)...,
	)
	ctx, _ = context.WithTimeout(ctx, 30*time.Second)
	ctx, _ = chromedp.NewContext(ctx, chromedp.WithDebugf(t.Logf))
	//defer cancel()
	if err := chromedp.Run(ctx, myTasks()); err != nil {
		t.Fatal(err)
	}
}

const loginURL = "https://account.wps.cn/"

func myTasks() chromedp.Tasks {
	return chromedp.Tasks{
		// open login view
		chromedp.Navigate(loginURL),
		// click wechat login button
		chromedp.Click(`#wechat > span:nth-child(2)`),
		// click confirm button
		chromedp.Click(`#dialog > div.dialog-wrapper > div > div.dialog-footer > div.dialog-footer-ok`),

		getCode(),
	}
}

func getCode() chromedp.ActionFunc {
	return func(ctx context.Context) (err error) {
		var code []byte
		if err = chromedp.Screenshot(`#wximport`, &code, chromedp.ByID).Do(ctx); err != nil {
			return
		}
		return printQRCode(code)
		//return saveQRCodeImg(code)
	}
}

func printQRCode(code []byte) (err error) {
	img, _, err := image.Decode(bytes.NewReader(code))
	if err != nil {
		return
	}
	bmp, err := gozxing.NewBinaryBitmapFromImage(img)
	if err != nil {
		return
	}
	res, err := qrcode.NewQRCodeReader().Decode(bmp, nil)
	if err != nil {
		return
	}
	qr, err := goQrcode.New(res.String(), goQrcode.High)
	if err != nil {
		return
	}
	fmt.Println(qr.ToSmallString(false))
	return
}

func saveQRCodeImg(code []byte) (err error) {
	return ioutil.WriteFile("code.png", code, 0755)
}
