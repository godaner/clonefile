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
    "runtime/debug"
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
    flag.StringVar(&exclude, "e", "cloneFile,cloneFile.exe", "exclude file, split by ,")
    flag.Parse()
}
func main() {
    checkFlag()
    initParam()
    go loopClone()
    go loopRemoveDir()
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
        fmt.Println("[Initing]Get src abs path err:", err)
        os.Exit(1)
    }
    // /Users/godaner/gomod/cloneFile/bin/darwin-arm64
    fmt.Println("[Initing]Src abs path:", src)
    if !fileutil.IsDir(src) {
        fmt.Println("[Initing]Src abs path is not dir")
        os.Exit(1)
    }
    // darwin-arm64
    srcLastDir = filepath.Base(src)
    fmt.Println("[Initing]Src abs last dir:", srcLastDir)

    dst, err = filepath.Abs(dst)
    if err != nil {
        fmt.Println("[Initing]Get dst abs path err:", err)
        os.Exit(1)
    }
    // /Users/godaner/gomod/cloneFile/bin
    fmt.Println("[Initing]Dst abs path:", dst)

    if !fileutil.IsDir(dst) {
        fmt.Println("[Initing]Dst abs path is not dir")
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
            fmt.Printf("[Removing]Recover remove dir err: %v, %v\n", err, string(debug.Stack()))
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
        fmt.Printf("[Removing]Walk remove dir: %v err: %v\n", dst, err)
        return
    }
    dirs.Sort()
    m := dirs.Len() - int(max)
    if m <= 0 {
        return
    }
    delDirs := (([]string)(dirs))[:m]
    for _, dd := range delDirs {
        fmt.Println("[Removing]Removing dir:", dd)
        err = os.RemoveAll(dd)
        if err != nil {
            fmt.Printf("[Removing]Remove dir: %v err: %v\n", dd, err)
            continue
        }
        fmt.Printf("[Removing]Remove dir: %v success\n", dd)
    }
}

func cloneFile() {
    defer func() {
        if err := recover(); err != nil {
            fmt.Printf("[Cloning]Recover clone file err: %v, %v\n", err, string(debug.Stack()))
        }
    }()
    fmt.Println("[Cloning]...")
    defer func() {
        fmt.Println("[Cloning]Finish!")
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
            fmt.Println("[Cloning]Ignore file:", d.Name())
            return nil
        }
        bs, err := ioutil.ReadFile(path)
        if err != nil {
            return err
        }
        nName := strings.ReplaceAll(path, src, dst+string(filepath.Separator)+prefix+"_"+ts+"_"+srcLastDir)
        fmt.Println("[Cloning]Clone", path, "to", nName)
        err = ioutil.WriteFile(nName, bs, 0777)
        if err != nil {
            return err
        }
        return nil
    })
    if err != nil {
        fmt.Printf("[Cloning]Walk clone dir: %v err: %v\n", dst, err)
    }
}
