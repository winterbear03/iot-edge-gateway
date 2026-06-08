package collector

import (
    "encoding/binary"
    "log"
    "net"
    "sync"
    "time"

    "github.com/winterbear03/iot-edge-gateway/internal/config"
)

// SerialReadWriter abstracts a serial port for RTU communication.
type SerialReadWriter interface {
    Read(p []byte) (n int, err error)
    Write(p []byte) (n int, err error)
    Close() error
    SetReadTimeout(d time.Duration) error
}

// CRC16 computes the Modbus RTU CRC16 over data.
// Polynomial 0xA001, initial value 0xFFFF.
// The return value is byte-swapped so that byte(crc) gives the LSB
// and byte(crc>>8) gives the MSB for on-wire transmission.
func CRC16(data []byte) uint16 {
    crc := uint16(0xFFFF)
    for _, b := range data {
        crc ^= uint16(b)
        for i := 0; i < 8; i++ {
            if crc&0x0001 != 0 {
                crc = (crc >> 1) ^ 0xA001
            } else {
                crc >>= 1
            }
        }
    }
    return (crc >> 8) | (crc << 8)
}

// buildReadHoldingRegistersPDU returns the Modbus PDU for Read Holding Registers (FC 0x03).
func buildReadHoldingRegistersPDU(startAddr, quantity uint16) []byte {
    pdu := make([]byte, 5)
    pdu[0] = 3
    binary.BigEndian.PutUint16(pdu[1:3], startAddr)
    binary.BigEndian.PutUint16(pdu[3:5], quantity)
    return pdu
}

// ---------- Modbus Collector ----------

type ModbusCollector struct {
    mu         sync.RWMutex
    devices    []config.DeviceConfig
    serialCfg  config.SerialConfig
    serialPort SerialReadWriter
    serialMu   sync.Mutex
}

func NewModbusCollector(devices []config.DeviceConfig, serialCfg config.SerialConfig) *ModbusCollector {
    return &ModbusCollector{devices: devices, serialCfg: serialCfg}
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
        if dev.Protocol != "modbus" {
            continue
        }
        switch dev.Transport {
        case "rtu":
            m.pollRTUDevice(dev)
        default:
            m.pollTCPDevice(dev)
        }
    }
}

func (m *ModbusCollector) pollTCPDevice(dev config.DeviceConfig) {
    conn, err := net.DialTimeout("tcp", "127.0.0.1:502", 2*time.Second)
    if err != nil {
        log.Printf("连接 %s 失败: %v", dev.ID, err)
        return
    }
    defer conn.Close()

    req := buildReadHoldingRegistersRequest(byte(dev.Address), 0, 1)
    if _, err := conn.Write(req); err != nil {
        log.Printf("写入 %s 失败: %v", dev.ID, err)
        return
    }
    resp := make([]byte, 256)
    n, _ := conn.Read(resp)
    if n < 9 {
        return
    }
    val := binary.BigEndian.Uint16(resp[9:11])
    log.Printf("设备 %s 采集值: %d", dev.ID, val)
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

// ---------- TCP 从站模拟器 ----------

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
    if err != nil {
        log.Fatal(err)
    }
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
        if err != nil {
            return
        }
        resp := s.processRequest(buf[:n])
        conn.Write(resp)
    }
}

func (s *ModbusSlaveSimulator) processRequest(req []byte) []byte {
    if len(req) < 8 || req[7] != 3 {
        return nil
    }
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
