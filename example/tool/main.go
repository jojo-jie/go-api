package main

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"text/template"
)

type Rest string

func (r *Rest) String() string {
	return fmt.Sprint(*r)
}

func (r *Rest) Set(value string) error {
	if len(*r) > 0 {
		return errors.New("name flag already set")
	}
	*r = Rest("年纪轻轻不干好事:" + value)
	return nil
}

func main() {
	//var name string
	/*flag.StringVar(&name, "n", "Go 语言编程之旅", "help info")
	flag.StringVar(&name, "name", "Go 语言编程之旅", "help info2")*/

	/*flag.Parse()
	//子命令使用
	goCmd := flag.NewFlagSet("go", flag.ExitOnError)
	goCmd.StringVar(&name, "n", "Go", "help info")
	phpCmd := flag.NewFlagSet("php", flag.ExitOnError)
	phpCmd.StringVar(&name, "name", "php", "help info")
	args := flag.Args()
	if len(args) <= 0 {
		return
	}
	switch args[0] {
	case "go":
		_ = goCmd.Parse(args[1:])
	case "php":
		_ = phpCmd.Parse(args[1:])
	}
	log.Printf("name %s", name)*/
	//ss := "中"
	//Go语言的字符有以下两种：
	//一种是 uint8 类型，或者叫 byte 型，代表了 ASCII 码的一个字符。
	//另一种是 rune 类型，代表一个 UTF-8 字符，当需要处理中文、日文或者其他复合字符时，则需要用到 rune 类型。rune 类型等价于 int32 类型。
	/*fmt.Println([]byte(ss))
	fmt.Println([]rune(ss))
	var r Rest
	flag.Var(&r, "r", "rest help info")
	flag.Parse()
	log.Println(r)*/

	funcMap := template.FuncMap{"title": strings.Title}
	tpl := template.New("go-programming-tour")
	tpl, _ = tpl.Funcs(funcMap).Parse(templateText)
	data := map[string]string{
		"Name1": "go",
		"Name2": "programming",
		"Name3": "tour",
	}
	_ = tpl.Execute(os.Stdout, data)
}

const templateText = `
Output 0: {{ title .Name1 }}
Output 1: {{ title .Name2 }}
Output 2: {{ .Name3 | title }}
`
