package main

import (
	"strings"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	t.Parallel()
	r := strings.NewReader(`---
emails:
  - foo@example.com
  - bar@example.com
notify_url: http://127.0.0.1:8080/minecraft/notify
log_file: latest.log
time_zone: America/Phoenix
muted_users:
  - ricky_ninja
`)
	conf, err := loadConfig(r)
	if err != nil {
		t.Fatal(err)
	}
	wantEmails := []string{
		"foo@example.com",
		"bar@example.com",
	}
	if len(conf.Emails) != len(wantEmails) {
		t.Errorf("wrong number of emails got %d want %d", len(conf.Emails), len(wantEmails))
	}
	for i, w := range wantEmails {
		if conf.Emails[i] != w {
			t.Errorf("wrong email got %s want %s\n", conf.Emails[i], w)
		}
	}
	wantURL := "http://127.0.0.1:8080/minecraft/notify"
	if conf.NotifyURL != wantURL {
		t.Errorf("wrong notify_url got %s want %s\n", conf.NotifyURL, wantURL)
	}
	wantLogFile := "latest.log"
	if conf.LogFile != wantLogFile {
		t.Errorf("wrong notify_url got %s want %s\n", conf.LogFile, wantLogFile)
	}
	wantTimeZone := "America/Phoenix"
	if conf.TimeZone != wantTimeZone {
		t.Errorf("wrong time_zone got %s want %s\n", conf.TimeZone, wantTimeZone)
	}
	wantMutes := []string{
		"ricky_ninja",
	}
	if len(conf.MutedUsers) != len(wantMutes) {
		t.Errorf("wrong number of excluded_users got %d want %d", len(conf.MutedUsers), len(wantMutes))
	}
	for i, w := range wantMutes {
		if conf.MutedUsers[i] != w {
			t.Errorf("wrong user got %s want %s\n", conf.MutedUsers[i], w)
		}
	}
}
