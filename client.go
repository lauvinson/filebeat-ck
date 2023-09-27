package clickhouse_20200328

import (
	"context"
	"database/sql"
	"errors"
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

func (c *client) Publish(_ context.Context, batch publisher.Batch) error {
	if c == nil {
		panic("no client")
	}
	if batch == nil {
		panic("no batch")
	}

	events := batch.Events()
	c.observer.NewBatch(len(events))
	rest, err := c.publish(events)
	if rest != nil {
		c.observer.Failed(len(rest))
		c.sleepBeforeRetry(err)
		batch.RetryEvents(rest)
		return err
	}

	batch.ACK()
	return err
}

func (c *client) String() string {
	return "clickhouse(" + c.url + ")"
}

// publish events
func (c *client) publish(data []publisher.Event) ([]publisher.Event, error) {
	ctx := context.Background()
	formatRows := make([][]interface{}, 0)
	// group events
	okFormatEvents, failFormatEvents, formatRows := extractDataFromEvent(c.log, formatRows, data, c.columns, c.grokStr, c.grokField, c.g)
	c.log.Infof("[check data format] ok-format-events: %d, fail-format-events: %d, format-rows: %d", len(okFormatEvents), len(failFormatEvents), len(formatRows))

	if len(okFormatEvents) == 0 {
		return failFormatEvents, errors.New("[check data format] all events match field fail")
	}
	tx, err := c.connect.BeginTx(ctx, nil)
	if err != nil {
		c.log.Errorf("[transaction] begin fail: {%+v}", err)
		return data, err
	}
	stmt, err := tx.PrepareContext(ctx, generateSql(c.table, c.columns))
	if err != nil {
		c.log.Errorf("[transaction] stmt prepare fail: {%+v}", err)
		return data, err
	}
	// defer
	defer func(stmt *sql.Stmt) {
		err = stmt.Close()
		if err != nil {
			c.log.Errorf("[transaction] stmt close fail: {%+v}", err)
		}
	}(stmt)
	var lastErr error
	var okExecEvents, failExecEvents []publisher.Event
	for k, row := range formatRows {
		_, err = stmt.ExecContext(ctx, row...)
		if err != nil {
			c.log.Errorf("[transaction] stmt exec fail: {%+v}", err)

			lastErr = err
			// fail
			failExecEvents = append(failExecEvents, okFormatEvents[k])
			continue
		}
		// ok
		okExecEvents = append(okExecEvents, okFormatEvents[k])
	}
	c.log.Infof("[check data type] ok-exec-events: %d, fail-exec-events: %d", len(okExecEvents), len(failExecEvents))
	// happen error, skip unexpected type row
	if lastErr != nil && c.skipUnexpectedTypeRow {
		// rollback
		err = tx.Rollback()
		if err != nil {
			return nil, err
		}
		if len(okExecEvents) > 0 {
			c.log.Infof("[skip unexpected type row] recall publish, ok-exec-events: %d, fail-exec-events: %d", len(okExecEvents), len(failExecEvents))
			_, err = c.publish(okExecEvents)
			if err != nil {
				return nil, err
			}
		}
		return nil, nil
	}
	if err = tx.Commit(); err != nil {
		err = tx.Rollback()
		if err != nil {
			return nil, err
		}
		c.log.Errorf("[transaction] commit failed, ok-exec-events: %d, fail-exec-events: %d, err: {%+v}", len(okExecEvents), len(failExecEvents), err)
		return data, err
	}

	c.log.Infof("[transaction] commit successed, ok-exec-events: %d, fail-exec-events: %d", len(okExecEvents), len(failExecEvents))
	return failExecEvents, lastErr
}
