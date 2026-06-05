package storage

import (
    "context"
    "log"
    "time"
    "github.com/winterbear03/iot-edge-gateway/internal/config"
    influxdb2 "github.com/influxdata/influxdb-client-go/v2"
)

type InfluxDBWriter struct {
    client influxdb2.Client
    org    string
    bucket string
}

func NewInfluxDBWriter(cfg config.StorageConfig) *InfluxDBWriter {
    client := influxdb2.NewClient(cfg.InfluxDBURL, cfg.InfluxDBToken)
    return &InfluxDBWriter{client: client, org: cfg.InfluxDBOrg, bucket: cfg.InfluxDBBucket}
}

func (w *InfluxDBWriter) Write(deviceID, field string, value float64) {
    writeAPI := w.client.WriteAPIBlocking(w.org, w.bucket)
    p := influxdb2.NewPointWithMeasurement("gateway").
        AddTag("device", deviceID).
        AddField(field, value).
        SetTime(time.Now())
    if err := writeAPI.WritePoint(context.Background(), p); err != nil {
        log.Printf("写入 InfluxDB 失败: %v", err)
    }
}
