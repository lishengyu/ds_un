package fileproc

import (
	"ds_tool/spec"
	"fmt"
	"log"
	"strconv"
	"strings"
	"sync"
)

type Md5HitKey struct {
	Sip   string
	Dip   string
	Sport string
	Dport string
	Md5   string
}

type Md5HitInfo struct {
	FileType    int
	FileSize    string
	AppProto    int
	DataProto   int
	HasCx       int
	HasC0       int
	DataInfo    []string
	HasC1       int
	EventInfo   []string
	HasC3       int
	HasC4       int
	KeywordInfo []string
}

var (
	SheetMd5HitTitle = []string{
		"文件md5", //key
		"文件类型",
		"大小",
		"应用层协议",
		"数据层协议",
		"DPI 日志",
		"识别日志",
		"监测日志",
		"取证文件",
		"关键字日志",
	}

	SheetFileLogTitle = []string{
		"日志类型",
		"文件数量",
		"日志数量",
		"有效日志",
		"空行",
		"错误日志",
	}
)

var (
	Md5HitMap sync.Map //MD5表，比对输出日志类型
)

func Md5HitMapStore(key any, value any) {
	Md5HitMap.Store(key, value)
}

func Md5HitMapLoad(key any) (any, bool) {
	return Md5HitMap.Load(key)
}

func Md5HitMapRange(f func(key, value any) bool) {
	Md5HitMap.Range(f)
}

func NewSheetMd5Line(key, value any) []string {
	md5 := key.(string)
	hit := value.(*Md5HitInfo)

	var fields []string
	fields = append(fields, md5)
	fields = append(fields, fmt.Sprintf("%d(%s)", hit.FileType, spec.C10_DICT[hit.FileType]))
	fields = append(fields, hit.FileSize)
	fields = append(fields, fmt.Sprintf("%d(%s)", hit.AppProto, spec.C3_DICT[hit.AppProto]))
	fields = append(fields, fmt.Sprintf("%d(%s)", hit.DataProto, spec.C9_DICT[hit.DataProto]))
	fields = append(fields, fmt.Sprintf("%v", hit.HasCx))
	fields = append(fields, fmt.Sprintf("%v", hit.HasC0))
	fields = append(fields, fmt.Sprintf("%v", hit.HasC1))
	fields = append(fields, fmt.Sprintf("%v", hit.HasC3))
	fields = append(fields, fmt.Sprintf("%v", hit.HasC4))

	return fields
}

func NewSheetFileLog(logname string, cnt StFileStat) []string {
	var fields []string
	fields = append(fields, logname)
	fields = append(fields, fmt.Sprintf("%05d", cnt.FileNum))
	fields = append(fields, fmt.Sprintf("%05d", cnt.LogNum.AllCnt))
	fields = append(fields, fmt.Sprintf("%05d", cnt.LogNum.ValidCnt))
	fields = append(fields, fmt.Sprintf("%05d", cnt.LogNum.NullCnt))
	fields = append(fields, fmt.Sprintf("%05d", cnt.LogNum.InvalidCnt))
	return fields
}

func getC0DataInfo(fs []string, group int) string {
	start := spec.C0_AssetsNum
	end := spec.C0_DataInfoNum + (group * 3)

	value := fs[start:end]

	var str string
	for _, v := range value {
		if str == "" {
			str = v
		} else {
			str += "|" + v
		}
	}

	return str
}

func checkExist(list []string, str string) bool {
	for _, v := range list {
		if v == str {
			return true
		}
	}
	return false
}

func statC0Fields(fs []string, group int) {
	offset := (group - 1) * 3
	md5 := fs[spec.C0_FileMD5+offset]

	key := &Md5HitKey{
		Sip:   fs[spec.C0_SrcIP+offset],
		Dip:   fs[spec.C0_DestIP+offset],
		Sport: fs[spec.C0_SrcPort+offset],
		Dport: fs[spec.C0_DestPort+offset],
		Md5:   md5,
	}

	datatype, _ := strconv.Atoi(fs[spec.C0_DataFileType])
	appid, _ := strconv.Atoi(fs[spec.C0_ApplicationProtocol+offset])
	dataInfo := getC0DataInfo(fs, group)
	if offset != 0 {
		log.Printf("test %d %s\n", spec.C0_FileMD5+offset, md5)
	}
	value, ok := Md5HitMapLoad(key)
	if ok {
		hit := value.(*Md5HitInfo)
		hit.HasC0 += 1
		if hit.FileType == 0 {
			hit.FileSize = fs[spec.C0_AssetsSize]
			hit.FileType = datatype
			hit.AppProto = appid
		}
		if !checkExist(hit.DataInfo, dataInfo) {
			hit.DataInfo = append(hit.DataInfo, dataInfo)
		}
	} else {
		value := &Md5HitInfo{
			HasC0:    1,
			FileType: datatype,
			FileSize: fs[spec.C0_AssetsSize],
			AppProto: appid,
		}
		value.DataInfo = append(value.DataInfo, dataInfo)
		Md5HitMapStore(key, value)
	}
}

func statC1Fields(fs []string) {
	md5 := fs[spec.C1_FileMD5]
	key := &Md5HitKey{
		Sip:   fs[spec.C1_SrcIP],
		Dip:   fs[spec.C1_DestIP],
		Sport: fs[spec.C1_SrcPort],
		Dport: fs[spec.C1_DestPort],
		Md5:   md5,
	}

	eventInfo := fmt.Sprintf("%s|%s", fs[spec.C1_EventTypeID]+fs[spec.C1_EventSubType])
	value, ok := Md5HitMapLoad(key)
	if ok {
		hit := value.(*Md5HitInfo)
		hit.HasC1 += 1
		if !checkExist(hit.DataInfo, eventInfo) {
			hit.EventInfo = append(hit.EventInfo, eventInfo)
		}

	} else {
		value := &Md5HitInfo{
			HasC1: 1,
		}
		value.EventInfo = append(value.EventInfo, eventInfo)
		Md5HitMapStore(key, value)
	}
}

/*
func statC3Fields(fs []string) {
	md5 := fs[6]

	value, ok := Md5HitMapLoad(md5)
	if ok {
		hit := value.(*Md5HitInfo)
		hit.HasC3 += 1
	} else {
		value := &Md5HitInfo{
			HasC3: 1,
		}
		Md5HitMapStore(md5, value)
	}
}
*/

func statC4Fields(fs []string) {
	md5 := fs[spec.C4_FileMD5]
	key := &Md5HitKey{
		Sip:   fs[spec.C4_SrcIP],
		Dip:   fs[spec.C4_DestIP],
		Sport: fs[spec.C4_ScrPort],
		Dport: fs[spec.C4_DestPort],
		Md5:   md5,
	}

	value, ok := Md5HitMapLoad(key)
	if ok {
		hit := value.(*Md5HitInfo)
		hit.HasC4 += 1
	} else {
		value := &Md5HitInfo{
			HasC4: 1,
		}
		Md5HitMapStore(key, value)
	}
}

func statC0Ctx(line string) {
	fs := strings.Split(line, "|")
	if len(fs) < spec.C0_Max {
		incLogInvalidCnt(spec.IndexC0)
		return
	}
	datainfoGroup, _ := strconv.Atoi(fs[spec.C0_DataInfoNum])
	if datainfoGroup > 1 {
		nums := spec.C0_Max + (datainfoGroup-1)*3
		if len(fs) != nums {
			incLogInvalidCnt(spec.IndexC0)
			return
		}
	} else {
		if len(fs) != spec.C0_Max {
			incLogInvalidCnt(spec.IndexC0)
			return
		}
	}

	incLogValidCnt(spec.IndexC0)
	statC0Fields(fs, datainfoGroup)
}

func statC1Ctx(line string) {
	fs := strings.Split(line, "|")
	if len(fs) != spec.C1_Max {
		incLogInvalidCnt(spec.IndexC1)
		return
	}

	incLogValidCnt(spec.IndexC1)
	statC1Fields(fs)
}

func statC2Ctx(line string) {

}

func statC3Ctx(line string) {
	basename := strings.TrimSuffix(line, ".zip")
	fs := strings.Split(basename, "+")
	if len(fs) != 7 {
		incLogInvalidCnt(spec.IndexC3)
		return
	}

	incLogValidCnt(spec.IndexC3)
	//statC3Fields(fs)
}

func statC4Ctx(line string) {
	fs := strings.Split(line, "|")
	if len(fs) != spec.C4_Max {
		incLogInvalidCnt(spec.IndexC4)
		return
	}

	incLogValidCnt(spec.IndexC4)
	statC4Fields(fs)
}

func statCxCtx(line string) {
	fs := strings.Split(line, "|")
	if len(fs) != spec.Cx_Max {
		incLogInvalidCnt(spec.IndexCx)
		return
	}

	incLogValidCnt(spec.IndexCx)
	statCxFields(fs)
}

func statLogData(ctx string, logType int) error {
	switch logType {
	case spec.IndexC0:
		statC0Ctx(ctx)
	case spec.IndexC1:
		statC1Ctx(ctx)
	case spec.IndexC2:
		//statC2Ctx(ctx)
	case spec.IndexC3:
		statC3Ctx(ctx)
	case spec.IndexC4:
		statC4Ctx(ctx)
	case spec.IndexCx:
		statCxCtx(ctx)
	default:
		log.Printf("Not support log type: %d\n", logType)
	}

	return nil
}
