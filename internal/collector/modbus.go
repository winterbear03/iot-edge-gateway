package collector

import (
    "encoding/binary"
    "log"
    "net"
    "sync"
    "time"
    "github.com/winterbear03/iot-edge-gateway/internal/config"
)

type ModbusCollector struct {
    mu      sync.RWMutex
    devices []config.DeviceConfig
}

func NewModbusCollector(devices []config.DeviceConfig) *ModbusCollector {
    return &ModbusCollector{devices: devices}
}

func (m *ModbusCollector) Start() {
    log.Println("Modbus 采集器启动")
    go func() {
        for {
            m.pollDevices()
            time.Sleep(5 * time.Second)
        }
    }()
}

func (m *ModbusCollector) pollDevices() {
    for _, dev := range m.devices {
        if dev.Protocol != "modbus" { continue }
        conn, err := net.DialTimeout("tcp", "127.0.0.1:502", 2*time.Second)
        if err != nil { log.Printf("连接 %s 失败: %v", dev.ID, err); continue }
        req := buildReadHoldingRegistersRequest(byte(dev.Address), 0, 1)
        conn.Write(req)
        resp := make([]byte, 256)
        n, _ := conn.Read(resp)
        conn.Close()
        if n < 9 { continue }
        val := binary.BigEndian.Uint16(resp[9:11])
        log.Printf("设备 %s 采集值: %d", dev.ID, val)
    }
}

func buildReadHoldingRegistersRequest(unitID byte, startAddr, quantity uint16) []byte {
    frame := make([]byte, 12)
    binary.BigEndian.PutUint16(frame[0:2], 0)
    binary.BigEndian.PutUint16(frame[2:4], 0)
    binary.BigEndian.PutUint16(frame[4:6], 6)
    frame[6] = unitID
    frame[7] = 3
    binary.BigEndian.PutUint16(frame[8:10], startAddr)
    binary.BigEndian.PutUint16(frame[10:12], quantity)
    return frame
}

// ---------- 从站模拟器 ----------
type ModbusSlaveSimulator struct {
    addr    string
    holding map[uint16]uint16
}

func NewModbusSlaveSimulator(addr string) *ModbusSlaveSimulator {
    s := &ModbusSlaveSimulator{addr: addr, holding: make(map[uint16]uint16)}
    s.holding[0] = 12345
    return s
}

func (s *ModbusSlaveSimulator) ListenAndServe() {
    ln, err := net.Listen("tcp", s.addr)
    if err != nil { log.Fatal(err) }
    defer ln.Close()
    for {
        conn, _ := ln.Accept()
        go s.handleConnection(conn)
    }
}

func (s *ModbusSlaveSimulator) handleConnection(conn net.Conn) {
    defer conn.Close()
    buf := make([]byte, 260)
    for {
        n, err := conn.Read(buf)
        if err != nil { return }
        resp := s.processRequest(buf[:n])
        conn.Write(resp)
    }
}

func (s *ModbusSlaveSimulator) processRequest(req []byte) []byte {
    if len(req) < 8 || req[7] != 3 { return nil }
    start := binary.BigEndian.Uint16(req[8:10])
    quantity := binary.BigEndian.Uint16(req[10:12])
    resp := make([]byte, 9+int(quantity)*2)
    copy(resp[0:7], req[0:7])
    resp[7] = 3
    resp[8] = byte(quantity * 2)
    for i := uint16(0); i < quantity; i++ {
        val := s.holding[start+i]
        binary.BigEndian.PutUint16(resp[9+int(i)*2:], val)
    }
    binary.BigEndian.PutUint16(resp[4:6], uint16(3+int(quantity)*2))
    return resp
}
