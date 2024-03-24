package sources

// Session is the struct for storing the session of the providers
// Todo currently each provider only support one session,
// a list of session will support for each provider in the future
type Session struct {
	QuakeToken string
	FofaKey    string
}