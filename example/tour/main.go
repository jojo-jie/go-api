package main

import (
	"log"
	"tour/cmd"
)

func main() {
	err := cmd.Execute()
	if err != nil {
		log.Fatalf("cmd.Execute err:%v", err)
	}
}
