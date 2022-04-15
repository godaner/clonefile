package main

import (
	"flag"
	"fmt"
	"io/fs"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

var src, dst, exclude, prefix string
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
	if src == "" || dst == "" || prefix == "" {
		flag.PrintDefaults()
		os.Exit(1)
	}
	excludeM = map[string]bool{}
	excludes := strings.Split(exclude, ",")
	for _, exc := range excludes {
		excludeM[exc] = true
	}
	sn, err := filepath.Abs(src)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	// /Users/godaner/gomod/clonefile/bin/darwin-arm64
	fmt.Println("src abs path", sn)
	snSplit := strings.Split(sn, string(filepath.Separator))
	sName := snSplit[len(snSplit)-1]
	// darwin-arm64
	fmt.Println("src abs dir", sName)
	dn, err := filepath.Abs(dst)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	// /Users/godaner/gomod/clonefile/bin/
	fmt.Println("dst abs path", dn)
	go func() {
		for {
			go func() {
				err = clonefile(sn, dn, sName)
				if err != nil {
					fmt.Println(err)
					os.Exit(1)
				}
			}()
			<-time.After(time.Duration(interval) * time.Second)
		}
	}()
	go func() {
		for {
			go func() {
				dirs := sort.StringSlice{}
				filepath.Dir(dst)
				err = filepath.WalkDir(dst, func(path string, d fs.DirEntry, err error) error {
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
					fmt.Println(err)
					os.Exit(1)
				}
				dirs.Sort()
				m := dirs.Len() - int(max)
				if m <= 0 {
					return
				}
				delDirs := (([]string)(dirs))[:m]
				for _, dd := range delDirs {
					fmt.Println("Delete dir", dd)
					err = os.RemoveAll(dd)
					if err != nil {
						fmt.Println(err)
					}
				}
			}()
			<-time.After(time.Duration(1) * time.Second)
		}
	}()
	select {}
}

func clonefile(src string, dst string, name string) error {
	fmt.Println("Clone...")
	defer func() {
		fmt.Println("Finish!")
	}()
	ts := time.Now().Format("2006_01_02_15_04_05")
	err := filepath.WalkDir(src, func(path string, d fs.DirEntry, err error) error {
		if d.IsDir() {
			nName := strings.ReplaceAll(path, src, dst+string(filepath.Separator)+prefix+"_"+ts+"_"+name)
			err = os.Mkdir(nName, 0777)
			if err != nil {
				return err
			}
			return nil
		}
		if excludeM[d.Name()] {
			fmt.Println("Ignore file", d.Name())
			return nil
		}
		bs, err := ioutil.ReadFile(path)
		if err != nil {
			return err
		}
		nName := strings.ReplaceAll(path, src, dst+string(filepath.Separator)+prefix+"_"+ts+"_"+name)
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
