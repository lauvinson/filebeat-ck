# filebeat-ck

本项目为filebeat输出插件，支持将事件推送到clickhouse数据表，您需要在filebeat中引入本项目并且重新编译filebeat。

## 一、安装编译

### 1、下载beats

```shell
git clone git@github.com:elastic/beats.git
```

### 2、更改beats outputs includes文件, 添加filebeat-ck

includes所在目录：beats/libbeat/publisher/includes/includes.go  
添加如下代码：

```go
import (
...
_ "github.com/lauvinson/filebeat-ck"
)
```

### 3、构建filebeat

filebeat所在目录：beats/filebeat  
执行如下命令：

```shell
make
```

## 二、配置

### filebeat-ck插件配置

```yml
#----------------------------- ClickHouse output --------------------------------
output.clickHouse:
  # clickhouse数据库配置
  # https://github.com/ClickHouse/clickhouse-go
  # 示例 tcp://host1:9000?username=user&password=qwerty&database=clicks&read_timeout=10&write_timeout=20&alt_hosts=host2:9000,host3:9000
  url: "tcp://127.0.0.1:9000?debug=true&read_timeout=10&write_timeout=20"
  # 接收数据的表名
  table: ck_test
  # 数据过滤器的表列，匹配日志文件中对应的键
  columns: [ "id", "name", "created_date" ]
  # 异常重试休眠时间 单位：秒
  retry_interval: 60
  # 是否跳过异常事件推送 true-表示跳过执行异常实践 false-会一直重试，重试间隔为retry_interval
  skip_unexpected_type_row: false
  # 批处理数据量，影响filebeat每次抓取数据行
  bulk_max_size: 1000
  # 如果配置了grok，可以使用grok的结果作为数据过滤器的表列
  grok: "remote_addr\[(?:%{NOTSPACE:remote_addr}?|%{DATA})\] \| time_local\[(?:%{HTTPDATE:time_local}?|%{DATA})\] \| server_name\[(?:%{DATA:server_name}?|%{DATA})\] \| request\[(?:%{WORD:verb} %{GREEDYDATA:path}\?%{GREEDYDATA:uri_query} HTTP/%{NUMBER:http_version}?|%{DATA})\] \| status\[(?:%{NUMBER:status}?|%{DATA})\] \| body_bytes_sent\[(?:%{NUMBER:body_bytes_sent}?|%{DATA})\] \| http_referer\[(?:%{NOTSPACE:http_referer}?|%{DATA})\] \| http_user_agent\[(?:%{QUOTEDSTRING:http_user_agent}?|%{DATA})\] \| http_x_forwarded_for\[(?:%{NOTSPACE:http_x_forwarded_for}?|%{DATA})\]"
  grok_field: "message"
```
