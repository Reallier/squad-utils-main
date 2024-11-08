package handler

import (
	"context"
	"errors"
	"github.com/docker/docker/api/types"
	"github.com/rs/zerolog/log"
	"squad-utils/config"
	"strings"
)

var ErrEnvHasNoEqual = errors.New("这个环境变量没有等于号无法分割")

// SpiltDockerEnv 分割来自 Docker 的环境变量
// 这么做可以方便我们自己检查用
func SpiltDockerEnv(env string) (key, value string, length int, err error) {
	// 分割成 K V
	kv := strings.Split(env, "=")
	length = len(kv)
	// 如果长度不足 2 说明没有 = 无法分割
	if length < 2 {
		err = ErrEnvHasNoEqual
		return
	}
	key = kv[0]
	// 剩下的部分全都是 value
	// 无论后面还有多少 =  全部给 value
	value = strings.Join(kv[1:], "=")
	return
}

// MergeDockerEnv 合并 KV 到 Docker 环境变量
// 毕竟你最后传递过去的时候还是要用一个完整的字符串的
func MergeDockerEnv(key, value string) string {
	return key + "=" + value
}

func HasImage(image string) bool {
	ctx := context.Background()
	_, _, err := config.DockerCli.ImageInspectWithRaw(ctx, image)
	if err != nil {
		return false
	}
	return true
}

// NeedRecreate 判断容器是否需要重新创建
// 如果容器不存在,也会返回 true
func NeedRecreate(name string, configHash string) bool {
	ctx := context.Background()
	container, err := config.DockerCli.ContainerInspect(ctx, name)
	if err != nil {
		// 任何错误都送去重建
		log.Err(err).Str("name", name).Msg("获取容器信息失败,标记重建")
		return true
	}
	// 读取容器 Label
	labels := container.Config.Labels
	// 如果没有 Label 也要重建
	if labels == nil {
		log.Debug().Str("name", name).Msg("容器没有 Label,需要重建")
		return true
	}
	// 从 Label 里面读取 sq.confighash
	h, ok := labels["sq.confighash"]
	if !ok {
		log.Debug().Str("name", name).Msg("容器没有配置哈希的标签,需要重建")
		return true
	}
	// 如果不相等也要重建
	if h != configHash {
		log.Debug().Str("name", name).
			Str("old", h).
			Str("new", configHash).
			Msg("容器配置哈希不相等,需要重建")
		return true
	}
	return false
}

func GetContainerByName(name string) (types.ContainerJSON, error) {
	ctx := context.Background()
	container, err := config.DockerCli.ContainerInspect(ctx, name)
	if err != nil {
		return types.ContainerJSON{}, err
	}
	return container, nil
}

func GetContainerLabelByNameAndKey(name, key string) (string, error) {
	ctx := context.Background()
	container, err := config.DockerCli.ContainerInspect(ctx, name)
	if err != nil {
		return "", err
	}
	labels := container.Config.Labels
	if labels == nil {
		return "", nil
	}
	value, ok := labels[key]
	if !ok {
		return "", nil
	}
	return value, nil
}
