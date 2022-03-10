package log_rotation

import (
	"fmt"
	"os"
	"path"
	"strings"
	"time"
)

type limitType uint8

const (
	UseTime limitType = iota
	UseFileSize
	UseMaxLines
)

type Option func(rotation *LogRotation)

type LogRotation struct {
	isReady      chan struct{}
	filePath     string
	fileName     string
	fileExt      string
	fileUniqueId int64
	fileSize     int64
	maxLines     uint64
	timeLimit    time.Duration
	limitType    limitType
	linkQueue    *uChan
	fpCreateTime time.Time
	fp           *os.File
	limitCount   struct {
		size  int64
		lines uint64
	}
}

func NewLogRotation(opts ...Option) *LogRotation {
	rotation := LogRotation{
		isReady:      make(chan struct{}),
		fileName:     "soft_log",
		fileExt:      ".log",
		fileUniqueId: time.Now().UnixMicro(),
		linkQueue:    newChan(),
	}
	for i := range opts {
		opts[i](&rotation)
	}
	return &rotation
}

// Launch 启动异步写入
func (rotation *LogRotation) Launch() {
	go func(rotation *LogRotation) {
		for {
			select {
			case <-rotation.isReady:
				//fmt.Println("in ready", rotation.linkQueue.data)
				rotation.checkInitFile()
				for logData := rotation.linkQueue.get(); logData != nil; logData = rotation.linkQueue.get() {
					currentLen := len(logData)
					rotation.checkFileLimit(currentLen)
					n, _ := rotation.fp.Write(logData)
					rotation.limitCount.size += int64(n)
					rotation.limitCount.lines += 1
				}
			}
		}
	}(rotation)
}

// 初始化日志文件，如果初始化失败，则写入stdout
func (rotation *LogRotation) checkInitFile() {
	if rotation.fp == nil {
		rotation.createFile()
	}
}

// checkLimit检查limit，如果触发限制则关闭旧fp并新建一个fp
func (rotation *LogRotation) checkFileLimit(currentSize int) {
	if rotation.limitType == UseFileSize {
		// 如果文件大小超过限制，则重新创建
		if rotation.limitCount.size+int64(currentSize) > rotation.fileSize*1024 {
			rotation.createFile()
		}
	}
	if rotation.limitType == UseMaxLines {
		// 如果文件行数已经写满，则重新创建
		if rotation.limitCount.lines+1 > rotation.maxLines {
			rotation.createFile()
		}
	}
	if rotation.limitType == UseTime {
		// 如果当前时间已经超过 timeLimit时间，则重新创建
		if time.Now().After(rotation.fpCreateTime.Add(rotation.timeLimit)) {
			rotation.createFile()
		}
	}
}

// 创建日志文件
func (rotation *LogRotation) createFile() {
	var err error
	if err = rotation.fp.Close(); err != nil {
		_, _ = os.Stderr.WriteString(fmt.Sprintf("日志文件关闭失败：%v", err))
	}
	rotation.fileUniqueId = time.Now().UnixMicro()
	if rotation.fp, err = os.OpenFile(fmt.Sprintf("%s-%s-%d%s", rotation.fileName, time.Now().Format("2006_01_02"), rotation.fileUniqueId, rotation.fileExt), os.O_CREATE|os.O_APPEND, 0755); err != nil {
		_, _ = os.Stderr.WriteString(fmt.Sprintf("日志文件创建失败：%v", err))
		rotation.fp = os.Stdout
	}
	rotation.fpCreateTime = time.Now()
	rotation.limitCount.size = 0
	rotation.limitCount.lines = 0
}

// SetLimitType 设置文件的处理方式
//func SetLimitType(arg limitType) Option {
//	return func(rotation *LogRotation) {
//		rotation.limitType = arg
//	}
//}

// SetLimitUseTime 设置文件按指定时间空隙分割
func SetLimitUseTime(limit time.Duration) Option {
	if limit <= 0 {
		panic(fmt.Sprintf("Invalid parameter %d time", limit))
	}
	return func(rotation *LogRotation) {
		rotation.limitType = UseTime
		rotation.timeLimit = limit
	}
}

// SetLimitUseMaxLines 设置文件按行数分割
func SetLimitUseMaxLines(maxLine uint64) Option {
	if maxLine <= 0 {
		panic(fmt.Sprintf("Invalid parameter %d lines", maxLine))
	}
	return func(rotation *LogRotation) {
		rotation.limitType = UseMaxLines
		rotation.maxLines = maxLine
	}
}

// SetLimitUseFileSize 设置文件按大小分割，单位（kb）
func SetLimitUseFileSize(kb int64) Option {
	if kb <= 0 {
		panic(fmt.Sprintf("Invalid parameter %d kb", kb))
	}
	return func(rotation *LogRotation) {
		rotation.limitType = UseFileSize
		rotation.fileSize = kb
	}
}

// SetFileName 设置日志文件名，可包含路径，保存格式：{path}/{fileBaseName}-{yyy_MM_dd}-{uniqueId}.{fileExt}
func SetFileName(f string) Option {
	return func(rotation *LogRotation) {
		rotation.filePath = path.Dir(f)
		if !strings.HasSuffix(rotation.filePath, "/") {
			rotation.filePath += "/"
		}
		fileName := strings.TrimSpace(path.Base(f))
		if fileName == "" || fileName == "/" || fileName == "." {
			panic(fmt.Sprintf("Invalid fileName %s", fileName))
		}
		rotation.fileExt = path.Ext(fileName)
		rotation.fileName = strings.TrimSuffix(fileName, rotation.fileExt)
		//如果文件夹不存在，则先创建
		if ok, _ := pathExists(rotation.filePath); !ok {
			if err := os.MkdirAll(rotation.filePath, os.ModeDir); err != nil {
				panic(err)
			}
		}
	}
}

// Writer 实现io.Writer
func (rotation *LogRotation) Write(p []byte) (n int, err error) {
	//fmt.Println(string(p))
	select {
	case rotation.isReady <- struct{}{}:
		rotation.linkQueue.put(p)
	default:
		rotation.linkQueue.put(p)
	}
	return len(p), nil
}
