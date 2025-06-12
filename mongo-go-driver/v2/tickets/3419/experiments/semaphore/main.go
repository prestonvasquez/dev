package main

import (
	"fmt"
	"sync"
	"time"
)

type conn struct {
	id int
}

type wantConn struct {
	ready chan *conn
}

type pool struct {
	mu     sync.Mutex    // protects conns and queues
	sem    chan struct{} // semaphore for maxConnecting
	conns  map[int]*conn // active connections
	queue  []*wantConn   // pending checkout requests
	nextID int           // for Conn IDs
}

func newPool(maxConnecting int) *pool {
	return &pool{
		sem:   make(chan struct{}, maxConnecting),
		conns: make(map[int]*conn),
		queue: make([]*wantConn, 0),
	}
}

func (p *pool) spawnConnection(w *wantConn) {
	// Relase slot when done.
	defer func() { <-p.sem }()

	// Simulate connection setup delay
	time.Sleep(1000 * time.Millisecond)

	// Register new conenction
	p.mu.Lock()
	p.nextID++
	conn := &conn{id: p.nextID}
	p.conns[conn.id] = conn
	p.mu.Unlock()

	w.ready <- conn // Notify the waiting goroutine that the connection is ready
}

// spawnConnectionIfNeeded takes on waiting waitConn (if any) and starts its
// connection creation subject to the semaphore limit.
func (p *pool) spawnConnectionIfNeeded() {
	// Pop a pending request from teh queue.
	p.mu.Lock()
	if len(p.queue) == 0 {
		p.mu.Unlock()
		return
	}

	w := p.queue[0]
	p.queue = p.queue[1:]
	p.mu.Unlock()

	// Aquire a semaphore slot (blocks if maxConnecting is reached).
	p.sem <- struct{}{}

	// Spawn the actual connection creator
	go p.spawnConnection(w)
}

func (p *pool) queueForNewConn(w *wantConn) {
	p.mu.Lock()
	p.queue = append(p.queue, w)
	p.mu.Unlock()

	// Try to spawn without blocking the caller.
	go p.spawnConnectionIfNeeded()
}

func main() {
	pool := newPool(2)

	wg := sync.WaitGroup{}
	wg.Add(5)

	// simluate 5 concurrent connections with maxConnecting = 2
	for i := 0; i < 5; i++ {
		w := &wantConn{ready: make(chan *conn)}
		pool.queueForNewConn(w)

		// Wait for the connection to be ready
		go func() {
			<-w.ready // Ensure the channel is closed after use
			wg.Done()

			fmt.Printf("Connection %d is ready\n", w.ready)
		}()
	}

	wg.Wait() // Wait for all connections to be ready
}
