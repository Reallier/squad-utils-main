package config

import (
	"github.com/docker/docker/api/types/events"
	"github.com/docker/docker/client"
)

var DockerCli *client.Client

var MsgCh <-chan events.Message
var ErrCh <-chan error
