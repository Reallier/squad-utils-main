package handler

import (
	"errors"
	"github.com/rs/zerolog"
)

var ErrNoAvailablePort = errors.New("没有可用端口")

type Common struct {
	// Image 所用镜像
	Image string
	// 容器名字,格式为 {UUID}-a2s-server
	Name string
	// Env 启动环境变量
	Env map[string]string
	// ConfigHash 配置哈希
	// 计算方法为 {Image}-{Name}-{ListenPort}-{QueryPort} 连起来做 sha256
	ConfigHash string
	// Logger 打日志用的
	Logger zerolog.Logger
	// Error 错误记录接口
	Error error
}
