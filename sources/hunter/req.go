package hunter

import (
	"encoding/base64"
	"fmt"
	"strconv"
	"time"
)

type HunterSearchFiled struct {
	search     string
	page       int
	pageSize   int
	start_time string
	end_time   string
}

func NewHunterSearchFiled(search string, pageIndex, pageSize int) HunterSearchFiled {
	return HunterSearchFiled{
		search:     search,
		page:       pageIndex,
		pageSize:   pageSize,
		start_time: fmt.Sprintf("%s+00%%3A00%%3A00", strconv.Itoa(time.Now().Year()-1)+time.Now().Format("2006-01-02 03:04:05")[4:10]),
		end_time:   fmt.Sprintf("%s+23%%3A59%%3A59", time.Now().Format("2006-01-02 03:04:05")[:10]),
	}
}

func hunterSearchTrans(h HunterSearchFiled) string {
	getParameter := fmt.Sprintf(
		"&search=%s&page=%v&page_size=%v&is_web=3&start_time=%s&end_time=%s",
		base64.URLEncoding.EncodeToString([]byte(h.search)),
		h.page,
		h.pageSize,
		h.start_time,
		h.end_time,
	)
	return getParameter
}
