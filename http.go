package main

import (
	"encoding/json"
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
	"strings"
	"time"
)

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
	err = os.RemoveAll(path.Join(dst, prefix+"_"+t.Format(timeFormat)+"_"+srcLastDir))
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
	err = filepath.WalkDir(src, func(p string, d fs.DirEntry, err error) error {
		if p == src {
			return nil
		}
		if excludeM[d.Name()] {
			logrus.Warn("[BackupUse]Ignore remove file:", d.Name())
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
	err = filepath.WalkDir(path.Join(dst, prefix+"_"+t.Format(timeFormat)+"_"+srcLastDir), func(p string, d fs.DirEntry, err error) error {
		if d.IsDir() {
			nName := strings.ReplaceAll(p, prefix+"_"+t.Format(timeFormat)+"_"+srcLastDir, srcLastDir)
			err = os.MkdirAll(nName, 0777)
			if err != nil {
				return err
			}
			return nil
		}
		if excludeM[d.Name()] {
			logrus.Warn("[BackupUse]Ignore clone file:", d.Name())
			return nil
		}
		bs, err := ioutil.ReadFile(p)
		if err != nil {
			return err
		}
		nName := strings.ReplaceAll(p, prefix+"_"+t.Format(timeFormat)+"_"+srcLastDir, srcLastDir)
		logrus.Info("[BackupUse]Clone", p, "to", nName)
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
	http.HandleFunc("/bk_list", backupList)
	http.HandleFunc("/bk_delete/", backupDelete)
	http.HandleFunc("/bk_use/", backupUse)
	logrus.Fatal(http.ListenAndServe(httpServerAddr, nil))
}

func renderBackupList(w io.Writer) error {
	dirs := make([]string, 0)
	err := filepath.WalkDir(dst, func(path string, d fs.DirEntry, err error) error {
		if d.IsDir() && strings.HasPrefix(d.Name(), prefix) {
			dirs = append(dirs, d.Name())
		}
		return nil
	})
	if err != nil {
		logrus.Error("[RenderBackupList]Walk dst dir: %v err: %v", dst, err)
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
	bs, _ := ioutil.ReadFile(path.Join(src, versionFile))
	_ = json.Unmarshal(bs, &versionJson)
	ver := cast.ToString(versionJson["version"])
	err = templateBackupList.Execute(w, map[string]interface{}{
		"Title":       fmt.Sprintf("Clone file: %v to %v", src, path.Join(dst, prefix+"_*_"+srcLastDir)),
		"Version":     lo.Ternary(ver == "", "-", ver),
		"TotalCnt":    len(dirs),
		"RefreshTime": time.Now().Format(timeFormat2),
		"Header":      []string{"File", "Time"},
		"Rows":        rows,
	})
	if err != nil {
		logrus.Error("[RenderBackupList]Exec backup list template err: %v", err)
		return err
	}
	return nil
}

func Style(version string, row []string) template.CSS {
	if row[1] == version {
		return "background: darkgreen;color: white;"
	}
	return "color: black;"
}
