package fileproc

import (
	"archive/tar"
	"bufio"
	"compress/gzip"
	"ds_tool/comm"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

var (
	UnrecoverPath = "./miss"
	ChangePath    = "./change"
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

func readTargzFile(filename, md5 string) error {
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
			if strings.Contains(line, md5) {
				hitMd5 = true
			}
		}
	}

	if hitMd5 {
		fmt.Printf("拷贝文件: %s\n", filename)
		dn := filepath.Join(ChangePath, strings.ToLower(md5), filepath.Base(filename))
		done := comm.CopyFile(filename, dn)
		if done {
			os.Remove(filename)
		} else {
			fmt.Printf("拷贝文件失败：%s\n", filename)
		}
	}

	return nil
}

func extractLogFile(path, md5 string) {
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
				readTargzFile(dir, md5)
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
	extractLogFile(filepath.Join(path, IdentifyName), md5)
	//监测
	extractLogFile(filepath.Join(path, MonitorName), md5)
	//关键字生成话单和备份话单文件名不一致，需要再次处理一下
	kfile := filepath.Join(path, KeywordName)
	if exist := comm.PathExists(kfile); !exist {
		kfile = filepath.Join(path, KeywordNameB)
	}
	extractLogFile(kfile, md5)

	md5path := filepath.Join(ChangePath, strings.ToLower(md5))
	//解压对应的压缩文件
	uncompressDir(md5path)

	fmt.Printf(">>>>>拷贝路径：%s\n", md5path)
}

func targzFile(filename string) {
	dstfile := strings.ReplaceAll(filename, "txt", "tar.gz")
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
