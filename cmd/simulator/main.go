package main

import (
    "flag"
    "log"
    "os"
    "os/signal"
    "syscall"

    "github.com/winterbear03/iot-edge-gateway/internal/collector"
)

func main() {
    mode := flag.String("mode", "tcp", "模拟器模式: tcp 或 rtu")
    tcpAddr := flag.String("tcp-addr", ":502", "TCP 监听地址")
    portName := flag.String("port", "COM1", "串口名称")
    baudRate := flag.Int("baud", 9600, "波特率")
    dataBits := flag.Int("data-bits", 8, "数据位")
    stopBits := flag.Int("stop-bits", 1, "停止位")
    parity := flag.String("parity", "none", "校验位: none, even, odd")
    flag.Parse()

    log.Println("启动 Modbus 从站模拟器...")
    if *mode == "rtu" {
        log.Printf("RTU 模式: 端口=%s, 波特率=%d, 数据位=%d, 停止位=%d, 校验=%s",
            *portName, *baudRate, *dataBits, *stopBits, *parity)
        srv := collector.NewModbusRTUSlaveSimulator(*portName, *baudRate, *dataBits, *stopBits, *parity)
        go srv.ListenAndServe()
    } else {
        log.Printf("TCP 模式: 地址=%s", *tcpAddr)
        srv := collector.NewModbusSlaveSimulator(*tcpAddr)
        go srv.ListenAndServe()
    }

    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    <-quit
    log.Println("模拟器已停止")
}
