package main

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io/fs"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"
)

func getFileToRun() (string, error) {
	if len(os.Args) < 2 {
		return "", errors.New("No go file specified.")
	} else if len(os.Args) > 2 {
		return "", errors.New("Too many command line arguments specified.")
	}
	return os.Args[1], nil
}

// Check if the file `f` is a go file.
// Works by checking if it has the `.go` extension
func isGoFile(f fs.FileInfo) bool {
	return strings.HasSuffix(strings.ToLower(f.Name()), ".go")
}

func getGoFiles() []fs.FileInfo {
	files, err := ioutil.ReadDir("./")
	if err != nil {
		log.Fatal(err)
	}

	res := make([]fs.FileInfo, 0)

	for _, f := range files {
		if isGoFile(f) {
			res = append(res, f)
		}
	}

	return res

}

func anyFilesUpdatedSince(t time.Time) bool {
	for _, f := range getGoFiles() {
		if f.ModTime().Sub(t).Seconds() > 0 {
			return true
		}
	}
	return false
}

func runAndGetCancel(fn string) context.CancelFunc {
	// Create the cancelable context
	ctx, cancel := context.WithCancel(context.Background())
	// Create the run command
	cmd := exec.CommandContext(ctx, "go", "run", fn)

	// Capture stdout and stderr
	stdout, _ := cmd.StdoutPipe()
	stderr, _ := cmd.StderrPipe()

	// Run the command
	cmd.Start()

	// Kick off the stdout capturing
	go func() {
		bufout := bufio.NewReader(stdout)
		defer stdout.Close()
		for {
			select {
			case <-ctx.Done():
				fmt.Println("stdout got a cancel message")
				return
			default:
				// Read/Print from stdout
				cmdout, _ := bufout.ReadString('\n')
				strout := string(cmdout)
				if len(strout) > 0 {
					fmt.Print("STDOUT:", strout)
				}
			}
		}
	}()
	// Kick off the stderr capturing
	go func() {
		buferr := bufio.NewReader(stderr)
		defer stderr.Close()
		for {
			select {
			case <-ctx.Done():
				fmt.Println("stderr got a cancel message")
				return
			default:
				// Read/Print from stderr
				cmderr, _ := buferr.ReadString('\n')
				strerr := string(cmderr)
				if len(strerr) > 0 {
					fmt.Print("STDERR:", strerr)
				}
			}
		}
	}()
	return cancel
}

func main() {
	// Get go file to run from cli arg
	fn, err := getFileToRun()
	if err != nil {
		log.Fatalln(err)
	}

	// Log the start
	lastReload := time.Now()
	log.Printf("Starting file \"%s\" at %s", fn, lastReload.Format("15:04:05"))

	// Create the context and cancel function
	cancel := runAndGetCancel(fn)
	defer cancel()

	for {
		// Check for file updates
		if anyFilesUpdatedSince(lastReload) {
			log.Println("Someone updated a file! Reloading...")
			lastReload = time.Now()
			cancel()

			cancel = runAndGetCancel(fn)
		}

		// Wait 100ms before checking again
		time.Sleep(time.Millisecond * 100)
	}
}
