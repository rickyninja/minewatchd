package main

import "testing"

func TestScanLogin(t *testing.T) {
	t.Parallel()
	line := "ricky_ninja joined the game"
	got := ScanLogin(line)
	want := "ricky_ninja"
	if got != want {
		t.Errorf("wrong ScanLogin got %s want %s\n", got, want)
	}
}

func TestScanLogout(t *testing.T) {
	t.Parallel()
	line := "ricky_ninja left the game"
	want := "ricky_ninja"
	got := ScanLogout(line)
	if got != want {
		t.Errorf("wrong ScanLogin got %s want %s\n", got, want)
	}
}

func TestIsChatMessage(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		line string
		want bool
	}{
		{"chat msg", "<ricky_ninja> blah blah ricky_ninja joined the game", true},
		{"join msg", "ricky_ninja joined the game", false},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := IsChatMessage(tc.line)
			if got != tc.want {
				t.Errorf("wrong IsChatMessage got %t want %t\n", got, tc.want)
				t.Log(tc.name)
				t.Log(tc.line)
			}
		})
	}
}
