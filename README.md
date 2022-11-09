<!--
Hey, thanks for using the awesome-readme-template template.  
If you have any enhancements, then fork this project and create a pull request 
or just open an issue with the label "enhancement".

Don't forget to give this project a star for additional support ;)
Maybe you can mention me or this repo in the acknowledgements too
-->
<div align="center">

  <img src="assets/logo.png" alt="logo" width="200" height="200" />
  <h1>Filebeat Output To Clickhouse</h1>

  <p>
    The output plugin for filebeat support push events to ClickHouse! 
  </p>


<!-- Badges -->
<p>
  <a href="https://github.com/lauvinson/filebeat-ck/graphs/contributors">
    <img src="https://img.shields.io/github/contributors/lauvinson/filebeat-ck" alt="contributors" />
  </a>
  <a href="">
    <img src="https://img.shields.io/github/last-commit/lauvinson/filebeat-ck" alt="last update" />
  </a>
  <a href="https://github.com/lauvinson/filebeat-ck/network/members">
    <img src="https://img.shields.io/github/forks/lauvinson/filebeat-ck" alt="forks" />
  </a>
  <a href="https://github.com/lauvinson/filebeat-ck/stargazers">
    <img src="https://img.shields.io/github/stars/lauvinson/filebeat-ck" alt="stars" />
  </a>
  <a href="https://github.com/lauvinson/filebeat-ck/issues/">
    <img src="https://img.shields.io/github/issues/lauvinson/filebeat-ck" alt="open issues" />
  </a>
  <a href="https://github.com/lauvinson/filebeat-ck/blob/master/LICENSE">
    <img src="https://img.shields.io/github/license/lauvinson/filebeat-ck.svg" alt="license" />
  </a>
</p>
</div>

# :package: Compile
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

# :running: Configure Output
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