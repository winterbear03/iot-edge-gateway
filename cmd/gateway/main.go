package main

import (
	"embed"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/gin-gonic/gin"
	"github.com/winterbear03/iot-edge-gateway/internal/api"
	"github.com/winterbear03/iot-edge-gateway/internal/config"
	"github.com/winterbear03/iot-edge-gateway/internal/daemon"
)

//go:embed web/*
var webFiles embed.FS

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}
	dm := daemon.NewManager(cfg)
	dm.Start()

	r := gin.Default()

	// API 路由
	r.GET("/api/health", api.HealthHandler)

	// 前端静态文件（Vue 构建产物在 web 目录中）
	distFS, err := fs.Sub(webFiles, "web")
	if err != nil {
		log.Fatalf("嵌入前端文件失败: %v", err)
	}

	// 静态资源（/assets/* 请求）
	r.GET("/assets/*filepath", gin.WrapH(http.FileServer(http.FS(distFS))))

	// SPA 回退：所有其他路径返回 index.html
	r.NoRoute(func(c *gin.Context) {
		data, err := fs.ReadFile(distFS, "index.html")
		if err != nil {
			c.String(http.StatusNotFound, "Not Found")
			return
		}
		c.Data(http.StatusOK, "text/html; charset=utf-8", data)
	})

	// 启动 HTTP 服务
	go func() {
		addr := fmt.Sprintf(":%d", cfg.Server.Port)
		log.Printf("网关启动成功，访问 http://localhost%s", addr)
		if err := http.ListenAndServe(addr, r); err != nil {
			log.Fatalf("HTTP 服务启动失败: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("正在关闭网关...")
	dm.Stop()
	log.Println("网关已退出")
}
