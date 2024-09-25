package fileproc

import (
	"ds_tool/spec"
	"fmt"
	"strconv"
	"strings"
)

type SampleCxInfo struct {
}

var (
	Cx_CheckMap map[string]CheckInfo //Dpi话单必填项校验
)

func procCxFields(fs []string) (int, bool) {
	attach := fs[spec.Cx_FileMd5]
	if attach == "" {
		return spec.Cx_FileMd5, false
	}
	fields := strings.Split(attach, "_")
	if valid := fieldsMd5(fields[0], spec.IndexCx); !valid {
		return spec.Cx_FileMd5, false
	}

	return 0, true
}

func recordCxInfo(fs []string) {

}

func procCxCtx(line, filename string) {
	fs := strings.Split(line, "|")
	if len(fs) != spec.Cx_Max {
		info := CheckInfo{
			Reason:   fmt.Sprintf("字段个数%d不符", len(fs)),
			Filenmae: filename,
		}
		Cx_CheckMap[line] = info
		incLogInvalidCnt(spec.IndexCx)
		return
	}

	if index, valid := procCxFields(fs); valid {
		incLogValidCnt(spec.IndexCx)
		recordCxInfo(fs)
	} else {
		info := CheckInfo{
			Reason:   fmt.Sprintf("第%d个字段非法", index+1),
			Filenmae: filename,
		}
		Cx_CheckMap[line] = info
		incLogInvalidCnt(spec.IndexCx)
	}
}

func statCxFields(fs []string) {
	md5 := GetMd5FromAttach(fs[spec.Cx_FileMd5])
	if md5 == "" {
		incLogInvalidCnt(spec.IndexCx)
		return
	}

	key := &Md5HitKey{
		Sip:   fs[spec.Cx_Sip],
		Dip:   fs[spec.Cx_Dip],
		Sport: fs[spec.Cx_Sport],
		Dport: fs[spec.Cx_Dport],
		Md5:   md5,
	}

	filetype, _ := strconv.Atoi(fs[spec.Cx_FileType])
	appproto, _ := strconv.Atoi(fs[spec.Cx_AppId])
	dataproto, _ := strconv.Atoi(fs[spec.Cx_DataPro])
	filesize := fs[spec.Cx_FileSize]

	value, ok := Md5HitMapLoad(key)
	if ok {
		hit := value.(*Md5HitInfo)
		hit.HasCx += 1
		hit.FileSize = filesize
		hit.FileType = filetype
		hit.AppProto = appproto
		hit.DataProto = dataproto
	} else {
		value := &Md5HitInfo{
			HasCx:     1,
			FileSize:  filesize,
			FileType:  filetype,
			AppProto:  appproto,
			DataProto: dataproto,
		}

		Md5HitMapStore(key, value)
	}
}
