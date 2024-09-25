package main

import (
	"archive/tar"
	"bufio"
	"compress/gzip"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

var (
	Lpath = flag.String("f", "", "查找路径")
	Lmap  map[int]int
)

func procCompress(filename string) error {
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
			line := strings.TrimSpace(scanner.Text())
			if line == "" {
				continue
			}
			fs := strings.Split(line, "|")
			if len(fs) != 13 {
				continue
			} else {
				cap := len(fs[12])
				v, ok := Lmap[cap]
				if ok {
					v++
					Lmap[cap] = v
				} else {
					Lmap[cap] = 1
				}
			}
		}
	}

	return nil
}

func printLmap() {
	var list []int
	for k, _ := range Lmap {
		list = append(list, k)
	}

	sort.Ints(list)

	for _, k := range list {
		fmt.Printf("长度：%05d, 次数：%d\n", k, Lmap[k])
	}
}

func main() {
	flag.Parse()

	if *Lpath == "" {
		flag.Usage()
		return
	}

	Lmap = make(map[int]int, 0)

	err := filepath.WalkDir(*Lpath, func(dir string, d fs.DirEntry, err error) error {
		if err != nil {
			fmt.Printf("filepath walk failed:%v\n", err)
			return err
		}

		if !d.IsDir() {
			if strings.HasSuffix(d.Name(), "tar.gz") {
				procCompress(dir)
			}
		}

		return nil
	})

	if err != nil {
		fmt.Println(err)
	}
	printLmap()

}
