package advertizer

import (
	"sync"
	"time"
	"path/filepath"
)

import (
	"github.com/BurntSushi/toml"
	"github.com/dghubble/oauth1"
	"github.com/dghubble/go-twitter/twitter"
)

type TwitterClient struct {
	cl   *twitter.Client

	last time.Time
	mtx  *sync.Mutex
}

type TwConfig struct {
	ConsumerKey    string
	ConsumerSecret string
	Token          string
	AccessSecret   string
}

func loadtwConfig(path string) (*TwConfig, error) {
	fpath := filepath.Clean(path)

	var conf TwConfig
	if _, err := toml.DecodeFile(fpath, &conf); err != nil {
		return nil, err
	}
	return &conf, nil
}

func NewTwitterClient(c_path string) (*TwitterClient, error) {
	path := filepath.Clean(c_path)
	c, err := loadtwConfig(path)
	if err != nil {
		return nil, err
	}

	config := oauth1.NewConfig(c.ConsumerKey, c.ConsumerSecret)
	token := oauth1.NewToken(c.Token, c.AccessSecret)
	httpClient := config.Client(oauth1.NoContext, token)

	client := twitter.NewClient(httpClient)

	return &TwitterClient{
		cl: client,
		mtx: new(sync.Mutex),
	}, nil
}

func (self *TwitterClient) Tweet(msg string) error {
	self.mtx.Lock()
	defer self.mtx.Unlock()

	_, _, err := self.cl.Statuses.Update(msg, nil)
	return err
}
