package sources

// Query is the struct for storing the query
// You can set corresponding query for different providers
type Query struct {
	Query         string `json:"query"`           // input query
	QuakeQuery    string `json:"quake_query"`     // input query to quake query grammar
	FofaQuery     string `json:"fofa_query"`      // input query to fofa query grammar
	HunterQuery   string `json:"hunter_query"`    // input query to hunter query grammar
	NumberOfQuery int    `json:"number_of_query"` // number of query
}

// Provider is the interface for all providers
type Provider interface {
	// Name returns the name of the provider
	Name() string

	// Auth checks if the provider is valid to use
	Auth(*Session) bool

	// Search the result with provider
	Search(*Query) (chan *Result, error)
}
