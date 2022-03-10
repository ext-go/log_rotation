# Go日志文件归档包（自动分割日志文件）

## 归档类型

- 按日志文件大小进行归档
- 按日志文件行数进行归档
- 按时间进行归档

## 使用说明
- 此包实现自定义io.Write接口，实现日志文件的异步写入以及分割
- 如果任何原因导致的文件创建失败，本包会自动以stdout（标准输出）来接管您的程序日志
    ### std log (go标准库内置日志包)
   ```go
    // 初始化logRotation
    logRota := NewLogRotation(SetFileName("./test.log"), SetLimitUseMaxLines(1000))
    // 启动日志异步写入，必写，否则会导致日志文件堆积，不写入文件！
    logRota.Launch()
    log.SetOutput(w)
    log.Println("test 1")
    ```
    ### zap uber开源的日志包
    ```go
    // 初始化logRotation
    logRota := NewLogRotation(SetFileName("./test.log"), SetLimitUseMaxLines(1000))
    // 启动日志异步写入，必写，否则会导致日志文件堆积，不写入文件！
    logRota.Launch()
    var encoder = zap.NewDevelopmentEncoderConfig()
    core := zapcore.NewCore(zapcore.NewConsoleEncoder(encoder), zapcore.AddSync(m), zap.DebugLevel)
    logger := zap.New(core)
    logger.Info("test msg", zap.Int("line", 22))
    ```
    ### logrus
    ```go
    // 初始化logRotation
    logRota := NewLogRotation(SetFileName("./test.log"), SetLimitUseMaxLines(1000))
    // 启动日志异步写入，必写，否则会导致日志文件堆积，不写入文件！
    logRota.Launch()
    logrus.SetOutput(m)
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