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
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/xuri/excelize/v2"
)

// ++++++++++++++++++++++++

type FileHitKey struct {
	FileType  int
	AppProto  int
	DataProto int
}

type HitNode struct {
	HitType  string
	HitCount int
}

type FileHitValue struct {
	HitC0 []HitNode
	HitC1 []HitNode
	HitC4 []HitNode
}

type HitC4Id struct {
	IdList string
}

// ++++++++++++++++++++++++

type StLogStat struct {
	AllCnt     int64
	ValidCnt   int64
	NullCnt    int64
	InvalidCnt int64
}

type StFileStat struct {
	FileNum int64
	LogNum  StLogStat
}

type StDictStat struct {
	Name string
	Cnt  int
}

type SampleMapValue struct {
	C0Info []SampleC0Info
	C1Info []SampleC1Info
	C4Info []SampleC4Info
}

var (
	Md5Map    [spec.IndexMax]sync.Map   //MD5表，比对话单和取证文件是否对应
	FileStat  [spec.IndexMax]StFileStat //统计各类话单上报情况
	SampleMap sync.Map
)

func incFileCnt(index int) {
	atomic.AddInt64(&FileStat[index].FileNum, 1)
}

func incLogAllCnt(index int) {
	atomic.AddInt64(&FileStat[index].LogNum.AllCnt, 1)
}

func incLogValidCnt(index int) {
	atomic.AddInt64(&FileStat[index].LogNum.ValidCnt, 1)
}

func incLogNullCnt(index int) {
	atomic.AddInt64(&FileStat[index].LogNum.NullCnt, 1)
}

func incLogInvalidCnt(index int) {
	atomic.AddInt64(&FileStat[index].LogNum.InvalidCnt, 1)
}

func procLogData(ctx string, logType int, filename string) error {
	switch logType {
	case spec.IndexC0:
		procC0Ctx(ctx, filename)
		//statC0Ctx(ctx)
	case spec.IndexC1:
		procC1Ctx(ctx, filename)
	case spec.IndexC2:
		procC2Ctx(ctx, filename)
	case spec.IndexC3:
		procC3Ctx(ctx, filename)
	case spec.IndexC4:
		procC4Ctx(ctx, filename)
	case spec.IndexCx:
		procCxCtx(ctx, filename)
		//statCxCtx(ctx)
	default:
		fmt.Printf("Not support log type: %d\n", logType)
	}

	return nil
}

func procLogtarFile(filename string, logType int) error {
	// 打开文件
	file, err := os.Open(filename)
	if err != nil {
		fmt.Println("Error opening file:", err)
		return err
	}
	defer file.Close()

	// 创建一个Scanner
	scanner := bufio.NewScanner(file)

	var count int
	// 逐行读取文件内容
	for scanner.Scan() {
		count++
		line := scanner.Text()
		incLogAllCnt(logType)
		if line == "" {
			incLogNullCnt(logType)
			continue
		}
		statLogData(line, logType)
	}

	return nil
}

func procTargzFile(filename string, logType int) error {
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
			incLogAllCnt(logType)
			if line == "" {
				incLogNullCnt(logType)
				continue
			}
			statLogData(line, logType)
			//procLogData(line, logType, filename)
		}
	}

	return nil
}

func ProcLogtarPath(path string, wg *sync.WaitGroup) error {
	defer wg.Done()

	if exist := comm.PathExists(path); !exist {
		fmt.Printf("Path %s not exist, skip it!\n", path)
		return nil
	}

	err := filepath.WalkDir(path, func(dir string, d fs.DirEntry, err error) error {
		if err != nil {
			fmt.Printf("filepath walk failed:%v\n", err)
			return err
		}

		if !d.IsDir() {
			if strings.HasSuffix(d.Name(), "logtar") {
				incFileCnt(spec.IndexCx)
				procLogtarFile(dir, spec.IndexCx)
			}
		}

		return nil
	})

	if err != nil {
		fmt.Printf("filepath walk failed:%v\n", err)
	}

	return err
}

func ProcLogPath(path string, wg *sync.WaitGroup, logType int) error {
	defer wg.Done()

	if exist := comm.PathExists(path); !exist {
		fmt.Printf("Path %s not exist, skip it!\n", path)
		return nil
	}
	//fmt.Printf("DS Identify path: [%s]\n", path)
	err := filepath.WalkDir(path, func(dir string, d fs.DirEntry, err error) error {
		if err != nil {
			fmt.Printf("filepath walk failed:%v\n", err)
			return err
		}

		if !d.IsDir() {
			if strings.HasSuffix(d.Name(), "tar.gz") {
				incFileCnt(logType)
				procTargzFile(dir, logType)
			}
		}

		return nil
	})

	if err != nil {
		fmt.Printf("filepath walk failed:%v\n", err)
	}

	return err
}

func ProcEvidencePath(dir string, wg *sync.WaitGroup) error {
	defer wg.Done()

	err := filepath.WalkDir(dir, func(dir string, d fs.DirEntry, err error) error {
		if err != nil {
			fmt.Printf("filepath walk failed:%v\n", err)
			return err
		}

		if !d.IsDir() {
			if strings.HasSuffix(d.Name(), "zip") {
				incFileCnt(spec.IndexC3)
				statLogData(filepath.Base(d.Name()), spec.IndexC3)
			}
		}

		return nil
	})

	if err != nil {
		fmt.Printf("filepath walk failed:%v\n", err)
	}

	return err
}

func AnalyzeLogFile(cx, c0, c1, c3, c4, dst string, verbose bool) {
	cur := time.Now()

	var wg sync.WaitGroup

	//获取Dpi备份logtar话单
	wg.Add(1)
	go ProcLogtarPath(cx, &wg)

	//获取取证文件MD5值
	wg.Add(1)
	go ProcEvidencePath(c3, &wg)

	//处理识别话单
	wg.Add(1)
	go ProcLogPath(c0, &wg, spec.IndexC0)

	//处理监测话单
	wg.Add(1)
	go ProcLogPath(c1, &wg, spec.IndexC1)

	//处理关键字话单
	wg.Add(1)
	go ProcLogPath(c4, &wg, spec.IndexC4)

	wg.Wait()

	//处理完以后，开始写文件
	excel := excelize.NewFile()
	defer func() {
		if err := excel.Close(); err != nil {
			fmt.Printf("close excel err: %v\n", err)
		}
	}()

	comm.PrintReportPrefix("Report Start")
	GenerateResult(excel, verbose)
	comm.PrintReportSuffix("Report End")
	//保存文件名
	os.Rename(dst, dst+"_bak")
	if err := excel.SaveAs(dst); err != nil {
		fmt.Printf("close xlsx file failed: %v\n", err)
		return
	}

	fmt.Printf("Check Complete, elapse %.2f 秒\n", time.Since(cur).Seconds())
}

func init() {
	C0_CheckMap = make(map[string]CheckInfo)
	C1_CheckMap = make(map[string]CheckInfo)
	C4_CheckMap = make(map[string]CheckInfo)
	Cx_CheckMap = make(map[string]CheckInfo)
}
