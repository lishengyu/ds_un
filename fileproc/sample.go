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
	"strconv"
	"strings"
	"sync"
)

var (
	UnrecoverPath = "./miss"
	ChangePath    = "./change"
)

type MapKey struct {
	Cmdid string
	Md5   string
}

type MapValue struct {
	Md5   string
	Count [LogIndexMax]int
}

var (
	MapEvidence sync.Map
	MapIdentify sync.Map
	MapMonitor  sync.Map
	MapKeyword  sync.Map
)

// CompareSampleFile 以 dst未标准，查找src，是否有不存在的文件
//
//	@param src
//	@param dst
func CompareSampleFile(src, dst string) error {
	files, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	existMap := make(map[string]string, 0)
	for _, file := range files {
		bn := filepath.Base(file.Name())
		fs := strings.Split(bn, "_")
		if len(fs) != 3 {
			fmt.Printf("文件格式不符: %s\n", bn)
			continue
		}
		md5 := strings.ToLower(fs[0])
		existMap[md5] = bn
	}

	df, err := os.ReadDir(dst)
	if err != nil {
		return err
	}

	var found []string
	var notFound []string

	os.MkdirAll(UnrecoverPath, 0644)

	for _, file := range df {
		bn := filepath.Base(file.Name())
		fs := strings.Split(bn, "_")
		if len(fs) != 3 {
			fmt.Printf("文件格式不符: %s\n", bn)
			continue
		}
		md5 := strings.ToLower(fs[0])
		value, ok := existMap[md5]
		if ok {
			found = append(found, value)
		} else {
			notFound = append(notFound, bn)
			dstfile := filepath.Join(dst, file.Name())
			comm.CopyFile(dstfile, filepath.Join(UnrecoverPath, bn))
		}
	}

	fmt.Printf("已还原样本：========> %d\n", len(found))
	for _, v := range found {
		fmt.Println(v)
	}

	fmt.Printf("未还原样本：========> %d  拷贝路径: %s\n", len(notFound), UnrecoverPath)
	for _, v := range notFound {
		fmt.Println(v)
	}

	return nil
}

func FoundBackMd5Line(filename, md5 string) string {
	// 打开文件
	file, err := os.Open(filename)
	if err != nil {
		fmt.Println("Error opening file:", err)
		return ""
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}
		if !strings.Contains(line, strings.ToLower(md5)) {
			continue
		}
		return line
	}

	return ""
}

func FoundBackMd5Dir(path, md5 string) (*LogSample, error) {
	var dlog LogSample

	if exist := comm.PathExists(path); !exist {
		return &dlog, fmt.Errorf("path %s not exist, skip it", path)
	}

	err := filepath.WalkDir(path, func(dir string, d fs.DirEntry, err error) error {
		if err != nil {
			fmt.Printf("filepath walk failed:%v\n", err)
			return err
		}

		if !d.IsDir() {
			if strings.HasSuffix(d.Name(), "logtar") {
				line := FoundBackMd5Line(dir, md5)
				if line != "" {
					fs := strings.Split(line, "|")
					dlog.PreName = getLogFilePrefix(dir)
					dlog.Fields = append(dlog.Fields, fs...)
					return io.EOF
				}
			}
		}

		return nil
	})

	if err != io.EOF {
		return &dlog, err
	}

	return &dlog, nil
}

func readTargzFile(filename, md5 string) (bool, error) {
	// 打开tar.gz文件
	f, err := os.Open(filename)
	if err != nil {
		fmt.Println("Error opening file:", err)
		return false, err
	}
	defer f.Close()

	// 创建gzip.Reader
	gr, err := gzip.NewReader(f)
	if err != nil {
		fmt.Println("Error creating gzip reader:", err)
		return false, err
	}
	defer gr.Close()

	var hitMd5 bool

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
			return false, err
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
			if strings.Contains(strings.ToLower(line), md5) {
				hitMd5 = true
				break
			}
		}
		if hitMd5 {
			break
		}
	}

	return hitMd5, nil
}

func extractAuditFile(path, filename, md5 string) {
	if exist := comm.PathExists(path); !exist {
		fmt.Printf("Path %s not exist, skip it!\n", path)
		return
	}

	err := filepath.WalkDir(path, func(dir string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if !d.IsDir() {
			if strings.HasSuffix(d.Name(), "tar.gz") {
				hit, err := readTargzFile(dir, filename)
				if err != nil {
					fmt.Printf("read targz file failed: %v\n", err)
				}
				if hit {
					fmt.Printf("拷贝文件: %s\n", dir)
					dn := filepath.Join(ChangePath, strings.ToLower(md5), filepath.Base(dir))
					done := comm.CopyFile(dir, dn)
					if done {
						os.Remove(dir)
					} else {
						fmt.Printf("拷贝文件失败：%s\n", dir)
					}
					return nil
				}
			}
		}

		return nil
	})

	if err != nil {
		fmt.Printf("filepath walk failed:%v\n", err)
	}
}

func extractLogFile(gpath, path, md5 string) {
	if exist := comm.PathExists(path); !exist {
		fmt.Printf("Path %s not exist, skip it!\n", path)
		return
	}

	err := filepath.WalkDir(path, func(dir string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if !d.IsDir() {
			if strings.HasSuffix(d.Name(), "tar.gz") {
				hit, err := readTargzFile(dir, md5)
				if err != nil {
					fmt.Printf("read targz file failed: %v\n", err)
				}
				if hit {
					fmt.Printf("拷贝文件: %s\n", dir)
					dn := filepath.Join(ChangePath, strings.ToLower(md5), filepath.Base(dir))
					done := comm.CopyFile(dir, dn)
					if done {
						os.Remove(dir)
					} else {
						fmt.Printf("拷贝文件失败：%s\n", dir)
					}
					extractAuditFile(filepath.Join(gpath, AuditName), dir, md5)
				}
			}
		}

		return nil
	})

	if err != nil {
		fmt.Printf("filepath walk failed:%v\n", err)
	}
}

func uncompressTargz(filename string) error {
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

		targetPath := filepath.Join(filepath.Dir(filename), header.Name)
		fileOut, err := os.Create(targetPath)
		if err != nil {
			fmt.Printf("解压文件失败：%v\n", err)
			return err
		}
		defer fileOut.Close()

		if _, err := io.Copy(fileOut, tr); err != nil {
			fmt.Printf("解压文件失败：%v\n", err)
			return err
		}
	}

	return nil
}

func uncompressDir(path string) {
	if exist := comm.PathExists(path); !exist {
		fmt.Printf("Path %s not exist, skip it!\n", path)
	}

	err := filepath.WalkDir(path, func(dir string, d fs.DirEntry, err error) error {
		if err != nil {
			fmt.Printf("filepath walk failed:%v\n", err)
			return err
		}

		if !d.IsDir() {
			if strings.HasSuffix(d.Name(), "tar.gz") {
				uncompressTargz(dir)
				os.Remove(dir)
			}
		}

		return nil
	})

	if err != nil {
		fmt.Printf("filepath walk failed:%v\n", err)
	}
}

func ExtractMd5File(path, md5 string) {
	LogFilePreEnv(filepath.Join(ChangePath, strings.ToLower(md5)))
	//识别
	extractLogFile(path, filepath.Join(path, IdentifyName), md5)
	//监测
	extractLogFile(path, filepath.Join(path, MonitorName), md5)
	//关键字生成话单和备份话单文件名不一致，需要再次处理一下
	kfile := filepath.Join(path, KeywordName)
	if exist := comm.PathExists(kfile); !exist {
		kfile = filepath.Join(path, KeywordNameB)
	}
	extractLogFile(path, kfile, md5)

	md5path := filepath.Join(ChangePath, strings.ToLower(md5))
	//解压对应的压缩文件
	uncompressDir(md5path)

	fmt.Printf(">>>>>拷贝路径：%s\n", md5path)
}

func targzFile(filename string) {
	dstfile := strings.ReplaceAll(filename, ".txt", ".tar.gz")
	defer fmt.Printf("生成压缩文件: %s\n", dstfile)

	d, _ := os.Create(dstfile)
	defer d.Close()
	gw := gzip.NewWriter(d)
	defer gw.Close()
	tw := tar.NewWriter(gw)
	defer tw.Close()

	fr, err := os.Open(filename)
	if err != nil {
		fmt.Println(err)
	}
	defer fr.Close()

	info, err := fr.Stat()
	if err != nil {
		fmt.Println(err)
	}

	h := new(tar.Header)
	h.Name = info.Name()
	h.Size = info.Size()
	h.Mode = int64(info.Mode())
	h.ModTime = info.ModTime()

	err = tw.WriteHeader(h)
	if err != nil {
		fmt.Println(err)
	}

	_, err = io.Copy(tw, fr)
	if err != nil {
		fmt.Println(err)
	}
}

func CompressLogtar(path string) {
	if exist := comm.PathExists(path); !exist {
		fmt.Printf("Path %s not exist, skip it!\n", path)
	}

	err := filepath.WalkDir(path, func(dir string, d fs.DirEntry, err error) error {
		if err != nil {
			fmt.Printf("filepath walk failed:%v\n", err)
			return err
		}

		if !d.IsDir() {
			if strings.HasSuffix(d.Name(), "txt") {
				targzFile(dir)
				//os.Remove(dir)
			}
		}

		return nil
	})

	if err != nil {
		fmt.Printf("filepath walk failed:%v\n", err)
	}
}

func getAuditFiles(path string) ([]string, error) {
	var files []string
	if exist := comm.PathExists(path); !exist {
		fmt.Printf("Path %s not exist, skip it!\n", path)
		return files, nil
	}

	err := filepath.WalkDir(path, func(dir string, d fs.DirEntry, err error) error {
		if err != nil {
			fmt.Printf("filepath walk failed:%v\n", err)
			return err
		}

		if !d.IsDir() {
			if strings.HasPrefix(d.Name(), "0x31+0x04a8") && strings.HasSuffix(d.Name(), ".tar.gz") {
				files = append(files, dir)
			}
		}

		return nil
	})

	if err != nil {
		fmt.Printf("filepath walk failed:%v\n", err)
	}

	return files, err
}

func RemoveAudit(path string) {
	files, err := getAuditFiles(ChangePath)
	if err != nil {
		fmt.Printf("get audit files failed: %v\n", err)
		return
	}

	if len(files) == 0 {
		fmt.Printf("no 04a8 files found!!\n")
		return
	}

	for _, file := range files {
		fn := filepath.Base(file)
		dn := filepath.Join(path, AuditName, fn)
		err := os.Remove(dn)
		if err != nil {
			fmt.Printf("Remove Audit File: %s fail!\n", dn)
		} else {
			fmt.Printf("Remove Audit File: %s succ!\n", dn)
		}
	}
}

func LoadEvidenceFile(path string, logType int) error {
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

		if !strings.HasSuffix(d.Name(), "zip") {
			return nil
		}

		deep2 := strings.Count(dir, string(os.PathSeparator))
		if deep2 > deep1+1 {
			//跳过子目录下的文件
			return nil
		}

		bn := filepath.Base(dir)
		fs := strings.Split(bn, "+")
		if len(fs) != 7 {
			return nil
		}

		key := MapKey{
			Cmdid: fs[2],
			Md5:   strings.TrimSuffix(fs[6], ".zip"),
		}

		value, ok := MapEvidence.Load(key)
		if ok {
			v := value.(*MapValue)
			v.Count[EvidenceIndex]++
		} else {
			v := &MapValue{
				Md5: fs[6],
			}
			v.Count[EvidenceIndex]++
			MapEvidence.Store(key, v)
		}

		return nil
	})

	if err != nil {
		fmt.Printf("filepath walk failed:%v\n", err)
	}
	return err
}

func LoadTarGzFile(filename string, logType int) error {
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
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}

			fs := strings.Split(line, "|")

			var key MapKey
			if logType == IdentifyIndex {
				if len(fs) < spec.C0_Max {
					continue
				}

				group, _ := strconv.Atoi(fs[spec.C0_DataInfoNum])
				offset := (group - 1) * 3
				key.Cmdid = fs[spec.C0_CommandID]
				key.Md5 = fs[spec.C0_FileMD5+offset]

				value, ok := MapIdentify.Load(key)
				if ok {
					v := value.(*MapValue)
					v.Count[IdentifyIndex]++
				} else {
					var v MapValue
					v.Md5 = key.Md5
					v.Count[IdentifyIndex]++
					MapIdentify.Store(key, &v)
				}

			} else if logType == MonitorIndex {
				key.Cmdid = fs[spec.C1_CommandId]
				key.Md5 = fs[spec.C1_FileMD5]

				if len(fs) < spec.C1_Max {
					continue
				}
				value, ok := MapMonitor.Load(key)
				if ok {
					v := value.(*MapValue)
					v.Count[MonitorIndex]++
				} else {
					var v MapValue
					v.Md5 = key.Md5
					v.Count[MonitorIndex]++
					MapMonitor.Store(key, &v)
				}
			} else if logType == KeywordIndex {
				key.Cmdid = fs[spec.C4_CommandId]
				key.Md5 = fs[spec.C4_FileMD5]

				if len(fs) < spec.C4_Max {
					continue
				}

				value, ok := MapKeyword.Load(key)
				if ok {
					v := value.(*MapValue)
					v.Count[KeywordIndex]++
				} else {
					var v MapValue
					v.Md5 = key.Md5
					v.Count[KeywordIndex]++
					MapKeyword.Store(key, &v)
				}
			} else {
				fmt.Printf("不支持的话单类型")
			}
		}
	}

	return nil
}

func walkTargzFile(path string, logType int) error {
	deep1 := strings.Count(path, string(os.PathSeparator))
	fmt.Printf("walk dir %s\n", path)
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

		if strings.HasSuffix(d.Name(), "tar.gz") {
			LoadTarGzFile(dir, logType)
		}

		return nil
	})

	return err
}

func VerifySampleRelation(dir, date string) error {
	//取证文件存表
	path := getPathByParam(dir, EvidenceName, date)
	err := LoadEvidenceFile(path, EvidenceIndex)
	if err != nil {
		fmt.Printf("读取取证文件失败：%v", err)
		return err
	}

	//识别文件存表
	path = getPathByParam(dir, IdentifyName, date)
	err = walkTargzFile(path, IdentifyIndex)
	if err != nil {
		fmt.Printf("读取识别日志失败：%v", err)
		return err
	}

	//监测文件存表
	path = getPathByParam(dir, MonitorName, date)
	err = walkTargzFile(path, MonitorIndex)
	if err != nil {
		fmt.Printf("读取监测日志失败：%v", err)
		return err
	}

	//关键字文件存表
	path = getPathByParam(dir, KeywordName, date)
	if exist := comm.PathExists(path); !exist {
		path = getPathByParam(dir, KeywordNameB, date)
	}

	err = walkTargzFile(path, KeywordIndex)
	if err != nil {
		fmt.Printf("读取关键字日志失败：%v", err)
		return err
	}

	//log话单 ==> 取证文件
	var recordC0 int
	var missSample1 int
	MapIdentify.Range(func(key, value interface{}) bool {
		k := key.(MapKey)
		value, ok := MapEvidence.Load(key)
		if ok {
			v := value.(*MapValue)
			v.Count[IdentifyIndex]++
			MapEvidence.Store(key, v)
		} else {
			missSample1++
			fmt.Printf("MD5:%s, CmdId:%s C0缺失取证文件\n", k.Md5, k.Cmdid)
		}
		recordC0++
		return true
	})
	var recordC1 int
	var missSample2 int
	MapMonitor.Range(func(key, value interface{}) bool {
		k := key.(MapKey)
		value, ok := MapEvidence.Load(key)
		if ok {
			v := value.(*MapValue)
			v.Count[MonitorIndex]++
		} else {
			missSample2++
			fmt.Printf("MD5:%s, CmdId:%s C1缺失取证文件\n", k.Md5, k.Cmdid)
		}
		recordC1++
		return true
	})
	var recordC4 int
	var missSample3 int
	MapKeyword.Range(func(key, value interface{}) bool {
		k := key.(MapKey)
		value, ok := MapEvidence.Load(key)
		if ok {
			v := value.(*MapValue)
			v.Count[KeywordIndex]++
		} else {
			missSample3++
			fmt.Printf("MD5:%s, CmdId:%s C4缺失取证文件\n", k.Md5, k.Cmdid)
		}
		recordC4++
		return true
	})

	var recordC3 int
	var missC0 int
	var missC1 int
	//var missC4 int
	MapEvidence.Range(func(key, value interface{}) bool {
		k := key.(MapKey)
		v := value.(*MapValue)
		recordC3++

		if v.Count[IdentifyIndex] > 0 && v.Count[MonitorIndex] > 0 && v.Count[KeywordIndex] > 0 {
			return true
		} else if v.Count[IdentifyIndex] > 0 && v.Count[MonitorIndex] > 0 {
			return true
		} else if v.Count[KeywordIndex] > 0 {
			return true
		} else if v.Count[IdentifyIndex] == 0 {
			missC0++
			fmt.Printf("取证文件 MD5:%s, CmdId:%s 缺失C0话单\n", k.Md5, k.Cmdid)
		} else if v.Count[MonitorIndex] == 0 {
			missC1++
			fmt.Printf("取证文件 MD5:%s, CmdId:%s 缺失C1话单\n", k.Md5, k.Cmdid)
		}

		return true
	})

	fmt.Printf("\n")
	fmt.Printf("================\n")
	fmt.Printf("C0话单缺取证文件：%d\n", missSample1)
	fmt.Printf("C1话单缺取证文件：%d\n", missSample2)
	fmt.Printf("C4话单缺取证文件：%d\n", missSample3)

	fmt.Printf("\n")
	fmt.Printf("================\n")
	fmt.Printf("取证文件缺C0话单：%d\n", missC0)
	fmt.Printf("取证文件缺C1话单：%d\n", missC1)
	//fmt.Printf("取证文件缺C4话单：%d\n", missC4)

	fmt.Printf("\n")
	fmt.Printf("================\n")
	fmt.Printf("识别  条数(md5+cmdid去重)：%d\n", recordC0)
	fmt.Printf("监测  条数(md5+cmdid去重)：%d\n", recordC1)
	fmt.Printf("关键字条数(md5+cmdid去重)：%d\n", recordC4)
	fmt.Printf("取证  条数(md5+cmdid去重)：%d\n", recordC3)
	return nil
}
