package store

import "os"

type Persistence struct {
	file *os.File
}

func NewAOF(filename string) (*Persistence, error) {
	file, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}
	return &Persistence{file: file}, nil
}

func (p *Persistence) Append(cmd string) error {
	_, err := p.file.WriteString(cmd + "\n")
	return err
}

func (p *Persistence) Close() error {
	return p.file.Close()
}
