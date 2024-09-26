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
