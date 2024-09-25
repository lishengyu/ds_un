package comm

import (
	"crypto/md5"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
)

/*
type Condition struct {
	Index int
	Value string
}

type MatchSplit struct {
	Condition []Condition
	Match     []string
}

type RangeSplit struct {
	Min int64
	Max int64
}

type ContentSplit struct {
	Ctx   []string
	Range []RangeSplit
}

type ContentSplits struct {
	Ctxs []ContentSplit
}
*/

// PathExists	检查路径或文件是否存在
//
//	@param path
//	@return bool
func PathExists(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true
	}
	if os.IsNotExist(err) {
		return false
	}
	return true
}

// GetFileMd5	计算文件md5
//
//	@param path
//	@return string	小写的十六进制
func GetFileMd5(path string) (string, error) {
	fd, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer fd.Close()

	md5h := md5.New()
	io.Copy(md5h, fd)

	return fmt.Sprintf("%x", md5h.Sum(nil)), nil
}

// GetFileSize 	计算文件大小
//
//	@param path
//	@return string	单位KB，保留小数点后两位
func GetFileSize(path string) (string, error) {
	fi, err := os.Stat(path)
	if err != nil {
		return "0", err
	}

	sizeKB := float64(fi.Size()) / 1024

	return fmt.Sprintf("%.2f", sizeKB), nil
}

// IsFile 	判断是否为文件
//
//	@param path
//	@return bool
func IsFile(path string) bool {
	fi, err := os.Stat(path)
	if err != nil {
		return false
	}

	if fi.IsDir() {
		return false
	}

	return true
}

// IsDir 判断是否为路径
//
//	@param path
//	@return bool
func IsDir(path string) bool {
	fi, err := os.Stat(path)
	if err != nil {
		return false
	}

	if fi.IsDir() {
		return true
	}

	return false
}

// FileNameFormat 	判断文件是否符合数安样本文件规范
//
//	@param name
//	@return string
func FileNameFormat(name string) bool {
	fn := filepath.Base(name)
	fs := strings.Split(fn, "_")
	return len(fs) == 3
}

// GetSuffix 	获取文件名后缀
//
//	@param name
//	@return string
func GetSuffix(name string) (string, error) {
	if strings.HasSuffix(name, ".tar.gz") {
		return "tar.gz", nil
	}

	fn := filepath.Base(name)
	index := strings.LastIndex(fn, ".")
	if index == -1 {
		fs := strings.Split(fn, "_")
		return fs[len(fs)-1], nil
	}
	return fn[index+1:], nil
}

// CopyFile 	拷贝文件
//
//	@param src
//	@param dst
//	@return bool	拷贝是否成功
func CopyFile(src, dst string) bool {
	input, err := ioutil.ReadFile(src)
	if err != nil {
		log.Printf("%v\n", err)
		return false
	}

	tmp := dst + ".temp"

	err = ioutil.WriteFile(tmp, input, 0666)
	if err != nil {
		log.Printf("%v\n", err)
		return false
	}

	err = os.Rename(tmp, dst)
	if err != nil {
		log.Printf("%v\n", err)
		return false
	}

	return true
}

// AppendFile	写文件内容，追加的方式
//
//	@param name
//	@param data
//	@return error
func AppendFile(name string, data []byte) error {
	f, err := os.OpenFile(name, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0664)
	if err != nil {
		return err
	}
	_, err = f.Write(data)
	if err1 := f.Close(); err1 != nil && err == nil {
		err = err1
	}
	return err
}

// ReadFile 		读取文件内容
//
//	@param fn
//	@return string
//	@return error
func ReadFile(fn string) ([]byte, error) {
	data, err := ioutil.ReadFile(fn)
	if err != nil {
		log.Printf("%v\n", err)
		return []byte{}, err
	}

	return data, nil
}

func PrintReportPrefix(ctx string) {
	fmt.Printf("======================================\n")
	fmt.Printf("%s\n", ctx)
	fmt.Printf("----------------\n")
}

func PrintReportSuffix(ctx string) {
	fmt.Printf("----------------\n")
	fmt.Printf("%s\n", ctx)
	fmt.Printf("======================================\n")
}
