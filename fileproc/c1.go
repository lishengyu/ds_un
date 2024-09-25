package fileproc

import (
	"ds_tool/spec"
	"fmt"
	"strconv"
	"strings"
	"sync"
)

type SampleC1Info struct {
	Event    string
	L7Proto  int
	FileType int
	FileSize string
}

var (
	C1_CheckMap map[string]CheckInfo //监测话单必填项校验
)

func SampleMapUpdateC1(m *sync.Map, md5 string, info SampleC1Info) {
	value, ok := m.Load(md5)
	if ok {
		sample := value.(*SampleMapValue)
		sample.C1Info = append(sample.C1Info, info)
	} else {
		sample := &SampleMapValue{}
		sample.C1Info = append(sample.C1Info, info)
		m.Store(md5, sample)
	}
}

func procC1Fields(fs []string) (int, bool) {
	if valid := feildsLogid(fs[spec.C1_LogID]); !valid {
		return spec.C1_LogID, false
	}

	if valid := fieldsNull(fs[spec.C1_CommandId]); !valid {
		return spec.C1_CommandId, false
	}

	if valid := fieldsNull(fs[spec.C1_House_ID]); !valid {
		return spec.C1_House_ID, false
	}

	if valid := fieldsDataProto(fs[spec.C1_Proto]); !valid {
		return spec.C1_Proto, false
	}

	if valid := fieldsHttp(fs[spec.C1_Proto], fs[spec.C1_Domain]); !valid {
		return spec.C1_Domain, false
	}

	if valid := fieldsHttp(fs[spec.C1_Proto], fs[spec.C1_Url]); !valid {
		return spec.C1_Url, false
	}

	if valid := fieldsEvent(fs[spec.C1_EventTypeID], fs[spec.C1_EventSubType]); !valid {
		return spec.C1_EventTypeID, false
	}

	if valid := fieldsNull(fs[spec.C1_SrcIP]); !valid {
		return spec.C1_SrcIP, false
	}

	if valid := fieldsNull(fs[spec.C1_DestIP]); !valid {
		return spec.C1_DestIP, false
	}

	if valid := fieldsNull(fs[spec.C1_SrcPort]); !valid {
		return spec.C1_SrcPort, false
	}

	if valid := fieldsNull(fs[spec.C1_DestPort]); !valid {
		return spec.C1_DestPort, false
	}

	if valid := fieldsFileType(fs[spec.C1_FileType]); !valid {
		return spec.C1_FileType, false
	}

	if valid := fieldsNull(fs[spec.C1_FileSize]); !valid {
		return spec.C1_FileSize, false
	}

	if valid := fieldsNull(fs[spec.C1_DataNum]); !valid {
		return spec.C1_DataNum, false
	}

	if valid := fieldsDataType(fs[spec.C1_DataType]); !valid {
		return spec.C1_DataType, false
	}

	if valid := fieldsMd5(fs[spec.C1_FileMD5], spec.IndexC1); !valid {
		return spec.C1_FileMD5, false
	}

	if valid := fieldsNull(fs[spec.C1_GatherTime]); !valid {
		return spec.C1_GatherTime, false
	}

	return 0, true
}

func recordC1Info(fs []string) {
	l7Proto, _ := strconv.Atoi(fs[spec.C1_Proto])
	fileType, _ := strconv.Atoi(fs[spec.C1_FileType])

	info := SampleC1Info{
		Event:    fs[spec.C1_EventTypeID] + "," + fs[spec.C1_EventSubType],
		L7Proto:  l7Proto,
		FileType: fileType,
		FileSize: fs[spec.C1_FileSize],
	}

	SampleMapUpdateC1(&SampleMap, fs[spec.C1_FileMD5], info)
}

func procC1Ctx(line, filename string) {
	fs := strings.Split(line, "|")
	if len(fs) != 21 {
		//fmt.Printf("invalid log:[%s]\n", line)
		info := CheckInfo{
			Reason:   fmt.Sprintf("字段个数%d不符", len(fs)),
			Filenmae: filename,
		}
		C1_CheckMap[line] = info
		incLogInvalidCnt(spec.IndexC1)
		return
	}

	if index, valid := procC1Fields(fs); valid {
		recordC1Info(fs)
		incLogValidCnt(spec.IndexC1)
	} else {
		//fmt.Printf("invalid log:[%s]\n", line)
		info := CheckInfo{
			Reason:   fmt.Sprintf("第%d个字段非法", index+1),
			Filenmae: filename,
		}
		C1_CheckMap[line] = info
		incLogInvalidCnt(spec.IndexC1)
	}
}
