package fofa

import "encoding/base64"

type FofaSearchFiled struct {
	Query  string
	Size   int
	Page   int
	Fields string
	Full   string // 默认搜索一年内的数据，指定为true即可搜索全部数据
}

func NewFofaSearchFiled(query string, pageIndex, pageSize int) *FofaSearchFiled {
	return &FofaSearchFiled{
		Query:  base64.StdEncoding.EncodeToString([]byte(query)),
		Size:   pageSize,
		Page:   pageIndex,
		Fields: "ip,host,port,domain,protocol,icp",
		Full:   "false",
	}
}
