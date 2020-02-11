package main

import (
	"fmt"
	"testing"
	"time"
)

func TestLogScanner_Scan_Now(t *testing.T) {
	t.Parallel()
	line := "[2020-02-08 16:10:39 MST] [INFO]: ricky_ninja joined the game"
	no := &mute{}
	nots := []Notifyer{no}
	ls := NewLogScanner("America/Phoenix", nots, nil)
	ls.Scan(line)
	if no.sent {
		t.Error("should not have sent notice because log line is not near current time")
	}
}

func TestLogScanner_Scan_NotSelf(t *testing.T) {
	t.Parallel()
	now := time.Now()
	line := fmt.Sprintf("[%s] [INFO]: ricky_ninja joined the game", now.Format(timeFormat))
	no := &mute{}
	nots := []Notifyer{no}
	ls := NewLogScanner("America/Phoenix", nots, []string{"ricky_ninja"})
	ls.Scan(line)
	if no.sent {
		t.Error("should not have sent notice because user ricky_ninja is configured not to send notices")
	}
}

type mute struct {
	sent bool
}

func (m *mute) Notify(msg string) error {
	m.sent = true
	return nil
}
