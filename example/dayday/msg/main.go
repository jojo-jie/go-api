package main

import (
	"context"
	"crypto/tls"
	"embed"
	"fmt"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"golang.org/x/sync/errgroup"
	"gopkg.in/gomail.v2"
	"html/template"
	"io"
	"os"
	"strings"
	"time"
)

//go:embed *
var configDirs embed.FS

//go:embed html/*
var htmlDirs embed.FS

type EmailConfig struct {
	FromName string
	From     string
	Password string
	Subject  string
	To       []string
	Content  string
}

var emailConfig EmailConfig

type Content struct {
	Name string
	Msg  string
}

type Message struct {
	St       string
	Et       string
	Interval int
	Date     string
	DateTime string
	Title    string
	List     []Content
}

var message Message

var WeekDayMap = map[string]string{
	"Monday":    "星期一",
	"Tuesday":   "星期二",
	"Wednesday": "星期三",
	"Thursday":  "星期四",
	"Friday":    "星期五",
	"Saturday":  "星期六",
	"Sunday":    "星期日",
}

var nowTime time.Time

func init() {
	f, err := configDirs.ReadFile("config.yaml")
	if os.IsNotExist(err) {
		panic(err)
	}
	v := viper.New()
	v.SetConfigType("yaml")
	err = v.ReadConfig(strings.NewReader(string(f)))
	if err != nil {
		panic(err)
	}
	err = v.UnmarshalKey("Email", &emailConfig)
	if err != nil {
		panic(err)
	}
	f, err = htmlDirs.ReadFile("html/index.html")
	if err != nil {
		panic(err)
	}
	emailConfig.Content = string(f)
	err = v.UnmarshalKey("Message", &message)
	if err != nil {
		panic(err)
	}
	location, _ := time.LoadLocation("Asia/Shanghai")
	nowTime = time.Now().In(location)
	message.DateTime = nowTime.Format("2006-01-02")
	t, err := time.Parse("2006-01-02", message.DateTime)
	if err != nil {
		panic(err)
	}
	message.Date = WeekDayMap[t.Weekday().String()]
	startTime, _ := time.Parse("2006-01-02", message.St)
	message.Interval = int(t.Sub(startTime).Hours() / 24)
}

func main() {
	fmt.Println("执行开始时间 " + nowTime.Format("2006-01-02 15:04:05"))
	g, _ := errgroup.WithContext(context.Background())
	g.Go(func() error {
		return sendEmail()
	})
	if err := g.Wait(); err != nil {
		<-time.After(20 * time.Second)
		err := sendEmail()
		if err != nil {
			fmt.Println("再次执行发送失败 ", errors.Wrap(fmt.Errorf("%v", err), "邮件发送错误").Error())
		}
	}
	fmt.Println("执行结束时间 " + nowTime.Format("2006-01-02 15:04:05"))
}

func sendEmail() error {
	m := gomail.NewMessage()
	m.SetHeader("From", m.FormatAddress(emailConfig.From, emailConfig.FromName))
	m.SetHeader("To", emailConfig.To...)
	m.SetHeader("Subject", emailConfig.Subject)
	t := template.Must(template.New("index").Parse(emailConfig.Content))
	m.AddAlternativeWriter("text/html", func(w io.Writer) error {
		return t.Execute(w, message)
	})
	d := gomail.NewDialer("smtp.163.com", 465, emailConfig.From, emailConfig.Password)
	d.TLSConfig = &tls.Config{InsecureSkipVerify: true}
	return d.DialAndSend(m)
}
