package fileexcel

import (
	"testing"
)

func TestReadXlsxFile(t *testing.T) {
	err := ReadXlsxFile("ds_code_dict.xlsx")
	if err != nil {
		t.Errorf("read xlsx file failed: %v\n", err)
	}
}
