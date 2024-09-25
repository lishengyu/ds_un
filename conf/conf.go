package conf

import (
	"ds_tool/comm"
	"encoding/json"
)

type DefaultConf struct {
	C3Code     int    `json:"c3"`
	C4Code     int    `json:"c4"`
	C9Code     int    `json:"c9"`
	DataDir    int    `json:"data_dir"`
	CmdId      string `json:"cmdid"`
	HouseId    string `json:"houseid"`
	AccessTime int    `json:"time"`
	Sip        string `json:"sip"`
	Dip        string `json:"dip"`
	Sport      int    `json:"sport"`
	Dport      int    `json:"dport"`
	Protocol   int    `json:"proto"`
	Url        string `json:"url"`
	Domain     string `json:"domain"`
	HttpMethod int    `json:"method"`
}

// LoadConf		加载配置文件内容
//
//	@param fn
//	@return DefaultConf
//	@return error
func LoadConf(fn string) (DefaultConf, error) {
	var cfg DefaultConf
	if !comm.IsFile(fn) {
		return cfg, nil
	}

	data, err := comm.ReadFile(fn)
	if err != nil {
		return cfg, err
	}

	err = json.Unmarshal(data, &cfg)
	if err != nil {
		return cfg, err
	}

	return cfg, nil
}
