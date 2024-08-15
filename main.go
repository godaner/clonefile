package main

import (
	"flag"
	"fmt"
	"github.com/duke-git/lancet/v2/fileutil"
	"github.com/samber/lo"
	"github.com/sirupsen/logrus"
	"io/fs"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime/debug"
	"sort"
	"strings"
	"time"
)

const (
	timeFormat = "2006_01_02_15_04_05"
)

var (
	src, dst, srcLastDir, exclude, prefix string
	interval, max                         int64
	excludeM                              map[string]bool
	showVersion                           bool
	gitHash                               string
	buildTime                             string
	goVersion                             string
)

func init() {
	flag.StringVar(&src, "s", "./", "src dir")  // /a/b
	flag.StringVar(&dst, "d", "../", "dst dir") // /a =>> /a/clonefile_2022_04_15_14_33_32_b
	flag.Int64Var(&interval, "i", 60, "interval, second")
	flag.Int64Var(&max, "m", 360, "max count")
	flag.StringVar(&prefix, "p", "f93851f4", "prefix")
	flag.StringVar(&exclude, "e", "clonefile,clonefile.exe", "exclude file, split by ,")
	flag.BoolVar(&showVersion, "v", false, "version info")
	flag.Parse()
}
func main() {
	version()
	checkFlag()
	initParam()
	initLog()
	go loopClone()
	go loopRemoveDir()
	select {}
}

func version() {
	if showVersion {
		fmt.Printf("Git Commit Hash: %s\n", gitHash)
		fmt.Printf("Build TimeStamp: %s\n", buildTime)
		fmt.Printf("GoLang Version: %s\n", goVersion)
		os.Exit(0)
	}
}
func checkFlag() {

	if src == "" || dst == "" || prefix == "" {
		flag.PrintDefaults()
		os.Exit(1)
	}
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
func initParam() {
	excludeM = lo.SliceToMap(strings.Split(exclude, ","), func(s string) (string, bool) {
		return s, true
	})
	var err error
	src, err = filepath.Abs(src)
	if err != nil {
		logrus.Error("[Initing]Get src abs path err:", err)
		os.Exit(1)
	}
	// /Users/godaner/gomod/clonefile/bin/darwin-arm64
	logrus.Info("[Initing]Src abs path:", src)
	if !fileutil.IsDir(src) {
		logrus.Error("[Initing]Src abs path is not dir")
		os.Exit(1)
	}
	// darwin-arm64
	srcLastDir = filepath.Base(src)
	logrus.Info("[Initing]Src abs last dir:", srcLastDir)

	dst, err = filepath.Abs(dst)
	if err != nil {
		logrus.Error("[Initing]Get dst abs path err:", err)
		os.Exit(1)
	}
	// /Users/godaner/gomod/clonefile/bin
	logrus.Info("[Initing]Dst abs path:", dst)

	if !fileutil.IsDir(dst) {
		logrus.Error("[Initing]Dst abs path is not dir")
		os.Exit(1)
	}
}

func loopRemoveDir() {
	for {
		removeDir()
		<-time.After(time.Duration(1) * time.Second)
	}
}
func loopClone() {
	for {
		cloneFile()
		<-time.After(time.Duration(interval) * time.Second)
	}
}
func removeDir() {
	defer func() {
		if err := recover(); err != nil {
			logrus.Infof("[Removing]Recover remove dir err: %v, %v\n", err, string(debug.Stack()))
		}
	}()
	dirs := sort.StringSlice{}
	filepath.Dir(dst)
	err := filepath.WalkDir(dst, func(path string, d fs.DirEntry, err error) error {
		if !strings.Contains(path, prefix) {
			return nil
		}
		ps := strings.Split(path, string(filepath.Separator))
		if len(ps) <= 0 {
			return nil
		}
		if !strings.Contains(ps[len(ps)-1], prefix) {
			return nil
		}
		dirs = append(dirs, path)
		return nil
	})
	if err != nil {
		logrus.Errorf("[Removing]Walk remove dir: %v err: %v\n", dst, err)
		return
	}
	dirs.Sort()
	m := dirs.Len() - int(max)
	if m <= 0 {
		return
	}
	delDirs := (([]string)(dirs))[:m]
	for _, dd := range delDirs {
		logrus.Info("[Removing]Removing dir:", dd)
		err = os.RemoveAll(dd)
		if err != nil {
			logrus.Errorf("[Removing]Remove dir: %v err: %v\n", dd, err)
			continue
		}
		logrus.Infof("[Removing]Remove dir: %v success\n", dd)
	}
}

func cloneFile() {
	defer func() {
		if err := recover(); err != nil {
			logrus.Infof("[Cloning]Recover clone file err: %v, %v\n", err, string(debug.Stack()))
		}
	}()
	logrus.Info("[Cloning]...")
	defer func() {
		logrus.Info("[Cloning]Finish!")
	}()
	ts := time.Now().Format(timeFormat)
	err := filepath.WalkDir(src, func(path string, d fs.DirEntry, err error) error {
		if d.IsDir() {
			nName := strings.ReplaceAll(path, src, dst+string(filepath.Separator)+prefix+"_"+ts+"_"+srcLastDir)
			err = os.Mkdir(nName, 0777)
			if err != nil {
				return err
			}
			return nil
		}
		if excludeM[d.Name()] {
			logrus.Warn("[Cloning]Ignore file:", d.Name())
			return nil
		}
		bs, err := ioutil.ReadFile(path)
		if err != nil {
			return err
		}
		nName := strings.ReplaceAll(path, src, dst+string(filepath.Separator)+prefix+"_"+ts+"_"+srcLastDir)
		logrus.Info("[Cloning]Clone", path, "to", nName)
		err = ioutil.WriteFile(nName, bs, 0777)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		logrus.Error("[Cloning]Walk clone dir: %v err: %v\n", dst, err)
	}
}
