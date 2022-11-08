# filebeat-ck
The output for filebeat support push events to ClickHouse，You need to recompile filebeat with the ClickHouse Output.

# Compile
## We need clone beats first
```$xslt
git clone git@github.com:elastic/beats.git
```

## then, Install clickhouse Output, under GOPATH directory
```
go get -u github.com/lauvinson/filebeat-ck
```

## modify beats outputs includes, add clickhouse output
```
cd {your beats directory}/github.com/elastic/beats/libbeat/publisher/includes/includes.go
```
```
import (
	...
	_ "github.com/lauvinson/filebeat-ck"
)
```
## build package, in filebeat
```
cd {your beats directory}/github.com/elastic/beats/filebeat
make
```

# Configure Output
## clickHouse output configuration
```yml
#----------------------------- ClickHouse output --------------------------------
output.clickHouse:
  # clickHouse tcp link address
  # https://github.com/ClickHouse/clickhouse-go
  # example tcp://host1:9000?username=user&password=qwerty&database=clicks&read_timeout=10&write_timeout=20&alt_hosts=host2:9000,host3:9000
  url: "tcp://127.0.0.1:9000?debug=true&read_timeout=10&write_timeout=20"
  # table name for receive data
  table: ck_test
  # table columns for data filter, match the keys in log file
  columns: [ "id", "name", "created_date" ]
  # will sleep the retry_interval seconds when unexpected exception, default 60s
  retry_interval: 60
  # whether to skip the unexpected type row, when true will skip unexpected type row, default false will always try again
  skip_unexpected_type_row: false
  # batch size
  bulk_max_size: 1000
  # if this grok is not blank, the result of the match will replace the log file column
  grok: "remote_addr\[(?:%{NOTSPACE:remote_addr}?|%{DATA})\] \| time_local\[(?:%{HTTPDATE:time_local}?|%{DATA})\] \| server_name\[(?:%{DATA:server_name}?|%{DATA})\] \| request\[(?:%{WORD:verb} %{GREEDYDATA:path}\?%{GREEDYDATA:uri_query} HTTP/%{NUMBER:http_version}?|%{DATA})\] \| status\[(?:%{NUMBER:status}?|%{DATA})\] \| body_bytes_sent\[(?:%{NUMBER:body_bytes_sent}?|%{DATA})\] \| http_referer\[(?:%{NOTSPACE:http_referer}?|%{DATA})\] \| http_user_agent\[(?:%{QUOTEDSTRING:http_user_agent}?|%{DATA})\] \| http_x_forwarded_for\[(?:%{NOTSPACE:http_x_forwarded_for}?|%{DATA})\]"
  # the matching field for grok
  grok_field: "message"
```
