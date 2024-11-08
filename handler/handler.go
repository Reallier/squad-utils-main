package handler

import (
	"github.com/docker/docker/api/types/events"
	"github.com/rs/zerolog/log"
	"squad-utils/config"
)

func NewHandler() {
	log.Info().Msg("现在启动事件监听")
	for {
		select {
		case event := <-config.MsgCh:
			if event.Type == events.ContainerEventType {
				switch event.Action {
				case "start":
					log.Debug().Msg("有容器启动")
					// 获取镜像
					if event.Actor.Attributes["image"] == config.SQConfig.Image {
						log.Debug().Msg("检测到容器启动事件,且镜像符合")
						go DeployContainers(event)
					}
				case "stop":
					log.Debug().Msg("有容器停止")
					if event.Actor.Attributes["image"] == config.SQConfig.Image {
						log.Debug().Msg("检测到容器停止事件,且镜像符合")
						go StopImages(event)
					}
				case "destroy":
					log.Debug().Msg("有容器删除")
					if event.Actor.Attributes["image"] == config.SQConfig.Image {
						log.Debug().Msg("检测到容器删除事件,且镜像符合")
						go DestroyContainers(event)
					}
				}
			}
		}
	}

}

func DestroyContainers(event events.Message) {
	containerUUID := event.Actor.Attributes["name"]
	log.Debug().Str("containerUUID", containerUUID).Msg("检测到翼龙删除了这个镜像")
	errs := make(chan error, 1)
	go func() {
		a2sHandler, err := NewA2SHandlerFromEvent(event)
		if err != nil {
			log.Err(err).Msg("创建 A2SHandler 失败")
			return
		}
		log.Trace().Msg("创建了 A2SHandler")
		go func() {
			errs <- a2sHandler.Remove()
		}()
	}()
	if err := <-errs; err != nil {
		log.Err(err).Msg("删除 A2S 容器失败")
	}

	errs = make(chan error, 1)
	go func() {
		promtailHandler, err := NewPromtailHandlerFromEvent(event)
		if err != nil {
			log.Err(err).Msg("创建 PromtailHandler 失败")
			return
		}
		log.Trace().Msg("创建了 PromtailHandler")
		go func() {
			errs <- promtailHandler.Remove()
		}()
	}()
}

func StopImages(event events.Message) {
	containerUUID := event.Actor.Attributes["name"]
	log.Debug().Str("containerUUID", containerUUID).Msg("检测到翼龙停止了这个镜像")
	errs := make(chan error, 1)
	go func() {
		a2sHandler, err := NewA2SHandlerFromEvent(event)
		if err != nil {
			log.Err(err).Msg("创建 A2SHandler 失败")
			return
		}
		log.Trace().Msg("创建了 A2SHandler")
		go func() {
			errs <- a2sHandler.Stop()
		}()
	}()
	if err := <-errs; err != nil {
		log.Err(err).Msg("停止 A2S 容器失败")
	}
	errs = make(chan error, 1)
	go func() {
		promtailHandler, err := NewPromtailHandlerFromEvent(event)
		if err != nil {
			log.Err(err).Msg("创建 PromtailHandler 失败")
			return
		}
		log.Trace().Msg("创建了 PromtailHandler")
		go func() {
			errs <- promtailHandler.Stop()
		}()
	}()
	if err := <-errs; err != nil {
		log.Err(err).Msg("停止 Promtail 容器失败")
	}

}
func DeployContainers(event events.Message) {
	containerUUID := event.Actor.Attributes["name"]
	log.Debug().Str("containerUUID", containerUUID).Msg("检测到翼龙启动了这个镜像")
	errs := make(chan error, 1)
	go func() {
		a2sHandler, err := NewA2SHandlerFromEvent(event)
		if err != nil {
			log.Err(err).Msg("创建 A2SHandler 失败")
			return
		}
		log.Trace().Msg("创建了 A2SHandler")
		go func() {
			errs <- a2sHandler.Deploy()
		}()
	}()
	if err := <-errs; err != nil {
		log.Err(err).Msg("启动 A2S 容器失败")
	}
	errs = make(chan error, 1)
	go func() {
		promtailHandler, err := NewPromtailHandlerFromEvent(event)
		if err != nil {
			log.Err(err).Msg("创建 PromtailHandler 失败")
			return
		}
		log.Trace().Msg("创建了 PromtailHandler")
		go func() {
			errs <- promtailHandler.Deploy()
		}()
	}()
	if err := <-errs; err != nil {
		log.Err(err).Msg("启动 Promtail 容器失败")
	}
}
