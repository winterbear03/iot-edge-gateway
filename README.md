# iot-edge-gateway

工业边缘网关系统（Go 实现），负责从工业现场设备采集数据，通过 MQTT 上报至云平台，同时提供 Web 管理仪表盘。

## 架构

```
设备层                     网关核心                          基础设施               云平台
┌──────────────┐      ┌──────────────────┐           ┌──────────────┐       ┌────────┐
│ Modbus TCP   │─────▶│                  │           │  InfluxDB    │       │        │
│ 模拟器/设备   │      │  Collector       │──────────▶│  (时序存储)    │       │  MQTT  │
└──────────────┘      │  (数据采集)        │           └──────────────┘       │ Broker │
                       │                  │           ┌──────────────┐       │        │
┌──────────────┐      │                  │           │  Redis       │       │        │
│ Modbus RTU   │─────▶│  DataBus         │◀─────────▶│  (消息总线)    │       │        │
│ 模拟器/设备   │      │  (内部数据总线)    │           └──────────────┘       └────────┘
└──────────────┘      │                  │                                       │
                       │  Reporter        │─────────────────────────────────────▶│
                       │  (MQTT 上报)      │
                       │                  │
                       │  HTTP Server     │
                       │  (API + 前端)     │
                       └──────────────────┘
                              │
                    浏览器 ◀───┘  (Web 管理仪表盘)
```

## 目录结构

```
iot-edge-gateway/
├── cmd/
│   ├── gateway/         # 网关主程序入口
│   └── simulator/       # Modbus 从站模拟器（用于开发调试）
├── config/
│   └── config.yaml      # 网关配置文件
├── deploy/
│   └── mosquitto.conf   # MQTT Broker 配置文件
├── frontend/            # Vue 3 前端（管理仪表盘）
│   └── src/
│       └── App.vue      # 单文件 SPA 应用
├── internal/
│   ├── api/             # HTTP API 接口
│   ├── bus/             # 内部数据总线（Redis Pub/Sub）
│   ├── collector/       # Modbus 数据采集器（TCP + RTU）
│   ├── config/          # 配置文件解析
│   ├── daemon/          # 守护进程管理器
│   ├── reporter/        # MQTT 上报模块
│   └── storage/         # InfluxDB 时序数据写入
├── docker-compose.yml   # 一键启动所有依赖服务
├── Dockerfile           # 网关服务镜像构建
└── go.mod
```

## 核心模块说明

### 1. 数据采集器 (collector)

负责与现场 Modbus 设备通信，支持两种传输模式：

| 模式 | 说明 |
|------|------|
| **Modbus TCP** | 通过 TCP/IP 网络连接设备，默认地址 `127.0.0.1:502` |
| **Modbus RTU** | 通过串口（RS-232/RS-485）连接设备，支持配置波特率、数据位、停止位、校验位 |

采集器每 5 秒轮询一次所有已配置的设备，读取保持寄存器的值。RTU 模式实现了完整的帧编解码（CRC16 校验、3.5 字符帧间隔）。

### 2. 数据总线 (bus)

基于 Redis Pub/Sub 的内部消息总线，用于模块间解耦通信。采集器将原始数据发布到总线，上报模块和存储模块可订阅相应的数据通道。

### 3. 时序存储 (storage)

将采集到的设备数据写入 InfluxDB 2.x，以 `gateway` 为 measurement，按设备 ID 打 tag，支持后续按设备和时间范围查询。

### 4. MQTT 上报 (reporter)

通过 MQTT 协议将处理后的数据上报至云平台 Broker，QoS 1，自动重连。

### 5. HTTP API + 前端

- **API**: 提供 `/api/health` 健康检查端点
- **前端**: Vue 3 单页应用，构建产物嵌入 Go 二进制中，单文件部署。包含仪表盘、设备管理、系统日志、系统设置四个页面

## 快速开始

### 前置条件

- Docker & Docker Compose
- Go 1.26+（仅本地编译需要）

### 1. 启动基础设施

```bash
docker-compose up -d mosquitto redis influxdb
```

这会启动：
- **Mosquitto** MQTT Broker → `localhost:7805`
- **Redis** → `localhost:7807`
- **InfluxDB 2.x** → `localhost:7809`

### 2. 配置设备

编辑 `config/config.yaml`，根据实际设备修改：

```yaml
devices:
  - id: "temp_sensor_01"     # 设备唯一标识
    protocol: modbus          # 协议（目前仅支持 modbus）
    transport: rtu            # 传输方式：tcp 或 rtu
    address: 1                # Modbus 从站地址
    points:
      - name: "temperature"   # 数据点名称
        register: 0           # 寄存器地址
        type: uint16          # 数据类型
        scale: 0.1            # 缩放系数（原始值 × scale = 实际值）
```

对于 TCP 设备，设置 `transport: tcp` 即可（默认连接 `127.0.0.1:502`）。

对于 RTU 设备，还需配置串口参数：

```yaml
serial:
  port: "COM3"      # Windows 串口；Linux 上如 /dev/ttyUSB0
  baud_rate: 9600
  data_bits: 8
  stop_bits: 1
  parity: "none"
```

### 3. 启动网关

**方式一：Docker Compose 一键启动**

```bash
docker-compose up -d
```

**方式二：本地编译运行**

```bash
# 构建前端
cd frontend && npm install && npm run build

# 将构建产物拷贝到 cmd/gateway/web/ 目录
# 然后编译网关
go build -o gateway ./cmd/gateway

# 运行
./gateway
```

启动后访问 `http://localhost:7819` 打开管理仪表盘。

### 4. 使用模拟器（开发调试）

如果没有真实 Modbus 设备，可以使用内置的模拟器：

```bash
# TCP 模式模拟器（默认监听 :502）
go run ./cmd/simulator -mode tcp

# RTU 模式模拟器（需要串口，可配合虚拟串口使用）
go run ./cmd/simulator -mode rtu -port COM1 -baud 9600
```

## 配置参考

完整的 `config/config.yaml` 配置项说明：

| 配置节 | 字段 | 说明 | 默认值 |
|--------|------|------|--------|
| `daemon` | `health_check_interval_sec` | 健康检查间隔（秒） | `10` |
| `bus` | `redis_addr` | Redis 地址 | `127.0.0.1:7807` |
| `serial` | `port` | 串口路径 | `COM3` |
| `serial` | `baud_rate` | 波特率 | `9600` |
| `serial` | `data_bits` | 数据位 | `8` |
| `serial` | `stop_bits` | 停止位 | `1` |
| `serial` | `parity` | 校验位 (none/even/odd) | `none` |
| `devices` | `id` | 设备唯一标识 | — |
| `devices` | `protocol` | 协议类型 | `modbus` |
| `devices` | `transport` | 传输方式 (tcp/rtu) | `tcp` |
| `devices` | `address` | Modbus 从站地址 | `1` |
| `devices[].points` | `name` | 数据点名称 | — |
| `devices[].points` | `register` | 保持寄存器地址 | `0` |
| `devices[].points` | `type` | 数据类型 | `uint16` |
| `devices[].points` | `scale` | 缩放系数 | `1.0` |
| `storage` | `influxdb_url` | InfluxDB 地址 | `http://127.0.0.1:7809` |
| `storage` | `influxdb_token` | InfluxDB 认证 Token | — |
| `storage` | `influxdb_org` | InfluxDB 组织名 | `gateway` |
| `storage` | `influxdb_bucket` | InfluxDB 存储桶 | `default` |
| `server` | `port` | HTTP 服务端口 | `7819` |

## 数据流

```
1. 网关启动 → 加载 config.yaml
2. Daemon Manager → 启动 Collector
3. Collector → 每 5s 轮询设备（Modbus TCP / RTU）
4. Collector → 将原始数据发布到 Redis DataBus
5. Storage 模块 → 订阅 DataBus → 写入 InfluxDB
6. Reporter 模块 → 订阅 DataBus → 通过 MQTT 上报云平台
7. HTTP Server → 提供 /api/health + Vue 前端界面
```

## 技术栈

| 层级 | 技术 |
|------|------|
| 语言 | Go 1.26 |
| HTTP 框架 | Gin |
| MQTT 客户端 | Eclipse Paho |
| 消息总线 | Redis Pub/Sub |
| 时序数据库 | InfluxDB 2.x |
| 串口通信 | go.bug.st/serial |
| 前端 | Vue 3 + Vite |
| 容器化 | Docker + Docker Compose |
