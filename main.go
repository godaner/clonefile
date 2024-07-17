package main

import (
	"flag"
	"fmt"
	"github.com/duke-git/lancet/v2/fileutil"
	"github.com/samber/lo"
	"io/fs"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

const (
	timeFormat = "2006_01_02_15_04_05"
)

var src, dst, srcLastDir, exclude, prefix string
var interval, max int64
var excludeM map[string]bool

func init() {
	flag.StringVar(&src, "s", "./", "src dir")  // /a/b
	flag.StringVar(&dst, "d", "../", "dst dir") // /a =>> /a/clonefile_2022_04_15_14_33_32_b
	flag.Int64Var(&interval, "i", 60, "interval, second")
	flag.Int64Var(&max, "m", 1000, "max count")
	flag.StringVar(&prefix, "p", "f93851f4", "prefix")
	flag.StringVar(&exclude, "e", "clonefile,clonefile.exe", "exclude file, split by ,")
	flag.Parse()
}
func main() {
	checkFlag()
	initParam()
	go loopClone()
	go loopDelete()
	select {}
}

func checkFlag() {
	if src == "" || dst == "" || prefix == "" {
		flag.PrintDefaults()
		os.Exit(1)
	}
}

func initParam() {
	excludeM = lo.SliceToMap(strings.Split(exclude, ","), func(s string) (string, bool) {
		return s, true
	})
	var err error
	src, err = filepath.Abs(src)
	if err != nil {
		fmt.Println("Get src abs path err:", err)
		os.Exit(1)
	}
	// /Users/godaner/gomod/clonefile/bin/darwin-arm64
	fmt.Println("Src abs path:", src)
	if !fileutil.IsDir(src) {
		fmt.Println("Src abs path is not dir")
		os.Exit(1)
	}
	// darwin-arm64
	srcLastDir = filepath.Base(src)
	fmt.Println("Src abs last dir:", srcLastDir)

	dst, err = filepath.Abs(dst)
	if err != nil {
		fmt.Println("Get dst abs path err:", err)
		os.Exit(1)
	}
	// /Users/godaner/gomod/clonefile/bin
	fmt.Println("Dst abs path:", dst)

	if !fileutil.IsDir(dst) {
		fmt.Println("Dst abs path is not dir")
		os.Exit(1)
	}
}

func loopDelete() {
	for {
		func() {
			defer func() {
				if err := recover(); err != nil {
					fmt.Println("Recover delete err:", err)
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
				fmt.Printf("Walk dir: %v err: %v\n", dst, err)
				return
			}
			dirs.Sort()
			m := dirs.Len() - int(max)
			if m <= 0 {
				return
			}
			delDirs := (([]string)(dirs))[:m]
			for _, dd := range delDirs {
				fmt.Println("Removing dir:", dd)
				err = os.RemoveAll(dd)
				if err != nil {
					fmt.Printf("Remove dir: %v err: %v\n", dd, err)
					continue
				}
				fmt.Printf("Remove dir: %v success\n", dd)
			}
		}()
		<-time.After(time.Duration(1) * time.Second)
	}
}

func loopClone() {
	for {
		func() {
			defer func() {
				if err := recover(); err != nil {
					fmt.Println("Recover clone file err:", err)
				}
			}()
			err := clonefile()
			if err != nil {
				fmt.Println("Clone file err:", err)
				return
			}
		}()
		<-time.After(time.Duration(interval) * time.Second)
	}
}

func clonefile() error {
	fmt.Println("Cloning...")
	defer func() {
		fmt.Println("Finish!")
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
			fmt.Println("Ignore file:", d.Name())
			return nil
		}
		bs, err := ioutil.ReadFile(path)
		if err != nil {
			return err
		}
		nName := strings.ReplaceAll(path, src, dst+string(filepath.Separator)+prefix+"_"+ts+"_"+srcLastDir)
		fmt.Println("Clone", path, "to", nName)
		err = ioutil.WriteFile(nName, bs, 0777)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}
