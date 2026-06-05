package config

import (
    "os"
    "gopkg.in/yaml.v3"
)

type Config struct {
    Daemon  DaemonConfig   `yaml:"daemon"`
    Bus     BusConfig      `yaml:"bus"`
    Devices []DeviceConfig `yaml:"devices"`
    Storage StorageConfig  `yaml:"storage"`
    Server  ServerConfig   `yaml:"server"`
}

type DaemonConfig struct { HealthCheckInterval int `yaml:"health_check_interval_sec"` }
type BusConfig struct { RedisAddr string `yaml:"redis_addr"` }
type DeviceConfig struct {
    ID       string  `yaml:"id"`
    Protocol string  `yaml:"protocol"`
    Address  int     `yaml:"address"`
    Points   []Point `yaml:"points"`
}
type Point struct {
    Name     string  `yaml:"name"`
    Register int     `yaml:"register"`
    Type     string  `yaml:"type"`
    Scale    float64 `yaml:"scale"`
}
type StorageConfig struct {
    InfluxDBURL    string `yaml:"influxdb_url"`
    InfluxDBToken  string `yaml:"influxdb_token"`
    InfluxDBOrg    string `yaml:"influxdb_org"`
    InfluxDBBucket string `yaml:"influxdb_bucket"`
}
type ServerConfig struct { Port int `yaml:"port"` }

func Load() (*Config, error) {
    data, err := os.ReadFile("config/config.yaml")
    if err != nil { return nil, err }
    var cfg Config
    err = yaml.Unmarshal(data, &cfg)
    return &cfg, err
}
