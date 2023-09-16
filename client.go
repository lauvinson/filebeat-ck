package clickhouse_20200328

import (
	"database/sql"
	_ "github.com/ClickHouse/clickhouse-go"
	"github.com/elastic/beats/v7/libbeat/logp"
	"github.com/elastic/beats/v7/libbeat/outputs"
	"github.com/vjeantet/grok"
	"sync"
)

type client struct {
	log                   *logp.Logger
	observer              outputs.Observer
	url                   string
	table                 string
	columns               []string
	grokStr               string
	grokField             string
	retryInterval         int
	connect               *sql.DB
	mutex                 sync.Mutex
	skipUnexpectedTypeRow bool
	g                     *grok.Grok
}

func newClient(
	observer outputs.Observer,
	url string,
	table string,
	columns []string,
	grokStr string,
	grokField string,
	retryInterval int,
	skipUnexpectedTypeRow bool,
) *client {
	client := &client{
		log:                   logp.NewLogger("clickhouse"),
		observer:              observer,
		url:                   url,
		table:                 table,
		columns:               columns,
		grokStr:               grokStr,
		grokField:             grokField,
		retryInterval:         retryInterval,
		skipUnexpectedTypeRow: skipUnexpectedTypeRow,
	}
	if grokStr != "" && grokField != "" {
		g, err := grok.NewWithConfig(&grok.Config{NamedCapturesOnly: true})
		if err != nil {
			client.log.Errorf("grok compile error: {%+v}", err)
			return nil
		}
		client.g = g
	}
	return client
}

func (c *client) Connect() error {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	if c.connect != nil {
		c.log.Infof("connection reuse")
		return nil
	}
	connect, err := sql.Open("clickhouse", c.url)
	if err != nil {
		c.connect = nil
		c.log.Errorf("open clickhouse connection fail: {%+v}", err)
		return err
	}
	if err = connect.Ping(); err != nil {
		c.connect = nil
		c.log.Errorf("ping clickhouse fail: {%+v}", err)
		return err
	}
	c.log.Infof("new connection")

	c.connect = connect
	return err
}

func (c *client) Close() error {
	c.log.Infof("close connection")
	return c.connect.Close()
}
