package main

import (
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/google/uuid"
)

type deployment interface{}

type topology struct {
	currentDriverInfo *atomic.Pointer[driverInfo]
}

var _ deployment = (*topology)(nil)

type driverInfo struct {
	name string
}

type client struct {
	deployemnt        deployment // Masks topology
	currentDriverInfo *atomic.Pointer[driverInfo]
}

func (c *client) appendDriverInfo(info driverInfo) {
	copy := new(driverInfo)
	*copy = info

	c.currentDriverInfo.Store(copy)
}

func main() {
	c := &client{
		currentDriverInfo: &atomic.Pointer[driverInfo]{},
	}

	t := &topology{
		currentDriverInfo: c.currentDriverInfo,
	}

	c.deployemnt = t

	// Add a new info and see if it's reflected in the topology.
	c.appendDriverInfo(driverInfo{name: "new-info"})
	fmt.Println("client driver info name:", c.currentDriverInfo.Load().name)

	top := c.deployemnt.(*topology)
	tDriverInfo := top.currentDriverInfo.Load()

	fmt.Println("topology driver info name:", tDriverInfo.name)

	// Two in a row
	c.appendDriverInfo(driverInfo{name: "newest-info"})
	c.appendDriverInfo(driverInfo{name: "newest-info-2"})

	fmt.Println("client driver info name:", c.currentDriverInfo.Load().name)

	// A bunch concurrently
	wg := sync.WaitGroup{}
	wg.Add(1000)
	for i := 0; i < 1000; i++ {
		go func() {
			defer wg.Done()

			c.appendDriverInfo(driverInfo{name: "concurrent-info" + " " + uuid.New().String()})
		}()
	}

	wg.Wait()

	fmt.Println("client driver info name:", c.currentDriverInfo.Load().name)
}
