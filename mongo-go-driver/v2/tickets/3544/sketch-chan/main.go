package main

import (
	"fmt"
)

type deployment interface{}

type topology struct {
	currentDriverInfoCh chan driverInfo
}

var _ deployment = (*topology)(nil)

type driverInfo struct {
	name string
}

type client struct {
	deployemnt          deployment // Masks topology
	currentDriverInfoCh chan driverInfo
}

func (c *client) appendDriverInfo(info driverInfo) {
	c.currentDriverInfoCh <- info
}

func main() {
	c := &client{
		currentDriverInfoCh: make(chan driverInfo, 1),
	}

	t := &topology{
		currentDriverInfoCh: c.currentDriverInfoCh,
	}

	c.deployemnt = t

	// Add a new info and see if it's reflected in the topology.
	c.appendDriverInfo(driverInfo{name: "new-info"})
	fmt.Println("client driver info name:", (<-t.currentDriverInfoCh).name)

	c.appendDriverInfo(driverInfo{name: "newer-info"})
	fmt.Println("client driver info name:", (<-t.currentDriverInfoCh).name)

	// Two in a row
	c.appendDriverInfo(driverInfo{name: "newest-info"})
	c.appendDriverInfo(driverInfo{name: "newest-info-2"})

	fmt.Println("client driver info name:", (<-t.currentDriverInfoCh).name)
}
