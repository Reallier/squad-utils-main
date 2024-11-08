package agent

import (
	"context"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"squad-utils/config"
	"squad-utils/handler"
)

var Cmd = &cobra.Command{
	Use:   "agent",
	Short: "开始侦听容器,并启动对应附属",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		viper.AddConfigPath(".")
		log.Debug().Str("path", viper.GetString("config")).Msg("配置文件路径")
		//viper.SetConfigName(viper.GetString("config"))
		viper.SetConfigFile(viper.GetString("config"))
		err := viper.ReadInConfig()
		if err != nil {
			// 判断错误类型
			if _, ok := err.(viper.ConfigFileNotFoundError); ok {
				log.Fatal().Msg("兄弟,配置文件呢?")
				return
			}
			log.Fatal().Err(err).Msg("配置文件读取失败")
			return
		}
		config.SQConfig = &config.SQConfigStruct{}
		err = viper.Unmarshal(config.SQConfig)
		if err != nil {
			log.Fatal().Err(err).Msg("配置文件解析失败")
			return
		}
		// 设置日志等级
		level, err := zerolog.ParseLevel(config.SQConfig.LogLevel)
		if err != nil {
			log.Fatal().Err(err).Msg("日志级别解析失败")
			return
		}
		log.Info().Str("level", config.SQConfig.LogLevel).Msg("已设定日志等级")
		zerolog.SetGlobalLevel(level)
		// 事前检查部分
		err = PreflightCheck()
		if err != nil {
			log.Fatal().Err(err).Msg("配置文件检查失败")
			return
		}
		// 连接容器部分
		err = ConnectDocker()
		if err != nil {
			log.Fatal().Err(err).Msg("连接Docker失败")
		}
		CreateHandler()
	},
}

func Execute() error {
	if err := Cmd.Execute(); err != nil {
		return err
	}
	return nil
}

func init() {
	var configPath string
	Cmd.Flags().StringVarP(&configPath, "config", "c", "config.yml", "配置文件路径")
	viper.BindPFlag("config", Cmd.Flags().Lookup("config"))
}

func ConnectDocker() error {
	cli, err := client.NewClientWithOpts(
		client.FromEnv, client.WithAPIVersionNegotiation(), client.WithHost(config.SQConfig.Socket),
	)
	// 传递过去
	config.DockerCli = cli
	if err != nil {
		return err
	}
	ctx := context.Background()
	options := types.EventsOptions{
		Filters: filters.NewArgs(filters.Arg("type", "container")),
	}

	config.MsgCh, config.ErrCh = cli.Events(ctx, options)
	log.Info().Str("path", config.SQConfig.Socket).Msg("Docker 连接成功,事件通道已建立")
	return nil
}
func CreateHandler() {
	handler.NewHandler()
}
