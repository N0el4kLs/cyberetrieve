package quake

// QuakeSearchResult Quake service data interface return data structure
type QuakeSearchResult struct {
	Code    int    `json:"code,omitempty"`
	Message string `json:"message,omitempty"`
	Data    []struct {
		Service struct {
			Http struct {
				Host            string                 `json:"host"`
				Title           string                 `json:"title"`
				MetaKeywords    string                 `json:"meta_keywords"`
				XPoweredBy      string                 `json:"x_powered_by"`
				HttpLoadUrl     []string               `json:"http_load_url"`
				Robots          string                 `json:"robots"`
				SitemapHash     string                 `json:"sitemap_hash"`
				Server          string                 `json:"server"`
				Body            string                 `json:"body"`
				RobotsHash      string                 `json:"robots_hash"`
				Sitemap         string                 `json:"sitemap"`
				Path            string                 `json:"path"`
				SecurityText    string                 `json:"security_text"`
				StatusCode      int                    `json:"status_code"`
				ResponseHeaders string                 `json:"response_headers"`
				Icp             map[string]interface{} `json:"icp"` // Todo more effective way to get icp info
			} `json:"http,omitempty"`
		} `json:"service,omitempty"`
		Port     int    `json:"port"`
		Asn      int    `json:"asn"`
		IP       string `json:"ip"`
		Hostname string `json:"hostname"`
		Domain   string `json:"domain"`
	} `json:"data,omitempty"`
	Meta struct {
		Pagination struct {
			Count     int `json:"count"`
			PageIndex int `json:"page_index"`
			PageSize  int `json:"page_size"`
			Total     int `json:"total"`
		} `json:"pagination"`
	} `json:"meta,omitempty"`
}

// NewQuakeSearchResult construct of QuakeSearchResult struct
func NewQuakeSearchResult() *QuakeSearchResult {
	return &QuakeSearchResult{}
}
