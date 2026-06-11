//go:build !linux

package metrics

// Collector reads system metrics from /proc.
type Collector struct{}

// NewCollector creates a new metrics collector.
func NewCollector() *Collector {
	return &Collector{}
}

// Collect gathers current system metrics.
func (c *Collector) Collect() (*Metrics, error) {
	return &Metrics{}, nil
}
