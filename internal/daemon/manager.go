package daemon

import (
    "log"

    "github.com/winterbear03/iot-edge-gateway/internal/collector"
    "github.com/winterbear03/iot-edge-gateway/internal/config"
)

type Manager struct {
    cfg       *config.Config
    collector *collector.ModbusCollector
}

func NewManager(cfg *config.Config) *Manager {
    return &Manager{
        cfg:       cfg,
        collector: collector.NewModbusCollector(cfg.Devices, cfg.Serial),
    }
}

func (m *Manager) Start() {
    log.Println("守护进程启动")
    m.collector.Start()
}

func (m *Manager) Stop() {
    log.Println("守护进程停止")
}
