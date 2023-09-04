## 简介

利用阿里云解析的 API 实现动态域名解析的功能（类似花生壳，例如定时地将自己的域名解析更新为家中当前的 IP 地址）。


## 快速开始


1. 下载依赖

运行命令: `go mod tidy`

1.1 如因网络问题下载失败，可设置模块代理。运行命令:
```
go env -w GOPROXY=https://goproxy.cn,direct
# 或者 go env -w GOPROXY=https://goproxy.io,direct
```

1.2 若出现依赖包版本冲突，请删除 `go.mod` 文件，再运行命令:
```
go mod init github.com/iotames/qddns
go mod tidy
```


2. 编译

```
# linux或mac 运行: go build -o qddns .
go build -o qddns.exe .
```


3. 使用

3.1 配置文件

复制 `env.default` 文件为 `.env`, 并更修改配置项，以覆盖 `env.default` 配置文件的默认值

3.2 运行

执行命令: `./start.sh` 或 `./qddns`


## 参考

> 阿里云解析API概览 https://help.aliyun.com/document_detail/2355661.html

> 获取解析记录列表 https://help.aliyun.com/document_detail/2357159.html

> 修改域名解析记录 https://help.aliyun.com/document_detail/2355677.html

> 解析记录 https://help.aliyun.com/document_detail/2355673.html