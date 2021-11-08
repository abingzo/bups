# bups

![Go Report Card](https://goreportcard.com/badge/github.com/abingzo/bups) ![GitHub](https://img.shields.io/github/license/abingzo/bups)

一个Go语言写的用于备份Wordpress/Typecho网站数据至云端的小工具，支持自定义插件，目前只支持`linux`&`darwin`

---

### 构建

```shell
make build-linux
```

### 基本配置

---

- [Config](./CONFIG.md)

### 启动

---

```shell
./bups
```

使用自带的守护进程插件

```shell
./bups --plugin daemon --args '<-s start>'
```

`Supervisor`启动，将`/Users/xx`替换为程序所在的绝对路径，[原模板文件](./bupsd.ini)

```ini
[program:bupsd]
# run folder
directory=/Users/xx
# run command
command=/Users/xx/bups

autostart=true
autorestart=false
startsecs=3

user=root
stdout_logfile=/Users/xx/bups/logs/stdout.log
redirect_stderr=true
// stdout log size
stdout_logfile_maxbytes=30MB
```

`Systemctl`启动，修改[service模板文件](./bupsd.service)，将模板中的目录改为自己的安装目录即可，**注意：**由于`service`文件的编写时间比较久远可能有问题，这一部分抽时间再改

### 编写自己的插件

---

> 首先，插件是使用原生的`plugin`构建的，插件内部要实现的原型函数

```go
type New func() Plugin
```

> 该原型函数返回的是`common.plugin.Plugin`接口，接口的定义如下

```go
type Plugin interface {
	// Start 插件启动时调用的方法
	Start(args []string)
	// Caller 接收到信号时调用的方法
	Caller(single Single)
	// GetName 主程序获取插件的名字
	GetName() string
	// GetType 主程序获取插件的类型
	GetType() Type
	// GetSupport 主程序获取插件需要的支持
	GetSupport() []int
	// SetStdout 设置Stdout
	SetStdout(writer io.Writer)
	// SetLogOut 设置日志接口
	SetLogOut(writer logger.Logger)
	// ConfRead 设置配置文件的读取接口
	ConfRead(reader io.Reader)
	// ConfWrite 设置配置文件的写入接口
	ConfWrite(writer io.Writer)
}
```

> 关于一些参数的释义

- 插件可以注册的类型，在`common/plugin`里定义的，这几种类型代表了在不同时期的调用，需要注意的是`Init`在应用程序运行时只会被调用一次

    ```go
    const (
    	Init      Type = 0 // 初始化时调用的插件
    	BCollect  Type = 1 // 需要搜集备份数据时调用的插件
    	BHandle   Type = 2 // 需要处理备份的数据时调用的插件
    	BCallBack Type = 3 // 处理完备份数据时调用的插件
    )
    ```

- 插件可以注册的支持，同样也是在`common/plugin`中定义，一个插件可以注册多种支持，有些支持需要设置方法接收，比如:`SupportLogger -> SetLogOut(writer logger.Logger)`

    ```go
    const (
    	// SupportArgs 命令行参数支持
    	SupportArgs int = iota
    	// SupportLogger 输出到内置日志的支持
    	SupportLogger
    	// SupportConfigRead 配置文件读取的支持
    	SupportConfigRead
    	// SupportConfigWrite 配置文件写入的支持
    	SupportConfigWrite
    	// SupportNativeStdout 共享输出缓冲区的支持
    	SupportNativeStdout
    )
    ```

- 插件接收包装的信号，在插件中接收的信号并不是来自于`syscall`包定义的信号，而是在`common/plugin`中定义的，实际上它是`int`类型的包装

    ```go
    type Single int
    ```

    > 目前只定义了少量信号

    ```go
    const (
    	// Exit 退出
    	Exit Single = iota
    )
    ```

---

> 更多的示例可以查看自带的插件源码，它们位于`plugins`内

