package fileexcel

import (
	"crypto/md5"
	"fmt"
	"io"
	"io/ioutil"
	"log"

	"github.com/xuri/excelize/v2"
)

func getStringMd5(ctx string) string {
	w := md5.New()
	io.WriteString(w, ctx)
	return fmt.Sprintf("%x", w.Sum(nil))
}

func loadSheet(index int, name string, excel *excelize.File) error {
	rows, err := excel.GetRows(name)
	if err != nil {
		return err
	}

	var num int
	for _, row := range rows {
		var buff string
		for index, r := range row {
			if index == 0 {
				buff = r
			} else {
				buff = fmt.Sprintf("%s|%s", buff, r)
			}
		}
		md5str := getStringMd5(buff)
		filename := fmt.Sprintf("./sample/%s_%05d_txt", md5str, num)
		num++
		err := ioutil.WriteFile(filename, []byte(buff), 0644)
		if err != nil {
			log.Printf("failed to write file: %v\n", err)
		}
		log.Printf("write file: %s\n", filename)
	}

	return err
}

func ReadXlsxFile(file string) error {
	excel, err := excelize.OpenFile(file)
	if err != nil {
		return err
	}

	defer func() {
		if err := excel.Close(); err != nil {
			log.Printf("excel close failed: %v\n", err)
		}
	}()

	for index, name := range excel.GetSheetMap() {
		err := loadSheet(index, name, excel)
		if err != nil {
			log.Printf("load sheet[%s] failed: %v\n", name, err)
			return err
		}
	}

	return nil
}
