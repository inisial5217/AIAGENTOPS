package integration

// Client base integration client
type Client interface {
	Ping() error
}
