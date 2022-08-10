package clickhouse_20200328

import (
	"github.com/elastic/beats/v7/libbeat/logp"
	"github.com/elastic/beats/v7/libbeat/outputs"
)

var logger = logp.NewLogger("ClickHouse")

func init() {
	outputs.RegisterType("clickHouse", makeClickHouse)
}
