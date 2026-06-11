package metrics

// Metrics represents a snapshot of system metrics.
type Metrics struct {
	CPU        float64 `json:"cpu"`
	RAMUsed    uint64  `json:"ram_used"`
	RAMTotal   uint64  `json:"ram_total"`
	DiskUsed   uint64  `json:"disk_used"`
	DiskTotal  uint64  `json:"disk_total"`
	LoadAvg    [3]float64 `json:"load_avg"`
	NetworkRx  uint64  `json:"network_rx"`
	NetworkTx  uint64  `json:"network_tx"`
	Processes  int     `json:"processes"`
	Uptime     uint64  `json:"uptime"`
}

// Process represents a single process.
type Process struct {
	PID       int     `json:"pid"`
	Name      string  `json:"name"`
	CPU       float64 `json:"cpu"`
	RAM       uint64  `json:"ram"`
	Status    string  `json:"status"`
	Command   string  `json:"command"`
}
