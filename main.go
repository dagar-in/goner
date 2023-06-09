package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/fsnotify/fsnotify"
)

func main() {

	if len(os.Args) < 2 {
		fmt.Println("Please provide the command to run.")
		return
	}

	defer func() {
		if r := recover(); r != nil {
			fmt.Println(": => ", r)
		}
		main();
	}()

	command := os.Args[1:]

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		fmt.Println("Error setting up file watcher:", err)
		return
	}
	defer watcher.Close()

	cwd, err := os.Getwd()
	if err != nil {
		fmt.Println("Error getting current working directory:", err)
		return
	}
	err = addWatchRecursively(watcher, cwd, 10)
	if err != nil {
		fmt.Println("Error adding current working directory and its subdirectories to file watcher:", err)
		return
	}

	cmd := exec.Command(command[0], command[1:]...)
	cmdSetup(cmd)
	go func() {
		if err := runCommand(cmd); err != nil {
			fmt.Println("\n:", err)
			return
		}
	}()


	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				fmt.Println("File watcher closed unexpectedly.")
				return
			}
			if event.Op&fsnotify.Write == fsnotify.Write || event.Op&fsnotify.Create == fsnotify.Create {
				fmt.Println("Detected change in file:", event.Name)
				if err := stopCommand(cmd); err != nil {
					fmt.Println(":-)", err)
				}
				if err := runCommand(cmd); err != nil {
					fmt.Println(";=) \n")
					return
				}
			}
		case err, ok := <-watcher.Errors:
			if !ok {
				fmt.Println("File watcher closed unexpectedly.")
				return
			}
			fmt.Println("Error watching file system:", err)
		}
	}
}

func cmdSetup(cmd *exec.Cmd) error {
	
    stdout, err := cmd.StdoutPipe()
    if err != nil {
        fmt.Println("Error setting up stdout pipe:", err)
		return err
    }

    stderr, err := cmd.StderrPipe()
    if err != nil {
        fmt.Println("Error setting up stderr pipe:", err)
		return err
    }
    stdoutScanner := bufio.NewScanner(stdout)
    stderrScanner := bufio.NewScanner(stderr)

    go func() {
        for stdoutScanner.Scan() {
            fmt.Println(stdoutScanner.Text())
        }
    }()

    go func() {
        for stderrScanner.Scan() {
            fmt.Println(stderrScanner.Text())
        }
    }()

	return nil
}

func runCommand(cmd *exec.Cmd) error {

    if err := cmd.Start(); err != nil {
		return nil
    }

    if err := cmd.Wait(); err != nil {
		return nil

    }
	return nil
}

func stopCommand(cmd *exec.Cmd) error {

	defer func() {
		panic("Restart")
	}()
	if cmd == nil || cmd.Process == nil {
		return nil
	}
	return cmd.Process.Kill()
}

func addWatchRecursively(watcher *fsnotify.Watcher, dir string, depth int) error {
	if depth == 0 {
		return nil
	}

	// Add watch for the directory itself
	if err := watcher.Add(dir); err != nil {
		return err
	}

	// Recursively add watch for subdirectories
	fileInfos, err := ioutil.ReadDir(dir)
	if err != nil {
		return err
	}

	for _, fileInfo := range fileInfos {
		if fileInfo.IsDir() {
			subDir := filepath.Join(dir, fileInfo.Name())
			if err := addWatchRecursively(watcher, subDir, depth-1); err != nil {
				return err
			}
		}
	}

	return nil
}
