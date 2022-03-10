# Go日志文件归档包（自动分割日志文件）
log_rotation、log_truncation、logrus、zap、std log
## 归档类型

- 按日志文件大小进行归档
- 按日志文件行数进行归档
- 按时间进行归档

## 功能说明
- 因go语言的多数日志包不提供日志文件分割功能，而大部分日志分割包已经久不维护，且大部分新手开发者（包含此包作者）会一股脑的将程序产生的output和err信息对入单一的文件，日积月累之后，文件体积越来越大，导致各种不便，因此开发本包，以供自己与所有go语言爱好者使用方便。
- 此包会将日志异步写入文件，如日志消息产生过多，包内的无限长队列会自动进行扩容操作（queueLength * 2），日志堆积处理至（logNum < queueLength / 4）时，内置队列会自动进行缩容操作。
- 因为队列的特性，如在每次pop之后进行手动内存对齐对性能产生过多压力，故这里采用一个阈值，如日志处理或日志累积数量各自达到阈值之后，队列中的数据会自动迁移到前置空余内存，方便后续追加内容（实现见 [queue.go](./queue.go) 文件中的`elastic`函数）。
- 初步设计中，包内的队列以链表形式实现，由于 *过多的小内存对象* 会对 *gc* 产生过多的压力，故这里目前采用slice实现（更好更快的实现方式仍在探索，以为每次扩容会对有原有数据进行一次copy，虽然全在内存中操作，但是要避免对于的性能损耗，后续可能会采用链表+slice进行重构）。

## logFile说明
- 在`NewLogRotation`函数中通过`SetFileName(filename)`进行设置，可包含已存在或不存在的路径，如`/home/mylog.log`、`/myblog/output.log`
- 自动创建的文件格式 `{path}/{fileBaseName}-{yyy_MM_dd}-{uniqueId}.{fileExt}`
- SetLimitUseMaxLines：通过写入日志的行数自动进行文件分割，*建议值 > 100*
- SetLimitUseTime: 通过时间进行文件分割，*建议值 > time.Hour * 1*
- SetLimitUseFileSize: 通过文件大小进行文件分割，*传入单位为kb，建议值 > 1024kb*
- 以上Limit类型需传入`NewLogRotation`函数，且只生效最后一个，故建议只使用一个限制类型

## 使用说明
- 此包实现自定义io.Write接口，实现日志文件的异步写入以及分割
- 如果任何原因导致的文件创建失败，本包会自动以stdout（标准输出）来接管您的程序日志
    ### std log (go标准库内置日志包)
   ```go
    // 初始化logRotation
    logRota := NewLogRotation(SetFileName("./test.log"), SetLimitUseMaxLines(1000))
    // 启动日志异步写入，必写，否则会导致日志文件堆积，不写入文件！
    logRota.Launch()
    log.SetOutput(logRota)
    log.Println("test 1")
    ```
    ### zap uber开源的日志包
    ```go
    // 初始化logRotation
    logRota := NewLogRotation(SetFileName("./test.log"), SetLimitUseMaxLines(1000))
    // 启动日志异步写入，必写，否则会导致日志文件堆积，不写入文件！
    logRota.Launch()
    var encoder = zap.NewDevelopmentEncoderConfig()
    core := zapcore.NewCore(zapcore.NewConsoleEncoder(encoder), zapcore.AddSync(logRota), zap.DebugLevel)
    logger := zap.New(core)
    logger.Info("test msg", zap.Int("line", 22))
    ```
    ### logrus
    ```go
    // 初始化logRotation
    logRota := NewLogRotation(SetFileName("./test.log"), SetLimitUseMaxLines(1000))
    // 启动日志异步写入，必写，否则会导致日志文件堆积，不写入文件！
    logRota.Launch()
    logrus.SetOutput(logRota)
    logrus.Println("test log")
    ```
## 基准测试
```shell
goos: windows
goarch: amd64
pkg: github.com/ext-go/log_rotation
cpu: Intel(R) Core(TM) i5-3570 CPU @ 3.40GHz
BenchmarkNewLogRotation
BenchmarkNewLogRotation/log
BenchmarkNewLogRotation/log-4            2562033               419.4 ns/op
             124 B/op          2 allocs/op
PASS
```
- 本包内置一个自动扩容、缩容的队列实现 [查看源码](./queue.go)

    - 基准测试

    ```shell
    goos: windows
    goarch: amd64
    pkg: github.com/ext-go/log_rotation
    cpu: Intel(R) Core(TM) i5-3570 CPU @ 3.40GHz
    BenchmarkQueue
    BenchmarkQueue/insert
    BenchmarkQueue/insert-4                 12903628                80.41 ns/op
    66 B/op          1 allocs/op
    BenchmarkQueue/get
    BenchmarkQueue/get-4                    58238360                20.96 ns/op
    0 B/op          0 allocs/op
    PASS
    
    Process finished with the exit code 0
    ```