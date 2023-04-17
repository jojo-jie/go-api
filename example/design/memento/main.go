package main

import (
	"errors"
	"fmt"
	"time"
)

// IEditor 编辑器接口定义
type IEditor interface {
	Title(title string)
	Content(content string)
	Save()
	Undo() error
	Redo() error
	Show()
}

type Memento struct {
	title      string
	content    string
	createTime int64
}

func newMemento(title string, content string) *Memento {
	return &Memento{
		title, content, time.Now().Unix(),
	}
}

// Editor 编辑器类, 实现IEditor接口
type Editor struct {
	title    string
	content  string
	versions []*Memento
	index    int
}

func NewEditor() IEditor {
	return &Editor{
		"", "", make([]*Memento, 0), 0,
	}
}

func (editor *Editor) Title(title string) {
	editor.title = title
}

func (editor *Editor) Content(content string) {
	editor.content = content
}

func (editor *Editor) Save() {
	it := newMemento(editor.title, editor.content)
	editor.versions = append(editor.versions, it)
	editor.index = len(editor.versions) - 1
}

func (editor *Editor) Undo() error {
	return editor.load(editor.index - 1)
}

func (editor *Editor) load(i int) error {
	size := len(editor.versions)
	if size <= 0 {
		return errors.New("no history versions")
	}

	if i < 0 || i >= size {
		return errors.New("no more history versions")
	}

	it := editor.versions[i]
	editor.title = it.title
	editor.content = it.content
	editor.index = i
	return nil
}

func (editor *Editor) Redo() error {
	return editor.load(editor.index + 1)
}

func (editor *Editor) Show() {
	fmt.Printf("MockEditor.Show, title=%s, content=%s\n", editor.title, editor.content)
}

// 备忘录模式中	主要有这三个角色的类
//
// Originator（发起者）：Originator是当前的基础对象，它会将自己的状态保存进备忘录，此角色可以类比博客系统中的文章对象
// 发起者中要有保存方法和从备忘录中恢复状态的方法，保存方法会返回当时状态组成的备忘录对象
// Memento（备忘录） ： 存储着Originator的状态的对象，类比理解即为文章对象的不同版本。
// Caretaker（管理人）：Caretaker是保存着多条备忘录的对象，并维护着备忘录的索引，在需要的时候会返回相应的备忘录 -- 类比理解为博客系统中的编辑器对象
// 管理者的保存和恢复操作，会代理其持有的发起者对象的保存和恢复操作，在这些代理方法中会增加对备忘录对象列表、当前备忘录版本的维护。
func main() {
	editor := NewEditor()

	// test save()
	editor.Title("唐诗")
	editor.Content("白日依山尽")
	editor.Save()

	editor.Title("唐诗 登鹳雀楼")
	editor.Content("白日依山尽, 黄河入海流. ")
	editor.Save()

	editor.Title("唐诗 登鹳雀楼 王之涣")
	editor.Content("白日依山尽, 黄河入海流。欲穷千里目, 更上一层楼。")
	editor.Save()

	// test show()
	fmt.Println("-------------Editor 当前内容-----------")
	editor.Show()

	fmt.Println("-------------Editor 回退内容-----------")
	for {
		e := editor.Undo()
		if e != nil {
			break
		} else {
			editor.Show()
		}
	}

	fmt.Println("-------------Editor 前进内容-----------")
	for {
		e := editor.Redo()
		if e != nil {
			break
		} else {
			editor.Show()
		}
	}
}
