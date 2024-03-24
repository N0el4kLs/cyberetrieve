package quake

import (
	"strconv"
	"time"
)

// QuakeSearchFiled quake query interface parameters
type QuakeSearchFiled struct {
	Query       string      `json:"query"` // Query sentence
	Start       int         `json:"start"` // Paging Index
	Size        int         `json:"size"`  // Paging Size
	IgnoreCache interface{} `json:"ignore_cache"`
	StartTime   string      `json:"start_time"` // Query start time
	EndTime     string      `json:"end_time"`   // Inquiry off time
}

// NewQuakeSearchFiled construct of QuakeSearchFiled struct
func NewQuakeSearchFiled(query string, pageIndex, pageSize int) *QuakeSearchFiled {
	return &QuakeSearchFiled{
		Query:       query,
		Start:       pageIndex,
		Size:        pageSize,
		IgnoreCache: false,
		StartTime:   strconv.Itoa(time.Now().Year()-1) + time.Now().Format("2006-01-02 03:04:05")[4:10] + " 00:00:00",
		EndTime:     time.Now().Format("2006-01-02 03:04:05")[:10] + " 00:00:00", // Data from the default query for the past year
	}
}
