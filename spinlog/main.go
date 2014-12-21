package main

import (
	"flag"
	"fmt"
	"github.com/unixpickle/spinlog"
	"io"
	"os"
	"os/user"
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
	flag.StringVar(&owner, "owner", "", "uid:gid or username owner for file")
	
	var dirPath string
	flag.StringVar(&dirPath, "dir", ".", "directory for log files")
	
	var prefix string
	flag.StringVar(&prefix, "prefix", "log", "prefix for log files")
	
	flag.Parse()
	
	cfg := *new(spinlog.Config)
	if !parseOwner(owner, &cfg) {
		fmt.Println("Invalid owner:", owner)
		os.Exit(1)
	}
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

func parseOwner(str string, cfg *spinlog.Config) bool {
	if parseOwnerNum(str, cfg) {
		return true
	}
	
	// Look up the named user
	info, _ := user.Lookup(str)
	if info == nil {
		return false
	}
	
	// Attempt to parse UID and GID as numbers
	uid, err := strconv.Atoi(info.Uid)
	if err != nil {
		return false
	}
	gid, err := strconv.Atoi(info.Gid)
	if err != nil {
		return false
	}
	
	// Set info on configuration
	cfg.SetOwner = true
	cfg.UID = uid
	cfg.GID = gid
	return true
}

func parseOwnerNum(str string, cfg *spinlog.Config) bool {
	if str == "" {
		// Empty string means use default
		return true
	}
	
	// If the string is not : separated, it's invalid.
	idx := strings.Index(str, ":")
	if idx < 0 {
		return false
	}
	uidStr := str[0:idx]
	gidStr := str[idx+1:]
	
	// Parse UID and GID strings
	uid, err := strconv.Atoi(uidStr)
	if err != nil {
		return false
	}
	gid, err := strconv.Atoi(gidStr)
	if err != nil {
		return false
	}
	
	// Set info on configuration
	cfg.SetOwner = true
	cfg.UID = uid
	cfg.GID = gid
	return true
}
