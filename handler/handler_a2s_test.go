package handler

import (
	"context"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
	"github.com/stretchr/testify/assert"
	"squad-utils/config"
	"testing"
)

func TestPullImage(t *testing.T) {
	cli, _ := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	ctx := context.Background()
	//pull, err := cli.ImagePull(ctx, "ccr.ccs.tencentyun.com/karasu/stck:a3s", types.ImagePullOptions{})
	pull, err := cli.ImagePull(ctx, "ubuntu", types.ImagePullOptions{})
	buf := make([]byte, 4096)
	for {
		n, err := pull.Read(buf)
		if err != nil && err.Error() == "EOF" {
			break
		}
		fmt.Print(string(buf[0:n]))
	}
	if err != nil {
		panic(err)
	}
	defer pull.Close()
}

func TestImageInspect(t *testing.T) {
	cli, _ := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	ctx := context.Background()
	//inspect, err := cli.ImageInspect(ctx, "ccr.ccs.tencentyun.com/karasu/stck:a3s")
	inspect, _, err := cli.ImageInspectWithRaw(ctx, "ubuntu")
	if err != nil {
		fmt.Println("是找不到吗", client.IsErrNotFound(err))
		fmt.Println(err)
	}
	fmt.Println(inspect)
}

func TestContainerInspectByName(t *testing.T) {
	cli, _ := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	ctx := context.Background()
	//inspect, err := cli.ContainerInspectByName(ctx, "ccr.ccs.tencentyun.com/karasu/stck:a3s")
	container, err := cli.ContainerInspect(ctx, "ubuntu")
	if err != nil {
		fmt.Println("是找不到吗", client.IsErrNotFound(err))
		fmt.Println(err.Error())
	}
	fmt.Println(container.Image)
}
func TestInspectContainerByID(t *testing.T) {
	cli, _ := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	ctx := context.Background()
	//containers, err := cli.ContainerList(ctx, types.ContainerListOptions{})
	container, _ := cli.ContainerInspect(ctx, "4c5f212ae08d996788547f741a32dc65fe5c6a01db8947501d1e868a1d33f8b7")
	// 去掉 name 开头的斜杠
	container.Name = container.Name[1:]
	fmt.Println(container.Name)

}

func TestListenContainerStart(t *testing.T) {
	cli, _ := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	ctx := context.Background()
	options := types.EventsOptions{
		Filters: filters.NewArgs(filters.Arg("type", "container")),
	}
	config.MsgCh, config.ErrCh = cli.Events(ctx, options)
	for {
		select {
		case event := <-config.MsgCh:
			fmt.Println(event.Actor.Attributes["name"])
		}
	}
}

func TestGetContainerLabel(t *testing.T) {
	cli, _ := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	ctx := context.Background()
	// 先创建一个容器
	container, err := cli.ContainerCreate(
		ctx, &container.Config{
			Volumes: map[string]struct{}{
				"/data": {},
			},
			Env:   nil,
			Image: "fedora",
			Labels: map[string]string{
				"test": "test",
			},
		}, nil, nil, nil, "scratch",
	)
	if err != nil {
		panic(err)
	}
	defer cli.ContainerRemove(ctx, container.ID, types.ContainerRemoveOptions{})
	// 获取容器的label
	label, err := cli.ContainerInspect(ctx, container.ID)
	if err != nil {
		panic(err)
	}
	assert.Equal(t, label.Config.Labels["test"], "test")
	fmt.Println(label)

}
