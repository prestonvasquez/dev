package main

import (
	"fmt"
	"sync/atomic"
)

type deployment interface{}

type topology struct {
	currentDriverInfo *atomic.Value
}

var _ deployment = (*topology)(nil)

type driverInfo struct {
	name string
}

type client struct {
	deployemnt        deployment // Masks topology
	currentDriverInfo *atomic.Value
}

func (c *client) appendDriverInfo(info driverInfo) {
	c.currentDriverInfo.Store(info)
}

func main() {
	c := &client{
		currentDriverInfo: &atomic.Value{},
	}

	t := &topology{
		currentDriverInfo: c.currentDriverInfo,
	}

	c.deployemnt = t

	// Add a new info and see if it's reflected in the topology.
	c.appendDriverInfo(driverInfo{name: "new-info"})
	fmt.Println("client driver info name:", c.currentDriverInfo.Load().(driverInfo).name)

	top := c.deployemnt.(*topology)
	tDriverInfo := top.currentDriverInfo.Load().(driverInfo)

	fmt.Println("topology driver info name:", tDriverInfo.name)

	// Try again
	c.appendDriverInfo(driverInfo{name: "newer-info"})
	fmt.Println("client driver info name:", c.currentDriverInfo.Load().(driverInfo).name)
	fmt.Println("topology driver info name:", top.currentDriverInfo.Load().(driverInfo).name)

}
