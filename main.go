package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
)

var cmdArg [2]string

func init() {
	if runtime.GOOS == "windows" {
		cmdArg[0] = "cmd"
		cmdArg[1] = "/c"
	} else {
		cmdArg[0] = "/bin/sh"
		cmdArg[1] = "-c"
	}
}

func main() {

	gopathEnv := os.Getenv("GOPATH")
	if gopathEnv == "" {
		fmt.Println("Please set GOPATH first.")
		return
	}
	gopaths := filepath.SplitList(gopathEnv)
	appPath := AppPath()
	inGopath := false

	for _, gopath := range gopaths {
		if strings.HasPrefix(appPath, gopath) {
			inGopath = true
			break
		}
	}

	if inGopath {
		update(appPath)
	} else {
		for _, gopath := range gopaths {
			update(gopath)
		}
	}

	fmt.Println("All repositories up to date. You're ready to Go :-)")
}

func update(rootDir string) {
	repositories, err := List(rootDir, nil)
	if err != nil {
		log.Fatal(err)
	}

	var wg sync.WaitGroup
	token := make(chan struct{}, 10)
	wg.Add(len(repositories))
	for _, repo := range repositories {
		go func(dir string) {
			defer wg.Done()
			token <- struct{}{}
			updateRepo(dir)
			<-token
		}(repo)
	}
	wg.Wait()
}

func isRepo(dir string) bool {
	return FolderExist(path.Join(dir, ".git"))
}

// FolderExist returns true if a specified folder exists; false if it does not.
func FolderExist(path string) bool {
	info, err := os.Stat(path)
	if err != nil || !info.IsDir() {
		return false
	}
	return true
}

// AppPath returns an absolute path of app.
func AppPath() string {
	dir, _ := filepath.Abs(filepath.Dir(os.Args[0]))
	return dir
}

// List returns all repositories under the directory
func List(dir string, list []string) ([]string, error) {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	for _, file := range files {
		if file.IsDir() {
			if isRepo(path.Join(dir, file.Name())) {
				list = append(list, filepath.ToSlash(path.Join(dir, file.Name())))
			} else {
				list, err = List(path.Join(dir, file.Name()), list)
				if err != nil {
					return list, err
				}
			}
		}
	}

	return list, nil
}

func updateRepo(dir string) {
	cmd := exec.Command(cmdArg[0], cmdArg[1], "cd "+dir+" && git pull")
	out, err := cmd.Output()
	if err != nil {
		log.Println(dir, err)
		return
	}
	log.Print(dir, " ", string(out))
}
