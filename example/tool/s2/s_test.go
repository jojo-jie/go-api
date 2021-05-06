package s2

import (
	"fmt"
	"testing"
)

func TestS(t *testing.T)  {
	// 是否可以寻址
	// 不可寻址 字面量 常量
	/*a1:=A{}
	a2:=&A{}
	t.Log(a1.Print())
	t.Log(a2.Print())*/

	// jack是个接口，被一个指针对象赋值，下面的方法都可以正确执行
	var jack Person  // 声明一个接口类型的对象
	//jack := &Person{"jack", 10} // Person实现了接口
	jack.SayHello()
	jack.SetAge(20)
	fmt.Println(jack.GetAge())

	// 值类型并没有实现SetAge的方法，所以赋值的时候会报错
	// cannot use Person literal (type Person) as type Human in assignment:
	//    Person does not implement Human (SetAge method has pointer receiver)
	var Tom Human
	Tom = &Person{"Tom", 12}
	Tom.SayHello()
	Tom.SetAge(10)
	fmt.Println(Tom.GetAge())

}

// 接口类型
type Human interface{
	SayHello()
	SetAge(age int)
	GetAge()int
}
// 结构体对象
type Person struct{
	Name string
	Age int
}
// 值方法
func(p Person) SayHello(){
	fmt.Printf("Hello, my name is %s\n", p.Name)
}
// 指针方法
func(p *Person) SetAge(age int){
	p.Age = age
}
// 值方法
func(p Person)GetAge()int{
	return p.Age
}

func TestD(t *testing.T)  {
	x:=int8(1) //-2^8~2^8-1
	y:=x-127
	t.Log(y)
}
