package fileproc

import (
	"ds_tool/spec"
	"fmt"
	"strings"
	"sync"
)

type SampleC4Info struct {
	Keyword string
}

var (
	C4_CheckMap map[string]CheckInfo //关键字话单必填项校验
)

func SampleMapUpdateC4(m *sync.Map, md5 string, info SampleC4Info) {
	value, ok := m.Load(md5)
	if ok {
		sample := value.(*SampleMapValue)
		sample.C4Info = append(sample.C4Info, info)
	} else {
		sample := &SampleMapValue{}
		sample.C4Info = append(sample.C4Info, info)
		m.Store(md5, sample)
	}
}

func procC4Fields(fs []string) (int, bool) {
	if valid := fieldsMd5(fs[spec.C4_FileMD5], spec.IndexC4); !valid {
		return spec.C4_FileMD5, false
	}

	return 0, true
}

func procC4Ctx(line, filename string) {
	fs := strings.Split(line, "|")
	if len(fs) != spec.C4_Max {
		info := CheckInfo{
			Reason:   fmt.Sprintf("字段个数%d不符", len(fs)),
			Filenmae: filename,
		}
		C4_CheckMap[line] = info
		incLogInvalidCnt(spec.IndexC4)
		return
	}

	if index, valid := procC4Fields(fs); valid {
		incLogValidCnt(spec.IndexC4)
	} else {
		info := CheckInfo{
			Reason:   fmt.Sprintf("第%d个字段非法", index+1),
			Filenmae: filename,
		}
		C4_CheckMap[line] = info
		incLogInvalidCnt(spec.IndexC4)
	}
}
