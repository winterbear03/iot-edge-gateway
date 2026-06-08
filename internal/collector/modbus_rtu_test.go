package collector

import (
    "bytes"
    "testing"
    "time"
)

func TestCRC16(t *testing.T) {
    // Known Modbus test vector: 01 03 00 00 00 01 → CRC = 0x840A
    data := []byte{0x01, 0x03, 0x00, 0x00, 0x00, 0x01}
    expected := uint16(0x840A)
    if got := CRC16(data); got != expected {
        t.Errorf("CRC16(%X) = 0x%04X, want 0x%04X", data, got, expected)
    }
    // Verify bytes: LSB=0x0A, MSB=0x84 for on-wire transmission
    if byte(expected) != 0x0A || byte(expected>>8) != 0x84 {
        t.Errorf("CRC16 wire bytes: got [%02X %02X], want [0A 84]", byte(expected), byte(expected>>8))
    }
}

func TestCRC16_SingleByte(t *testing.T) {
    // CRC of single byte 0x00, byte-swapped for wire transmission
    expected := uint16(0xBF40)
    if got := CRC16([]byte{0x00}); got != expected {
        t.Errorf("CRC16([0x00]) = 0x%04X, want 0x%04X", got, expected)
    }
    if byte(expected) != 0x40 || byte(expected>>8) != 0xBF {
        t.Errorf("CRC16([0x00]) wire bytes: got [%02X %02X], want [40 BF]", byte(expected), byte(expected>>8))
    }
}

func TestCRC16_Empty(t *testing.T) {
    // CRC of empty data = 0xFFFF
    if got := CRC16([]byte{}); got != 0xFFFF {
        t.Errorf("CRC16([]) = 0x%04X, want 0xFFFF", got)
    }
}

func TestEncodeRTUFrame(t *testing.T) {
    frame := encodeRTUFrame(0x01, 0x03, []byte{0x00, 0x00, 0x00, 0x01})
    if len(frame) != 8 {
        t.Fatalf("frame length = %d, want 8", len(frame))
    }
    if frame[0] != 0x01 {
        t.Errorf("unitID = 0x%02X", frame[0])
    }
    if frame[1] != 0x03 {
        t.Errorf("funcCode = 0x%02X", frame[1])
    }
    // Verify CRC
    crc := CRC16(frame[:len(frame)-2])
    expected := uint16(frame[len(frame)-2]) | uint16(frame[len(frame)-1])<<8
    if crc != expected {
        t.Errorf("CRC in frame: computed 0x%04X, embedded 0x%04X", crc, expected)
    }
}

func TestDecodeFrame_Valid(t *testing.T) {
    frame := encodeRTUFrame(0x0A, 0x03, []byte{0x02, 0x12, 0x34})
    unitID, funcCode, data, err := DecodeFrame(frame)
    if err != nil {
        t.Fatalf("DecodeFrame error: %v", err)
    }
    if unitID != 0x0A {
        t.Errorf("unitID = 0x%02X", unitID)
    }
    if funcCode != 0x03 {
        t.Errorf("funcCode = 0x%02X", funcCode)
    }
    if !bytes.Equal(data, []byte{0x02, 0x12, 0x34}) {
        t.Errorf("data = %X", data)
    }
}

func TestDecodeFrame_CRCMismatch(t *testing.T) {
    frame := encodeRTUFrame(0x01, 0x03, []byte{0x00, 0x01})
    frame[len(frame)-1] ^= 0xFF // corrupt CRC
    _, _, _, err := DecodeFrame(frame)
    if err == nil {
        t.Error("expected CRC mismatch error")
    }
}

func TestDecodeFrame_TooShort(t *testing.T) {
    _, _, _, err := DecodeFrame([]byte{0x01, 0x03})
    if err == nil {
        t.Error("expected error for short frame")
    }
}

func TestCharTimeout_Below19200(t *testing.T) {
    // 9600 baud: t3.5 = 3.5 * 11 / 9600 ≈ 4.01ms
    timeout := charTimeout(9600)
    if timeout < 4*time.Millisecond || timeout > 5*time.Millisecond {
        t.Errorf("charTimeout(9600) = %v, expected ~4ms", timeout)
    }
}

func TestCharTimeout_Above19200(t *testing.T) {
    timeout := charTimeout(38400)
    if timeout != 1750*time.Microsecond {
        t.Errorf("charTimeout(38400) = %v, want 1.75ms", timeout)
    }
}

func TestRoundTrip(t *testing.T) {
    original := []byte{0x00, 0x00, 0x00, 0x01}
    frame := encodeRTUFrame(0x01, 0x03, original)
    unitID, funcCode, data, err := DecodeFrame(frame)
    if err != nil {
        t.Fatalf("round-trip error: %v", err)
    }
    if unitID != 0x01 || funcCode != 0x03 || !bytes.Equal(data, original) {
        t.Errorf("round-trip mismatch: addr=%X func=%X data=%X", unitID, funcCode, data)
    }
}
