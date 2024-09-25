package fileproc

import (
	"ds_tool/spec"
	"fmt"
	"strconv"
	"strings"
	"sync"
)

type SampleC0Info struct {
	Data        string
	MatchNum    string
	Application int
	Business    int
	CrossBoard  string
}

var (
	C0_CheckMap map[string]CheckInfo //识别话单必填项校验
)

func SampleMapUpdateC0(m *sync.Map, md5 string, info SampleC0Info) {
	value, ok := m.Load(md5)
	if ok {
		sample := value.(*SampleMapValue)
		sample.C0Info = append(sample.C0Info, info)
	} else {
		sample := &SampleMapValue{}
		sample.C0Info = append(sample.C0Info, info)
		m.Store(md5, sample)
	}
}

func procC0Fields(fs []string) (int, bool) {
	if valid := feildsLogid(fs[spec.C0_LogID]); !valid {
		return spec.C0_LogID, false
	}

	if valid := fieldsNull(fs[spec.C0_CommandID]); !valid {
		return spec.C0_CommandID, false
	}

	if valid := fieldsNull(fs[spec.C0_House_ID]); !valid {
		return spec.C0_House_ID, false
	}

	if valid := fieldsFileType(fs[spec.C0_DataFileType]); !valid {
		return spec.C0_DataFileType, false
	}

	if valid := fieldsNull(fs[spec.C0_AssetsSize]); !valid {
		return spec.C0_AssetsSize, false
	}

	if valid := fieldsNull(fs[spec.C0_AssetsNum]); !valid {
		return spec.C0_AssetsNum, false
	}

	datainfoGroup, _ := strconv.Atoi(fs[spec.C0_DataInfoNum])
	if valid := fieldsIntZero(datainfoGroup); !valid {
		return spec.C0_DataInfoNum, false
	}

	offset := 0
	if datainfoGroup > 0 {
		offset = (datainfoGroup - 1) * 3
	}

	for i := 0; i < offset+3; i++ {
		if valid := fieldsDataInfo(fs[spec.C0_DataType+i], i%3); !valid {
			return spec.C0_DataType + i%3, false
		}
	}

	if valid := fieldsUpload(fs[spec.C0_IsUploadFile+offset]); !valid {
		return spec.C0_IsUploadFile, false
	}

	if valid := fieldsMd5(fs[spec.C0_FileMD5+offset], spec.IndexC0); !valid {
		return spec.C0_FileMD5, false
	}

	if valid := fieldsNull(fs[spec.C0_CurTime+offset]); !valid {
		return spec.C0_CurTime, false
	}

	if valid := fieldsNull(fs[spec.C0_SrcIP+offset]); !valid {
		return spec.C0_SrcIP, false
	}

	if valid := fieldsNull(fs[spec.C0_DestIP+offset]); !valid {
		return spec.C0_DestIP, false
	}

	if valid := fieldsNull(fs[spec.C0_SrcPort+offset]); !valid {
		return spec.C0_SrcPort, false
	}

	if valid := fieldsNull(fs[spec.C0_DestPort+offset]); !valid {
		return spec.C0_DestPort, false
	}

	if valid := fieldsL4Proto(fs[spec.C0_ProtocolType+offset]); !valid {
		return spec.C0_ProtocolType, false
	}

	if valid := fieldsAppProto(fs[spec.C0_ApplicationProtocol+offset]); !valid {
		return spec.C0_ApplicationProtocol, false
	}

	if valid := fieldsBusProto(fs[spec.C0_BusinessProtocol+offset]); !valid {
		return spec.C0_BusinessProtocol, false
	}

	if valid := fieldsMatch(fs[spec.C0_IsMatchEvent+offset]); !valid {
		return spec.C0_IsMatchEvent, false
	}

	return 0, true
}

func recordC0Info(fs []string) {
	datainfoGroup, _ := strconv.Atoi(fs[spec.C0_DataInfoNum])
	var data string
	for i := 0; i < datainfoGroup; i++ {
		if i == 0 {
			data = fs[spec.C0_DataType+i*3]
		} else {
			data = data + "|" + fs[spec.C0_DataType+i*3]
		}
		data += fs[spec.C0_DataType+i*3]
		data += "," + fs[spec.C0_DataType+i*3+1]
		data += "," + fs[spec.C0_DataType+i*3+2]
	}
	offset := 0
	if datainfoGroup > 0 {
		offset = (datainfoGroup - 1) * 3
	}

	matchNum := fs[spec.C0_AssetsNum]
	application, _ := strconv.Atoi(fs[spec.C0_ApplicationProtocol+offset])
	business, _ := strconv.Atoi(fs[spec.C0_BusinessProtocol+offset])
	cross := fs[spec.C0_IsMatchEvent+offset]

	info := SampleC0Info{
		Data:        data,
		MatchNum:    matchNum,
		Application: application,
		Business:    business,
		CrossBoard:  cross,
	}

	SampleMapUpdateC0(&SampleMap, fs[spec.C0_FileMD5+offset], info)
}

func procC0Ctx(line, filename string) {
	fs := strings.Split(line, "|")
	if len(fs) < spec.C0_Max {
		info := CheckInfo{
			Reason:   fmt.Sprintf("字段个数%d不符", len(fs)),
			Filenmae: filename,
		}
		C0_CheckMap[line] = info
		incLogInvalidCnt(spec.IndexC0)
		return
	}
	datainfoGroup, _ := strconv.Atoi(fs[spec.C0_DataInfoNum])
	if datainfoGroup > 1 {
		nums := spec.C0_Max + (datainfoGroup-1)*3
		if len(fs) != nums {
			info := CheckInfo{
				Reason:   fmt.Sprintf("字段个数%d不符", len(fs)),
				Filenmae: filename,
			}
			C0_CheckMap[line] = info
			incLogInvalidCnt(spec.IndexC0)
			return
		}
	} else {
		if len(fs) != spec.C0_Max {
			//fmt.Printf("invalid log:[%s]\n", line)
			info := CheckInfo{
				Reason:   fmt.Sprintf("字段个数%d不符", len(fs)),
				Filenmae: filename,
			}
			C0_CheckMap[line] = info
			incLogInvalidCnt(spec.IndexC0)
			return
		}
	}

	if index, valid := procC0Fields(fs); valid {
		recordC0Info(fs)
		incLogValidCnt(spec.IndexC0)
	} else {
		//fmt.Printf("invalid log:[%s]\n", line)
		info := CheckInfo{
			Reason:   fmt.Sprintf("第%d个字段非法", index+1),
			Filenmae: filename,
		}
		C0_CheckMap[line] = info
		incLogInvalidCnt(spec.IndexC0)
	}
}
