package recorder

import (
	"encoding/csv"
	"os"
	"path/filepath"
	"strings"

	"github.com/gocarina/gocsv"
	"github.com/wonderstone/QuantKit/config"
)

type CsvRecorder struct {
	needHeader bool
	filename   string
	file       *os.File
	writer     *gocsv.SafeCSVWriter
	channel    chan any
}

func (c *CsvRecorder) QueryRecord(query ...WithQuery) []any {
	config.ErrorF("CsvRecorder目前不支持查询")
	return nil
}

func (c *CsvRecorder) Read(data any) error {
	file, err := os.OpenFile(c.filename, os.O_RDONLY, os.ModePerm)
	if err != nil {
		return err
	}

	err = gocsv.UnmarshalFile(file, data)
	if err != nil {
		return err
	}

	return nil
}

func NewCsvRecorder(option ...WithOption) *CsvRecorder {
	op := NewOp(option...)
	if !strings.HasSuffix(op.file, ".csv") {
		op.file += ".csv"
	}

	// 文件夹不存在则创建
	dir := filepath.Dir(op.file)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		err := os.MkdirAll(dir, 0777)
		if err != nil {
			config.ErrorF("创建文件夹失败: %v", err)
			return nil
		}
	}

	if op.plusMode {
		create, err := os.OpenFile(op.file, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			config.ErrorF("创建文件失败: %v", err)
			return nil
		}

		writer := gocsv.NewSafeCSVWriter(csv.NewWriter(create))

		// 如果文件为空，则写入表头
		stat, err := create.Stat()
		if err != nil {
			config.ErrorF("获取文件信息失败: %v", err)
			return nil
		}

		return &CsvRecorder{
			needHeader: stat.Size() == 0, filename: op.file, file: create, writer: writer, channel: make(chan any),
		}
	}

	create, err := os.Create(op.file)
	if err != nil {
		config.ErrorF("创建文件失败: %v", err)
		return nil
	}

	writer := gocsv.NewSafeCSVWriter(csv.NewWriter(create))

	return &CsvRecorder{needHeader: true, filename: op.file, file: create, writer: writer, channel: make(chan any)}
}

func (c *CsvRecorder) setColumn(column []string) {
	err := c.writer.Write(column)
	if err != nil {
		config.ErrorF("写入数据失败: %v", err)
		return
	}

	c.writer.Flush()
}

func (c *CsvRecorder) RecordChan() error {
	if c.needHeader {
		return gocsv.MarshalChan(c.channel, c.writer)
	} else {
		return gocsv.MarshalChanWithoutHeaders(c.channel, c.writer)
	}
}

func (c *CsvRecorder) GetChannel() chan any {
	return c.channel
}

func (c *CsvRecorder) Release() {
	close(c.channel)
}
