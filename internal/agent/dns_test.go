package agent

import (
	"testing"
	"time"
)

func TestNextSerial(t *testing.T) {
	tests := []struct {
		name     string
		current  int64
		wantLen  int
		wantDate string
	}{
		{
			name:     "zero serial generates today",
			current:  0,
			wantLen:  10,
			wantDate: time.Now().Format("20060102"),
		},
		{
			name:     "same day increments sequence",
			current:  2025061501,
			wantLen:  10,
			wantDate: time.Now().Format("20060102"),
		},
		{
			name:     "old date resets to today 01",
			current:  2025060105,
			wantLen:  10,
			wantDate: time.Now().Format("20060102"),
		},
		{
			name:     "malformed serial generates today",
			current:  12345,
			wantLen:  10,
			wantDate: time.Now().Format("20060102"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := nextSerial(tt.current)
			s := formatInt64(got)
			if len(s) != tt.wantLen {
				t.Errorf("nextSerial(%d) len = %d, want %d", tt.current, len(s), tt.wantLen)
			}
			if s[:8] != tt.wantDate {
				t.Errorf("nextSerial(%d) date = %s, want %s", tt.current, s[:8], tt.wantDate)
			}
		})
	}
}

func formatInt64(n int64) string {
	if n == 0 {
		return "0"
	}
	if n < 0 {
		return "-" + formatInt64(-n)
	}
	digits := []byte{}
	for n > 0 {
		digits = append([]byte{byte('0' + n%10)}, digits...)
		n /= 10
	}
	return string(digits)
}

func TestValidateDomainName(t *testing.T) {
	tests := []struct {
		name    string
		domain  string
		wantErr bool
	}{
		{"valid simple", "example.com", false},
		{"valid with subdomain", "www.example.com", false},
		{"valid with hyphen", "my-site.com", false},
		{"valid with numbers", "site123.com", false},
		{"valid with wildcard", "*.example.com", false},
		{"valid multi-label", "a.b.c.example.com", false},
		{"empty", "", true},
		{"single label", "example", true},
		{"too long", string(make([]byte, 254)), true},
		{"label too long", "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa.example.com", true},
		{"invalid char underscore", "my_site.com", true},
		{"invalid char space", "my site.com", true},
		{"empty label", "example..com", true},
		{"label starts with hyphen", "-example.com", true},
		{"label ends with hyphen", "example-.com", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateDomainName(tt.domain)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateDomainName(%q) error = %v, wantErr %v", tt.domain, err, tt.wantErr)
			}
		})
	}
}

func TestGenerateZoneFromRecords(t *testing.T) {
	domain := "example.com"
	serverIP := "1.2.3.4"
	serial := int64(2025061501)

	tests := []struct {
		name          string
		records       []DNSRecordData
		wantSubstring []string
		notWant       []string
	}{
		{
			name:    "empty records generates defaults",
			records: []DNSRecordData{},
			wantSubstring: []string{
				"$TTL 3600",
				"ns1.example.com",
				"ns2.example.com",
				"1.2.3.4",
				"mail.example.com",
				"v=spf1 mx ~all",
				"v=DMARC1",
			},
			notWant: []string{},
		},
		{
			name: "A record replaces default",
			records: []DNSRecordData{
				{Type: "A", Name: "@", Value: "5.6.7.8", TTL: 7200},
			},
			wantSubstring: []string{
				"5.6.7.8",
				"$TTL 3600",
			},
			notWant: []string{"1.2.3.4"},
		},
		{
			name: "MX record with priority",
			records: []DNSRecordData{
				{Type: "MX", Name: "@", Value: "mail.example.com.", Priority: 20, TTL: 3600},
			},
			wantSubstring: []string{
				"MX     20    mail.example.com.",
			},
		},
		{
			name: "TXT record with SPF",
			records: []DNSRecordData{
				{Type: "TXT", Name: "@", Value: "v=spf1 a ~all", TTL: 3600},
			},
			wantSubstring: []string{
				`"v=spf1 a ~all"`,
			},
		},
		{
			name: "CNAME record",
			records: []DNSRecordData{
				{Type: "CNAME", Name: "ftp", Value: "example.com.", TTL: 3600},
			},
			wantSubstring: []string{
				"CNAME  example.com.",
			},
		},
		{
			name: "AAAA record",
			records: []DNSRecordData{
				{Type: "AAAA", Name: "@", Value: "::1", TTL: 3600},
			},
			wantSubstring: []string{
				"::1",
			},
		},
		{
			name: "CAA record",
			records: []DNSRecordData{
				{Type: "CAA", Name: "@", Value: `0 issue "letsencrypt.org"`, TTL: 3600},
			},
			wantSubstring: []string{
				`0 issue "letsencrypt.org"`,
			},
		},
		{
			name: "SRV record",
			records: []DNSRecordData{
				{Type: "SRV", Name: "_http._tcp", Value: "10 5 443 web.example.com.", TTL: 3600},
			},
			wantSubstring: []string{
				"_http._tcp",
			},
		},
		{
			name: "NS record",
			records: []DNSRecordData{
				{Type: "NS", Name: "@", Value: "ns1.example.com.", TTL: 3600},
			},
			wantSubstring: []string{
				"ns1.example.com.",
			},
		},
		{
			name: "TTL is written per record",
			records: []DNSRecordData{
				{Type: "A", Name: "@", Value: "1.2.3.4", TTL: 300},
				{Type: "A", Name: "www", Value: "1.2.3.4", TTL: 7200},
			},
			wantSubstring: []string{
				"IN    A      1.2.3.4",
				"300",
				"7200",
			},
		},
		{
			name: "default TTL when zero",
			records: []DNSRecordData{
				{Type: "A", Name: "@", Value: "1.2.3.4", TTL: 0},
			},
			wantSubstring: []string{
				"IN    A      1.2.3.4",
				"3600",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := generateZoneFromRecords(domain, serverIP, serial, tt.records)

			for _, want := range tt.wantSubstring {
				if !containsString(got, want) {
					t.Errorf("generateZoneFromRecords missing %q\ngot:\n%s", want, got)
				}
			}
			for _, notWant := range tt.notWant {
				if containsString(got, notWant) {
					t.Errorf("generateZoneFromRecords should NOT contain %q\ngot:\n%s", notWant, got)
				}
			}
		})
	}
}

func TestSerialFromTime(t *testing.T) {
	t.Run("first serial of day", func(t *testing.T) {
		tm := time.Date(2025, 6, 15, 10, 0, 0, 0, time.UTC)
		got := serialFromTime(tm, 0)
		s := formatInt64(got)
		if s != "2025061501" {
			t.Errorf("serialFromTime first = %s, want 2025061501", s)
		}
	})

	t.Run("increment within same day", func(t *testing.T) {
		tm := time.Date(2025, 6, 15, 14, 0, 0, 0, time.UTC)
		got := serialFromTime(tm, 2025061503)
		s := formatInt64(got)
		if s != "2025061504" {
			t.Errorf("serialFromTime increment = %s, want 2025061504", s)
		}
	})

	t.Run("new day resets to 01", func(t *testing.T) {
		tm := time.Date(2025, 6, 16, 1, 0, 0, 0, time.UTC)
		got := serialFromTime(tm, 2025061510)
		s := formatInt64(got)
		if s != "2025061601" {
			t.Errorf("serialFromTime new day = %s, want 2025061601", s)
		}
	})
}

func TestSplitTXTRecords(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []string
	}{
		{
			name:  "single quoted",
			input: `"v=spf1 mx ~all"`,
			want:  []string{"v=spf1 mx ~all"},
		},
		{
			name:  "multiple quoted",
			input: `"v=spf1 mx ~all" "dmarc"`,
			want:  []string{"v=spf1 mx ~all", "dmarc"},
		},
		{
			name:  "unquoted",
			input: "v=spf1 mx ~all",
			want:  []string{"v=spf1 mx ~all"},
		},
		{
			name:  "dmarc record",
			input: `"v=DMARC1; p=quarantine; rua=mailto:dmarc@example.com"`,
			want:  []string{"v=DMARC1; p=quarantine; rua=mailto:dmarc@example.com"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := splitTXTRecords(tt.input)
			if len(got) != len(tt.want) {
				t.Errorf("splitTXTRecords(%q) count = %d, want %d", tt.input, len(got), len(tt.want))
				return
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("splitTXTRecords[%d] = %q, want %q", i, got[i], tt.want[i])
				}
			}
		})
	}
}

func TestSanitizeForZone(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"normal", "normal"},
		{"with\nnewline", "withnewline"},
		{"with\r carriage", "with carriage"},
		{"with\ttab", "with tab"},
		{"multi\n\r\n\t", "multi "},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := sanitizeForZone(tt.input)
			if got != tt.want {
				t.Errorf("sanitizeForZone(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestJoinFirstNonEmpty(t *testing.T) {
	tests := []struct {
		name      string
		ss        []string
		fallback  []string
		want      string
	}{
		{"first non-empty", []string{"a", "b"}, []string{}, "a"},
		{"fallback to second", []string{"", "b"}, []string{}, "b"},
		{"fallback to fallback", []string{""}, []string{"c"}, "c"},
		{"all empty fallback", []string{""}, []string{""}, ""},
		{"empty both", []string{}, []string{}, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := joinFirstNonEmpty(tt.ss, tt.fallback)
			if got != tt.want {
				t.Errorf("joinFirstNonEmpty = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestDefaultTTL(t *testing.T) {
	tests := []struct {
		input int
		want  int
	}{
		{0, 3600},
		{-1, 3600},
		{300, 300},
		{3600, 3600},
		{86400, 86400},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			got := defaultTTL(tt.input)
			if got != tt.want {
				t.Errorf("defaultTTL(%d) = %d, want %d", tt.input, got, tt.want)
			}
		})
	}
}

func containsString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
