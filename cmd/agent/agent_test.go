package agent

import (
	"github.com/stretchr/testify/assert"
	"squad-utils/config"
	"testing"
)

func TestConnectToDockerSocket(t *testing.T) {
	config.SQConfig = &config.SQConfigStruct{
		Socket: "unix:///var/run/docker.sock",
	}
	err := ConnectDocker()
	assert.Nil(t, err)
}
