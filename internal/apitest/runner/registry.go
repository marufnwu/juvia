package runner

import "sync"

type ResourceRegistry struct {
	mu sync.Mutex

	Websites         []int64
	Databases        []int64
	DatabaseUsers    map[int64][]int64
	Users            []int64
	Mailboxes        []int64
	Aliases          []int64
	Forwarders       []int64
	FirewallRules    []int64
	CronJobs         []int64
	BackupSchedules  []int64
	Backups          []int64
	DNSRecords       []int64
	TerminalSessions []string
}

func NewResourceRegistry() *ResourceRegistry {
	return &ResourceRegistry{
		DatabaseUsers: make(map[int64][]int64),
	}
}

func (r *ResourceRegistry) AddWebsite(id int64) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.Websites = append(r.Websites, id)
}

func (r *ResourceRegistry) GetWebsite() int64 {
	r.mu.Lock()
	defer r.mu.Unlock()
	if len(r.Websites) == 0 {
		return 0
	}
	return r.Websites[0]
}

func (r *ResourceRegistry) AddDatabase(id int64) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.Databases = append(r.Databases, id)
}

func (r *ResourceRegistry) GetDatabase() int64 {
	r.mu.Lock()
	defer r.mu.Unlock()
	if len(r.Databases) == 0 {
		return 0
	}
	return r.Databases[0]
}

func (r *ResourceRegistry) AddDatabaseUser(dbID, userID int64) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.DatabaseUsers[dbID] = append(r.DatabaseUsers[dbID], userID)
}

func (r *ResourceRegistry) GetDatabaseUser(dbID int64) int64 {
	r.mu.Lock()
	defer r.mu.Unlock()
	if users, ok := r.DatabaseUsers[dbID]; ok && len(users) > 0 {
		return users[0]
	}
	return 0
}

func (r *ResourceRegistry) AddMailbox(id int64) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.Mailboxes = append(r.Mailboxes, id)
}

func (r *ResourceRegistry) GetMailbox() int64 {
	r.mu.Lock()
	defer r.mu.Unlock()
	if len(r.Mailboxes) == 0 {
		return 0
	}
	return r.Mailboxes[0]
}

func (r *ResourceRegistry) AddAlias(id int64) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.Aliases = append(r.Aliases, id)
}

func (r *ResourceRegistry) GetAlias() int64 {
	r.mu.Lock()
	defer r.mu.Unlock()
	if len(r.Aliases) == 0 {
		return 0
	}
	return r.Aliases[0]
}

func (r *ResourceRegistry) AddForwarder(id int64) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.Forwarders = append(r.Forwarders, id)
}

func (r *ResourceRegistry) GetForwarder() int64 {
	r.mu.Lock()
	defer r.mu.Unlock()
	if len(r.Forwarders) == 0 {
		return 0
	}
	return r.Forwarders[0]
}

func (r *ResourceRegistry) AddFirewallRule(id int64) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.FirewallRules = append(r.FirewallRules, id)
}

func (r *ResourceRegistry) GetFirewallRule() int64 {
	r.mu.Lock()
	defer r.mu.Unlock()
	if len(r.FirewallRules) == 0 {
		return 0
	}
	return r.FirewallRules[0]
}

func (r *ResourceRegistry) AddCronJob(id int64) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.CronJobs = append(r.CronJobs, id)
}

func (r *ResourceRegistry) GetCronJob() int64 {
	r.mu.Lock()
	defer r.mu.Unlock()
	if len(r.CronJobs) == 0 {
		return 0
	}
	return r.CronJobs[0]
}

func (r *ResourceRegistry) AddBackupSchedule(id int64) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.BackupSchedules = append(r.BackupSchedules, id)
}

func (r *ResourceRegistry) GetBackupSchedule() int64 {
	r.mu.Lock()
	defer r.mu.Unlock()
	if len(r.BackupSchedules) == 0 {
		return 0
	}
	return r.BackupSchedules[0]
}

func (r *ResourceRegistry) AddDNSRecord(id int64) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.DNSRecords = append(r.DNSRecords, id)
}

func (r *ResourceRegistry) GetDNSRecord() int64 {
	r.mu.Lock()
	defer r.mu.Unlock()
	if len(r.DNSRecords) == 0 {
		return 0
	}
	return r.DNSRecords[0]
}

func (r *ResourceRegistry) AddUser(id int64) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.Users = append(r.Users, id)
}

func (r *ResourceRegistry) GetUser() int64 {
	r.mu.Lock()
	defer r.mu.Unlock()
	if len(r.Users) == 0 {
		return 0
	}
	return r.Users[0]
}

func (r *ResourceRegistry) AddBackup(id int64) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.Backups = append(r.Backups, id)
}

func (r *ResourceRegistry) GetBackup() int64 {
	r.mu.Lock()
	defer r.mu.Unlock()
	if len(r.Backups) == 0 {
		return 0
	}
	return r.Backups[0]
}

func (r *ResourceRegistry) AddTerminalSession(id string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.TerminalSessions = append(r.TerminalSessions, id)
}

func (r *ResourceRegistry) GetTerminalSession() string {
	r.mu.Lock()
	defer r.mu.Unlock()
	if len(r.TerminalSessions) == 0 {
		return ""
	}
	return r.TerminalSessions[0]
}
