# Metric-Index

## 项目介绍
**一个remoteWrite的中间处理层，转存指标名，过滤不规范Metric**

```go
Metric = MetricName + Labels
Labels = []{Label.Name=Label.Value}）
```

- **接收RemoteWrite，处理后转存远程时序库存储**
- **过滤不规范Metric，丢弃对应Metric**
- **将Metric存入Redis和ES中**
    - Redis缓存Metric，提供Exist判断，过期时间可配置
    - ES存储Metric，提供Metric搜索，通过ES生命周期控制过期时间


## 项目结构

- routers 路由层，URL配置，项目接口入口
- controller 控制器层，验证提交的数据，将验证完成的数据传递给 service
- service 业务层，只完成业务逻辑的开发，不进行操作数据库
- dao 数据库层，操作数据的CURD

## 项目组件

- Viper 配置管理，监听配置，自动加载更新
- Zap 日志管理
- go-elasticsearch/v7 elasticsearch驱动
- go-redis/redis redis驱动

## 项目编译

使用 golangci-lint和golint做代码规范检测，推荐使用golint

```bash
bash build_linux.sh
```

## 项目配置

- 环境变量配置

    目前只区分开发环境和线上环境，通过配置环境变量`export GO_ENV=dev`或`export GO_ENV=prod`

- 配置文件

```yaml
app: # 项目配置
  name: "metric-index"  # 项目名称
  port: 7001 # 项目端口

log: # 日志配置
  level: "debug"  # 日志级别
  filename: "logs/metric-index.log" # 日志文件相对路径
  max_size: 200 # 日志文件大小，单位MB
  max_age: 30 # 日志保存时长，单位天
  max_backups: 7 # 日志文件保存个数

redis:
  host: "127.0.0.1"
  port: 6379
  password: ""
  db: 0
  pool_size: 100

metric_filter:
  metricNameRegex: ".*"
  metricValueRegex: ".*"
  labelNameRegex: ".*"
  labelValueRegex: ".*"

metric_store:
  cache:
    isexpire: true
    expire: 604800
    defaultExpire: 604800
    cleanupInterval: 600
    dist_interval: 172800
    worker_num: 50
    flush_lens: 1000
    flush_interval: 10
  store:
    url:
      - "http://elasticsearch-host:9200"
    username:
    password:
    index_name: "metric-test"
    worker_num: 50 # 写入ES工作协程数
    flush_bytes: 5242880 # 批量写入ES的缓存队列大小
    flush_interval: 10 # 超时刷新，时间内如果未写满队列则主动刷入ES

remote: # remote wirte转发配置，将接收到的数据转发到指定接口
  write:
    url: "http://vminsert-host:8480/insert/1/prometheus/api/v1/write"
    content_type: "application/x-protobuf"
  send:
    url: "http://vminsert-host:4242/insert/1/opentsdb/api/put"
    content_type: "application/json"

```



## 项目启动

#### 测试环境
- goland运行

```
项目配置（Edit Configurations） → Configuration → Environment → 添加：GO_ENV=dev
```

- 本地开发环境二进制启动(MAC)
    - 配置环境变量

    ```bash
    sudo echo 'export GO_ENV=dev' >> ~/.zshrc
    ```
    - config目录中创建配置文件：dev.yml
    - 执行脚本build_linux.sh打包编译
    - 启动项目：./metric-index


- 线上环境
    - 配置环境变量

    ```bash
    sudo echo 'export GO_ENV=prod' >> ~/.zshrc
    ```
    - config目录中创建配置文件：prod.yml
    - 执行脚本build_linux.sh打包编译
    - 启动项目：./metric-index
 