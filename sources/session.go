package sources

// Session is the struct for storing the session of the providers
// Todo currently each provider only support one session,
// a list of session will support for each provider in the future
type Session struct {
	QuakeToken string
	FofaKey    string
	HunterKey  string
}

const (
	DEFAULT_PAGE_SIZE     = 30
	DEFAULT_PAGE_SIZE_MAX = 100
)
