package latency

import (
	"time"

	"go.mongodb.org/mongo-driver/mongo/options"
)

type config struct {
	targetLatency      time.Duration          // Desired latency target
	windowDuration     time.Duration          // Time window for aggregating latency
	runDuration        time.Duration          // Duration to run the test
	maxWorkers         int32                  // Maximum number of workers allowed
	experimentTimeout  time.Duration          // Duration for experiment queries
	initialWorkerCount int                    // Initial number of workers
	clientOpts         *options.ClientOptions // Additional options to apply to client used for experiment
}

type configOpt func(*config)

func withTargetLatency(latency time.Duration) configOpt {
	return func(cfg *config) {
		cfg.targetLatency = latency
	}
}

func withWindowDuration(duration time.Duration) configOpt {
	return func(cfg *config) {
		cfg.windowDuration = duration
	}
}

func withRunDuration(duration time.Duration) configOpt {
	return func(cfg *config) {
		cfg.runDuration = duration
	}
}

func withMaxWorkers(workers int32) configOpt {
	return func(cfg *config) {
		cfg.maxWorkers = workers
	}
}

func withExperimentTimeout(timeout time.Duration) configOpt {
	return func(cfg *config) {
		cfg.experimentTimeout = timeout
	}
}

func withInitialWorkerCount(count int) configOpt {
	return func(cfg *config) {
		cfg.initialWorkerCount = count
	}
}

func withClientOptions(opts *options.ClientOptions) configOpt {
	return func(cfg *config) {
		cfg.clientOpts = opts
	}
}

func newConfig(opts ...configOpt) config {
	cfg := config{
		targetLatency:      1 * time.Millisecond,
		windowDuration:     10 * time.Second,
		runDuration:        1 * time.Minute,
		maxWorkers:         2000,
		experimentTimeout:  50 * time.Millisecond,
		initialWorkerCount: 20,
	}

	return cfg
}
