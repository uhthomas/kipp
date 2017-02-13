package worker

import (
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/6f7262/conf/model"
)

func init() {
	c := &cleanup{}
	c.Start()
}

type cleanup struct {
	sync.Once
	l *log.Logger
	t <-chan time.Time
}

func (c *cleanup) Start() {
	c.Do(func() {
		c.l = log.New(os.Stdout, "cleanup", log.LstdFlags)
		c.t = time.Tick(5 * time.Minute)
		go c.loop()
	})
}

func (c *cleanup) Error(err error) {
	if err == nil {
		return
	}
	c.l.Printf("error: %s", err)
}

func (c *cleanup) loop() {
	for {
		go c.Clean()
		<-c.t
	}
}

func (c *cleanup) Clean() {
	if err := model.DB.Delete(model.Content{}, "expires < ?", time.Now()).Error; err != nil {
		c.Error(err)
		return
	}
	dir, err := ioutil.ReadDir(filepath.Join("_", "files"))
	if err != nil {
		c.Error(err)
		return
	}
	for _, f := range dir {
		if !model.DB.Where("hash = ?", f.Name()).Find(&model.Content{}).RecordNotFound() {
			continue
		}
		if err := os.Remove(filepath.Join("_", "files", f.Name())); err != nil {
			c.Error(err)
			continue
		}
	}
}
