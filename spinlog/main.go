package main

import (
	"flag"
	"fmt"
	"github.com/unixpickle/spinlog"
	"io"
	"os"
	"strconv"
	"strings"
)

func main() {
	maxSize := flag.Int("size", 1048576, "max log file size")
	maxCount := flag.Int("count", 3, "max number of log files")
	perms := flag.Int("permissions", 0600,
		"octal file permissions for log files")
	lineSize := flag.Int("line", -1,
		"max line size for buffering (-1 means no line buffering)")
	
	var owner string
	flag.StringVar(&owner, "owner", "", "uid:gid owner for file")
	
	var dirPath string
	flag.StringVar(&dirPath, "dir", ".", "directory for log files")
	
	var prefix string
	flag.StringVar(&prefix, "prefix", "log", "prefix for log files")
	
	flag.Parse()
	
	uid, gid, enable, ok := parseOwner(owner)
	if !ok {
		fmt.Println("Invalid owner:", owner)
		os.Exit(1)
	}
	
	cfg := *new(spinlog.Config)
	cfg.SetOwner = enable
	cfg.UID = uid
	cfg.GID = gid
	cfg.SetPerm = (*perms != -1)
	cfg.Perm = *perms
	cfg.LogDir.Directory = dirPath
	cfg.LogDir.Prefix = prefix
	cfg.MaxCount = *maxCount
	cfg.MaxSize = int64(*maxSize)
	var writer io.Writer
	var err error
	if *lineSize > 0 {
		lc := spinlog.LineConfig{cfg, *lineSize}
		writer, err = spinlog.NewLineLog(lc)
	} else {
		writer, err = spinlog.NewLog(cfg)
	}
	if err != nil {
		fmt.Println("Error opening log:", err)
		os.Exit(1)
	}
	io.Copy(writer, os.Stdin)
}

func parseOwner(str string) (uid int, gid int, enable bool, ok bool) {
	if str == "" {
		return 0, 0, false, true
	}
	idx := strings.Index(str, ":")
	if idx < 0 {
		return 0, 0, false, false
	}
	uidStr := str[0:idx]
	gidStr := str[idx+1:]
	uid, err := strconv.Atoi(uidStr)
	if err != nil {
		return 0, 0, false, false
	}
	gid, err = strconv.Atoi(gidStr)
	if err != nil {
		return 0, 0, false, false
	}
	return uid, gid, true, true
}
