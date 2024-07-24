package config

import (
	"github.com/gocarina/gocsv"
	"os"
)

// WriteCsvFile 写入csv文件
func WriteCsvFile[T any](filename string, data []T) error {
	file, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE, os.ModePerm)
	if err != nil {
		return err
	}
	defer file.Close()

	// 将结构体切片编组到 CSV 文件
	if err := gocsv.MarshalFile(&data, file); err != nil {
		return err
	}

	return nil
}

// ReadCsvFile 读取csv文件
func ReadCsvFile[T any](filename string, data *[]T) error {
	if _, err := os.Stat(filename); err != nil {
		return err
	}

	file, err := os.OpenFile(filename, os.O_RDWR, os.ModePerm)
	if err != nil {
		return err
	}

	defer file.Close()

	// 将 CSV 文件解组到结构体切片
	if err := gocsv.UnmarshalFile(file, data); err != nil {
		return err
	}

	return nil
}
