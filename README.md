# Eagleeye

该项目致力于构建一个综合性的安全检测系统，通过多种扫描和检测技术，帮助用户发现和识别网络资产，评估安全状况，并识别潜在的安全风险。系统主要包括资产在线检测、资产识别、端口扫描、弱口令爆破和漏洞扫描五大核心功能。

# Warning

❗️❗️❗️***如模版中包含Headless协议，切记需先安装chrome (If templates include the Headless protocol, remember to install Chrome first)*** 

# Preinstalled

## Task CLI (for windows)

```bash
# Go Modules
go install github.com/go-task/task/v3/cmd/task@latest
# Chocolatey
choco install go-task
# Scoop
scoop install task
```

## Swag CLI

```bash
go install github.com/swaggo/swag/cmd/swag@latest
```

Or download a pre-compiled binary from the [release page](https://github.com/swaggo/swag/releases).

# Usage

## Download and get help

```bash
go install codeup.aliyun.com/66d825f8c06a2fdac7bbfe8c/eagleeye

eagleeye --help

eagleeye scanner

Usage:
  eagleeye [flags]
  eagleeye [command]

Available Commands:
  apiserver   start api server
  completion  Generate the autocompletion script for the specified shell
  help        Help about any command
  iface       output net interfaces
  meta        print program metadata
  testserver  start a http test (ansible-awx-detect) server

Flags:
      --cfg string        config file
      --da string         探活输出格式 (default "csv")
      --dc int            探活并发数 (default 150)
      --de string         探活超时时间 (default "1s")
  -d, --discovery         设备探活
      --dn int            探活轮次 (default 1)
      --dr int            探活频率 (default 150)
  -h, --help              help for eagleeye
  -j, --job goflag        任务配置 (default Usage of job:
                            -a string
                                任务输出格式 (default "csv")
                            -b  开启headless模式
                            -c int
                                任务并发数 (default 150)
                            -e string
                                任务超时时间 (default "1s")
                            -m string
                                任务名称
                            -n int
                                任务执行轮次 (default 1)
                            -r int
                                任务执行频率 (default 150)
                            -t string
                                任务模版（目录/文件）
                          )
      --mi string         监控频率 (default "5s")
  -m, --monitor           监控日志
  -l, --out_log           任务执行日志
      --pa string         端口扫描输出格式 (default "csv")
      --pc int            端口扫描并发数 (default 150)
      --pe string         端口扫描超时时间 (default "1s")
      --pn int            端口扫描轮次 (default 1)
  -p, --port_scanning     端口扫描
      --pp string         端口扫描端口 (default "http")
      --pr int            端口扫描频率 (default 150)
  -u, --targets strings   目标地址/文件
      --ue strings        排除目标地址/文件
  -v, --version           version for eagleeye
  -z, --vuln string       漏洞映射文件

Use "eagleeye [command] --help" for more information about a command.
```

## Use the binary

### Linux/MacOS

```bash
git clone http://codeup.aliyun.com/66d825f8c06a2fdac7bbfe8c/eagleeye.git

cd eagleeye

make

./eagleeye --help
```

### Windows

```powershell
git clone http://codeup.aliyun.com/66d825f8c06a2fdac7bbfe8c/eagleeye.git

cd eagleeye

go install github.com/go-task/task/v3/cmd/task@latest

task build

.\eagleeye.exe --help
```

## Run

```bash
# Description
./eagleeye -u [目标：文件/CIDR/IP] -j "-a [任务输出格式] -c [任务并发数] -e [任务超时时间] -m [任务名称] -n [任务执行轮次] -r [任务执行频率] -t [任务模版（目录/文件）]" -p -d

# Demo
./eagleeye -u 192.168.1.0/24 -j "-a csv -m 资产识别 -c 2000 -r 2000 -e 1s -t ./templates/资产识别" -j "-a csv -m 漏洞扫描 -c 2000 -r 2000 -e 1s -t ./templates/漏洞扫描"

#use config file
./eagleeye --cfg plan.demo.yaml
```

## Config

```yaml
targets:
  - 192.168.1.177:9080
exclude_targets:
  - 192.168.1.108
out_log: true
monitor:
  use: true
  interval: 5s
mapping:
  vuln: ./vm.demo.yaml
port_scanning:
  use: true
  timeout: 1s
  count: 1
  ports: top100
  concurrency: 100
  rate_limit: 1000
  format: excel
host_discovery:
  use: true
  timeout: 1s
  count: 1
  concurrency: 100
  rate_limit: 1000
  format: excel
jobs:
  - name: 漏洞扫描
    headless: false
    concurrency: 1000
    rate_limit: 1000
    format: console
    timeout: 1s
    count: 1
    template: ./templates/漏洞扫描
  - name: 资产扫描
    headless: true
    concurrency: 100
    rate_limit: 1000
    format: console
    timeout: 1s
    count: 1
    template: ./templates/资产识别
```

### Description

| 名称              | 类型              | 说明                       | 可选值                                   | 实例值                     |
| --------------- | --------------- | ------------------------ | ------------------------------------- | ----------------------- |
| targets         | array\<string\> | 目标(CIDR/IP/IPRange)      | CIDR<br>IP<br>IPRange<br>File(.txt)   | 192.168.1.0/24          |
| exclude_targets | array\<string\> | 需忽略的目标(CIDR/IP/IPRange)  | CIDR<br>IP<br>IPRange<br>File(.txt)   | 192.168.1.1-192.168.1.8 |
| mapping         | object          | 映射相关                     |                                       |                         |
| >vuln           | string          | 漏洞映射文件(yaml格式)           |                                       | ./vm.demo.yaml          |
| out_log         | boolean         | job输出日志                  |                                       | false                   |
| monitor         | object          | 监控(CPU/内存)               |                                       |                         |
| >use            | boolean         | 是否开启                     |                                       | false                   |
| >interval       | string          | 监控周期                     |                                       | 5s                      |
| port_scanning   | object          | 端口扫描                     |                                       |                         |
| >use            | boolean         | 是否开启                     |                                       | false                   |
| >timeout        | string          | 超时时间                     |                                       | 1s                      |
| >count          | integer         | 轮次(重试次数)                 |                                       | 1                       |
| >ports          | string          | 端口(http,top100,top1000,) | http<br>top100<br>top1000<br>80,81-90 | top100                  |
| >concurrency    | integer         | 并发数                      |                                       | 1000                    |
| >rate_limit     | integer         | 频率                       |                                       | 1000                    |
| >format         | string          | 输出格式                     | csv<br>excel<br>                      | csv                     |
| host_discovery  | object          | 在线检测                     |                                       |                         |
| >use            | boolean         | 是否开启                     |                                       | false                   |
| >timeout        | string          | 超时时间                     |                                       | 1s                      |
| >count          | integer         | 轮次(重试次数)                 |                                       | 1                       |
| >concurrency    | integer         | 并发数                      |                                       | 1000                    |
| >rate_limit     | integer         | 频率                       |                                       | 1000                    |
| >format         | string          | 输出格式                     | csv<br>excel<br>                      | csv                     |
| jobs            | array\<object\> | 任务列表                     |                                       |                         |
| >name           | string          | 任务名称                     |                                       | 漏洞扫描                    |
| >headless       | boolean         | 开启headless模式             |                                       | false                   |
| >concurrency    | integer         | 并发数                      |                                       | 1000                    |
| >rate_limit     | integer         | 频率                       |                                       | 1000                    |
| >format         | string          | 输出格式                     | console<br>csv<br>excel               | console                 |
| >timeout        | string          | 超时时间                     |                                       | 1s                      |
| >count          | integer         | 轮次(重试次数)                 |                                       | 1                       |
| >template       | string          | 任务模板/文件夹                 |                                       | ./templates/pocs        |

### Mapping - Vuln

```yaml
- template_id: SampleTemplateID
  vulnerabilities:
    - id: SampleActualID
      name: SampleName
      severity: critical
      description: SampleDescription
      remediation: SampleRemediation
      expressions:
        - ">7.5, <9.6"

    - id: SampleActualID2
      name: SampleName2
      severity: critical
      description: SampleDescription2
      remediation: SampleRemediation2
      expressions:
        - ">5.0"
        - "<3.2"
```

# Using as library

## Setting

❗️❗️❗️***重要 (IMPORTANT)***

```bash
go env -w GOPRIVATE="47.103.136.241"
go env -w GOINSECURE="47.103.136.241"

go get -u codeup.aliyun.com/66d825f8c06a2fdac7bbfe8c/eagleeye@dev
```

## Example

详见 [example](./example/sdk/example.go)

```go
// 实例化引擎 instantiate engine（❗️❗️❗️单例 singleton）
engine, err := eagleeye.NewEngine()
if err != nil {
  log.Fatalln("error:", err)
}

// 构建配置 build options
options := &types.Options{
    // 省略
    ...
}

// 构建运行条目 build entry
entry, err := engine.NewEntry(options)
if err != nil {
  log.Fatalln("error:", err)
}

c, cancel := context.WithCancel(context.Background())

// 运行 run
err = entry.Run(c)
if err != nil {
  log.Fatalln("error:", err)
}

// 停止运行 stop
entry.Stop()
// 或者 or
engine.Entry(entry.EntryID).Stop()
// 或者 or
cancel()

// Run没有出错，可通过Result获取结果
// Run出错，Result返回nil
entry.Result()

// 关闭引擎 close（❗️❗️❗️切记 remember）
engine.Close()
```

# License generate

## Command
```bash
./eagleeye license -e 2024-10-31 -w 00:16:3e:29:98:a5 -a curescan_test
```

## Description
+ -e 证书超时时间
+ -w 网卡信息（可有多个，逗号分割）
+ -a 受众/接收者（谁使用）
