package collector

import (
    "encoding/binary"
    "fmt"
    "log"

    "github.com/winterbear03/iot-edge-gateway/internal/config"
    "go.bug.st/serial"
)

func parityFromString(p string) serial.Parity {
    switch p {
    case "even":
        return serial.EvenParity
    case "odd":
        return serial.OddParity
    default:
        return serial.NoParity
    }
}

func stopBitsFromInt(n int) serial.StopBits {
    if n == 2 {
        return serial.TwoStopBits
    }
    return serial.OneStopBit
}

func (m *ModbusCollector) ensureSerialPort() error {
    m.serialMu.Lock()
    defer m.serialMu.Unlock()

    if m.serialPort != nil {
        return nil
    }

    mode := &serial.Mode{
        BaudRate: m.serialCfg.BaudRate,
        DataBits: m.serialCfg.DataBits,
        StopBits: stopBitsFromInt(m.serialCfg.StopBits),
        Parity:   parityFromString(m.serialCfg.Parity),
    }
    port, err := serial.Open(m.serialCfg.Port, mode)
    if err != nil {
        return fmt.Errorf("打开串口 %s 失败: %w", m.serialCfg.Port, err)
    }
    m.serialPort = port
    log.Printf("串口 %s 已打开 (波特率=%d, 数据位=%d, 停止位=%d, 校验=%s)",
        m.serialCfg.Port, m.serialCfg.BaudRate, m.serialCfg.DataBits, m.serialCfg.StopBits, m.serialCfg.Parity)
    return nil
}

func (m *ModbusCollector) pollRTUDevice(dev config.DeviceConfig) {
    if err := m.ensureSerialPort(); err != nil {
        log.Printf("RTU 初始化失败: %v", err)
        return
    }

    m.serialMu.Lock()
    defer m.serialMu.Unlock()

    timeout := charTimeout(m.serialCfg.BaudRate)

    pdu := buildReadHoldingRegistersPDU(0, 1)
    req := encodeRTUFrame(byte(dev.Address), 3, pdu)

    if err := writeRTUFrame(m.serialPort, req, timeout); err != nil {
        log.Printf("RTU 写入 %s 失败: %v", dev.ID, err)
        return
    }

    resp, err := readRTUFrame(m.serialPort, timeout)
    if err != nil {
        log.Printf("RTU 读取 %s 失败: %v", dev.ID, err)
        return
    }

    unitID, funcCode, data, err := DecodeFrame(resp)
    if err != nil {
        log.Printf("RTU 解码 %s 失败: %v", dev.ID, err)
        return
    }
    if unitID != byte(dev.Address) {
        log.Printf("RTU %s 响应地址不匹配: 期望 %d, 收到 %d", dev.ID, dev.Address, unitID)
        return
    }
    if funcCode&0x80 != 0 {
        log.Printf("RTU %s 返回异常: 功能码 0x%02X", dev.ID, funcCode)
        return
    }
    if len(data) < 2 || data[0] < 2 {
        log.Printf("RTU %s 响应数据不足", dev.ID)
        return
    }
    val := binary.BigEndian.Uint16(data[1:3])
    log.Printf("RTU 设备 %s (地址=%d) 采集值: %d", dev.ID, dev.Address, val)
}
