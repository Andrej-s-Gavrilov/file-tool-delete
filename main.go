package main

import (
	"flag"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strconv"

	log "github.com/sirupsen/logrus"
)

var (
	count_files_limit int
	period_of_check   int // in minutes
	mon_path          string
	file_extention    string
	log_level         log.Level
	log_filename      string
	application_name  string
	version           string
	action            string
)

func init() {
	const (
		default_application_name  = "file-tool-delete"
		default_version           = "0.01"
		default_log_level         = log.InfoLevel
		default_count_files_limit = 5
		default_period_of_check   = 60
		default_file_extention    = ".avi"
		default_mon_path          = "none"
		default_log_file_path     = default_application_name + ".log"
		default_action            = "check" //to do nothing
	)

	//TODO: Need to add option verification
	flag.IntVar(&count_files_limit, "count-files-limit", default_count_files_limit, "Count of files limit")
	flag.IntVar(&period_of_check, "period-of-check", default_period_of_check, "Interval for checking directory")
	flag.StringVar(&mon_path, "path", default_mon_path, "Path to monitoring directory")
	flag.StringVar(&file_extention, "file-ext", default_file_extention, "Monitoring file extention")
	flag.StringVar(&action, "act", default_action, "Action for files: delete - delete. Default: print result without deleting")
	flag.Parse()
	file_extention = "." + file_extention

	if mon_path == default_mon_path {
		fmt.Println("Flag \"path\" is requared")
		os.Exit(1)
	}

	log_filename = default_log_file_path

	if file, err := os.OpenFile(log_filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644); err != nil {
		fmt.Println("Error during opening log-file. Error: " + err.Error())
		os.Exit(1)
	} else {
		log.SetOutput(file)
		log.SetLevel(log.InfoLevel)
		//TODO: need to close log-file
	}

	application_name = default_application_name
	version = default_version
	log_level = default_log_level

	log.WithFields(log.Fields{"modul": "main"}).Info("Start: " + application_name + ". Ver: " + version)
	log.WithFields(log.Fields{"modul": "main"}).Info("Logleve: ", log_level.String())

}

func find(root, ext string) []fs.FileInfo {
	var list_of_fi []fs.FileInfo
	filepath.WalkDir(root, func(s string, d fs.DirEntry, e error) error {
		if e != nil {
			return e
		}
		if filepath.Ext(d.Name()) == ext {
			if fi, err := os.Stat(s); err != nil {
				log.WithFields(log.Fields{"modul": "main", "file": s}).Error("Error: " + err.Error())
			} else {
				//Getting ctime

				//for Unix
				//stat := fi.Sys().(*syscall.Stat_s)
				//atime := time.Unix(int64(stat.Atim.Sec), int64(stat.Atim.Nsec))
				//ctime := time.Unix(int64(stat.Ctim.Sec), int64(stat.Ctim.Nsec))

				//for Windows
				//stat := fi.Sys().(*syscall.Win32FileAttributeData)
				//atime := time.Unix(0, stat.LastAccessTime.Nanoseconds())
				//ctime := time.Unix(0, stat.CreationTime.Nanoseconds())

				list_of_fi = append(list_of_fi, fi)
			}
		}
		return nil
	})
	return list_of_fi
}

func main() {
	list := find(mon_path, file_extention)
	count_of_files := len(list)
	log.WithFields(log.Fields{"modul": "main"}).Info("Count of files: " + strconv.Itoa(count_of_files))

	if count_of_files > count_files_limit {
		sort.Slice(list, func(i, j int) bool {
			return list[i].ModTime().Before(list[j].ModTime())
		})

		for i, f := range list {
			if i < count_of_files-count_files_limit {
				log.WithFields(log.Fields{"modul": "main", "file": f.Name()}).Info("Deleted")
				if action == "check" {
					fmt.Printf("DELETED -> # %d File - %s%s\n", i, mon_path, f.Name())
				} else if action == "delete" {
					_ = 1
					if err := os.Remove(mon_path + f.Name()); err != nil {
						log.WithFields(log.Fields{"modul": "main", "file": f.Name()}).Info("File can't be deletered. Error:" + err.Error())
					}
				}
			} else {
				if action == "check" {
					fmt.Printf("KEEP -> # %d File - %s%s\n", i, mon_path, f.Name())
				} else if action == "delete" {
					_ = 1
				}
			}
		}
	}
}
