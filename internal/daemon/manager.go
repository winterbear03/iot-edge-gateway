package daemon

import (
    "log"
    "github.com/winterbear03/iot-edge-gateway/internal/config"
)

type Manager struct { cfg *config.Config }

func NewManager(cfg *config.Config) *Manager { return &Manager{cfg: cfg} }
func (m *Manager) Start() { log.Println("守护进程启动") }
func (m *Manager) Stop()  { log.Println("守护进程停止") }
