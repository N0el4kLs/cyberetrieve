package sources

import (
	"time"

	"github.com/imroc/req/v3"
)

var (
	DefaultClient *Client
)

func init() {
	DefaultClient = NewClient()
	DefaultClient.SetDevMode(false)
}

type Client struct {
	devMode bool
	cl      *req.Client
}

func NewClient() *Client {
	return &Client{
		devMode: false,
		cl:      req.C(),
	}
}
func (c *Client) SetDevMode(devMode bool) {
	c.devMode = devMode
}

func (c *Client) Get(url string, headers map[string]string) (*req.Response, error) {
	cl := c.cl.Clone()
	if c.devMode {
		cl.DevMode()
	}
	cl.SetBaseURL(url).SetTimeout(10 * time.Second)
	resp := cl.SetCommonHeaders(headers).Get().Do()
	return resp, resp.Err
}

func (c *Client) Post(url string, headers map[string]string, body interface{}) (*req.Response, error) {
	cl := c.cl.Clone()
	if c.devMode {
		cl.DevMode()
	}
	cl.SetBaseURL(url)
	cl.SetCommonHeaders(headers)

	// Todo Post body haven't been wrapped yet
	resp := cl.Post().SetBodyJsonMarshal(body).Do()
	return resp, resp.Err
}
