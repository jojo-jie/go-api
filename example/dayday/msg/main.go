package main

import (
	"context"
	"crypto/tls"
	"embed"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"golang.org/x/sync/errgroup"
	"gopkg.in/gomail.v2"
	"html/template"
	"io"
	"net/http"
	"os"
	"strconv"
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
	Content  string // html 模板
}

var emailConfig EmailConfig

type App struct {
	AppId        string
	AppSecret    string
	Host         string
	Registration string
	Subject      string
}

var app App

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
var e string

func init() {
	flag.StringVar(&e, "e", "medicine", "help info")
	flag.Parse()
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
	err = v.UnmarshalKey("App", &app)
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

type Info520 struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data []struct {
		SysDept struct {
			Id               int         `json:"id"`
			DeptCode         string      `json:"deptCode"`
			DeptId           string      `json:"deptId"`
			Pid              int         `json:"pid"`
			CertNoCode       string      `json:"certNoCode"`
			RegionCode       string      `json:"regionCode"`
			Name             string      `json:"name"`
			ShortName        interface{} `json:"shortName"`
			BookCode         interface{} `json:"bookCode"`
			BookCodeNo       interface{} `json:"bookCodeNo"`
			BookHeadRegion   interface{} `json:"bookHeadRegion"`
			DocCode          string      `json:"docCode"`
			OverPrintFlag    string      `json:"overPrintFlag"`
			DressFlag        string      `json:"dressFlag"`
			AdminType        string      `json:"adminType"`
			Type             string      `json:"type"`
			FeeType          interface{} `json:"feeType"`
			OpMode           interface{} `json:"opMode"`
			SortNo           string      `json:"sortNo"`
			CreatDate        string      `json:"creatDate"`
			AbolishDate      string      `json:"abolishDate"`
			OfficeTime       string      `json:"officeTime"`
			OfficeEmail      interface{} `json:"officeEmail"`
			DeptRoute        interface{} `json:"deptRoute"`
			DeptAddress      string      `json:"deptAddress"`
			DeptPost         string      `json:"deptPost"`
			DeptNumBz        interface{} `json:"deptNumBz"`
			DeptNumBzGov     interface{} `json:"deptNumBzGov"`
			DeptNumBzGov1    interface{} `json:"deptNumBzGov1"`
			DeptNumBzAid     interface{} `json:"deptNumBzAid"`
			DeptNumBzSelf    interface{} `json:"deptNumBzSelf"`
			DeptNumSj        interface{} `json:"deptNumSj"`
			DeptNumSjGov     interface{} `json:"deptNumSjGov"`
			DeptNumSjGov1    interface{} `json:"deptNumSjGov1"`
			DeptNumSjAid     interface{} `json:"deptNumSjAid"`
			DeptNumSjSelf    interface{} `json:"deptNumSjSelf"`
			DeptNumSjHire    interface{} `json:"deptNumSjHire"`
			OfficeArea       string      `json:"officeArea"`
			OfficeAreaHd     interface{} `json:"officeAreaHd"`
			OfficeAreaJh     interface{} `json:"officeAreaJh"`
			OfficeAreaLh     interface{} `json:"officeAreaLh"`
			OfficeAreaDa     interface{} `json:"officeAreaDa"`
			OfficeAreaBz     interface{} `json:"officeAreaBz"`
			DeptLeador       interface{} `json:"deptLeador"`
			DeptDuty         interface{} `json:"deptDuty"`
			DeptTel          string      `json:"deptTel"`
			DeptFax          interface{} `json:"deptFax"`
			ComplainTel      string      `json:"complainTel"`
			HomeUrl          interface{} `json:"homeUrl"`
			SkinValue        interface{} `json:"skinValue"`
			DocPageNumJh     int         `json:"docPageNumJh"`
			DocPageNumLh     int         `json:"docPageNumLh"`
			DocPageNumBf     int         `json:"docPageNumBf"`
			DocPageNumCx     int         `json:"docPageNumCx"`
			DegreeMan        interface{} `json:"degreeMan"`
			DegreeWoman      interface{} `json:"degreeWoman"`
			Descript         interface{} `json:"descript"`
			ModifyDate       interface{} `json:"modifyDate"`
			ModifyId         interface{} `json:"modifyId"`
			ValidFlag        int         `json:"validFlag"`
			Remark           interface{} `json:"remark"`
			DeptCodeDesk     string      `json:"deptCodeDesk"`
			Allowyd          int         `json:"allowyd"`
			FileDate         interface{} `json:"fileDate"`
			OldDeptCodes     interface{} `json:"oldDeptCodes"`
			DefaultPageNumJh int         `json:"defaultPageNumJh"`
			DefaultPageNumLh int         `json:"defaultPageNumLh"`
			DefaultPageNumBj int         `json:"defaultPageNumBj"`
			DefaultPageNumBl int         `json:"defaultPageNumBl"`
			CheckModify      interface{} `json:"checkModify"`
			NumberPrefix     string      `json:"numberPrefix"`
			Moulage          interface{} `json:"moulage"`
			DivisionName     string      `json:"divisionName"`
			CreditCode       string      `json:"creditCode"`
			ArchivesCode     string      `json:"archivesCode"`
			WsyyNum          int         `json:"wsyyNum"`
			WsyyMobile       string      `json:"wsyyMobile"`
		} `json:"sysDept"`
		Num int `json:"num"`
	} `json:"data"`
}

func sendEmail() error {
	m := gomail.NewMessage()
	m.SetHeader("From", m.FormatAddress(emailConfig.From, emailConfig.FromName))
	m.SetHeader("To", emailConfig.To...)
	if e == "medicine" {
		m.SetHeader("Subject", emailConfig.Subject)
		t := template.Must(template.New("index").Parse(emailConfig.Content))
		m.AddAlternativeWriter("text/html", func(w io.Writer) error {
			return t.Execute(w, message)
		})
	} else {
		resp := request()
		index := strings.LastIndex(resp, "num")
		if -1 == index {
			fmt.Println(resp)
			return nil
		}
		var info520 Info520
		err := json.Unmarshal([]byte(resp), &info520)
		if err != nil {
			return err
		}
		m.SetHeader("Subject", app.Subject+" "+strconv.Itoa(info520.Data[0].Num))
		m.SetBody("", "登记地址 "+app.Registration)
	}

	d := gomail.NewDialer("smtp.163.com", 465, emailConfig.From, emailConfig.Password)
	d.TLSConfig = &tls.Config{InsecureSkipVerify: true}
	return d.DialAndSend(m)
}

func request() (r string) {
	url := app.Host
	method := "GET"
	client := &http.Client{}
	req, err := http.NewRequest(method, url, nil)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/100.0.4896.127 Safari/537.36")
	if err != nil {
		fmt.Println(err)
		return
	}
	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return
	}
	return string(body)
}
