package metrics

import (
	"time"

	"go.mongodb.org/mongo-driver/mongo/options"
)

type Config struct {
	targetLatency                    time.Duration          // Desired latency target
	windowDuration                   time.Duration          // Time window for aggregating latency
	initialWorkerCount               int                    // Initial number of workers
	runDuration                      time.Duration          // Duration to run the test
	maxWorkers                       int32                  // Maximum number of workers allowed
	experimentTimeout                *time.Duration         // Deadline applied to context for experiment queries
	experimentClientOpts             *options.ClientOptions // Additional options to apply to client used for experiment
	experimentPoolMonitorCallback    PoolMonitorCallback    // Override the pool monitor
	experimentCommandMonitorCallback CommandMonitorCallback // Override the command monitor
	preloadCollectionSize            int
}

type ConfigOpt func(*Config)

// WithTargetLatency sets the target operation latency before starting the
// experiment. The deafult is 1ms.
func WithTargetLatency(latency time.Duration) ConfigOpt {
	return func(cfg *Config) {
		cfg.targetLatency = latency
	}
}

// WithWindowDuration is the time window for aggregating latency. I.e. "check
// every {duration} if we are at the target latency".
func WithWindowDuration(duration time.Duration) ConfigOpt {
	return func(cfg *Config) {
		cfg.windowDuration = duration
	}
}

// WithRunDuration sets how long to run the experiment. The default is 1
// minute.
func WithRunDuration(duration time.Duration) ConfigOpt {
	return func(cfg *Config) {
		cfg.runDuration = duration
	}
}

// WithMaxWorkers sets the maximum number of workers allowed to generate
// latency.
func WithMaxWorkers(workers int32) ConfigOpt {
	return func(cfg *Config) {
		cfg.maxWorkers = workers
	}
}

// WithExperimentTimeout applies a context deadline to the context passed to
// the experimental function.
func WithExperimentTimeout(timeout time.Duration) ConfigOpt {
	return func(cfg *Config) {
		cfg.experimentTimeout = ptr(timeout)
	}
}

// WithExpPoolMonitorCallback will override the pool monitor in the experiment.
func WithExpPoolMonitorCallback(cb PoolMonitorCallback) ConfigOpt {
	return func(cfg *Config) {
		cfg.experimentPoolMonitorCallback = cb
	}
}

// WithExpCommandMonitorCallback will override the command monitor in the
// experiment.
func WithExpCommandMonitorCallback(cb CommandMonitorCallback) ConfigOpt {
	return func(cfg *Config) {
		cfg.experimentCommandMonitorCallback = cb
	}
}

// WithInitialWorkerCount is how many workers to start rnning operation latency
// with. The default is 1.
func WithInitialWorkerCount(count int) ConfigOpt {
	return func(cfg *Config) {
		cfg.initialWorkerCount = count
	}
}

// WithExperimentClientOptions will extend the client options for the client
// used to run the experiment.
func WithExperimentClientOptions(opts *options.ClientOptions) ConfigOpt {
	return func(cfg *Config) {
		cfg.experimentClientOpts = opts
	}
}

// WithPreloadCollectionSize sets the size of the pre-loaded collection.
func WithPreloadCollectionSize(size int) ConfigOpt {
	return func(cfg *Config) {
		cfg.preloadCollectionSize = size
	}
}

// Ptr will return the memory location of the given value.
func ptr[T any](val T) *T {
	return &val
}

func NewConfig(opts ...ConfigOpt) Config {
	cfg := Config{
		runDuration:           1 * time.Minute,
		maxWorkers:            2000,
		initialWorkerCount:    5,
		targetLatency:         1,
		windowDuration:        100 * time.Millisecond,
		preloadCollectionSize: 1,
	}

	return cfg
}
