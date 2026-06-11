package alerts

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"juvia/internal/db"
	"juvia/internal/events"
	"juvia/internal/metrics"
)

type Engine struct {
	db       *db.DB
	log      *slog.Logger
	bus      *events.Bus
	metrics  *metrics.Collector
	interval time.Duration
	stopCh   chan struct{}
}

func NewEngine(database *db.DB, log *slog.Logger, bus *events.Bus, m *metrics.Collector) *Engine {
	return &Engine{
		db:       database,
		log:      log,
		bus:      bus,
		metrics:  m,
		interval: 60 * time.Second,
		stopCh:   make(chan struct{}),
	}
}

func (e *Engine) Start(ctx context.Context) {
	e.log.Info("alert engine started")
	ticker := time.NewTicker(e.interval)
	defer ticker.Stop()

	e.checkAll()

	for {
		select {
		case <-ticker.C:
			e.checkAll()
		case <-e.stopCh:
			e.log.Info("alert engine stopped")
			return
		case <-ctx.Done():
			return
		}
	}
}

func (e *Engine) Stop() {
	close(e.stopCh)
}

func (e *Engine) checkAll() {
	ctx := context.Background()

	e.checkDiskUsage(ctx)
	e.checkServices(ctx)
	e.checkWebsites(ctx)
	e.checkSSLExpiry(ctx)
	e.checkBackups(ctx)
}

func (e *Engine) checkDiskUsage(ctx context.Context) {
	enabled, _ := e.db.GetAlertSetting(ctx, "alert_disk_usage_enabled")
	if enabled != "true" {
		return
	}

	thresholdStr, _ := e.db.GetAlertSetting(ctx, "alert_disk_usage_threshold")
	threshold := 80
	if thresholdStr != "" {
		threshold, _ = strconv.Atoi(thresholdStr)
	}

	m, err := e.metrics.Collect()
	if err != nil {
		return
	}

	if m.DiskTotal == 0 {
		return
	}

	usage := float64(m.DiskUsed) / float64(m.DiskTotal) * 100

	if int(usage) >= threshold {
		recent, _ := e.db.GetRecentAlertByType(ctx, "disk_usage", time.Hour)
		if recent != nil {
			return
		}

		e.db.CreateAlert(ctx, "disk_usage", "warning",
			fmt.Sprintf("Disk usage is at %.1f%%", usage), nil, "")
		e.emitAlert("disk_usage", "warning", fmt.Sprintf("Disk usage is at %.1f%%", usage))
	}
}

func (e *Engine) checkServices(ctx context.Context) {
	enabled, _ := e.db.GetAlertSetting(ctx, "alert_service_down_enabled")
	if enabled != "true" {
		return
	}

	services := []string{"nginx", "mysql", "postfix", "dovecot"}

	for _, svc := range services {
		cmd := exec.Command("systemctl", "is-active", svc)
		output, _ := cmd.Output()
		status := strings.TrimSpace(string(output))

		if status != "active" {
			recent, _ := e.db.GetRecentAlertByType(ctx, "service_down:"+svc, time.Hour)
			if recent != nil {
				continue
			}

			e.db.CreateAlert(ctx, "service_down:"+svc, "critical",
				fmt.Sprintf("Service %s is not running", svc), nil, "service")
			e.emitAlert("service_down", "critical", fmt.Sprintf("Service %s is not running", svc))
		}
	}
}

func (e *Engine) checkWebsites(ctx context.Context) {
	enabled, _ := e.db.GetAlertSetting(ctx, "alert_website_down_enabled")
	if enabled != "true" {
		return
	}

	websites, err := e.db.ListWebsites(ctx, "", "", "created_at", "desc", 1, 100)
	if err != nil {
		return
	}

	for _, site := range websites.Websites {
		if site.Status != "active" {
			continue
		}

		if site.DocumentRoot == "" {
			continue
		}

		url := "http://" + site.Domain
		if site.SSLEnabled {
			url = "https://" + site.Domain
		}

		client := &http.Client{Timeout: 5 * time.Second}
		resp, err := client.Get(url)
		if err != nil {
			e.createWebsiteAlert(ctx, site.ID, "Website down: "+err.Error())
			continue
		}
		resp.Body.Close()

		if resp.StatusCode >= 500 {
			e.createWebsiteAlert(ctx, site.ID, fmt.Sprintf("Website returned HTTP %d", resp.StatusCode))
		}
	}
}

func (e *Engine) createWebsiteAlert(ctx context.Context, websiteID int64, message string) {
	recent, _ := e.db.GetRecentAlertByType(ctx, "website_down:"+strconv.FormatInt(websiteID, 10), time.Hour)
	if recent != nil {
		return
	}

	e.db.CreateAlert(ctx, "website_down:"+strconv.FormatInt(websiteID, 10), "critical", message, &websiteID, "website")
	e.emitAlert("website_down", "critical", message)
}

func (e *Engine) checkSSLExpiry(ctx context.Context) {
	enabled, _ := e.db.GetAlertSetting(ctx, "alert_ssl_expiry_enabled")
	if enabled != "true" {
		return
	}

	daysStr, _ := e.db.GetAlertSetting(ctx, "alert_ssl_expiry_days")
	days := 14
	if daysStr != "" {
		days, _ = strconv.Atoi(daysStr)
	}

	websites, err := e.db.ListWebsites(ctx, "", "", "created_at", "desc", 1, 100)
	if err != nil {
		return
	}

	expiryThreshold := time.Now().AddDate(0, 0, days)

	for _, site := range websites.Websites {
		if !site.SSLEnabled || site.SSLExpiry == nil {
			continue
		}

		if site.SSLExpiry.Before(expiryThreshold) {
			recent, _ := e.db.GetRecentAlertByType(ctx, "ssl_expiry:"+strconv.FormatInt(site.ID, 10), 24*time.Hour)
			if recent != nil {
				continue
			}

			msg := fmt.Sprintf("SSL certificate for %s expires on %s", site.Domain, site.SSLExpiry.Format("2006-01-02"))
			e.db.CreateAlert(ctx, "ssl_expiry:"+strconv.FormatInt(site.ID, 10), "warning", msg, &site.ID, "website")
			e.emitAlert("ssl_expiry", "warning", msg)
		}
	}
}

func (e *Engine) checkBackups(ctx context.Context) {
	enabled, _ := e.db.GetAlertSetting(ctx, "alert_backup_overdue_enabled")
	if enabled != "true" {
		return
	}

	daysStr, _ := e.db.GetAlertSetting(ctx, "alert_backup_overdue_days")
	days := 7
	if daysStr != "" {
		days, _ = strconv.Atoi(daysStr)
	}

	websites, err := e.db.ListWebsites(ctx, "", "", "created_at", "desc", 1, 100)
	if err != nil {
		return
	}

	for _, site := range websites.Websites {
		if site.Status != "active" {
			continue
		}

		backup, err := e.db.GetLastBackupForWebsite(ctx, site.ID)
		if err != nil || backup == nil {
			recent, _ := e.db.GetRecentAlertByType(ctx, "backup_overdue:"+strconv.FormatInt(site.ID, 10), 24*time.Hour)
			if recent != nil {
				continue
			}

			msg := fmt.Sprintf("No backup found for %s in the last %d days", site.Domain, days)
			e.db.CreateAlert(ctx, "backup_overdue:"+strconv.FormatInt(site.ID, 10), "warning", msg, &site.ID, "website")
			e.emitAlert("backup_overdue", "warning", msg)
			continue
		}

		if time.Since(backup.CreatedAt) > time.Duration(days)*24*time.Hour {
			recent, _ := e.db.GetRecentAlertByType(ctx, "backup_overdue:"+strconv.FormatInt(site.ID, 10), 24*time.Hour)
			if recent != nil {
				continue
			}

			msg := fmt.Sprintf("Last backup for %s was %s", site.Domain, backup.CreatedAt.Format("2006-01-02"))
			e.db.CreateAlert(ctx, "backup_overdue:"+strconv.FormatInt(site.ID, 10), "warning", msg, &site.ID, "website")
			e.emitAlert("backup_overdue", "warning", msg)
		}
	}
}

func (e *Engine) emitAlert(alertType, severity, message string) {
	if e.bus == nil {
		return
	}

	e.bus.Publish(events.Event{
		Type: events.EventAlertFired,
		Data: map[string]interface{}{
			"type":     alertType,
			"severity": severity,
			"message":  message,
			"time":     time.Now(),
		},
	})
	e.log.Info("alert fired", "type", alertType, "severity", severity)
}