package fileproc

import (
	"ds_tool/spec"
	"fmt"
	"strconv"
	"strings"
	"sync"

	"github.com/xuri/excelize/v2"
)

func printfItemResult(item, res string, succ int) {
	if succ == 0 {
		fmt.Printf("\t[PASS] %s:\t%s\n", item, res)
	} else {
		fmt.Printf("\t[FAIL] %s:\t%s\n", item, res)
	}
}

func CheckMd5(ex *excelize.File, index int) {
	fmt.Printf("Check Item %03d [校验话单和取证文件]\n", index)

	_, err := ex.NewSheet("MD5")
	if err != nil {
		fmt.Printf("new sheet failed:%v\n", err)
		return
	}

	streamWriter, err := ex.NewStreamWriter("MD5")
	if err != nil {
		fmt.Printf("new stream writer failed: %v\n", err)
		return
	}
	defer func() {
		if err = streamWriter.Flush(); err != nil {
			fmt.Printf("结束流式写入失败: %v\n", err)
		}
		if err = ex.SetColWidth("MD5", "A", "B", 32); err != nil {
			fmt.Printf("set col width failed: %v\n", err)
		}
	}()

	if err := streamWriter.SetRow("A1", []interface{}{"MD5", "原因"}); err != nil {
		fmt.Printf("stream writer write failed: %v\n", err)
		return
	}

	record := 0
	invalid := 0
	row := 0
	Md5Map[spec.IndexC0].Range(func(key, value interface{}) bool {
		record++
		_, ok := Md5Map[spec.IndexC3].Load(key)
		if !ok {
			tmp := []interface{}{
				key,
				"识别话单C0 缺失取证文件",
			}
			invalid++
			row++
			_ = streamWriter.SetRow("A"+strconv.Itoa(row+1), tmp)
		}
		return true
	})

	var res string
	if invalid == 0 {
		res = fmt.Sprintf("total %d records", record)
	} else {
		res = fmt.Sprintf("total %d recrods; invalid %d records", record, invalid)
	}
	printfItemResult("识别话单C0", res, invalid)

	record = 0
	invalid = 0
	Md5Map[spec.IndexC1].Range(func(key, value interface{}) bool {
		record++
		_, ok := Md5Map[spec.IndexC3].Load(key)
		if !ok {
			tmp := []interface{}{
				key,
				"监测话单C1 缺失取证文件",
			}
			invalid++
			row++
			_ = streamWriter.SetRow("A"+strconv.Itoa(row+1), tmp)
		}
		return true
	})

	if invalid == 0 {
		res = fmt.Sprintf("total %d records", record)
	} else {
		res = fmt.Sprintf("total %d recrods; invalid %d records", record, invalid)
	}
	printfItemResult("识别话单C1", res, invalid)

	record = 0
	invalid = 0
	Md5Map[spec.IndexC4].Range(func(key, value interface{}) bool {
		record++
		_, ok := Md5Map[spec.IndexC3].Load(key)
		if !ok {
			tmp := []interface{}{
				key,
				"关键词话单C4 缺失取证文件",
			}
			invalid++
			row++
			_ = streamWriter.SetRow("A"+strconv.Itoa(row+1), tmp)
		}
		return true
	})

	if invalid == 0 {
		res = fmt.Sprintf("total %d records", record)
	} else {
		res = fmt.Sprintf("total %d recrods; invalid %d records", record, invalid)
	}
	printfItemResult("关键词话单C4", res, invalid)

	//C3
	record = 0
	invalid = 0
	Md5Map[spec.IndexC3].Range(func(key, value interface{}) bool {
		record++
		_, ok1 := Md5Map[spec.IndexC0].Load(key)
		_, ok2 := Md5Map[spec.IndexC1].Load(key)
		_, ok3 := Md5Map[spec.IndexC4].Load(key)

		if (ok1 && ok2) || ok3 {
			return true
		}

		var str string
		if !ok1 && !ok2 {
			str = "取证文件C3 缺失C0/C1话单文件"
		} else if !ok1 {
			str = "取证文件C3 缺失C0话单文件"
		} else if !ok2 {
			str = "取证文件C3 缺失C1话单文件"
		}
		tmp := []interface{}{
			key,
			str,
		}
		invalid++
		row++
		_ = streamWriter.SetRow("A"+strconv.Itoa(row+1), tmp)
		return true
	})

	if invalid == 0 {
		res = fmt.Sprintf("total %d records", record)
	} else {
		res = fmt.Sprintf("total %d recrods; invalid %d records", record, invalid)
	}
	printfItemResult("取证文件C3", res, invalid)

	return
}

func writeStatLogCnt(ex *excelize.File, index *int, sheetName string) {
	*index += 1
	fmt.Printf("Check Item %03d [日志条目数统计]\n", index)

	err := ex.SetSheetName("Sheet1", sheetName)
	if err != nil {
		fmt.Printf("new sheet failed:%v\n", err)
		return
	}

	streamWriter, err := ex.NewStreamWriter(sheetName)
	if err != nil {
		fmt.Printf("new stream writer failed: %v\n", err)
		return
	}
	defer func() {
		if err = streamWriter.Flush(); err != nil {
			fmt.Printf("结束流式写入失败: %v\n", err)
		}
	}()

	if err := setSheetLine(streamWriter, "A1", SheetFileLogTitle); err != nil {
		fmt.Printf("stream writer write failed: %v\n", err)
		return
	}

	record := 0
	for i, cnt := range FileStat {
		record++
		fields := NewSheetFileLog(spec.LogName[i], cnt)
		setSheetLine(streamWriter, "A"+strconv.Itoa(record+1), fields)
	}

	printfItemResult(sheetName, fmt.Sprintf("total %d records", record), 0)
}

func CheckLogId(ex *excelize.File, index int) {
	fmt.Printf("Check Item %03d [LogId唯一性校验]\n", index)

	_, err := ex.NewSheet("Logid")
	if err != nil {
		fmt.Printf("new sheet failed:%v\n", err)
		return
	}

	streamWriter, err := ex.NewStreamWriter("Logid")
	if err != nil {
		fmt.Printf("new stream writer failed: %v\n", err)
		return
	}
	defer func() {
		if err = streamWriter.Flush(); err != nil {
			fmt.Printf("结束流式写入失败: %v\n", err)
		}
	}()

	if err := streamWriter.SetRow("A1", []interface{}{"Logid", "出现次数"}); err != nil {
		fmt.Printf("stream writer write failed: %v\n", err)
		return
	}

	record := 0
	invalid := 0
	LogidMap.Range(func(key, value interface{}) bool {
		record++
		if value != 1 {
			tmp := []interface{}{
				key,
				value,
			}
			invalid++
			_ = streamWriter.SetRow("A"+strconv.Itoa(invalid+1), tmp)
		}
		return true
	})

	var res string
	if invalid == 0 {
		res = fmt.Sprintf("total %d records", record)
	} else {
		res = fmt.Sprintf("total %d records; invalid %d records", record, invalid)
	}
	printfItemResult("Logid唯一性校验", res, invalid)

	return
}

func checkDict(ex *excelize.File, m sync.Map, name string) (int, int) {
	_, err := ex.NewSheet(name)
	if err != nil {
		fmt.Printf("new sheet failed:%v\n", err)
		return 0, 0
	}

	streamWriter, err := ex.NewStreamWriter(name)
	if err != nil {
		fmt.Printf("new stream writer failed: %v\n", err)
		return 0, 0
	}
	defer func() {
		if err = streamWriter.Flush(); err != nil {
			fmt.Printf("结束流式写入失败: %v\n", err)
		}
	}()

	if err := streamWriter.SetRow("A1", []interface{}{"类型", "计数"}); err != nil {
		fmt.Printf("stream writer write failed: %v\n", err)
		return 0, 0
	}

	record := 0
	invalid := 0
	m.Range(func(key, value interface{}) bool {
		record++
		tmp := []interface{}{
			key,
			value,
		}
		if strings.Contains(key.(string), "illegal:") {
			invalid++
		}
		_ = streamWriter.SetRow("A"+strconv.Itoa(record+1), tmp)
		return true
	})

	return record, invalid
}

func CheckDict(ex *excelize.File, index int) {
	fmt.Printf("Check Item %03d [附录表校验]\n", index)

	total, invalid := checkDict(ex, AppProtoStat, "C3表")
	res := fmt.Sprintf(" total %d records; invalid %d records", total, invalid)
	printfItemResult("应用层协议类型代码表C3", res, invalid)

	total, invalid = checkDict(ex, BusProtoStat, "C4表")
	res = fmt.Sprintf(" total %d records; invalid %d records", total, invalid)
	printfItemResult("业务层协议类型代码表C4", res, invalid)

	total, invalid = checkDict(ex, DataProtoStat, "C9表")
	res = fmt.Sprintf(" total %d records; invalid %d records", total, invalid)
	printfItemResult("数据识别协议列表C9", res, invalid)

	total, invalid = checkDict(ex, FileTypeStat, "C10表")
	res = fmt.Sprintf(" total %d records; invalid %d records", total, invalid)
	printfItemResult("数据识别文件格式类别C10", res, invalid)

	return
}

func CheckC0LogMap(ex *excelize.File, index int) {
	fmt.Printf("Check Item %03d [C0话单必填项校验]\n", index)

	_, err := ex.NewSheet("识别话单")
	if err != nil {
		fmt.Printf("new sheet failed:%v\n", err)
		return
	}

	streamWriter, err := ex.NewStreamWriter("识别话单")
	if err != nil {
		fmt.Printf("new stream writer failed: %v\n", err)
		return
	}
	defer func() {
		if err = streamWriter.Flush(); err != nil {
			fmt.Printf("结束流式写入失败: %v\n", err)
		}
	}()

	if err := streamWriter.SetRow("A1", []interface{}{"原始日志", "错误原因", "文件名"}); err != nil {
		fmt.Printf("stream writer write failed: %v\n", err)
		return
	}

	total := 0
	for key, value := range C0_CheckMap {
		tmp := []interface{}{
			key,
			value.Reason,
			value.Filenmae,
		}
		total++
		_ = streamWriter.SetRow("A"+strconv.Itoa(total+1), tmp)
	}

	var res string
	if total == 0 {
		res = fmt.Sprintf("Pass")
	} else {
		res = fmt.Sprintf("invalid %d records", total)
	}
	printfItemResult("C0话单合法性校验", res, total)

	return
}

func CheckC1LogMap(ex *excelize.File, index int) {
	fmt.Printf("Check Item %03d [C1话单必填项校验]\n", index)

	_, err := ex.NewSheet("监测话单")
	if err != nil {
		fmt.Printf("new sheet failed:%v\n", err)
		return
	}

	streamWriter, err := ex.NewStreamWriter("监测话单")
	if err != nil {
		fmt.Printf("new stream writer failed: %v\n", err)
		return
	}
	defer func() {
		if err = streamWriter.Flush(); err != nil {
			fmt.Printf("结束流式写入失败: %v\n", err)
		}
	}()

	if err := streamWriter.SetRow("A1", []interface{}{"原始日志", "错误原因", "文件名"}); err != nil {
		fmt.Printf("stream writer write failed: %v\n", err)
		return
	}

	total := 0
	for key, value := range C1_CheckMap {
		tmp := []interface{}{
			key,
			value.Reason,
			value.Filenmae,
		}
		total++
		_ = streamWriter.SetRow("A"+strconv.Itoa(total+1), tmp)
	}

	var res string
	if total == 0 {
		res = fmt.Sprintf("Pass")
	} else {
		res = fmt.Sprintf("invalid %d records", total)
	}
	printfItemResult("C1话单合法性校验", res, total)

	return
}

func genExlTitle() []interface{} {
	title := []interface{}{
		"MD5",
		"文件类型",
		"文件大小(KB)",
		"L7Proto",
		"Application",
		"Business",
		"是否跨境",
		"匹配次数",
		"识别结果",
		"监测风险",
	}

	return title
}

func genExlLine(key, value any) []interface{} {
	md5 := key.(string)
	info := value.(*SampleMapValue)

	var data string
	var matchnum string
	var app string
	var business string
	var cross string

	for i, v := range info.C0Info {
		if i == 0 {
			data = v.Data
		} else {
			data = data + "|" + v.Data
		}
		matchnum = v.MatchNum
		app = spec.C3_DICT[v.Application]
		business = spec.C4_DICT[v.Business]
		if v.CrossBoard == "0" {
			cross = "是"
		} else {
			cross = "否"
		}
	}

	var risk string
	var l7Proto string
	var fileType string
	var fileSize string
	for i, v := range info.C1Info {
		if i == 0 {
			risk = v.Event
		} else {
			risk = risk + "|" + v.Event
		}
		l7Proto = spec.C9_DICT[v.L7Proto]
		fileType = spec.C10_DICT[v.FileType]
		fileSize = v.FileSize
	}

	tmp := []interface{}{
		md5,
		fileType,
		fileSize,
		l7Proto,
		app,
		business,
		cross,
		matchnum,
		data,
		risk,
	}

	return tmp
}

func RecordSample(ex *excelize.File, index int) {
	fmt.Printf("Check Item %03d [记录所有样本扫描信息]\n", index)

	_, err := ex.NewSheet("Sample")
	if err != nil {
		fmt.Printf("new sheet failed:%v\n", err)
		return
	}

	streamWriter, err := ex.NewStreamWriter("Sample")
	if err != nil {
		fmt.Printf("new stream writer failed: %v\n", err)
		return
	}
	defer func() {
		if err = streamWriter.Flush(); err != nil {
			fmt.Printf("结束流式写入失败: %v\n", err)
		}
	}()

	title := genExlTitle()

	if err := streamWriter.SetRow("A1", title); err != nil {
		fmt.Printf("stream writer write failed: %v\n", err)
		return
	}

	record := 0
	SampleMap.Range(func(key, value interface{}) bool {
		record++
		tmp := genExlLine(key, value)
		_ = streamWriter.SetRow("A"+strconv.Itoa(record+1), tmp)
		return true
	})

	res := fmt.Sprintf("total %d records", record)
	printfItemResult("样本扫描信息", res, 0)

	return
}

func setSheetLine(stream *excelize.StreamWriter, row string, title []string) error {
	var interfaces []interface{}

	for _, str := range title {
		interfaces = append(interfaces, str)
	}

	err := stream.SetRow(row, interfaces)
	return err
}

func writeStatMd5Info(ex *excelize.File, index *int, sheetName string) {
	*index += 1
	fmt.Printf("Stat Item %03d [MD5统计]\n", index)

	_, err := ex.NewSheet(sheetName)
	if err != nil {
		fmt.Printf("new sheet failed:%v\n", err)
		return
	}

	streamWriter, err := ex.NewStreamWriter(sheetName)
	if err != nil {
		fmt.Printf("new stream writer failed: %v\n", err)
		return
	}
	defer func() {
		if err = streamWriter.Flush(); err != nil {
			fmt.Printf("结束流式写入失败: %v\n", err)
		}
		if err = ex.SetColWidth("MD5", "A", "B", 32); err != nil {
			fmt.Printf("set col width failed: %v\n", err)
		}
	}()

	if err := setSheetLine(streamWriter, "A1", SheetMd5HitTitle); err != nil {
		fmt.Printf("stream writer write failed: %v\n", err)
		return
	}

	record := 0
	f := func(key, value any) bool {
		record++
		fields := NewSheetMd5Line(key, value)
		setSheetLine(streamWriter, "A"+strconv.Itoa(record+1), fields)
		return true
	}

	Md5HitMapRange(f)
	printfItemResult(sheetName, fmt.Sprintf("total %d records", record), 0)
}

func GenerateResult(excel *excelize.File, verbose bool) {
	SheetId := 0
	SheetName := "stat_FileLog"
	writeStatLogCnt(excel, &SheetId, SheetName)
	SheetName = "stat_MD5"
	writeStatMd5Info(excel, &SheetId, SheetName)

	/*
		CheckMd5(excel, checkNum)
		checkNum++
		CheckLogCnt(excel, checkNum)
		checkNum++
		CheckLogId(excel, checkNum)
		checkNum++
		CheckDict(excel, checkNum)
		checkNum++
		CheckC0LogMap(excel, checkNum)
		checkNum++
		CheckC1LogMap(excel, checkNum)
		checkNum++
		RecordSample(excel, checkNum)
	*/
}
