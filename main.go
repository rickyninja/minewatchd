package main

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"gopkg.in/yaml.v3"
)

const timeFormat = "2006-01-02 15:04:05 MST"

func main() {
	var confFile string
	var defaultConfFile string = filepath.Join(os.Getenv("HOME"), ".minewatchd.yaml")
	flag.StringVar(&confFile, "conf-file", defaultConfFile, "path to minewatchd's yaml config file")
	flag.Parse()
	conf, err := loadConfigFile(confFile)
	if err != nil {
		log.Fatal(err)
	}
	if conf.LogFile == "" {
		log.Fatalf("you must provide log_file value in %s\n", confFile)
	}

	notifiers := make([]Notifyer, 0)
	for _, email := range conf.Emails {
		notifiers = append(notifiers, NewNotifyHttp(email, conf.NotifyURL))
	}
	logScanner := NewLogScanner(conf.TimeZone, notifiers, conf.MutedUsers)
	tail(conf.LogFile, logScanner)
}

type Config struct {
	Emails     []string `yaml:"emails"`
	MutedUsers []string `yaml:"muted_users"`
	NotifyURL  string   `yaml:"notify_url"`
	LogFile    string   `yaml:"log_file"`
	TimeZone   string   `yaml:"time_zone"`
}

func loadConfig(r io.Reader) (Config, error) {
	var conf Config
	err := yaml.NewDecoder(r).Decode(&conf)
	return conf, err
}

func loadConfigFile(file string) (Config, error) {
	var conf Config
	fd, err := os.Open(file)
	if err != nil {
		return conf, err
	}
	defer fd.Close()
	return loadConfig(fd)
}

type scanner interface {
	Scan(line string)
}

type LogScanner struct {
	TimeZone  string
	Notifiers []Notifyer
	MutedUser map[string]bool
}

func NewLogScanner(tz string, notifiers []Notifyer, muted []string) *LogScanner {
	mute := make(map[string]bool)
	for _, u := range muted {
		mute[u] = true
	}
	return &LogScanner{
		TimeZone:  tz,
		Notifiers: notifiers,
		MutedUser: mute,
	}
}

func (l *LogScanner) Scan(line string) {
	ll, err := NewLogLine(line, l.TimeZone)
	if err != nil {
		log.Println(err)
		return
	}
	now := time.Now()
	loc, err := time.LoadLocation(l.TimeZone)
	if err != nil {
		fmt.Printf("failed to load time location: %s\n", err)
	} else {
		now = now.In(loc)
	} // Skip lines that are not close to current time, so there isn't a flood of notices
	// every time the service starts.
	if now.Sub(ll.Time) > 10*time.Second {
		return
	}
	user := ScanLogin(ll.Line)
	if l.MutedUser[user] {
		return
	}
	if user != "" {
		l.SendNotices(fmt.Sprintf("%s\n%s online\n", ll.Time, user))
	}
	user = ScanLogout(ll.Line)
	if l.MutedUser[user] {
		return
	}
	if user != "" {
		l.SendNotices(fmt.Sprintf("%s\n%s offline\n", ll.Time, user))
	}
}

func (l *LogScanner) SendNotices(msg string) {
	for _, n := range l.Notifiers {
		err := n.Notify(msg)
		if err != nil {
			log.Println(err)
		}
	}
}

func tail(filename string, sc scanner) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	fd, err := syscall.InotifyInit()
	if err != nil {
		return err
	}
	_, err = syscall.InotifyAddWatch(fd, filename, syscall.IN_MODIFY)
	r := bufio.NewReader(file)
	for {
		by, err := r.ReadBytes('\n')
		if err != nil && err != io.EOF {
			return err
		}
		by = bytes.TrimSuffix(by, []byte{'\n'})
		//fmt.Printf("looking at line: %s\n", string(by))
		if len(by) > 0 {
			sc.Scan(string(by))
		}
		if err != io.EOF {
			continue
		}
		if err = waitForChange(fd); err != nil {
			return err
		}
	}
}

func waitForChange(fd int) error {
	for {
		var buf [syscall.SizeofInotifyEvent]byte
		_, err := syscall.Read(fd, buf[:])
		if err != nil {
			return err
		}
		r := bytes.NewReader(buf[:])
		var ev = syscall.InotifyEvent{}
		err = binary.Read(r, binary.LittleEndian, &ev)
		if err != nil {
			return err
		}
		if ev.Mask&syscall.IN_MODIFY == syscall.IN_MODIFY {
			return nil
		}
	}
}

type Notice struct {
	Recipient string
	Message   string
}

type Notifyer interface {
	Notify(msg string) error
}

type NotifyHttp struct {
	Email  string
	URL    string
	Client *http.Client
}

func NewNotifyHttp(email, url string) *NotifyHttp {
	return &NotifyHttp{
		Email:  email,
		URL:    url,
		Client: &http.Client{},
	}
}

func (n *NotifyHttp) Notify(msg string) error {
	notice := Notice{
		Recipient: n.Email,
		Message:   msg,
	}
	body := new(bytes.Buffer)
	err := json.NewEncoder(body).Encode(&notice)
	if err != nil {
		return err
	}
	req, err := http.NewRequest("POST", n.URL, body)
	if err != nil {
		return err
	}
	resp, err := n.Client.Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode != 200 {
		return fmt.Errorf("non-200 response: %s", resp.Status)
	}
	return nil
}

func ScanLogin(line string) string {
	if IsChatMessage(line) {
		return ""
	}
	// UUID of player ricky_ninja is 2ca2332f-429c-46b9-926f-03f9d7dc8049
	// ricky_ninja[/10.157.96.1:40090] logged in with entity id 257991 at (-166.82215288071876, 51.5, -342.1396285533262)
	// ricky_ninja joined the game
	if strings.HasSuffix(line, "joined the game") {
		f := strings.Fields(line)
		return f[0]
	}
	return ""
}

func ScanLogout(line string) string {
	if IsChatMessage(line) {
		return ""
	}
	// ricky_ninja lost connection: Disconnected
	// ricky_ninja left the game
	if strings.HasSuffix(line, "left the game") {
		f := strings.Fields(line)
		return f[0]
	}
	return ""
}

func IsChatMessage(line string) bool {
	// <ricky_ninja> rain keeps em from burning =/
	// user chat messages enclose username in angle brackets <>
	f := strings.Fields(line)
	if strings.HasPrefix(f[0], "<") && strings.HasSuffix(f[0], ">") {
		return true
	}
	return false
}

/*
minecraft@minecraft> cat worlds/rickyninja/logs/latest.log
[2020-02-08 16:10:39 MST] [INFO]: UUID of player ricky_ninja is 2ca2332f-429c-46b9-926f-03f9d7dc8049
[2020-02-08 16:10:39 MST] [INFO]: ricky_ninja[/10.157.96.1:33404] logged in with entity id 150207 at (-171.69999998807907, 51.0, -343.69999998807907)
[2020-02-08 16:10:39 MST] [INFO]: ricky_ninja joined the game
[2020-02-08 16:11:05 MST] [INFO]: <ricky_ninja> -156 64 -330
[2020-02-08 16:11:09 MST] [INFO]: ricky_ninja lost connection: Disconnected
[2020-02-08 16:11:09 MST] [INFO]: ricky_ninja left the game
*/

type LogLevel int

const (
	LogInfo LogLevel = iota
	LogWarn
)

type LogLine struct {
	Time     time.Time
	Level    LogLevel
	Line     string
	Location *time.Location
}

func NewLogLine(line, tz string) (*LogLine, error) {
	loc, err := time.LoadLocation(tz)
	if err != nil {
		return nil, err
	}
	l := &LogLine{
		//Time: ts,
		Location: loc,
	}
	err = l.Parse(line)
	if err != nil {
		return nil, err
	}
	return l, nil
}

func (l *LogLine) String() string {
	return fmt.Sprintf("%s: %s", l.Time, l.Line)
}

func (l *LogLine) Parse(line string) error {
	// [2020-02-08 16:10:39 MST] [INFO]: ricky_ninja joined the game
	parts := strings.SplitAfterN(line, "]: ", 2)
	if len(parts) != 2 {
		return fmt.Errorf("can't parse line: %s", line)
	}
	timeloglevel := strings.SplitAfterN(parts[0], "]", 2)
	if len(timeloglevel) != 2 {
		return fmt.Errorf("can't parse time: %s", parts[0])
	}
	timestr := removeBrackets(timeloglevel[0])
	// time.Parse will parse MST log line as UTC local time with GMT offset +0000, which
	// will do date math incorrectly.  I don't see a way to test this without changing system
	// time for the test.
	ts, err := time.ParseInLocation(timeFormat, timestr, l.Location)
	if err != nil {
		return err
	}
	l.Time = ts
	l.Line = parts[1]
	return nil
}

func removeBrackets(s string) string {
	s = strings.Replace(s, "[", "", -1)
	s = strings.Replace(s, "]", "", -1)
	return s
}
