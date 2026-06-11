package main

import (
	"context"
	"encoding/json"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"juvia/internal/agent"
	"juvia/internal/agent/apps"
	"juvia/internal/agent/backup"
	"juvia/internal/agent/cron"
	"juvia/internal/agent/database"
	"juvia/internal/agent/email"
	"juvia/internal/agent/firewall"
	"juvia/internal/agent/git"
	"juvia/internal/agent/terminal"
	"juvia/internal/socket"
)

func main() {
	log := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	socketPath := os.Getenv("JUVIA_SOCKET")
	if socketPath == "" {
		socketPath = "/var/run/juvia/agent.sock"
	}

	server := socket.NewServer(socketPath, log)
	server.RegisterMethod("system.ping", func(ctx context.Context, params json.RawMessage) (interface{}, error) {
		return map[string]interface{}{"pong": true}, nil
	})
	server.RegisterMethod("website.create", agent.HandleWebsiteCreate)
	server.RegisterMethod("website.delete", agent.HandleWebsiteDelete)
	server.RegisterMethod("website.suspend", agent.HandleWebsiteSuspend)
	server.RegisterMethod("ssl.issue", agent.HandleSSLIssue)
	server.RegisterMethod("ssl.renew", agent.HandleSSLRenew)
	server.RegisterMethod("ssl.check", agent.HandleSSLCheck)
	server.RegisterMethod("ssl.remove", agent.HandleSSLRemove)
	server.RegisterMethod("dns.zone.create", agent.HandleZoneCreate)
	server.RegisterMethod("dns.zone.update", agent.HandleZoneUpdate)
	server.RegisterMethod("dns.zone.delete", agent.HandleZoneDelete)
	server.RegisterMethod("files.list", agent.HandleFilesList)
	server.RegisterMethod("files.upload", agent.HandleFilesUpload)
	server.RegisterMethod("files.download", agent.HandleFilesDownload)
	server.RegisterMethod("files.delete", agent.HandleFilesDelete)
	server.RegisterMethod("files.rename", agent.HandleFilesRename)
	server.RegisterMethod("files.extract", agent.HandleFilesExtract)
	server.RegisterMethod("email.create", email.HandleMailboxCreate)
	server.RegisterMethod("email.delete", email.HandleMailboxDelete)
	server.RegisterMethod("email.update", email.HandleMailboxUpdate)
	server.RegisterMethod("email.alias.create", email.HandleAliasCreate)
	server.RegisterMethod("email.alias.delete", email.HandleAliasDelete)
	server.RegisterMethod("email.forwarder.create", email.HandleForwarderCreate)
	server.RegisterMethod("email.forwarder.delete", email.HandleForwarderDelete)
	server.RegisterMethod("email.deliverability", email.HandleDeliverability)
	server.RegisterMethod("database.create", database.HandleDatabaseCreate)
	server.RegisterMethod("database.delete", database.HandleDatabaseDelete)
	server.RegisterMethod("database.user.create", database.HandleDBUserCreate)
	server.RegisterMethod("database.user.delete", database.HandleDBUserDelete)
	server.RegisterMethod("database.export", database.HandleExport)
	server.RegisterMethod("database.tables", database.HandleListTables)
	server.RegisterMethod("database.rows", database.HandleGetRows)
	server.RegisterMethod("database.query", database.HandleQuery)
	server.RegisterMethod("firewall.rule.create", firewall.HandleRuleCreate)
	server.RegisterMethod("firewall.rule.update", firewall.HandleRuleUpdate)
	server.RegisterMethod("firewall.rule.delete", firewall.HandleRuleDelete)
	server.RegisterMethod("backup.create", backup.HandleBackupCreate)
	server.RegisterMethod("backup.restore", backup.HandleBackupRestore)
	server.RegisterMethod("backup.delete", backup.HandleBackupDelete)
	server.RegisterMethod("cron.create", cron.HandleCronCreate)
	server.RegisterMethod("cron.delete", cron.HandleCronDelete)
	server.RegisterMethod("cron.run", cron.HandleCronRun)
	server.RegisterMethod("apps.install", apps.HandleAppInstall)
	server.RegisterMethod("apps.update", apps.HandleAppUpdate)
	server.RegisterMethod("apps.check-update", apps.HandleAppCheckUpdate)
	server.RegisterMethod("git.setup", git.HandleGitSetup)
	server.RegisterMethod("git.pull", git.HandleGitPull)
	server.RegisterMethod("webmail.install", email.HandleWebmailInstall)
	server.RegisterMethod("webmail.uninstall", email.HandleWebmailUninstall)
	server.RegisterMethod("webmail.status", email.HandleWebmailStatus)
	server.RegisterMethod("terminal.start", terminal.HandleTerminalStart)
	server.RegisterMethod("terminal.stop", terminal.HandleTerminalStop)

	if err := server.Listen(); err != nil {
		log.Error("failed to listen", "error", err)
		os.Exit(1)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGTERM, syscall.SIGINT)

	go func() {
		if err := server.Serve(ctx); err != nil {
			log.Error("server error", "error", err)
		}
	}()

	log.Info("agent started")
	<-sigCh
	log.Info("agent shutting down")
	server.Close()
}