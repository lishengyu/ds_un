package fileexcel

import (
	"fmt"
	"log"

	"github.com/xuri/excelize/v2"
)

func loadSheet(index int, name string, excel *excelize.File) error {
	rows, err := excel.GetRows(name)
	if err != nil {
		return err
	}

	for _, row := range rows {
		var buff string
		for index, r := range row {
			if index == 0 {
				buff = r
			} else {
				buff = fmt.Sprintf("%s|%s", buff, r)
			}
		}
		fmt.Println(buff)
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
