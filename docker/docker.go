package docker

import (
	"context"
	"fmt"
	"sync"

	"github.com/docker/docker/api/types/events"
	"github.com/docker/docker/client"
)

type containerInfo struct {
	Port      int
	IpAddress string
}

var mutex sync.RWMutex
var containerCache map[string]containerInfo

func put(name string, payload containerInfo) {
	mutex.Lock()
	defer mutex.Unlock()
	containerCache[name] = payload
}

func Get(name string) (containerInfo, bool) {
	mutex.RLock()
	defer mutex.RUnlock()

	v, ok := containerCache[name]
	return v, ok
}

func SaveContainerData(cli *client.Client) {
	eventsCh, errCh := cli.Events(context.Background(), events.ListOptions{})

	for {
		select {
		case val := <-eventsCh:
			if val.Action == events.ActionStart && val.Type == events.ContainerEventType {
				info, err := cli.ContainerInspect(context.Background(), val.Actor.ID)
				var port int

				for k := range info.Config.ExposedPorts {
					if k.Proto() == "tcp" {
						port = k.Int()
					}
				}

				if err == nil {
					put(info.Name[1:], containerInfo{Port: port, IpAddress: info.NetworkSettings.IPAddress})
				}
			}

		case err := <-errCh:
			fmt.Printf("%+v Error getting event", err)
		}
	}
}
