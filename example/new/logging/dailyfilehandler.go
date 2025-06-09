package logging

import (
	"context"
	"log/slog"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// DailyFileHandler 是一个支持按天滚动日志文件的 Handler
type DailyFileHandler struct {
	dir        string       // 日志目录
	prefix     string       // 文件前缀，如 "app"
	ext        string       // 文件扩展名，如 ".log"
	file       *os.File     // 当前打开的日志文件
	handler    slog.Handler // 实际使用的 Handler（比如 JSON）
	currentDay string       // 当前是哪一天
	mu         sync.Mutex   // 确保并发安全
}

// NewDailyFileHandler 创建一个新的按天日志处理器
func NewDailyFileHandler(dir, prefix, ext string) (*DailyFileHandler, error) {
	d := &DailyFileHandler{
		dir:    dir,
		prefix: prefix,
		ext:    ext,
	}

	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}

	err := d.rotate()
	if err != nil {
		return nil, err
	}

	d.handler = slog.NewJSONHandler(d.file, nil)
	return d, nil
}

// rotate 检查是否需要更换日志文件（即是否跨天）
func (d *DailyFileHandler) rotate() error {
	now := time.Now()
	day := now.Format("2006-01-02")

	d.mu.Lock()
	defer d.mu.Unlock()

	if d.currentDay == day && d.file != nil {
		return nil // 同一天，不需要换文件
	}

	if d.file != nil {
		_ = d.file.Close()
	}

	filename := filepath.Join(d.dir, d.prefix+"."+day+d.ext)
	file, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return err
	}

	d.file = file
	d.currentDay = day
	d.handler = slog.NewJSONHandler(d.file, nil)
	return nil
}

// 实现 slog.Handler 接口

func (d *DailyFileHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return d.handler.Enabled(ctx, level)
}

func (d *DailyFileHandler) Handle(ctx context.Context, r slog.Record) error {
	d.rotate() // 每次写入前检查是否跨天
	return d.handler.Handle(ctx, r)
}

func (d *DailyFileHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &DailyFileHandler{
		dir:        d.dir,
		prefix:     d.prefix,
		ext:        d.ext,
		file:       d.file,
		handler:    d.handler.WithAttrs(attrs),
		currentDay: d.currentDay,
		mu:         sync.Mutex{},
	}
}

func (d *DailyFileHandler) WithGroup(name string) slog.Handler {
	return &DailyFileHandler{
		dir:        d.dir,
		prefix:     d.prefix,
		ext:        d.ext,
		file:       d.file,
		handler:    d.handler.WithGroup(name),
		currentDay: d.currentDay,
		mu:         sync.Mutex{},
	}
}

// Close 关闭当前文件
func (d *DailyFileHandler) Close() error {
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.file != nil {
		return d.file.Close()
	}
	return nil
}
