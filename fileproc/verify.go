package fileproc

import (
	"archive/tar"
	"bufio"
	"compress/gzip"
	"ds_tool/comm"
	"ds_tool/spec"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
)

var verifyMap sync.Map

type verifyRes struct {
	res     []string
	correct bool
	errinfo string
}

type AuditStat struct {
	Files [LogIndexMax]int
	Lines [LogIndexMax]int
}

type AuditInfo struct {
	FileName string
	FileFlag uint64
	HasFile  bool
}

var auditStat AuditStat
var auditMap sync.Map
var auditExtraLog []string

var auditFlag uint64 = (1 << 1) | (1 << 3)

func loadMd5Dict(mpath string) error {
	if exist := comm.PathExists(mpath); !exist {
		return fmt.Errorf(fmt.Sprintf("path %s not exist, skip it", mpath))
	}

	// 打开文件
	file, err := os.Open(mpath)
	if err != nil {
		return err
	}
	defer file.Close()

	var count int
	// 创建一个Scanner
	scanner := bufio.NewScanner(file)
	// 逐行读取文件内容
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		fs := strings.Split(line, "|")
		if len(fs) < 3 {
			fmt.Printf("格式错误：%s\n", line)
			continue
		}
		nums, err := strconv.Atoi(fs[2])
		if err != nil {
			fmt.Printf("格式错误：%s\n", line)
			continue
		}
		if nums == 0 || len(fs) != 3+nums {
			fmt.Printf("格式错误：%s\n", line)
			continue
		}

		var res []string
		for i := 3; i < len(fs); i++ {
			res = append(res, fs[i])
		}

		rs := &verifyRes{
			res: res,
		}

		count++
		md5 := strings.ToLower(strings.TrimSpace(fs[0]))
		verifyMap.Store(md5, rs)
	}

	fmt.Printf(">>> Load Map Nums: %d\n", count)
	return nil
}

func checkStringSliceEqual(a, b []string) bool {
	sort.StringSlice.Sort(a)
	sort.StringSlice.Sort(b)

	if len(a) != len(b) {
		return false
	}

	lena := len(a)
	for i := 0; i < lena; i++ {
		if a[i] != b[i] {
			return false
		}
	}

	return true
}

func extractC0Res(line string) {
	fs := strings.Split(line, "|")
	if len(fs) < spec.C0_Max {
		fmt.Printf("字段数有误：%s\n", line)
		return
	}
	datainfoGroup, err := strconv.Atoi(fs[spec.C0_DataInfoNum])
	if err != nil || datainfoGroup == 0 {
		fmt.Printf("识别结果组别提取有误：%s\n", line)
		return
	}

	nums := spec.C0_Max + (datainfoGroup-1)*3
	if len(fs) != nums {
		fmt.Printf("字段数有误：%s\n", line)
		return
	}

	md5 := strings.ToLower(fs[spec.C0_FileMD5+(datainfoGroup-1)*3])
	value, ok := verifyMap.Load(md5)
	if ok {
		v := value.(*verifyRes)
		var res []string
		for i := 0; i < datainfoGroup; i++ {
			res = append(res, fs[spec.C0_DataContent+(i)*3])
		}
		valid := checkStringSliceEqual(v.res, res)
		if valid {
			v.correct = true
		} else {
			v.errinfo = fmt.Sprintf("校验失败，预期：%v，话单：%v", v.res, res)
		}
	}
}

func procC0TargzFile(filename string) error {
	// 打开tar.gz文件
	f, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	// 创建gzip.Reader
	gr, err := gzip.NewReader(f)
	if err != nil {
		return err
	}
	defer gr.Close()

	// 创建tar.Reader
	tr := tar.NewReader(gr)

	// 遍历tar文件中的每个文件
	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		if header.Typeflag != tar.TypeReg {
			continue
		}

		scanner := bufio.NewScanner(tr)

		for scanner.Scan() {
			line := scanner.Text()
			if line == "" {
				continue
			}
			extractC0Res(line)
		}
	}

	return err
}

func FindLogC0Path(path string) error {
	if exist := comm.PathExists(path); !exist {
		return fmt.Errorf(fmt.Sprintf("Path %s not exist, skip it!\n", path))
	}

	err := filepath.WalkDir(path, func(dir string, d fs.DirEntry, err error) error {
		if err != nil {
			fmt.Printf("filepath walk failed:%v\n", err)
			return err
		}

		if !d.IsDir() {
			if strings.HasSuffix(d.Name(), "tar.gz") {
				err := procC0TargzFile(dir)
				if err != nil {
					fmt.Printf("读取tar.gz文件失败：%v\n", err)
				}
			}
		}

		return nil
	})

	if err != nil {
		fmt.Printf("filepath walk failed:%v\n", err)
	}

	return err
}

func PrintVerifyRes() {
	var count int
	var succ int
	var fail int
	var nofound int
	verifyMap.Range(func(key, value interface{}) bool {
		count++
		md5 := key.(string)
		res := value.(*verifyRes)
		if res.correct {
			succ++
		} else if res.errinfo != "" {
			fail++
			fmt.Printf(">>> md5: %s, %s\n", md5, res.errinfo)
		} else {
			nofound++
			fmt.Printf(">>> md5: %s, not found!!!\n", md5)
		}
		return true
	})
	fmt.Printf("----------------------------------------\n")
	fmt.Printf("All: %d|Succ: %d|Fail: %d|Miss %d\n", count, succ, fail, nofound)
}

func VerifyRecogResult(lpath, mpath string) error {
	err := loadMd5Dict(mpath)
	if err != nil {
		fmt.Printf("读取md5表失败：%v\n", err)
		return err
	}

	FindLogC0Path(lpath)

	PrintVerifyRes()
	return nil
}

type Callback func(string, int, string) error

func CallFileProc(filename string, logType int, fn string) error {
	value, ok := auditMap.Load(filename)
	if ok {
		node := value.(*AuditInfo)
		if node.FileFlag != auditFlag {
			fmt.Printf("audit file '%s' create & upload flag error\n", filename)
		}
		node.HasFile = true
	} else {
		fmt.Printf("file '%s' miss audit log\n", filename)
	}
	return nil
}

func CallAuditProc(line string, logType int, fn string) error {
	fs := strings.Split(line, "|")
	if len(fs) != 9 {
		fmt.Printf("invalid audit log: %s\n", line)
		return nil
	}
	if fs[5] == "" {
		fmt.Printf("invalid audit log: %s\n", line)
		return nil
	}

	filetype, err := strconv.ParseUint(fs[7], 10, 64)
	if err != nil {
		fmt.Printf("audit log type invalid: %v\n", line)
		return nil
	}

	filename := filepath.Base(fs[5])
	if strings.HasSuffix(filename, "eu.xml") {
		auditExtraLog = append(auditExtraLog, filename)
		return nil
	}

	auditStat.Lines[logType]++
	value, ok := auditMap.Load(filename)
	if ok {
		node := value.(*AuditInfo)
		res := node.FileFlag & (1 << filetype)
		if res == 1 {
			fmt.Printf("audit log type repeat: %v\n", line)
		} else {
			node.FileFlag |= 1 << filetype
		}
	} else {
		node := &AuditInfo{
			FileName: filename,
			FileFlag: 1 << filetype,
		}
		auditMap.Store(filename, node)
	}

	return nil
}

func TargzFileProc(filename string, logType int, call Callback) error {
	// 打开tar.gz文件
	f, err := os.Open(filename)
	if err != nil {
		fmt.Println("Error opening file:", err)
		return err
	}
	defer f.Close()

	// 创建gzip.Reader
	gr, err := gzip.NewReader(f)
	if err != nil {
		fmt.Println("Error creating gzip reader:", err)
		return err
	}
	defer gr.Close()

	// 创建tar.Reader
	tr := tar.NewReader(gr)

	// 遍历tar文件中的每个文件
	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			fmt.Println("Error reading tar file:", err)
			return err
		}

		if header.Typeflag != tar.TypeReg {
			continue
		}

		scanner := bufio.NewScanner(tr)

		for scanner.Scan() {
			line := scanner.Text()
			if line == "" {
				continue
			}
			call(line, logType, filepath.Base(filename))
		}
	}

	return nil
}

func LoadAuditMd5Map(path string) error {
	if exist := comm.PathExists(path); !exist {
		return fmt.Errorf("Path %s not exist, skip it!\n", path)
	}

	deep1 := strings.Count(path, string(os.PathSeparator))
	err := filepath.WalkDir(path, func(dir string, d fs.DirEntry, err error) error {
		if err != nil {
			fmt.Printf("filepath walk failed:%v\n", err)
			return err
		}

		if d.IsDir() {
			return nil
		}

		deep2 := strings.Count(dir, string(os.PathSeparator))
		if deep2 > deep1+1 {
			//跳过子目录下的文件
			return nil
		}

		if strings.HasPrefix(d.Name(), "0x31+0x04a8") && strings.HasSuffix(d.Name(), "tar.gz") {
			TargzFileProc(dir, AuditIndex, CallAuditProc)
		}

		return nil
	})

	if err != nil {
		fmt.Printf("filepath walk failed:%v\n", err)
	}
	return err
}

func LoadLogPathFile(path string, logType int) error {
	if exist := comm.PathExists(path); !exist {
		return fmt.Errorf("Path %s not exist, skip it!\n", path)
	}

	deep1 := strings.Count(path, string(os.PathSeparator))
	err := filepath.WalkDir(path, func(dir string, d fs.DirEntry, err error) error {
		if err != nil {
			fmt.Printf("filepath walk failed:%v\n", err)
			return err
		}

		if d.IsDir() {
			return nil
		}

		deep2 := strings.Count(dir, string(os.PathSeparator))
		if deep2 > deep1+1 {
			//跳过子目录下的文件
			return nil
		}

		if strings.HasSuffix(d.Name(), "tar.gz") || strings.HasSuffix(d.Name(), "zip") {
			auditStat.Files[logType]++
			CallFileProc(filepath.Base(dir), logType, filepath.Base(dir))
		}

		return nil
	})

	if err != nil {
		fmt.Printf("filepath walk failed:%v\n", err)
	}
	return err
}

func getPathByParam(lpath, logpath, date string) string {
	if date == "" {
		return filepath.Join(lpath, logpath)
	}
	return filepath.Join(lpath, logpath, date, "success")
}

func VerifyAuditFile(lpath string, date string) error {
	//解析审计日志，存表
	path := getPathByParam(lpath, AuditName, date)
	err := LoadAuditMd5Map(path)
	if err != nil {
		fmt.Printf("读取审计日志失败：%v", err)
		return err
	}
	//读取审计日志文件
	path = getPathByParam(lpath, AuditName, date)
	err = LoadLogPathFile(path, AuditIndex)
	if err != nil {
		fmt.Printf("读取审计话单文件失败：%v", err)
		return err
	}
	//读取识别日志文件
	path = getPathByParam(lpath, IdentifyName, date)
	err = LoadLogPathFile(path, IdentifyIndex)
	if err != nil {
		fmt.Printf("读取识别话单文件失败：%v", err)
		return err
	}
	//读取监测日志文件
	path = getPathByParam(lpath, MonitorName, date)
	err = LoadLogPathFile(path, MonitorIndex)
	if err != nil {
		fmt.Printf("读取监测话单文件失败：%v", err)
		return err
	}
	//读取关键字日志文件
	path = getPathByParam(lpath, KeywordName, date)
	if exist := comm.PathExists(path); !exist {
		path = getPathByParam(lpath, KeywordNameB, date)
	}
	err = LoadLogPathFile(path, KeywordIndex)
	if err != nil {
		fmt.Printf("读取关键字话单文件失败：%v", err)
		return err
	}
	//读取取证文件
	path = getPathByParam(lpath, EvidenceName, date)
	err = LoadLogPathFile(path, EvidenceIndex)
	if err != nil {
		fmt.Printf("读取关键字话单文件失败：%v", err)
		return err
	}
	//读取规则库文件
	path = getPathByParam(lpath, RulesName, date)
	err = LoadLogPathFile(path, RulesIndex)
	if err != nil {
		fmt.Printf("读取规则库话单文件失败：%v", err)
		return err
	}

	//遍历map，查询是否所有审计日志都有对应的话单文件
	auditMap.Range(func(key, value interface{}) bool {
		node := value.(*AuditInfo)
		if !node.HasFile {
			fmt.Printf("audit log miss file '%s'\n", node.FileName)
		}
		return true
	})

	nums := 0
	for i := 0; i < LogIndexMax; i++ {
		nums += auditStat.Files[i]
	}

	//打印统计信息
	fmt.Printf("\n")
	for _, v := range auditExtraLog {
		fmt.Printf("手动操作审计日志记录：%s\n", v)
	}
	fmt.Printf("\n")
	fmt.Printf("日志类型\t文件数量\t日志数量\n")
	fmt.Printf("审计日志\t%010d\t%010d\n", auditStat.Files[AuditIndex], auditStat.Lines[AuditIndex])
	fmt.Printf("识别日志\t%010d\t%010d\n", auditStat.Files[IdentifyIndex], auditStat.Files[IdentifyIndex])
	fmt.Printf("监测日志\t%010d\t%010d\n", auditStat.Files[MonitorIndex], auditStat.Files[MonitorIndex])
	fmt.Printf("取证文件\t%010d\t%010d\n", auditStat.Files[EvidenceIndex], auditStat.Files[EvidenceIndex])
	fmt.Printf("规则库  \t%010d\t%010d\n", auditStat.Files[RulesIndex], auditStat.Files[RulesIndex])
	fmt.Printf("关键词  \t%010d\t%010d\n", auditStat.Files[KeywordIndex], auditStat.Files[KeywordIndex])
	fmt.Printf("++++++++++++++++++++++++++++++++++++++++++++\n")
	fmt.Printf("总文件数\t总审计日志数\t是否匹配\n")
	fmt.Printf("%010d\t%010d\t%v\n", nums, auditStat.Lines[AuditIndex], nums*2 == auditStat.Lines[AuditIndex])
	return nil
}
