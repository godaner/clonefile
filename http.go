package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/duke-git/lancet/v2/fileutil"
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
	"path"
	"path/filepath"
	"runtime/debug"
	"sort"
	"strings"
	"time"
)

const (
	configFileName = ".clonefile_config"
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
	err = os.RemoveAll(path.Join(conf.Dst, conf.P+"_"+t.Format(timeFormat)+"_"+conf.SrcLastDir))
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
	err = filepath.WalkDir(conf.Src, func(p string, d fs.DirEntry, err error) error {
		if p == conf.Src {
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
	err = filepath.WalkDir(path.Join(conf.Dst, conf.P+"_"+t.Format(timeFormat)+"_"+conf.SrcLastDir), func(p string, d fs.DirEntry, err error) error {
		if d.IsDir() {
			nName := strings.ReplaceAll(p, conf.P+"_"+t.Format(timeFormat)+"_"+conf.SrcLastDir, conf.SrcLastDir)
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
		nName := strings.ReplaceAll(p, conf.P+"_"+t.Format(timeFormat)+"_"+conf.SrcLastDir, conf.SrcLastDir)
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
	http.HandleFunc("/bk_list", backupList)
	http.HandleFunc("/bk_delete/", backupDelete)
	http.HandleFunc("/bk_use/", backupUse)
	logrus.Fatal(http.ListenAndServe(httpServerAddr, nil))
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
	S          string          `json:"s"`
	D          string          `json:"d"`
	I          int64           `json:"i"`
	M          int64           `json:"m"`
	P          string          `json:"p"`
	E          string          `json:"e"`
	Src        string          `json:"-"`
	Dst        string          `json:"-"`
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
			S: "./",
			D: "../",
			I: 60,
			M: 360,
			P: "f93851f4",
			E: "clonefile,clonefile.exe",
		}
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
	c.ExcludeM = lo.SliceToMap(strings.Split(c.E, ","), func(s string) (string, bool) {
		return s, true
	})
	c.Src, err = filepath.Abs(c.S)
	if err != nil {
		return fmt.Errorf("get src abs path err: %v", err)
	}
	// /Users/godaner/gomod/clonefile/bin/darwin-arm64
	if !fileutil.IsDir(c.Src) {
		return errors.New("src abs path is not dir")
	}
	// darwin-arm64
	c.SrcLastDir = filepath.Base(c.Src)
	c.Dst, err = filepath.Abs(c.D)
	if err != nil {
		return fmt.Errorf("get dst abs path err: %v", err)
	}
	// /Users/godaner/gomod/clonefile/bin
	if !fileutil.IsDir(c.D) {
		return errors.New("dst abs path is not dir")
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
	p := r.Form.Get("p")
	e := r.Form.Get("e")
	nc := &config{
		S: s,
		D: d,
		I: i,
		M: m,
		P: p,
		E: e,
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
	err = filepath.WalkDir(conf.Dst, func(path string, d fs.DirEntry, err error) error {
		if d.IsDir() &&
			strings.HasPrefix(d.Name(), conf.P) &&
			strings.HasSuffix(d.Name(), "_"+conf.SrcLastDir) {
			dirs = append(dirs, d.Name())
		}
		return nil
	})
	if err != nil {
		logrus.Errorf("[RenderBackupList]Walk dst dir: %v err: %v", conf.Dst, err)
	}
	rows := lo.Map(dirs, func(item string, index int) []string {
		ts := item[:strings.LastIndex(item, "_")]
		ts = ts[strings.Index(ts, "_")+1:]
		t, _ := time.Parse(`2006_01_02_15_04_05`, ts)
		return []string{item, t.Format(timeFormat2)}
	})
	slices.SortFunc(rows, func(a, b []string) bool {
		return a[1] > b[1]
	})
	versionJson := map[string]any{}
	bs, _ := ioutil.ReadFile(path.Join(conf.Src, versionFile))
	_ = json.Unmarshal(bs, &versionJson)
	ver := cast.ToString(versionJson["version"])
	nexState := "Stop"
	select {
	case <-exitCh:
		nexState = "Start"
	default:

	}
	err = templateBackupList.Execute(w, map[string]interface{}{
		"Title":       fmt.Sprintf("Clone file: %v to %v", conf.Src, path.Join(conf.Dst, conf.P+"_*_"+conf.SrcLastDir)),
		"SfVersion":   versionString(),
		"Version":     lo.Ternary(ver == "", "-", ver),
		"TotalCnt":    len(dirs),
		"State":       nexState,
		"Conf":        conf,
		"RefreshTime": time.Now().Format(timeFormat2),
		"Header":      []string{"File", "Time"},
		"Rows":        rows,
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
			<-time.After(time.Duration(conf.I) * time.Second)
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
	err := filepath.WalkDir(conf.Dst, func(path string, d fs.DirEntry, err error) error {
		if !strings.Contains(path, conf.P) {
			return nil
		}
		ps := strings.Split(path, string(filepath.Separator))
		if len(ps) <= 0 {
			return nil
		}
		if !strings.Contains(ps[len(ps)-1], conf.P) {
			return nil
		}
		dirs = append(dirs, path)
		return nil
	})
	if err != nil {
		logrus.Errorf("[Removing]Walk remove dir: %v err: %v", conf.Dst, err)
		return
	}
	dirs.Sort()
	m := dirs.Len() - int(conf.M)
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

func cloneFile() {
	defer func() {
		if err := recover(); err != nil {
			logrus.Infof("[Cloning]Recover clone file err: %v, %v", err, string(debug.Stack()))
		}
	}()
	logrus.Infoln("[Cloning]...")
	defer func() {
		logrus.Infoln("[Cloning]Finish!")
	}()
	now := time.Now()
	ts := now.Format(timeFormat)
	err := filepath.WalkDir(conf.Src, func(p string, d fs.DirEntry, err error) error {
		if d.IsDir() {
			nName := strings.ReplaceAll(p, conf.Src, path.Join(conf.Dst, conf.P+"_"+ts+"_"+conf.SrcLastDir))
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
		nName := strings.ReplaceAll(p, conf.Src, path.Join(conf.Dst, conf.P+"_"+ts+"_"+conf.SrcLastDir))
		logrus.Infoln("[Cloning]Clone", p, "to", nName)
		err = ioutil.WriteFile(nName, bs, 0777)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		logrus.Errorln("[Cloning]Walk clone dir: %v err: %v", conf.Dst, err)
	} else {
		versionJson, _ := json.Marshal(map[string]any{
			"version": now.Format(timeFormat2),
		})
		versionFile := path.Join(conf.Dst, conf.P+"_"+ts+"_"+conf.SrcLastDir, versionFile)
		err = ioutil.WriteFile(versionFile, versionJson, 0777)
		if err != nil {
			logrus.Errorln("[Cloning]Write config json: %v err: %v", versionFile, err)
		}
	}
}
