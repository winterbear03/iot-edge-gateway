package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/winterbear03/iot-edge-gateway/internal/collector"
)

func main() {
	log.Println("启动 Modbus 从站模拟器...")
	srv := collector.NewModbusSlaveSimulator(":502")
	go srv.ListenAndServe()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("模拟器已停止")
}
