# go2rpcx


**方便快捷自动生成[RPCX](https://rpcx.io/)代码工具**

go2rpcx 可以很轻松的根据Golang定义的接口生成rpcx文件，很大程度简化rpcx服务的开发工作。

## show code

- 创建一个user.go, 写入如下内容

```go
package example

type User interface {
	// 获取用户
	GetUser(request Request) Response
}

type Request struct {
	Name string `json:"name"` // 用户名
}
type Response struct {
	Result string `json:"result"` // 返回结果
}
```

- 生成rpcx文件

在user.go 同目录下执行  ` go2rpcx -f user.go` 就会自动在当前目录的rpc文件夹生成user.rpc.go 文件
生成的文件参考 /example

## 安装

```shell
go get -u github.com/zsinx/go2rpcx
```

## 使用

安装完执行 go2rpcx  如果能输出以下内容则说明安装成功

```shell
➜  go2rpcx
version: 1.0
Usage: go2rpcx [-f] [-t]

Options:
  -f string
        source file path
  -t string
        rpc file target path (default "rpc")
```

-f 参数用于指定 go接口文件

-t 参数用于指定生成的rpc文件存储目录

## 参考资料：

github.com/akkagao/go2proto

github.com/rpcxio/protoc-gen-gogorpcx
