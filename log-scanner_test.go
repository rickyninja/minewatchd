package main

import (
	"fmt"
	"testing"
	"time"
)

func TestLogScanner_Scan_Notification(t *testing.T) {
	t.Parallel()
	now := time.Now()
	mutes := []string{"ricky_ninja"}
	tests := []struct {
		name  string
		line  string
		mutes []string
		want  bool
	}{
		{"old log", "[2020-02-08 16:10:39 MST] [INFO]: ricky_ninja joined the game", mutes, false},
		{"muted user joined", fmt.Sprintf("[%s] [INFO]: ricky_ninja joined the game", now.Format(timeFormat)), mutes, false},
		{"muted user left", fmt.Sprintf("[%s] [INFO]: ricky_ninja left the game", now.Format(timeFormat)), mutes, false},
		{"user joined", fmt.Sprintf("[%s] [INFO]: ricky_ninja joined the game", now.Format(timeFormat)), nil, true},
		{"user left", fmt.Sprintf("[%s] [INFO]: ricky_ninja left the game", now.Format(timeFormat)), nil, true},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			no := &notifyRecorder{}
			nots := []Notifyer{no}
			ls := NewLogScanner("America/Phoenix", nots, tc.mutes)
			ls.Scan(tc.line)
			if no.sent != tc.want {
				if tc.want {
					t.Errorf("should have sent notice for %s", tc.name)
				} else {
					t.Errorf("should not have sent notice for %s", tc.name)
				}
			}
		})
	}
}

type notifyRecorder struct {
	sent bool
}

func (m *notifyRecorder) Notify(msg string) error {
	m.sent = true
	return nil
}
