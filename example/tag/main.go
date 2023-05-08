package main

import (
	"fmt"
	"reflect"
)

type User struct {
	Name  string `mytag:"MyName"`
	Email string `mytag:"MyEmail"`
}

func main() {
	u := User{"Bob", "bob@mycompany.com"}
	t := reflect.TypeOf(u)

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		fmt.Printf("Field: User.%s\n", field.Name)
		fmt.Printf("\tWhole tag value : %s\n", field.Tag)
		fmt.Printf("\tValue of 'mytag': %s\n", field.Tag.Get("mytag"))
	}
}
