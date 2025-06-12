package logging

import (
	"context"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// DailyFileHandler 是一个支持按天滚动日志文件的 Handler，并支持保留 N 天日志
type DailyFileHandler struct {
	dir        string               // 日志目录
	prefix     string               // 文件前缀，如 "app"
	ext        string               // 文件扩展名，如 ".log"
	file       *os.File             // 当前打开的日志文件
	handler    slog.Handler         // 实际使用的 Handler（比如 JSON）
	currentDay string               // 当前是哪一天
	opts       *slog.HandlerOptions // 可选配置
	keepDays   int                  // 保留天数，0 表示不清理
	mu         sync.Mutex           // 确保并发安全
}

// NewDailyFileHandler 创建一个新的按天日志处理器
func NewDailyFileHandler(dir, prefix, ext string, opts *slog.HandlerOptions, keepDays int) (*DailyFileHandler, error) {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}

	d := &DailyFileHandler{
		dir:      dir,
		prefix:   prefix,
		ext:      ext,
		opts:     opts,
		keepDays: keepDays,
	}

	if err := d.rotate(); err != nil {
		return nil, err
	}

	return d, nil
}

// rotate 检查是否需要更换日志文件（即是否跨天），并清理过期日志
func (d *DailyFileHandler) rotate() error {
	now := time.Now()
	day := now.Format("2006-01-02")

	d.mu.Lock()
	defer d.mu.Unlock()

	if d.currentDay == day && d.file != nil {
		return nil // 同一天，不需要换文件
	}

	// 关闭旧文件
	if d.file != nil {
		if err := d.file.Close(); err != nil {
			return err
		}
	}

	// 打开新文件
	filename := filepath.Join(d.dir, d.prefix+"."+day+d.ext)
	file, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return err
	}

	// 更新状态
	d.file = file
	d.currentDay = day
	d.handler = slog.NewJSONHandler(d.file, d.opts)

	// 清理过期日志文件
	if d.keepDays > 0 {
		if err := d.cleanupOldLogs(now); err != nil {
			return err
		}
	}

	return nil
}

// cleanupOldLogs 删除超过 keepDays 的日志文件
func (d *DailyFileHandler) cleanupOldLogs(now time.Time) error {
	files, err := os.ReadDir(d.dir)
	if err != nil {
		return err
	}

	cutoffDate := now.Add(-time.Duration(d.keepDays*24) * time.Hour).Format("2006-01-02")

	for _, file := range files {
		if !file.IsDir() && strings.HasPrefix(file.Name(), d.prefix+".") && strings.HasSuffix(file.Name(), d.ext) {
			datePart := file.Name()[len(d.prefix)+1 : len(file.Name())-len(d.ext)]
			if datePart < cutoffDate {
				fullPath := filepath.Join(d.dir, file.Name())
				if err := os.Remove(fullPath); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

// 实现 slog.Handler 接口

func (d *DailyFileHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return d.handler.Enabled(ctx, level)
}

func (d *DailyFileHandler) Handle(ctx context.Context, r slog.Record) error {
	if err := d.rotate(); err != nil {
		return err
	}
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
		opts:       d.opts,
		keepDays:   d.keepDays,
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
		opts:       d.opts,
		keepDays:   d.keepDays,
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
