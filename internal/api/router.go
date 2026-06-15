package api

import (
	"embed"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"juvia/internal/auth"
	"juvia/internal/db"
	"juvia/internal/socket"
)

type RouterConfig struct {
	DB          *db.DB
	JWT         *auth.JWTManager
	Sessions    *auth.SessionStore
	AgentClient *socket.Client
	Log         *slog.Logger
	StaticFS    embed.FS
	CSRFSecret  string
}

func NewRouter(cfg RouterConfig) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(loggerMiddleware(cfg.Log))
	r.Use(corsMiddleware())

	api := r.Group("/api/v1")
	{
		api.GET("/health", healthHandler(cfg))
		api.GET("/setup/status", setupStatusHandler(cfg))
		api.POST("/setup/first-run", firstRunHandler(cfg))

		authGroup := api.Group("/auth")
		{
			authGroup.POST("/login", loginHandler(cfg))
			authGroup.POST("/refresh", refreshHandler(cfg))
			authGroup.POST("/logout", authWithCSRF(cfg.JWT, cfg.Sessions), logoutHandler(cfg))
			authGroup.GET("/me", authWithCSRF(cfg.JWT, cfg.Sessions), meHandler(cfg))
			authGroup.POST("/2fa/enable", authWithCSRF(cfg.JWT, cfg.Sessions), enable2FAHandler(cfg))
			authGroup.POST("/2fa/verify", authWithCSRF(cfg.JWT, cfg.Sessions), verify2FAHandler(cfg))
			authGroup.DELETE("/sessions/:id", authWithCSRF(cfg.JWT, cfg.Sessions), revokeSessionHandler(cfg))
		}

		users := api.Group("/users", authWithCSRF(cfg.JWT, cfg.Sessions), adminMiddleware())
		{
			users.GET("", listUsersHandler(cfg))
			users.POST("", createUserHandler(cfg))
			users.GET("/:id", getUserHandler(cfg))
			users.PUT("/:id", updateUserHandler(cfg))
			users.DELETE("/:id", deleteUserHandler(cfg))
		}

		websites := api.Group("/websites", authWithCSRF(cfg.JWT, cfg.Sessions))
		{
			websites.GET("", listWebsitesHandler(cfg))
			websites.POST("", createWebsiteHandler(cfg))
			websites.GET("/trash", listTrashedWebsitesHandler(cfg))
			websites.DELETE("/trash/purge", purgeTrashHandler(cfg))
			websites.GET("/:id", getWebsiteHandler(cfg))
			websites.PUT("/:id", updateWebsiteHandler(cfg))
			websites.DELETE("/:id", deleteWebsiteHandler(cfg))
			websites.POST("/:id/suspend", suspendWebsiteHandler(cfg))
			websites.POST("/:id/restore", restoreWebsiteHandler(cfg))
			websites.DELETE("/:id/permanent", permanentDeleteWebsiteHandler(cfg))
			websites.PUT("/:id/php", changePHPHandler(cfg))
			websites.GET("/:id/ssl/check", sslCheckHandler(cfg))
			websites.POST("/:id/ssl", issueSSLHandler(cfg))
			websites.POST("/:id/ssl-renew", renewSSLHandler(cfg))
			websites.DELETE("/:id/ssl", removeSSLHandler(cfg))
		websites.GET("/:id/dns/zone", getWebsiteZoneHandler(cfg))
		websites.GET("/:id/dns/records", listDNSRecordsHandler(cfg))
		websites.POST("/:id/dns/records", createDNSRecordHandler(cfg))
		websites.POST("/:id/dns/suggest", suggestDNSHandler(cfg))
		websites.GET("/:id/domains", listDomainsHandler(cfg))
		websites.POST("/:id/domains", createDomainHandler(cfg))
		websites.DELETE("/:id/domains/:domainId", deleteDomainHandler(cfg))
		websites.POST("/:id/domains/:domainId/ssl", issueDomainSSLHandler(cfg))
		websites.POST("/:id/domains/:domainId/ssl-renew", renewDomainSSLHandler(cfg))
		websites.POST("/:id/domains/:domainId/dns-verify", verifyDomainDNSHandler(cfg))
			websites.GET("/:id/files", listFilesHandler(cfg))
			websites.POST("/:id/files/upload", uploadFilesHandler(cfg))
			websites.GET("/:id/files/download", downloadFilesHandler(cfg))
			websites.PUT("/:id/files/rename", renameFilesHandler(cfg))
			websites.DELETE("/:id/files/delete", deleteFilesHandler(cfg))
			websites.POST("/:id/files/extract", extractFilesHandler(cfg))
			websites.POST("/:id/files/folder", createFolderHandler(cfg))
			websites.GET("/:id/files/edit", getFileContentHandler(cfg))
			websites.PUT("/:id/files/edit", saveFileContentHandler(cfg))
			websites.GET("/:id/logs/access", accessLogHandler(cfg))
			websites.GET("/:id/logs/error", errorLogHandler(cfg))
		}

		dns := api.Group("/dns", authWithCSRF(cfg.JWT, cfg.Sessions))
		{
			dns.GET("/nameservers", getNameserversHandler(cfg))
			dns.GET("/zones", listZonesHandler(cfg))
			dns.POST("/zones", createZoneHandler(cfg))
			dns.DELETE("/zones/:id", deleteZoneHandler(cfg))
			dns.GET("/health", nameserverHealthHandler(cfg))
			dns.GET("/check-all", checkAllDomainsDNSHandler(cfg))
			dns.POST("/restart", restartBindHandler(cfg))
			dns.PUT("/records/:id", updateDNSRecordHandler(cfg))
			dns.DELETE("/records/:id", deleteDNSRecordHandler(cfg))
		}

		databases := api.Group("/databases", authWithCSRF(cfg.JWT, cfg.Sessions))
		{
			databases.GET("", listDatabasesHandler(cfg))
			databases.POST("", createDatabaseHandler(cfg))
			databases.GET("/:id", getDatabaseHandler(cfg))
			databases.DELETE("/:id", deleteDatabaseHandler(cfg))
			databases.POST("/:id/users", createDBUserHandler(cfg))
			databases.DELETE("/:id/users/:user_id", deleteDBUserHandler(cfg))
			databases.POST("/:id/export", exportDatabaseHandler(cfg))
			databases.GET("/:id/tables", listTablesHandler(cfg))
			databases.GET("/:id/tables/:table/rows", getTableRowsHandler(cfg))
			databases.POST("/:id/query", queryDatabaseHandler(cfg))
		}

		email := api.Group("/email", authWithCSRF(cfg.JWT, cfg.Sessions))
		{
			email.GET("/mailboxes", listMailboxesHandler(cfg))
			email.POST("/mailboxes", createMailboxHandler(cfg))
			email.GET("/mailboxes/:id", getMailboxHandler(cfg))
			email.DELETE("/mailboxes/:id", deleteMailboxHandler(cfg))
			email.PUT("/mailboxes/:id", updateMailboxHandler(cfg))
			email.GET("/aliases", listAliasesHandler(cfg))
			email.POST("/aliases", createAliasHandler(cfg))
			email.DELETE("/aliases/:id", deleteAliasHandler(cfg))
			email.GET("/forwarders", listForwardersHandler(cfg))
			email.POST("/forwarders", createForwarderHandler(cfg))
			email.DELETE("/forwarders/:id", deleteForwarderHandler(cfg))
			email.GET("/catch-all", getCatchAllHandler(cfg))
			email.PUT("/catch-all", setCatchAllHandler(cfg))
			email.GET("/deliverability", deliverabilityHandler(cfg))
		}

		firewall := api.Group("/firewall", authWithCSRF(cfg.JWT, cfg.Sessions))
		{
			firewall.GET("/rules", listFirewallRulesHandler(cfg))
			firewall.POST("/rules", createFirewallRuleHandler(cfg))
			firewall.GET("/rules/:id", getFirewallRuleHandler(cfg))
			firewall.PUT("/rules/:id", updateFirewallRuleHandler(cfg))
			firewall.DELETE("/rules/:id", deleteFirewallRuleHandler(cfg))
		}

		backups := api.Group("/backups", authWithCSRF(cfg.JWT, cfg.Sessions))
		{
			backups.GET("", listBackupsHandler(cfg))
			backups.POST("", createBackupHandler(cfg))
			backups.GET("/:id", getBackupHandler(cfg))
			backups.POST("/:id/restore", restoreBackupHandler(cfg))
			backups.DELETE("/:id", deleteBackupHandler(cfg))
		}

		backupSchedules := api.Group("/backup-schedules", authWithCSRF(cfg.JWT, cfg.Sessions))
		{
			backupSchedules.GET("", listBackupSchedulesHandler(cfg))
			backupSchedules.POST("", createBackupScheduleHandler(cfg))
			backupSchedules.DELETE("/:id", deleteBackupScheduleHandler(cfg))
		}

		cron := api.Group("/cron", authWithCSRF(cfg.JWT, cfg.Sessions))
		{
			cron.GET("", listCronJobsHandler(cfg))
			cron.POST("", createCronJobHandler(cfg))
			cron.GET("/:id", getCronJobHandler(cfg))
			cron.PUT("/:id", updateCronJobHandler(cfg))
			cron.DELETE("/:id", deleteCronJobHandler(cfg))
			cron.POST("/:id/enable", enableCronJobHandler(cfg))
			cron.POST("/:id/disable", disableCronJobHandler(cfg))
			cron.GET("/:id/logs", getCronJobLogsHandler(cfg))
		}

		logs := api.Group("/logs", authWithCSRF(cfg.JWT, cfg.Sessions))
		{
			logs.GET("/system", systemLogsHandler(cfg))
		}

		alerts := api.Group("/alerts", authWithCSRF(cfg.JWT, cfg.Sessions))
		{
			alerts.GET("", listAlertsHandler(cfg))
			alerts.POST("/:id/acknowledge", acknowledgeAlertHandler(cfg))
			alerts.DELETE("/:id", deleteAlertHandler(cfg))
			alerts.GET("/settings", getAlertSettingsHandler(cfg))
			alerts.PUT("/settings", updateAlertSettingsHandler(cfg))
		}

		metrics := api.Group("/metrics", authWithCSRF(cfg.JWT, cfg.Sessions))
		{
			metrics.GET("/current", currentMetricsHandler(cfg))
			metrics.GET("/history", historyMetricsHandler(cfg))
		}

		apps := api.Group("/websites/:id/apps", authWithCSRF(cfg.JWT, cfg.Sessions))
		{
			apps.GET("", listAppsHandler(cfg))
			apps.POST("/install", installAppHandler(cfg))
			apps.POST("/:app_id/update", updateAppHandler(cfg))
		}

		git := api.Group("/websites/:id/git", authWithCSRF(cfg.JWT, cfg.Sessions))
		{
			git.GET("", getGitConfigHandler(cfg))
			git.POST("", setupGitHandler(cfg))
			git.POST("/pull", gitPullHandler(cfg))
			git.GET("/webhook", getWebhookURLHandler(cfg))
		}

		webmail := api.Group("/webmail", authWithCSRF(cfg.JWT, cfg.Sessions))
		{
			webmail.GET("/status", webmailStatusHandler(cfg))
			webmail.POST("/install", installWebmailHandler(cfg))
			webmail.POST("/uninstall", uninstallWebmailHandler(cfg))
			webmail.GET("/url", webmailURLHandler(cfg))
		}

		terminal := api.Group("/terminal", authWithCSRF(cfg.JWT, cfg.Sessions))
		{
			terminal.POST("/session", createTerminalSessionHandler(cfg))
			terminal.GET("/sessions", listTerminalSessionsHandler(cfg))
			terminal.DELETE("/sessions/:id", closeTerminalSessionHandler(cfg))
			terminal.GET("/recordings", listTerminalRecordingsHandler(cfg))
			terminal.GET("/recordings/:id", getTerminalRecordingHandler(cfg))
			terminal.DELETE("/recordings/:id", deleteTerminalRecordingHandler(cfg))
		}

		updates := api.Group("/updates", authWithCSRF(cfg.JWT, cfg.Sessions))
		{
			updates.GET("/check", checkUpdateHandler(cfg))
			updates.POST("/download", downloadUpdateHandler(cfg))
			updates.POST("/apply", applyUpdateHandler(cfg))
			updates.POST("/rollback", rollbackHandler(cfg))
		}

		api.GET("/server/ip", authWithCSRF(cfg.JWT, cfg.Sessions), getServerIPHandler(cfg))
		api.POST("/server/ip/refresh", authWithCSRF(cfg.JWT, cfg.Sessions), refreshServerIPHandler(cfg))

		settingsGroup := api.Group("/settings", authWithCSRF(cfg.JWT, cfg.Sessions))
		{
			settingsGroup.GET("", getSettingsHandler(cfg))
			settingsGroup.PUT("", updateSettingsHandler(cfg))
			settingsGroup.GET("/export", exportConfigHandler(cfg))
			settingsGroup.POST("/import", importConfigHandler(cfg))
			settingsGroup.GET("/ssh", getSSHSettingsHandler(cfg))
			settingsGroup.PUT("/ssh", updateSSHSettingsHandler(cfg))
		}

		auditLog := api.Group("/audit-log", authWithCSRF(cfg.JWT, cfg.Sessions))
		{
			auditLog.GET("", getAuditLogHandler(cfg))
		}

		services := api.Group("/services", authMiddleware(cfg.JWT, cfg.Sessions))
		{
			services.GET("", listServicesHandler(cfg))
			services.GET("/:name", getServiceHandler(cfg))
			services.POST("/:name/restart", restartServiceHandler(cfg))
		}

		api.GET("/version", getVersionHandler(cfg))

		api.POST("/webhooks/git/:token", gitWebhookHandler(cfg))
	}

	staticServer := http.FileServer(http.FS(cfg.StaticFS))
	r.GET("/assets/*filepath", func(c *gin.Context) {
		c.Request.URL.Path = "/dist" + c.Request.URL.Path
		staticServer.ServeHTTP(c.Writer, c.Request)
	})
	r.NoRoute(func(c *gin.Context) {
		data, err := cfg.StaticFS.ReadFile("dist/index.html")
		if err != nil {
			c.String(http.StatusNotFound, "not found")
			return
		}
		c.Data(http.StatusOK, "text/html; charset=utf-8", data)
	})

	return r
}