package sources

// Result is a struct for storing results from providers
type Result struct {
	IP         string
	URL        string
	Host       string
	Domain     string
	Port       int
	ICPUnit    string // ICP unit,like 北京百度网讯科技有限公
	ICPLicence string // ICP licence, like 京ICP证030173号
}
