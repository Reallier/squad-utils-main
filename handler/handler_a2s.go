package handler

import (
	"context"
	"crypto/sha256"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/events"
	"github.com/docker/docker/client"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"io"
	"os"
	"path"
	"squad-utils/config"
)

type A2SHandler struct {
	// Image 所用镜像
	Image string
	// 容器名字,格式为 {UUID}-a2s-server
	Name string
	// Env 启动环境变量
	Env map[string]string
	// ConfigHash 配置哈希
	// 计算方法为配置文件的 SHA256
	ConfigHash string
	// ContainerID 将会在初始化的时候记录 A2S 容器的 ID
	// 如果为空就意味着没有一个匹配的容器启动,此时直接删除会跳过
	// 任何更新容器后都必须更新此值 (即创建后更新他)
	// 删除容器后必须将此值置空
	ContainerID string
	// UUID 记录翼龙服务器的 UUID
	// 格式是带 - 分割的那种
	UUID string
	// 记录此次容器事件共后续共享使用
	Event          events.Message
	ReCreate       bool
	ConfigFilePath string
}

func NewA2SHandlerFromEvent(event events.Message) (*A2SHandler, error) {
	//ctx := context.Background()
	handler := &A2SHandler{
		Image: config.SQConfig.StartupImages.A2SServer,
		Env:   make(map[string]string),
		Event: event,
		UUID:  event.Actor.Attributes["name"],
	}
	// 这一步可以检查 A2S 容器在不在
	a2sContainerName := fmt.Sprintf("%s-a2s-server", handler.UUID)
	handler.Name = a2sContainerName
	log.Debug().Str("name", handler.Name).Msg("期望的 A2S 容器名字")
	a2sContainer, err := GetContainerByName(handler.Name)
	if err != nil {
		// 发生一个错误,看看是不是不存在
		if client.IsErrNotFound(err) {
			log.Info().Err(err).Str("name", handler.Name).Msg("A2S 容器不存在,标记重建")
			handler.ReCreate = true
		} else {
			// 其他错误,退出
			log.Err(err).Str("name", handler.Name).Msg("获取 A2S 容器信息失败")
			return nil, err
		}
	} else {
		// 无错误,继续
		handler.ContainerID = a2sContainer.ID
	}
	// 例如 /var/lib/pterodactyl/volumes/{UUID}/Steam/config/steam.yaml
	handler.ConfigFilePath = path.Join(config.SQConfig.Volumes.Pter, handler.UUID, "Steam", "config", "steam.yaml")
	// 先看文件在不在
	if _, err := os.Stat(handler.ConfigFilePath); errors.Is(err, os.ErrNotExist) {
		// 不存在应当报错退出
		log.Err(err).Str("name", handler.Name).Msg("A2S 配置文件不存在")
		return nil, err
	}
	// 读取文件计算 SHA256
	// 用于后续判断是否需要重建容器
	f, err := os.Open(handler.ConfigFilePath)
	if err != nil {
		log.Err(err).Str("name", handler.Name).Msg("A2S 配置文件打开出错")
		return nil, err
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		log.Err(err).Str("name", handler.Name).Msg("A2S 配置文件哈希计算出错")
		return nil, err
	}
	handler.ConfigHash = fmt.Sprintf("%x", h.Sum(nil))
	return handler, nil
}

// Deploy 部署一套容器,包括全部检测流程
func (a *A2SHandler) Deploy() error {
	log.Info().Str("name", a.Name).Msg("开始部署容器")
	log.Debug().Str("name", a.Name).Msg("检测是否需要重建")
	needRecreate := NeedRecreate(a.Name, a.ConfigHash)
	log.Info().Str("name", a.Name).Bool("need", needRecreate).Msg("需要重建吗")
	if !needRecreate {
		log.Info().Str("name", a.Name).Msg("不需要重建,只需重新启动一下")
		a.Restart()
		return nil
	}
	log.Debug().Str("name", a.Name).Msg("需要重建,开始干活")
	// 启动之前要 Pull 一下新镜像
	err := a.Pull(false)
	if err != nil {
		log.Err(err).Str("name", a.Name).Msg("拉取镜像失败,且本地也没有镜像")
		return err
	}
	log.Debug().Str("name", a.Name).Msg("删除旧容器")
	// 删除旧容器
	a.Remove()
	// 创建与启动
	err = a.Create()
	if err != nil {
		log.Err(err).Str("name", a.Name).Msg("创建容器失败")
		return err
	}
	a.Start()
	return nil
}
func (a *A2SHandler) Pull(force bool) error {
	ctx := context.Background()
	log.Info().Str("name", a.Name).Msg("检查镜像更新")
	pull, err := config.DockerCli.ImagePull(ctx, config.SQConfig.StartupImages.A2SServer, types.ImagePullOptions{})
	defer pull.Close()
	// 读取输出内容并打印到控制台
	buf := make([]byte, 4096)
	for {
		log.Debug().Str("name", a.Name).Msg("等待镜像拉取中")
		_, err := pull.Read(buf)
		if err != nil && err.Error() == "EOF" {
			break
		}
		//fmt.Print(string(buf[0:n]))
		log.Debug().Str("name", a.Name).Msg(string(buf))
	}
	if err != nil {
		// 看看本地是否有镜像
		if HasImage(a.Image) {
			log.Warn().Str("name", a.Name).Msg("拉取失败,但是本地有镜像,继续启动")
			// 本地有镜像就不报错
			return nil
		} else {
			log.Err(err).Str("name", a.Name).Msg("拉取失败,且本地也无镜像")
			return err
		}
	}
	log.Info().Str("name", a.Name).Msg("镜像更新完成")
	return nil
}

// Create 创建容器
func (a *A2SHandler) Create() error {
	ctx := context.Background()

	log.Info().Str("name", a.Name).Msg("现在创建容器")
	mount := fmt.Sprintf("%s:%s", config.SQConfig.Volumes.Pter, config.SQConfig.Volumes.Pter)
	volumeConfig := map[string]struct{}{}
	volumeConfig[mount] = struct{}{}
	hostConfig := &container.HostConfig{
		Binds:      []string{mount},
		AutoRemove: false,
		RestartPolicy: container.RestartPolicy{
			Name: "always",
		},
		NetworkMode: "host",
	}
	containerConfig := &container.Config{
		Volumes: volumeConfig,
		Env:     nil,
		Image:   config.SQConfig.StartupImages.A2SServer,
		Labels: map[string]string{
			"sq.confighash": a.ConfigHash,
		},
		Cmd: []string{"-c", a.ConfigFilePath},
	}
	create, err := config.DockerCli.ContainerCreate(
		ctx, containerConfig, hostConfig, nil, nil, a.Name,
	)
	if err != nil {
		log.Err(err).Str("name", a.Name).Msg("创建容器失败")
		return err
	}
	a.ContainerID = create.ID
	return nil
}

// Remove 删除容器
func (a *A2SHandler) Remove() error {
	log.Info().Str("name", a.Name).Str("id", a.ContainerID).Msg("现在删除容器")
	log.Debug().Str("name", a.Name).Str("id", a.ContainerID).Msg("先停止容器")
	a.Stop()
	ctx := context.Background()
	err := config.DockerCli.ContainerRemove(ctx, a.Name, types.ContainerRemoveOptions{})
	if err != nil {
		return err
	}
	log.Info().Str("name", a.Name).Msg("已删除容器")
	return nil
}

// Start 仅仅是启动容器,不在乎容器存在与否
func (a *A2SHandler) Start() error {
	log.Info().Str("name", a.Name).Str("id", a.ContainerID).Msg("现在启动容器")
	ctx := context.Background()
	if err := config.DockerCli.ContainerStart(ctx, a.ContainerID, types.ContainerStartOptions{}); err != nil {
		log.Err(err).Str("name", a.Name).Msg("启动容器失败")
		return err
	}
	return nil
}

func (a *A2SHandler) Restart() {
	log.Info().Str("name", a.Name).Msg("现在重启容器")
	ctx := context.Background()
	config.DockerCli.ContainerRestart(ctx, a.ContainerID, container.StopOptions{})
}

func (a *A2SHandler) Stop() error {
	log.Info().Str("name", a.Name).Msg("现在停止容器")
	ctx := context.Background()
	err := config.DockerCli.ContainerStop(ctx, a.ContainerID, container.StopOptions{})
	if err != nil {
		log.Err(err).Str("name", a.Name).Msg("停止容器失败")
		return err
	}
	return nil
}
