package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
)

type Rest string

func (r *Rest) String() string {
	return fmt.Sprint(*r)
}

func (r *Rest) Set(value string) error {
	if len(*r)>0 {
		return errors.New("name flag already set")
	}
	*r=Rest("年纪轻轻不干好事:"+value)
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

	var r Rest
	flag.Var(&r, "r", "rest help info")
	flag.Parse()
	log.Println(r)
}
