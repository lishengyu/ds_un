package fileproc

import (
	"ds_tool/comm"
	"ds_tool/conf"

	//"ds_tool/pkt"
	"ds_tool/spec"
	"encoding/base64"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type LogSample struct {
	PreName string
	Fields  []string
}

var (
	OutputPath = "./output"
	OutputLog  = "./output/new.logtar"
	OutputMap  = "./output/filemap.txt"
)

func LogFilePreEnv(out string) {
	//os.RemoveAll(out)
	os.MkdirAll(out, 0666)
}

func getLogFilePrefix(fn string) string {
	name := filepath.Base(fn)
	index := strings.LastIndex(name, "+")
	if index == -1 {
		return fn
	}

	return name[:index]
}

// UpdateLogFields		更新话单文件字段，字段内容来源于conf.json
//
//	@param fields
//	@param cfg
//	@return []string
func UpdateLogFields(dlog *LogSample, cfg conf.DefaultConf) {
	if cfg.CmdId != "" {
		dlog.Fields[spec.Cx_CmdId] = cfg.CmdId
	}

	if cfg.HouseId != "" {
		dlog.Fields[spec.Cx_HouseId] = cfg.HouseId
	}

	if cfg.AccessTime != 0 {
		dlog.Fields[spec.Cx_Time] = fmt.Sprintf("%d", cfg.AccessTime)
	}

	if cfg.Sip != "" {
		dlog.Fields[spec.Cx_Sip] = cfg.Sip
	}

	if cfg.Dip != "" {
		dlog.Fields[spec.Cx_Dip] = cfg.Dip
	}

	if cfg.Sport != 0 {
		dlog.Fields[spec.Cx_Sport] = fmt.Sprintf("%d", cfg.Sport)
	}

	if cfg.Dport != 0 {
		dlog.Fields[spec.Cx_Dport] = fmt.Sprintf("%d", cfg.Dport)
	}

	if cfg.Protocol != 0 {
		dlog.Fields[spec.Cx_L4] = fmt.Sprintf("%d", cfg.Protocol)
	}

	if cfg.Url != "" {
		dlog.Fields[spec.Cx_HttpUrl] = cfg.Url
	}

	if cfg.Domain != "" {
		dlog.Fields[spec.Cx_HttpDomain] = cfg.Domain
	}

	if cfg.HttpMethod != 0 {
		dlog.Fields[spec.Cx_HttpMethod] = fmt.Sprintf("%d", cfg.HttpMethod)
	}

	if cfg.C3Code != 0 {
		dlog.Fields[spec.Cx_AppId] = fmt.Sprintf("%d", cfg.C3Code)
	}

	if cfg.C4Code != 0 {
		dlog.Fields[spec.Cx_AppType] = fmt.Sprintf("%d", cfg.C4Code)
	}

	if cfg.C9Code != 0 {
		dlog.Fields[spec.Cx_DataPro] = fmt.Sprintf("%d", cfg.C9Code)
	}

	if cfg.DataDir != 0 {
		dlog.Fields[spec.Cx_DataDir] = fmt.Sprintf("%d", cfg.DataDir)
	}
}

func ReadSampleLog(logfile string) (*LogSample, error) {
	var dlog LogSample
	data, err := comm.ReadFile(logfile)
	if err != nil {
		return &dlog, err
	}

	line := strings.TrimRight(string(data), "\n")

	fields := strings.Split(string(line), "|")
	if len(fields) != spec.Cx_Max {
		return &dlog, fmt.Errorf(fmt.Sprintf("话单字段有误[%s]\n", string(line)))
	}

	dlog.PreName = getLogFilePrefix(logfile)
	dlog.Fields = append(dlog.Fields, fields...)

	return &dlog, nil
}

func sampleProc(fn string, index int, fields []string, newlogfile string) {
	md5, err := comm.GetFileMd5(fn)
	if err != nil {
		log.Printf("计算文件[%s]Md5失败: %v\n", fn, err)
		return
	}

	size, err := comm.GetFileSize(fn)
	if err != nil {
		log.Printf("获取文件[%s]大小失败：%v\n", fn, err)
		return
	}

	suffix, err := comm.GetSuffix(fn)
	if err != nil {
		log.Printf("获取文件后缀[%s]大小失败：%v\n", fn, err)
		return
	}

	isFormat := comm.FileNameFormat(fn)
	var dst string
	if isFormat {
		//文件名不变
		dst = filepath.Base(fn)
	} else {
		//生成新的文件名
		dst = fmt.Sprintf("%s_%05d_%s", md5, index, suffix)
	}

	comm.CopyFile(fn, filepath.Join(OutputPath, dst))
	//writelogfile
	newline := genLogLine(fields, dst, suffix, size, index)
	index++
	err = comm.AppendFile(newlogfile, []byte(newline))
	if err != nil {
		log.Printf("Error: write file: %v\n", err)
		return
	}

	data := []byte(fn + "|" + dst + "\n")
	err = comm.AppendFile(OutputMap, data)
	if err != nil {
		log.Printf("Error: write map file: %v\n", err)
		return
	}
}

func genNewLogFile(pre string) string {
	tstr := strings.ReplaceAll(time.Now().Format("20060102150405.000000"), ".", "")
	fn := fmt.Sprintf("%s/%s+%s.logtar", OutputPath, pre, tstr)
	return fn
}

func generatePreLog(fields []string, pi *pkt.PktInfo) []string {
	var plog []string

	fields[spec.Cx_Time] = fmt.Sprintf("%d", pi.TimeStamp)
	fields[spec.Cx_Sip] = pi.Tuple.Sip
	fields[spec.Cx_Dip] = pi.Tuple.Dip
	fields[spec.Cx_Sport] = pi.Tuple.Sport
	fields[spec.Cx_Dport] = pi.Tuple.Dport

	if pi.Tuple.Proto == 17 {
		fields[spec.Cx_L4] = "2"
	} else {
		fields[spec.Cx_L4] = "1"
	}

	url := fmt.Sprintf("http://%s%s", pi.Http.Domain, pi.Http.Uri)
	fields[spec.Cx_HttpUrl] = base64.StdEncoding.EncodeToString([]byte(url))
	fields[spec.Cx_HttpDomain] = pi.Http.Domain

	if pi.Http.Method == "GET" {
		fields[spec.Cx_HttpMethod] = "1"
	} else if pi.Http.Method == "POST" {
		fields[spec.Cx_HttpMethod] = "2"
	} else {
		fields[spec.Cx_HttpMethod] = "0"
	}

	fields[spec.Cx_ReportType] = "0"

	plog = append(plog, fields[:spec.Cx_Max]...)

	return plog
}

// GenerateFromSampleLog 	通过样本和话单生成新的logtar话单和符合规范的样本
//
//	@param samplepath
//	@param logFields
//	@return error
func GenerateFromSampleLog(samplepath string, dlog *LogSample) error {
	var index int

	logFn := genNewLogFile(dlog.PreName)
	if ok := comm.IsFile(samplepath); ok {
		sampleProc(samplepath, index, dlog.Fields, logFn)
		return nil
	}

	err := filepath.WalkDir(samplepath, func(dir string, d fs.DirEntry, err error) error {
		if err != nil {
			fmt.Printf("filepath walk failed:%v\n", err)
			return err
		}

		if !d.IsDir() {
			index++
			sampleProc(dir, index, dlog.Fields, logFn)
		}
		return nil
	})

	return err
}

func GenerateFromPcap(sapmlepath, pcap string, dlog *LogSample) error {
	pi, ok := pkt.DissectPcap(pcap)
	if !ok {
		return fmt.Errorf(fmt.Sprintf("解析pcap包[%s]失败\n", pcap))
	}

	prelog := generatePreLog(dlog.Fields, pi)
	logFn := genNewLogFile(dlog.PreName)
	var index int
	if ok := comm.IsFile(sapmlepath); ok {
		sampleProc(sapmlepath, index, prelog, logFn)
		return nil
	}

	err := filepath.WalkDir(sapmlepath, func(dir string, d fs.DirEntry, err error) error {
		if err != nil {
			fmt.Printf("filepath walk failed:%v\n", err)
			return err
		}

		if !d.IsDir() {
			index++
			sampleProc(dir, index, prelog, logFn)
		}
		return nil
	})

	return err
}

// GenLogLine 	构造新的ds话单，更新logid字段，更新文件相关字段（md5/size/filetype）
//
//	@param fs
//	@param md5		更新话单中的md5字段
//	@param suffix	更新话单中的文件类型字段
//	@param size		更新话单中的文件大小字段
//	@param index
//	@return string
func genLogLine(fs []string, md5, suffix, size string, index int) string {
	var buf string
	filecode := spec.GetFileCode(suffix)
	cur := time.Now().Format("20060102150405")
	logid := fmt.Sprintf("%s%06d", cur[2:], index)
	for i, v := range fs {
		if i == spec.Cx_LogId {
			buf = logid
			continue
		}

		buf += "|"
		if i == spec.Cx_FileType {
			buf += filecode
		} else if i == spec.Cx_FileSize {
			buf += size
		} else if i == spec.Cx_FileMd5 {
			buf += md5
		} else {
			buf += v
		}
	}

	return buf + "\n"
}
