package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"ds_tool/conf"
	"ds_tool/fileproc"
)

var (
	//通用字段
	SamplePath = flag.String("f", "/home/file_repo", "样本文件路径")
	gPath      = flag.String("l", "/home/udpi_log", "话单文件路径，各话单路径参考标准格式查找")

	//md5还原文件比对参数
	FindMd5  = flag.Bool("miss", false, "查找还原缺失样本")
	DemoPath = flag.String("s", "", "还原文件全量样本，作为标准答案，查找本地是否存在未还原文件")

	//生成话单文件参数
	Generate  = flag.Bool("gen", false, "生成话单，用于数据补报")
	UseCfg    = flag.Bool("useCfg", false, "是否需要使用./conf/conf.json文件内容更新话单字段")
	Md5Record = flag.String("md5", "", "需要从备份文件里面查找的md5话单样例")

	//从上报话单文件中提取有误的文件
	Extract = flag.Bool("extract", false, "从数安的话单文件路径下找到指定的md5文件")
	//md5取上面的Md5Record变量

	//对当前路径下的文件进行压缩tar.gz
	Compress = flag.Bool("compress", false, "对指定路径下的logtar文件进行压缩")

	//删除指定路径下的审计日志文件
	RemoveAudit = flag.Bool("rmaudit", false, "删除指定路径下生成的审计日志文件")

	//对补报的话单和人工答案进行核对
	Verify = flag.Bool("verify", false, "对已生成答案进行校验")

	Audit = flag.Bool("audit", false, "对审计日志和话单文件进行校验")

	DataTime = flag.String("date", "", "查询指定日期的文件")

	UsageFlag = flag.Bool("usage", false, "打印使用场景示例")

	Stat    = flag.Bool("stat", false, "统计相关上报信息")
	oPath   = flag.String("o", "report.xlsx", "话单文件核查，生成报告文件名")
	Verbose = flag.Bool("verbose", false, "用于话单核查，是否输出详细信息")

	Ver = flag.Bool("v", false, "查看工具版本信息")
)

var (
	version = "version"
)

const (
	ConfFile = "./conf/conf.json"
)

func GenerateLogtar(samplefile, backPath, md5 string, useCfg bool) {
	dlog, err := fileproc.FoundBackMd5Dir(backPath, md5)
	if err != nil {
		log.Printf("读取日志模板[%s]失败：%v\n", backPath, err)
		return
	}

	if useCfg {
		cfg, err := conf.LoadConf(ConfFile)
		if err != nil {
			log.Printf("加载[%s]失败：%v\n", ConfFile, err)
			return
		}

		fileproc.UpdateLogFields(dlog, cfg)
	}

	fileproc.LogFilePreEnv(fileproc.OutputPath)

	err = fileproc.GenerateFromSampleLog(samplefile, dlog)
	if err != nil {
		log.Printf("%v\n", err)
	}
	/*
		if pcapfile == "" {
			err = fileproc.GenerateFromSampleLog(samplefile, dlog)
		} else {
			err = fileproc.GenerateFromPcap(samplefile, pcapfile, dlog)
		}
		if err != nil {
			log.Printf("%v\n", err)
		}
	*/

	log.Printf("话单文件和样本文件路径：%s", fileproc.OutputPath)
}

func printFindMd5Usage() {
	fmt.Fprintf(os.Stderr, "\n>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>\n")
	fmt.Fprintf(os.Stderr, "使用说明：\n")
	fmt.Fprintf(os.Stderr, "    查找本地文件还原是否有缺失（和全量还原文件进行md5比对）\n")
	fmt.Fprintf(os.Stderr, "使用示例：\n")
	fmt.Fprintf(os.Stderr, "    %s -miss -f /home/file_repo -s ./sample\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "输入：\n")
	fmt.Fprintf(os.Stderr, "    -f 本设备还原样本路径：/home/file_repo\n")
	fmt.Fprintf(os.Stderr, "    -s 全量还原样本路径\n")
	fmt.Fprintf(os.Stderr, "输出:\n")
	fmt.Fprintf(os.Stderr, "    1) 文件比对查找结果\n")
	fmt.Fprintf(os.Stderr, "    2) 拷贝未还原的样本到指定目录下\n")
}

func printGenLogtar() {
	fmt.Fprintf(os.Stderr, "\n>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>\n")
	fmt.Fprintf(os.Stderr, "使用说明：\n")
	fmt.Fprintf(os.Stderr, "    未还原样本生成logtar文件，指定md5作为标准话单生成模板\n")
	fmt.Fprintf(os.Stderr, "使用示例：\n")
	fmt.Fprintf(os.Stderr, "    %s -generate -f ./miss -l /home/udpi_log/ds_data_bak/logtar/ -md5 9a6f185cea481b34de8a268068f0be4f\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "输入：\n")
	fmt.Fprintf(os.Stderr, "    -f 未还原样本文件路径\n")
	fmt.Fprintf(os.Stderr, "    -l 06cx话单的备份路径\n")
	fmt.Fprintf(os.Stderr, "    -md5 指定md5作为标准话单模板\n")
	fmt.Fprintf(os.Stderr, "    -userCfg[可选] ./conf/conf.json用于话单自定义字段的更新\n")
	fmt.Fprintf(os.Stderr, "输出:\n")
	fmt.Fprintf(os.Stderr, "    生成话单和样本文件路径: %s\n", fileproc.OutputPath)
}

func printExtractMd5() {
	fmt.Fprintf(os.Stderr, "\n>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>\n")
	fmt.Fprintf(os.Stderr, "使用说明：\n")
	fmt.Fprintf(os.Stderr, "    从上报话单中提取指定的MD5话单文件\n")
	fmt.Fprintf(os.Stderr, "使用示例：\n")
	fmt.Fprintf(os.Stderr, "    %s -extract -md5 9a6f185cea481b34de8a268068f0be4f\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "输入：\n")
	fmt.Fprintf(os.Stderr, "    -l 指定的话单文件路径\n")
	fmt.Fprintf(os.Stderr, "    -md5 需要从备份话单查找的md5记录\n")
	fmt.Fprintf(os.Stderr, "输出:\n")
	fmt.Fprintf(os.Stderr, "    移动找到的文件到指定目录: %s\n", fileproc.ChangePath)
}

func printCompressMd5() {
	fmt.Fprintf(os.Stderr, "\n>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>\n")
	fmt.Fprintf(os.Stderr, "使用说明：\n")
	fmt.Fprintf(os.Stderr, "    打包指定路径下的txt文件\n")
	fmt.Fprintf(os.Stderr, "使用示例：\n")
	fmt.Fprintf(os.Stderr, "    %s -compress -l ./change\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "输入：\n")
	fmt.Fprintf(os.Stderr, "    -l 指定的话单文件路径\n")
}

func printRmAudit() {
	fmt.Fprintf(os.Stderr, "\n>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>\n")
	fmt.Fprintf(os.Stderr, "使用说明：\n")
	fmt.Fprintf(os.Stderr, "    删除指定路径下生成的审计日志，删除的文件为./change目录下的同名0x04a8文件\n")
	fmt.Fprintf(os.Stderr, "使用示例：\n")
	fmt.Fprintf(os.Stderr, "    %s -rmaudit\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "输入：\n")
	fmt.Fprintf(os.Stderr, "    -l 指定的话单文件路径\n")
}

func printVerifyResult() {
	fmt.Fprintf(os.Stderr, "\n>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>\n")
	fmt.Fprintf(os.Stderr, "使用说明：\n")
	fmt.Fprintf(os.Stderr, "    校验已上报内容和人工结果是否一致（当前只取识别结果）\n")
	fmt.Fprintf(os.Stderr, "使用示例：\n")
	fmt.Fprintf(os.Stderr, "    %s -verify -l /home/udpi_log/ds_data_identify -md5 ./1.dict\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "输入：\n")
	fmt.Fprintf(os.Stderr, "    -l   话单文件路径，示例 /home/udpi_log/ds_data_identify\n")
	fmt.Fprintf(os.Stderr, "    -md5 人工确认的md5结果表\n")
}

func printVerifyAuditResult() {
	fmt.Fprintf(os.Stderr, "\n>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>\n")
	fmt.Fprintf(os.Stderr, "使用说明：\n")
	fmt.Fprintf(os.Stderr, "    校验日志文件和审计日志是否一致\n")
	fmt.Fprintf(os.Stderr, "使用示例：\n")
	fmt.Fprintf(os.Stderr, "    %s -verify -audit -l /home/udpi_log/\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "    %s -verify -audit -l /home/data/ -date 20241025\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "输入：\n")
	fmt.Fprintf(os.Stderr, "    -l   话单文件路径，示例 /home/udpi_log/\n")
	fmt.Fprintf(os.Stderr, "    -verify 核对文件\n")
	fmt.Fprintf(os.Stderr, "    -audit  核对审计日志文件\n")
	fmt.Fprintf(os.Stderr, "    -date   备份话单查询时，需要添加日期，否则查询所有备份的话单文件\n")
}

func printGenUsage1() {
	fmt.Fprintf(os.Stderr, "构建话单和样本示例：(来源于话单字段+样本内容)\n")
	fmt.Fprintf(os.Stderr, "++++++++++++++++++++++++++++++++++++++++++++++++++++++++\n")
	fmt.Fprintf(os.Stderr, "%s -f /home/file_repo -l sample.logtar\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "注意：\n")
	fmt.Fprintf(os.Stderr, "\t样本文件，保证后缀正确即可\n")
	fmt.Fprintf(os.Stderr, "\t话单文件，保证文件名符合规范即可\n")
	fmt.Fprintf(os.Stderr, "输出:\n")
	fmt.Fprintf(os.Stderr, "\t@%s\t生成的样本文件路径，包括logtar话单文件\n", fileproc.OutputPath)
	fmt.Fprintf(os.Stderr, "\t@%s\t原始文件和生成的文件映射关系表\n", fileproc.OutputMap)
	fmt.Fprintf(os.Stderr, "++++++++++++++++++++++++++++++++++++++++++++++++++++++++\n")
}

func printGenUsage2() {
	fmt.Fprintf(os.Stderr, "构建话单和样本示例：(来源于话单字段+样本内容+pcap包内容)\n")
	fmt.Fprintf(os.Stderr, "++++++++++++++++++++++++++++++++++++++++++++++++++++++++\n")
	fmt.Fprintf(os.Stderr, "%s -f /home/file_repo -l sample.logtar -p sample.logtar\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "注意：\n")
	fmt.Fprintf(os.Stderr, "\t隐藏输入，conf/conf.json 需要自定义的字段\n")
	fmt.Fprintf(os.Stderr, "\t样本文件，保证后缀正确即可\n")
	fmt.Fprintf(os.Stderr, "\t话单文件，保证文件名符合规范即可\n")
	fmt.Fprintf(os.Stderr, "\tpcap文件，保证只有一条流，同时http只提取第一个url/domain\n")
	fmt.Fprintf(os.Stderr, "输出:\n")
	fmt.Fprintf(os.Stderr, "\t@%s\t生成的样本文件路径，包括logtar话单文件\n", fileproc.OutputPath)
	fmt.Fprintf(os.Stderr, "\t@%s\t原始文件和生成的文件映射关系表\n", fileproc.OutputMap)
	fmt.Fprintf(os.Stderr, "++++++++++++++++++++++++++++++++++++++++++++++++++++++++\n")
}

// printUsage 自定义的Usage函数
func printUsage() {
	printFindMd5Usage()
	printGenLogtar()
	printExtractMd5()
	printCompressMd5()
	printRmAudit()
	printVerifyResult()
	printVerifyAuditResult()
}

// main
func main() {
	// 设置使用自定义的Usage函数
	//flag.Usage = printUsage
	flag.Parse()

	if *UsageFlag || (flag.NArg() == 0 && flag.NFlag() == 0) {
		printUsage()
		return
	}

	if *Ver {
		fmt.Printf("Version: %s\n\n", version)
		return
	}

	//查找还原的md5文件
	if *FindMd5 {
		if *SamplePath == "" || *DemoPath == "" {
			printFindMd5Usage()
			return
		}

		err := fileproc.CompareSampleFile(*SamplePath, *DemoPath)
		if err != nil {
			fmt.Println(err)
		}
		return
	}

	//通过样本文件和备份话单生成补报话单
	if *Generate {
		if *SamplePath == "" || *Md5Record == "" {
			printGenLogtar()
			return
		}
		GenerateLogtar(*SamplePath, *gPath, *Md5Record, *UseCfg)
		return
	}

	//从已生成话单中提取指定md5话单
	if *Extract {
		if *Md5Record == "" {
			printExtractMd5()
			return
		}
		fileproc.ExtractMd5File(*gPath, *Md5Record)
		return
	}

	//压缩指定路径下的话单模板
	if *Compress {
		fileproc.CompressLogtar(*gPath)
	}

	if *RemoveAudit {
		fileproc.RemoveAudit(*gPath)
	}

	if *Verify {
		if *Audit {
			fileproc.VerifyAuditFile(*gPath, *DataTime)
		} else {
			if *Md5Record == "" {
				printVerifyResult()
				return
			}
			fileproc.VerifyRecogResult(*gPath, *Md5Record)
		}
	}

	if *Stat {
		if *gPath == "" {
			flag.Usage()
			return
		}

		log.Printf("开始核查上报话单...\n")
		if *gPath != "" {
			pathcx := filepath.Join(*gPath, fileproc.DpiLogtarName)
			pathc0 := filepath.Join(*gPath, fileproc.IdentifyName)
			pathc1 := filepath.Join(*gPath, fileproc.MonitorName)
			pathc3 := filepath.Join(*gPath, fileproc.EvidenceName)
			pathc4 := filepath.Join(*gPath, fileproc.KeywordName)
			fileproc.AnalyzeLogFile(pathcx, pathc0, pathc1, pathc3, pathc4, *oPath, *Verbose)
		}
		return
	}
}
