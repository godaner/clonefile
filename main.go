package main

import (
	"flag"
	"fmt"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"html/template"
	"os"
	"os/exec"
	"runtime"
)

const (
	timeFormat  = "2006_01_02_15_04_05"
	timeFormat2 = "2006-01-02 15:04:05"
	versionFile = ".clonefile_version"
)

var (
	gitHash   string
	buildTime string
	goVersion string
)
var (
	showVersion    bool
	httpServerAddr string
)
var (
	templateBackupList *template.Template
	err                error
)

func init() {
	flag.BoolVar(&showVersion, "v", false, "version info")
	flag.StringVar(&httpServerAddr, "h", "127.0.0.1:31555", "http server address")
	flag.Parse()
}
func main() {
	version()
	initLog()
	checkFlag()
	initTemplate()
	openBrowser()
	go httpServer()
	select {}
}

func initTemplate() {
	templateBackupList, err = template.New("").Funcs(template.FuncMap{
		"StateStyle": StateStyle,
		"Style":      Style,
		"UUID": func() string {
			return uuid.NewString()
		},
	}).Parse(templateBackupListHtml)
	if err != nil {
		logrus.Fatalf("[InitTemplate]Parse backup list template html err: %v", err)
	}
}

func openBrowser() {
	var cmd string
	var args []string
	url := "http://" + httpServerAddr
	switch runtime.GOOS {
	case "windows":
		cmd = "cmd"
		args = []string{"/c", "start"}
	case "darwin":
		cmd = "open"
	default: // "linux", "freebsd", "openbsd", "netbsd"
		cmd = "xdg-open"
	}
	args = append(args, url)
	err = exec.Command(cmd, args...).Start()
	if err != nil {
		logrus.Errorf("[OpenBrowser]Open browser err: %v", err)
	}
}

func versionString() string {
	return fmt.Sprintf("Git Commit Hash: %s\nBuild TimeStamp: %s\nGoLang Version: %s\n", gitHash, buildTime, goVersion)
}
func version() {
	if showVersion {
		fmt.Println(versionString())
		os.Exit(0)
	}
}
func checkFlag() {

}

func initLog() {
	// Log as JSON instead of the default ASCII formatter.
	logrus.SetFormatter(&logrus.TextFormatter{
		ForceColors:      true,
		DisableTimestamp: true,
	})
	//logrus.SetReportCaller(true)

	// Output to stdout instead of the default stderr
	// Can be any io.Writer, see below for File example
	logrus.SetOutput(os.Stdout)

	// Only log the warning severity or above.
	logrus.SetLevel(logrus.DebugLevel)
}
