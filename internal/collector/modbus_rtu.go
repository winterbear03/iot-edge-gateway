package collector

import (
    "encoding/binary"
    "fmt"
    "time"
)

// encodeRTUFrame builds an RTU ADU: [unitID][funcCode][data...][CRC_L][CRC_H].
func encodeRTUFrame(unitID byte, funcCode byte, data []byte) []byte {
    frame := make([]byte, 2+len(data)+2)
    frame[0] = unitID
    frame[1] = funcCode
    copy(frame[2:], data)
    crc := CRC16(frame[:len(frame)-2])
    frame[len(frame)-2] = byte(crc)
    frame[len(frame)-1] = byte(crc >> 8)
    return frame
}

// DecodeFrame validates and splits an RTU frame.
// Returns unitID, functionCode, data payload, and any error.
func DecodeFrame(frame []byte) (byte, byte, []byte, error) {
    if len(frame) < 4 {
        return 0, 0, nil, fmt.Errorf("frame too short: %d bytes", len(frame))
    }
    computed := CRC16(frame[:len(frame)-2])
    received := binary.LittleEndian.Uint16(frame[len(frame)-2:])
    if computed != received {
        return 0, 0, nil, fmt.Errorf("CRC mismatch: computed 0x%04X, received 0x%04X", computed, received)
    }
    return frame[0], frame[1], frame[2 : len(frame)-2], nil
}

// charTimeout returns the 3.5-character inter-frame silence duration.
// For baud >= 19200, Modbus spec mandates a fixed 1.75ms.
func charTimeout(baudRate int) time.Duration {
    if baudRate >= 19200 {
        return 1750 * time.Microsecond
    }
    return time.Duration(float64(3.5*11) / float64(baudRate) * float64(time.Second))
}

// readRTUFrame reads one complete RTU frame from the serial port using
// 3.5-char inter-frame silence for frame delimiting.
func readRTUFrame(port SerialReadWriter, interFrameTimeout time.Duration) ([]byte, error) {
    buf := make([]byte, 256)

    // Read first byte with a generous timeout.
    port.SetReadTimeout(time.Second)
    n, err := port.Read(buf[0:1])
    if err != nil {
        return nil, fmt.Errorf("read first byte: %w", err)
    }
    total := n

    // Read remaining bytes until inter-frame silence.
    port.SetReadTimeout(interFrameTimeout)
    for total < len(buf) {
        n, err := port.Read(buf[total:])
        if err != nil {
            break // timeout signals end of frame
        }
        total += n
    }
    return buf[:total], nil
}

// writeRTUFrame sends a complete RTU frame with pre-send inter-frame silence.
func writeRTUFrame(port SerialReadWriter, frame []byte, interFrameTimeout time.Duration) error {
    time.Sleep(interFrameTimeout)
    _, err := port.Write(frame)
    return err
}
