package gochromedp

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
	"github.com/makiuchi-d/gozxing"
	"github.com/makiuchi-d/gozxing/qrcode"
	goQrcode "github.com/skip2/go-qrcode"
	"image"
	"os"
	"strings"
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
	ctx, _ = context.WithTimeout(ctx, 300*time.Second)
	ctx, _ = chromedp.NewContext(ctx, chromedp.WithDebugf(t.Logf))
	//defer cancel()
	if err := chromedp.Run(ctx, myTasks()); err != nil {
		err := chromedp.Run(ctx, myReadyLogin())
		if err != nil {
			t.Fatal(err)
		}
	} else {
		t.Log("success !!!!")
	}

}

const loginURL = "https://account.wps.cn/"

const userCenterURl = "https://account.wps.cn/usercenter/apps"

func myTasks() chromedp.Tasks {
	return chromedp.Tasks{
		loadCookies(),
		chromedp.Navigate(userCenterURl),
		checkLoginStatus(),
	}
}

func myReadyLogin() chromedp.Tasks {
	return chromedp.Tasks{
		// open login view
		chromedp.Navigate(loginURL),
		// click wechat login button
		chromedp.Click(`#wechat > span:nth-child(2)`),
		// click confirm button
		chromedp.Click(`#dialog > div.dialog-wrapper > div > div.dialog-footer > div.dialog-footer-ok`),
		getCode(),
		saveCookies(),
	}
}

func getCode() chromedp.ActionFunc {
	return func(ctx context.Context) (err error) {
		var code []byte
		if err = chromedp.Screenshot(`#wximport`, &code, chromedp.ByID).Do(ctx); err != nil {
			return
		}
		//return printQRCode(code)
		return saveQRCodeImg(code)
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
	return os.WriteFile("code.png", code, 0755)
}

func saveCookies() chromedp.ActionFunc {
	return func(ctx context.Context) (err error) {
		// wait qrcode login
		fmt.Println("wait qrcode login")
		if err = chromedp.WaitVisible(`#app`, chromedp.ByID).Do(ctx); err != nil {
			return
		}
		cookies, err := network.GetAllCookies().Do(ctx)
		if err != nil {
			return
		}
		cookiesData, err := network.GetAllCookiesReturns{Cookies: cookies}.MarshalJSON()
		if err != nil {
			return
		}
		// save tmp file
		if err = os.WriteFile("cookies.tmp", cookiesData, 0755); err != nil {
			return
		}
		return
	}
}

func loadCookies() chromedp.ActionFunc {
	return func(ctx context.Context) (err error) {
		if _, err = os.Stat("cookies.tmp"); err != nil {
			return
		}
		cookiesData, err := os.ReadFile("cookies.tmp")
		if err != nil {
			return
		}
		cookiesParams := network.SetCookiesParams{}
		if err = cookiesParams.UnmarshalJSON(cookiesData); err != nil {
			return
		}
		return network.SetCookies(cookiesParams.Cookies).Do(ctx)
	}
}

func checkLoginStatus() chromedp.ActionFunc {
	return func(ctx context.Context) (err error) {
		var url string
		if err = chromedp.Evaluate(`window.location.href`, &url).Do(ctx); err != nil {
			return
		}
		if strings.Contains(url, userCenterURl) {
			fmt.Println("login success")
			return nil
		}
		return errors.New("login fail")
	}
}
