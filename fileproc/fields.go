package fileproc

import (
	"ds_tool/spec"
	"fmt"
	"strconv"
	"strings"
	"sync"
)

type CheckInfo struct {
	Reason   string
	Filenmae string
}

var (
	LogidMap      sync.Map //logid是否重复判断
	AppProtoStat  sync.Map //统计应用层协议情况
	BusProtoStat  sync.Map //统计业务层协议情况
	FileTypeStat  sync.Map //文件类型，文件后缀类型
	DataProtoStat sync.Map //数据识别协议类型
)

func fieldsNull(key string) bool {
	return key != ""
}

func fieldsNullZero(key string) bool {
	if key == "" || key == "0" {
		return false
	}
	return true
}

func fieldsIntZero(num int) bool {
	return num != 0
}

func fieldsUpload(key string) bool {
	return key == "0"
}

func fieldsDataInfo(key string, index int) bool {
	flag := false
	switch index {
	case 0:
		id, _ := strconv.Atoi(key)
		if id >= 1 && id <= 2 {
			flag = true
		}
	case 1:
		id, _ := strconv.Atoi(key)
		if id >= 1 && id <= 6 {
			flag = true
		}
	case 2:
		fs := strings.Split(key, ",")
		if len(fs) == 2 {
			flag = true
		}
	}
	return flag
}

func fieldsL4Proto(key string) bool {
	if key != "1" && key != "2" {
		return false
	}
	return true
}

func fieldsMatch(key string) bool {
	if key != "0" && key != "1" {
		return false
	}

	return true
}

func fieldsHttp(proto, key string) bool {
	if proto == "1" {
		if key == "" {
			return false
		}
	}

	return true
}

func fieldsEvent(id, subid string) bool {
	if id == "" || subid == "" {
		return false
	}

	//校验正确性，存map表

	return true
}

func fieldsDataType(id string) bool {
	if id != "1" && id != "2" {
		return false
	}
	return true
}

func feildsLogid(key string) bool {
	if key == "" || len(key) != 32 {
		return false
	}

	LogidMapStoreInc(&LogidMap, key)

	return true
}

func fieldsMd5(key string, logType int) bool {
	md5 := strings.ToUpper(key)
	Md5Map[logType].Store(md5, 1)
	return true
}

func fieldsFileType(key string) bool {
	id, err := strconv.Atoi(key)
	if err != nil {
		fmt.Printf("transfer string to int failed: %v\n", err)
		return false
	}

	value, ok := spec.C10_DICT[id]
	if !ok {
		//fmt.Printf("app proto value is not in rfc: [%d]\n", id)
		LogidMapStoreInc(&FileTypeStat, "illegal:"+key)
		return false
	}

	LogidMapStoreInc(&FileTypeStat, value)
	return true
}

func fieldsAppProto(key string) bool {
	id, err := strconv.Atoi(key)
	if err != nil {
		fmt.Printf("transfer string to int failed: %v\n", err)
		return false
	}

	value, ok := spec.C3_DICT[id]
	if !ok {
		//fmt.Printf("app proto value is not in rfc: [%d]\n", id)
		LogidMapStoreInc(&AppProtoStat, "illegal:"+key)
		return false
	}

	LogidMapStoreInc(&AppProtoStat, value)
	return true
}

func fieldsBusProto(key string) bool {
	id, err := strconv.Atoi(key)
	if err != nil {
		fmt.Printf("transfer string to int failed: %v\n", err)
		return false
	}

	value, ok := spec.C4_DICT[id]
	if !ok {
		//fmt.Printf("business proto value is not in rfc: [%d]\n", id)
		LogidMapStoreInc(&BusProtoStat, "illegal:"+key)
		return false
	}

	LogidMapStoreInc(&BusProtoStat, value)
	return true
}

func fieldsDataProto(key string) bool {
	id, err := strconv.Atoi(key)
	if err != nil {
		fmt.Printf("transfer string to int failed: %v\n", err)
		return false
	}

	value, ok := spec.C9_DICT[id]
	if !ok {
		//fmt.Printf("business proto value is not in rfc: [%d]\n", id)
		LogidMapStoreInc(&DataProtoStat, "illegal:"+key)
		return false
	}

	LogidMapStoreInc(&DataProtoStat, value)
	return true
}

func LogidMapStoreInc(m *sync.Map, id string) {
	value, ok := m.LoadOrStore(id, 1)
	if ok {
		cnt := value.(int) + 1
		m.Store(id, cnt)
	}
}

func GetMd5FromAttach(fn string) string {
	if fn == "" {
		return fn
	}
	fs := strings.Split(fn, "_")
	return strings.ToUpper(fs[0])
}
