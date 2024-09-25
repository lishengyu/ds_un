package spec

var (
	FileCodeMap = map[string]string{
		"html":   "1",
		"txt":    "2",
		"xml":    "3",
		"json":   "4",
		"csv":    "5",
		"doc":    "7",
		"docx":   "8",
		"xls":    "9",
		"xlsx":   "10",
		"ppt":    "11",
		"pptx":   "12",
		"pdf":    "13",
		"xlsb":   "14",
		"odt":    "15",
		"rtf":    "16",
		"tar":    "18",
		"gz":     "19",
		"tar.gz": "20",
		"zip":    "21",
		"7z":     "22",
		"rar":    "23",
		"bz2":    "24",
		"jar":    "25",
		"war":    "26",
		"arj":    "27",
		"lzh":    "28",
		"xz":     "29",
		"jpeg":   "31",
		"jpg":    "31",
		"png":    "32",
		"tif":    "33",
		"tiff":   "33",
		"webp":   "34",
		"wbmp":   "35",
		"vsdx":   "201",
		"vsd":    "202",
		"fpx":    "401",
		"pbm":    "402",
		"pgm":    "403",
		"bmp":    "404",
	}
)

var C3_DICT = map[int]string{
	1:    "FTP",
	2:    "SSH",
	3:    "TELNET",
	4:    "SMTP",
	5:    "HTTP",
	6:    "POP3",
	7:    "LDAP",
	8:    "HTTPS",
	9:    "RDP",
	10:   "DNS",
	11:   "SNMP",
	12:   "SSDP",
	13:   "VNC",
	14:   "MDNS",
	15:   "PPTP VPN",
	16:   "L2TP VPN",
	17:   "IPSEC VPN",
	18:   "ICMP",
	19:   "IMAP",
	20:   "DHCP",
	21:   "H.323",
	22:   "SOCKS4",
	23:   "SOCKS5",
	24:   "RADIUS",
	25:   "HTTP FLV",
	26:   "NTP",
	27:   "SNTP",
	28:   "TFTP",
	29:   "RTMP",
	30:   "SIP",
	31:   "RTP",
	32:   "RTCP",
	33:   "RTSP",
	34:   "XMPP",
	35:   "HLS",
	36:   "HDS",
	37:   "POP3S",
	38:   "SMTPS",
	39:   "TLS",
	40:   "JT/T808",
	41:   "JT/T809",
	9999: "其他",
}

var C4_DICT = map[int]string{
	1:  "即时通信",
	2:  "阅读",
	3:  "微博",
	4:  "地图导航",
	5:  "视频",
	6:  "音乐",
	7:  "应用商店",
	8:  "网上商城",
	9:  "影像处理",
	10: "直播业务",
	11: "游戏",
	12: "支付",
	13: "动漫",
	14: "邮箱",
	15: "P2P业务",
	16: "VoIP业务",
	17: "彩信",
	18: "浏览下载",
	19: "财经",
	20: "安全杀毒",
	21: "购物",
	22: "出行旅游",
	23: "VPN类应用",
	24: "WAP类应用",
	25: "网盘云服务",
	26: "自营业务",
	27: "公共流量",
	28: "其它",
}

var C9_DICT = map[int]string{
	1:   "HTTP",
	2:   "SMTP",
	3:   "POP3",
	4:   "IMAP",
	5:   "FTP",
	6:   "mysql",
	7:   "tds",
	8:   "tns",
	9:   "PostgreSQL",
	10:  "其他通用协议",
	11:  "PPTP VPN",
	12:  "L2TP VPN",
	13:  "IPSEC VPN",
	14:  "其他vpn协议",
	201: "MongoDB",
	202: "Redis",
	203: "Cassandra",
	204: "ElasticSearch",
	114: "其他",
}

var C10_DICT = map[int]string{
	1:   "html",
	2:   "txt",
	3:   "xml",
	4:   "json",
	5:   "csv",
	6:   "其他文本类",
	7:   "doc",
	8:   "docx",
	9:   "xls",
	10:  "xlsx",
	11:  "ppt",
	12:  "pptx",
	13:  "pdf",
	14:  "xlsb",
	15:  "odt",
	16:  "rtf",
	17:  "其他文件类",
	201: "vsdx",
	202: "vsd",
	18:  "tar",
	19:  "gz",
	20:  "tar.gz",
	21:  "zip",
	22:  "7z",
	23:  "rar",
	24:  "bz2",
	25:  "jar",
	26:  "war",
	27:  "arj",
	28:  "lzh",
	29:  "xz",
	30:  "其他压缩文件类",
	31:  "jpeg/jpg",
	32:  "png",
	33:  "tif/tiff",
	34:  "webp",
	35:  "wbmp",
	36:  "其他图片类",
	401: "fpx",
	402: "pbm",
	403: "pgm",
	404: "bmp",
	114: "其他类",
}

const (
	IndexC0 = iota
	IndexC1
	IndexC2
	IndexC3
	IndexC4
	IndexCx
	IndexMax
)

// Cx字段索引
const (
	Cx_LogId = iota
	Cx_CmdId
	Cx_HouseId
	Cx_FileType
	Cx_FileSize
	Cx_FileMd5
	Cx_Time
	Cx_Sip
	Cx_Dip
	Cx_Sport
	Cx_Dport
	Cx_L4
	Cx_AppId
	Cx_AppType
	Cx_DataPro
	Cx_HttpDomain
	Cx_HttpUrl
	Cx_HttpMethod
	Cx_DataDir
	Cx_ReportType
	Cx_Max
)

// C0字段索引
const (
	C0_LogID = iota
	C0_CommandID
	C0_House_ID
	C0_RuleID
	C0_Rule_Desc
	C0_AssetsIP
	C0_DataFileType
	C0_AssetsSize
	C0_AssetsNum
	C0_DataInfoNum
	C0_DataType
	C0_DataLevel
	C0_DataContent
	C0_IsUploadFile
	C0_FileMD5
	C0_CurTime
	C0_SrcIP
	C0_DestIP
	C0_SrcPort
	C0_DestPort
	C0_ProtocolType
	C0_ApplicationProtocol
	C0_BusinessProtocol
	C0_IsMatchEvent
	C0_Max
)

// C1字段索引
const (
	C1_LogID = iota
	C1_CommandId
	C1_House_ID
	C1_RuleID
	C1_Rule_Desc
	C1_Proto
	C1_Domain
	C1_Url
	C1_Title
	C1_EventTypeID
	C1_EventSubType
	C1_SrcIP
	C1_DestIP
	C1_SrcPort
	C1_DestPort
	C1_FileType
	C1_FileSize
	C1_DataNum
	C1_DataType
	C1_FileMD5
	C1_GatherTime
	C1_Max
)

// C4字段索引
const (
	C4_CommandId = iota
	C4_LogID
	C4_HouseID
	C4_StrategyId
	C4_KeyWord
	C4_Features
	C4_AssetsNum
	C4_SrcIP
	C4_DestIP
	C4_ScrPort
	C4_DestPort
	C4_Domain
	C4_Url
	C4_DataDirection
	C4_Proto
	C4_FileType
	C4_FileSize
	C4_AttachMent
	C4_FileMD5
	C4_GatherTime
	C4_Max
)

var (
	LogName = [IndexMax]string{
		"识别",
		"监测",
		"规则",
		"样本",
		"关键字",
		"DPI话单",
	}
)

// GetFileCode 		通过后缀名获取文件类型代码表
//
//	@param suffix
//	@return string
func GetFileCode(suffix string) string {
	value, exist := FileCodeMap[suffix]
	if exist {
		return value
	}
	return "114"
}
