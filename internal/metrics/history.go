package metrics

import (
	"sync"
	"time"
)

type HistoryPoint struct {
	Timestamp time.Time `json:"timestamp"`
	CPU       float64   `json:"cpu"`
	RAMUsed   uint64    `json:"ram_used"`
	RAMTotal  uint64    `json:"ram_total"`
	DiskUsed  uint64    `json:"disk_used"`
	DiskTotal uint64    `json:"disk_total"`
	NetworkIn uint64    `json:"network_in"`
	NetworkOut uint64   `json:"network_out"`
}

type HistoryStore struct {
	mu      sync.RWMutex
	points  []HistoryPoint
	maxLen  int
}

var store = &HistoryStore{
	points: make([]HistoryPoint, 0, 43200),
	maxLen: 43200,
}

func RecordSnapshot(m *Metrics) {
	store.mu.Lock()
	defer store.mu.Unlock()
	store.points = append(store.points, HistoryPoint{
		Timestamp: time.Now(),
		CPU:       m.CPU,
		RAMUsed:   m.RAMUsed,
		RAMTotal:  m.RAMTotal,
		DiskUsed:  m.DiskUsed,
		DiskTotal: m.DiskTotal,
		NetworkIn: m.NetworkRx,
		NetworkOut: m.NetworkTx,
	})
	if len(store.points) > store.maxLen {
		store.points = store.points[len(store.points)-store.maxLen:]
	}
}

func GetHistory(rangeStr string) []HistoryPoint {
	store.mu.RLock()
	defer store.mu.RUnlock()

	var cutoff time.Time
	now := time.Now()
	switch rangeStr {
	case "1h":
		cutoff = now.Add(-1 * time.Hour)
	case "6h":
		cutoff = now.Add(-6 * time.Hour)
	case "24h":
		cutoff = now.Add(-24 * time.Hour)
	case "7d":
		cutoff = now.Add(-7 * 24 * time.Hour)
	default:
		cutoff = now.Add(-1 * time.Hour)
	}

	result := make([]HistoryPoint, 0)
	for _, p := range store.points {
		if p.Timestamp.After(cutoff) {
			result = append(result, p)
		}
	}
	return result
}