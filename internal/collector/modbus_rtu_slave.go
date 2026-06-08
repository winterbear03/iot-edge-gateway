package collector

import (
    "encoding/binary"
    "log"
    "sync"

    "go.bug.st/serial"
)

// ModbusRTUSlaveSimulator is a Modbus RTU slave that listens on a serial port.
type ModbusRTUSlaveSimulator struct {
    portName string
    baudRate int
    dataBits int
    stopBits int
    parity   string
    holding  map[uint16]uint16
    mu       sync.RWMutex
}

func NewModbusRTUSlaveSimulator(portName string, baudRate, dataBits, stopBits int, parity string) *ModbusRTUSlaveSimulator {
    s := &ModbusRTUSlaveSimulator{
        portName: portName,
        baudRate: baudRate,
        dataBits: dataBits,
        stopBits: stopBits,
        parity:   parity,
        holding:  make(map[uint16]uint16),
    }
    s.holding[0] = 12345
    return s
}

func (s *ModbusRTUSlaveSimulator) ListenAndServe() {
    mode := &serial.Mode{
        BaudRate: s.baudRate,
        DataBits: s.dataBits,
        StopBits: stopBitsFromInt(s.stopBits),
        Parity:   parityFromString(s.parity),
    }
    port, err := serial.Open(s.portName, mode)
    if err != nil {
        log.Fatalf("RTU 从站: 打开串口 %s 失败: %v", s.portName, err)
    }
    defer port.Close()
    log.Printf("RTU 从站: 串口 %s 已打开 (波特率=%d)", s.portName, s.baudRate)

    timeout := charTimeout(s.baudRate)
    for {
        frame, err := readRTUFrame(port, timeout)
        if err != nil {
            continue
        }

        unitID, funcCode, data, err := DecodeFrame(frame)
        if err != nil {
            log.Printf("RTU 从站: 帧解码失败: %v", err)
            continue
        }

        if funcCode != 3 {
            continue
        }

        resp := s.processRTURequest(unitID, data)
        if resp != nil {
            if err := writeRTUFrame(port, resp, timeout); err != nil {
                log.Printf("RTU 从站: 响应写入失败: %v", err)
            }
        }
    }
}

func (s *ModbusRTUSlaveSimulator) processRTURequest(unitID byte, data []byte) []byte {
    if len(data) < 4 {
        return nil
    }
    start := binary.BigEndian.Uint16(data[0:2])
    quantity := binary.BigEndian.Uint16(data[2:4])
    if quantity == 0 || quantity > 125 {
        return nil
    }

    s.mu.RLock()
    defer s.mu.RUnlock()

    pdu := make([]byte, 1+int(quantity)*2)
    pdu[0] = byte(quantity * 2)
    for i := uint16(0); i < quantity; i++ {
        val := s.holding[start+i]
        binary.BigEndian.PutUint16(pdu[1+int(i)*2:], val)
    }
    return encodeRTUFrame(unitID, 3, pdu)
}
