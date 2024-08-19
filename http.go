package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/duke-git/lancet/v2/fileutil"
	"github.com/duke-git/lancet/v2/formatter"
	"github.com/samber/lo"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cast"
	"golang.org/x/exp/slices"
	"html/template"
	"io"
	"io/fs"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"sort"
	"strings"
	"time"
)

const (
	configFileName = ".clonefile_config"
	minRefresh     = 60
)

var conf *config
var exitCh chan struct{}

func init() {
	exitCh = make(chan struct{}, 0)
	close(exitCh)

	conf = new(config)
	err = conf.load()
	if err != nil {
		logrus.Fatalf("Load config err: %v", err)
	}
	err = conf.validate()
	if err != nil {
		logrus.Fatalf("Validate config err: %v", err)
	}
}
func backupList(w http.ResponseWriter, r *http.Request) {
	err = renderBackupList(w)
	if err != nil {
		_, _ = w.Write([]byte(err.Error()))
		return
	}
}
func backupDelete(w http.ResponseWriter, r *http.Request) {
	var err error
	defer func() {
		errMsg := "Success"
		if err != nil {
			errMsg = err.Error()
		}
		http.Redirect(w, r, "/bk_list?errMsg="+errMsg, http.StatusMovedPermanently)
	}()
	version := strings.TrimPrefix(r.URL.Path, "/bk_delete/")
	t, err := time.Parse(timeFormat2, version)
	if err != nil {
		err = fmt.Errorf("[BackupDelete]Parse version: %v to format: %v err: %v", version, timeFormat2, err)
		return
	}
	err = os.RemoveAll(path.Join(conf.DstAbs, conf.Prefix+"_"+t.Format(timeFormat)+"_"+conf.SrcLastDir))
	if err != nil {
		return
	}
}
func backupUse(w http.ResponseWriter, r *http.Request) {
	var err error
	defer func() {
		errMsg := "Success"
		if err != nil {
			errMsg = err.Error()
		}
		http.Redirect(w, r, "/bk_list?errMsg="+errMsg, http.StatusMovedPermanently)
	}()
	version := strings.TrimPrefix(r.URL.Path, "/bk_use/")
	t, err := time.Parse(timeFormat2, version)
	if err != nil {
		err = fmt.Errorf("[BackupUse]Parse version: %v to format: %v err: %v", version, timeFormat2, err)
		return
	}

	// remove file
	err = filepath.WalkDir(conf.SrcAbs, func(p string, d fs.DirEntry, err error) error {
		if p == conf.SrcAbs {
			return nil
		}
		if conf.ExcludeM[d.Name()] {
			logrus.Warnln("[BackupUse]Ignore remove file:", d.Name())
			return nil
		}
		if d.IsDir() {
			return os.RemoveAll(p)
		}
		return fileutil.RemoveFile(p)
	})
	if err != nil {
		return
	}

	// copy
	err = filepath.WalkDir(path.Join(conf.DstAbs, conf.Prefix+"_"+t.Format(timeFormat)+"_"+conf.SrcLastDir), func(p string, d fs.DirEntry, err error) error {
		if d.IsDir() {
			nName := strings.ReplaceAll(p, conf.Prefix+"_"+t.Format(timeFormat)+"_"+conf.SrcLastDir, conf.SrcLastDir)
			err = os.MkdirAll(nName, 0777)
			if err != nil {
				return err
			}
			return nil
		}
		if conf.ExcludeM[d.Name()] {
			logrus.Warnln("[BackupUse]Ignore clone file:", d.Name())
			return nil
		}
		bs, err := ioutil.ReadFile(p)
		if err != nil {
			return err
		}
		nName := strings.ReplaceAll(p, conf.Prefix+"_"+t.Format(timeFormat)+"_"+conf.SrcLastDir, conf.SrcLastDir)
		logrus.Infoln("[BackupUse]Clone", p, "to", nName)
		err = ioutil.WriteFile(nName, bs, 0777)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return
	}
}
func httpServer() {
	http.HandleFunc("/", backupList)
	http.HandleFunc("/cf_set", setConfig)
	http.HandleFunc("/start", start)
	http.HandleFunc("/stop", stop)
	http.HandleFunc("/clone", clone)
	http.HandleFunc("/bk_list", backupList)
	http.HandleFunc("/bk_delete/", backupDelete)
	http.HandleFunc("/bk_use/", backupUse)
	http.HandleFunc("/browser_file/", browserFile)
	logrus.Fatal(http.ListenAndServe(httpServerAddr, nil))
}

func browserFile(w http.ResponseWriter, r *http.Request) {
	var err error
	defer func() {
		errMsg := "Success"
		if err != nil {
			errMsg = err.Error()
		}
		http.Redirect(w, r, "/bk_list?errMsg="+errMsg, http.StatusMovedPermanently)
	}()
	// 获取当前系统类型
	var openCmd string
	switch runtime.GOOS {
	case "windows":
		openCmd = "explorer.exe"
	case "darwin":
		openCmd = "open"
	case "linux":
		openCmd = "xdg-open"
	default:
		err = errors.New("unsupported platform")
		return
	}

	version := strings.TrimPrefix(r.URL.Path, "/browser_file/")
	t, err := time.Parse(timeFormat2, version)
	if err != nil {
		err = fmt.Errorf("[BrowserFile]Parse version: %v to format: %v err: %v", version, timeFormat2, err)
		return
	}
	// 打开文件浏览器
	err = exec.Command(openCmd, path.Join(conf.Dst, conf.Prefix+"_"+t.Format(timeFormat)+"_"+conf.SrcLastDir)).Start()
	if err != nil {
		return
	}
}

func clone(w http.ResponseWriter, r *http.Request) {
	var err error
	defer func() {
		errMsg := "Success"
		if err != nil {
			errMsg = err.Error()
		}
		http.Redirect(w, r, "/bk_list?errMsg="+errMsg, http.StatusMovedPermanently)
	}()
	cloneFile()
}
func stop(w http.ResponseWriter, r *http.Request) {
	var err error
	defer func() {
		errMsg := "Success"
		if err != nil {
			errMsg = err.Error()
		}
		http.Redirect(w, r, "/bk_list?errMsg="+errMsg, http.StatusMovedPermanently)
	}()
	select {
	case <-exitCh:
		err = errors.New("already stopped")
		return
	default:
		close(exitCh)
	}
}

func start(w http.ResponseWriter, r *http.Request) {
	var err error
	defer func() {
		errMsg := "Success"
		if err != nil {
			errMsg = err.Error()
		}
		http.Redirect(w, r, "/bk_list?errMsg="+errMsg, http.StatusMovedPermanently)
	}()
	select {
	case <-exitCh:
		exitCh = make(chan struct{})
		go loopClone()
		go loopRemoveDir()
	default:
		err = errors.New("already running")
		return
	}
}

type config struct {
	Src        string          `json:"src"`
	Dst        string          `json:"dst"`
	Interval   int64           `json:"interval"`
	MaxCount   int64           `json:"max_count"`
	Prefix     string          `json:"prefix"`
	Exclude    string          `json:"exclude"`
	Refresh    int64           `json:"refresh"` // web刷新时间
	SrcAbs     string          `json:"-"`
	DstAbs     string          `json:"-"`
	SrcLastDir string          `json:"-"`
	ExcludeM   map[string]bool `json:"-"`
}

func (c *config) save() error {
	bs, _ := json.Marshal(c)
	err = ioutil.WriteFile(configFileName, bs, 0777)
	if err != nil {
		return err
	}
	return nil
}

func (c *config) load() error {
	if !fileutil.IsExist(configFileName) {
		*c = config{
			Src:      "./",
			Dst:      "../",
			Interval: 60,
			MaxCount: 360,
			Prefix:   "f93851f4",
			Exclude:  "clonefile,clonefile.exe," + configFileName,
			Refresh:  minRefresh,
		}
		_ = c.save()
		return nil
	}
	bs, err := ioutil.ReadFile(configFileName)
	if err != nil {
		return err
	}
	err = json.Unmarshal(bs, c)
	if err != nil {
		return err
	}
	return nil
}
func (c *config) validate() error {
	c.ExcludeM = lo.SliceToMap(strings.Split(c.Exclude, ","), func(s string) (string, bool) {
		return s, true
	})
	c.SrcAbs, err = filepath.Abs(c.Src)
	if err != nil {
		return fmt.Errorf("get src abs path err: %v", err)
	}
	// /Users/godaner/gomod/clonefile/bin/darwin-arm64
	if !fileutil.IsDir(c.SrcAbs) {
		return errors.New("src abs path is not dir")
	}
	// darwin-arm64
	c.SrcLastDir = filepath.Base(c.SrcAbs)
	c.DstAbs, err = filepath.Abs(c.Dst)
	if err != nil {
		return fmt.Errorf("get dst abs path err: %v", err)
	}
	// /Users/godaner/gomod/clonefile/bin
	if !fileutil.IsDir(c.DstAbs) {
		return errors.New("dst abs path is not dir")
	}
	if c.Refresh < minRefresh {
		return fmt.Errorf("refresh must>=%v", minRefresh)
	}
	return nil
}
func setConfig(w http.ResponseWriter, r *http.Request) {
	var err error
	defer func() {
		errMsg := "Success"
		if err != nil {
			errMsg = err.Error()
		}
		http.Redirect(w, r, "/bk_list?errMsg="+errMsg, http.StatusMovedPermanently)
	}()
	// 解析表单数据
	_ = r.ParseForm()
	// 获取表单字段值
	s := r.Form.Get("s")
	d := r.Form.Get("d")
	i := cast.ToInt64(r.Form.Get("i"))
	m := cast.ToInt64(r.Form.Get("m"))
	re := cast.ToInt64(r.Form.Get("r"))
	p := r.Form.Get("p")
	e := r.Form.Get("e")
	nc := &config{
		Src:      s,
		Dst:      d,
		Interval: i,
		MaxCount: m,
		Prefix:   p,
		Exclude:  e,
		Refresh:  re,
	}
	err = nc.validate()
	if err != nil {
		return
	}
	err = nc.save()
	if err != nil {
		return
	}
	conf = nc
	select {
	case <-exitCh:
	default:
		close(exitCh)
		exitCh = make(chan struct{})
		go loopClone()
		go loopRemoveDir()
	}

}

func renderBackupList(w io.Writer) error {
	dirs := make([]string, 0)
	err = filepath.WalkDir(conf.DstAbs, func(path string, d fs.DirEntry, err error) error {
		if d.IsDir() &&
			strings.HasPrefix(d.Name(), conf.Prefix) &&
			strings.HasSuffix(d.Name(), "_"+conf.SrcLastDir) {
			dirs = append(dirs, d.Name())
		}
		return nil
	})
	if err != nil {
		logrus.Errorf("[RenderBackupList]Walk dst dir: %v err: %v", conf.DstAbs, err)
	}
	rows := lo.Map(dirs, func(item string, index int) []string {
		ts := item[:strings.LastIndex(item, "_")]
		ts = ts[strings.Index(ts, "_")+1:]
		t, _ := time.Parse(`2006_01_02_15_04_05`, ts)
		fileCount, dirSize := countFilesAndSize(path.Join(conf.DstAbs, item))
		return []string{item, t.Format(timeFormat2), cast.ToString(fileCount), formatter.BinaryBytes(float64(dirSize))}
	})
	slices.SortFunc(rows, func(a, b []string) bool {
		return a[1] > b[1]
	})
	versionJson := map[string]any{}
	bs, _ := ioutil.ReadFile(path.Join(conf.SrcAbs, versionFile))
	_ = json.Unmarshal(bs, &versionJson)
	ver := cast.ToString(versionJson["version"])
	nexState := "Stop"
	select {
	case <-exitCh:
		nexState = "Start"
	default:

	}
	nextBackupIn := int64(0)
	if nexState == "Stop" {
		if lastBackupTime.IsZero() {
			nextBackupIn = conf.Interval
		} else {
			nextBackupIn = int64(lastBackupTime.Add(time.Duration(conf.Interval)*time.Second).Sub(time.Now()).Seconds()) + 1
		}
	}

	err = templateBackupList.Execute(w, map[string]interface{}{
		"Title":        fmt.Sprintf("Clone file: %v to %v", conf.SrcAbs, path.Join(conf.DstAbs, conf.Prefix+"_*_"+conf.SrcLastDir)),
		"SfVersion":    versionString(),
		"Version":      lo.Ternary(ver == "", "-", ver),
		"TotalCnt":     len(dirs),
		"NextBackupIn": nextBackupIn,
		"NextState":    nexState,
		"Conf":         conf,
		"Header":       []string{"Dir", "Time", "FileCount", "DirSize"},
		"Rows":         rows,
	})
	if err != nil {
		logrus.Errorf("[RenderBackupList]Exec backup list template err: %v", err)
		return err
	}
	return nil
}

func StateStyle(state string) template.CSS {
	if state == "Start" {
		return "color: green;"
	}
	return "color: red;"
}
func Style(version string, row []string) template.CSS {
	if row[1] == version {
		return "background: darkgreen;color: white;"
	}
	return "color: black;"
}

func loopRemoveDir() {
	for {
		select {
		case <-exitCh:
			return
		default:
			removeDir()
			<-time.After(time.Duration(1) * time.Second)
		}
	}
}
func loopClone() {
	for {
		select {
		case <-exitCh:
			return
		default:
			cloneFile()
			<-time.After(time.Duration(conf.Interval) * time.Second)
		}

	}
}
func removeDir() {
	defer func() {
		if err := recover(); err != nil {
			logrus.Infof("[Removing]Recover remove dir err: %v, %v", err, string(debug.Stack()))
		}
	}()
	dirs := sort.StringSlice{}
	err := filepath.WalkDir(conf.DstAbs, func(path string, d fs.DirEntry, err error) error {
		if !strings.Contains(path, conf.Prefix) {
			return nil
		}
		ps := strings.Split(path, string(filepath.Separator))
		if len(ps) <= 0 {
			return nil
		}
		if !strings.Contains(ps[len(ps)-1], conf.Prefix) {
			return nil
		}
		dirs = append(dirs, path)
		return nil
	})
	if err != nil {
		logrus.Errorf("[Removing]Walk remove dir: %v err: %v", conf.DstAbs, err)
		return
	}
	dirs.Sort()
	m := dirs.Len() - int(conf.MaxCount)
	if m <= 0 {
		return
	}
	delDirs := (([]string)(dirs))[:m]
	for _, dd := range delDirs {
		logrus.Infoln("[Removing]Removing dir:", dd)
		err = os.RemoveAll(dd)
		if err != nil {
			logrus.Errorf("[Removing]Remove dir: %v err: %v", dd, err)
			continue
		}
		logrus.Infof("[Removing]Remove dir: %v success", dd)
	}
}

var lastBackupTime time.Time

func cloneFile() {
	defer func() {
		if err := recover(); err != nil {
			logrus.Infof("[Cloning]Recover clone file err: %v, %v", err, string(debug.Stack()))
		}
	}()
	defer func() {
		lastBackupTime = time.Now()
	}()
	logrus.Infoln("[Cloning]...")
	defer func() {
		logrus.Infoln("[Cloning]Finish!")
	}()
	now := time.Now()
	ts := now.Format(timeFormat)
	err := filepath.WalkDir(conf.SrcAbs, func(p string, d fs.DirEntry, err error) error {
		if d.IsDir() {
			nName := strings.ReplaceAll(p, conf.SrcAbs, path.Join(conf.DstAbs, conf.Prefix+"_"+ts+"_"+conf.SrcLastDir))
			err = os.Mkdir(nName, 0777)
			if err != nil {
				return err
			}
			return nil
		}
		if conf.ExcludeM[d.Name()] {
			logrus.Warnln("[Cloning]Ignore clone file:", d.Name())
			return nil
		}
		bs, err := ioutil.ReadFile(p)
		if err != nil {
			return err
		}
		nName := strings.ReplaceAll(p, conf.SrcAbs, path.Join(conf.DstAbs, conf.Prefix+"_"+ts+"_"+conf.SrcLastDir))
		logrus.Infoln("[Cloning]Clone", p, "to", nName)
		err = ioutil.WriteFile(nName, bs, 0777)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		logrus.Errorln("[Cloning]Walk clone dir: %v err: %v", conf.DstAbs, err)
	} else {
		versionJson, _ := json.Marshal(map[string]any{
			"version": now.Format(timeFormat2),
		})
		versionFile := path.Join(conf.DstAbs, conf.Prefix+"_"+ts+"_"+conf.SrcLastDir, versionFile)
		err = ioutil.WriteFile(versionFile, versionJson, 0777)
		if err != nil {
			logrus.Errorln("[Cloning]Write config json: %v err: %v", versionFile, err)
		}
	}
}

func countFilesAndSize(dirPath string) (int, int64) {
	var fileCount int
	var totalSize int64

	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() {
			fileCount++
			totalSize += info.Size()
		}

		return nil
	})

	if err != nil {
		logrus.Errorln("Error:", err)
		return 0, 0
	}

	return fileCount, totalSize
}
