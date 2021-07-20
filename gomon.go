package main

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"
)

// Get the main go file to be run from command line
// arguments
func getFileToRun() (string, error) {
	if len(os.Args) < 2 {
		return "", errors.New("no go file specified")
	} else if len(os.Args) > 2 {
		return "", errors.New("too many command line arguments specified")
	}
	return os.Args[1], nil
}

// Check if the file `f` is a go file.
// Works by checking if it has the `.go` extension
func isGoFile(f fs.FileInfo) bool {
	return strings.HasSuffix(strings.ToLower(f.Name()), ".go")
}

// Get a slice of go file
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

// Check for file updates since `t`
func anyFilesUpdatedSince(t time.Time) bool {
	for _, f := range getGoFiles() {
		if f.ModTime().Sub(t).Seconds() > 0 {
			return true
		}
	}
	return false
}

// Format text as red
func toRed(text string) string {
	return fmt.Sprintf("\033[31m%s\033[0m", text)
}

// Format text as blue
func toBlue(text string) string {
	return fmt.Sprintf("\033[94m%s\033[0m", text)
}

// Run in a goroutine. Captures output from stderr/stdout pipe and
// passes to the logger to print. If EOF is reached, call cancel.
// Or if cancel is already called, stop.
func readOutputs(ctx context.Context, cancel context.CancelFunc, r io.Reader, l *log.Logger, logname string) {
	bufout := bufio.NewReader(r)
	// defer stdout.Close()
	for {
		select {
		case <-ctx.Done():
			return
		default:
			// Read/Print from stdout
			cmdout, err := bufout.ReadString('\n')
			if err == io.EOF {
				cancel()
				return
			}
			if err != nil {
				log.Printf("Error reading from %s : %s", logname, err)
			}
			strout := string(cmdout)
			if len(strout) > 0 {
				l.Printf("%s: %s", logname, strout)
			}
		}
	}
}

// Run the go file with filename `fn` and start goroutines to pipe
// stdout/stderr to the loggers. Returns the command context's
// cancel function.
func runAndGetCancel(fn string, logout *log.Logger, logerr *log.Logger) context.CancelFunc {
	// Create the cancelable context
	ctx, cancel := context.WithCancel(context.Background())
	// Create the run command
	cmd := exec.CommandContext(ctx, "go", "run", fn)

	// Capture stdout and stderr
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Panicf("Error Connecting to StdOut : %s", err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		log.Panicf("Error Connecting to StdErr : %s", err)
	}

	// Run the command
	err = cmd.Start()
	if err != nil {
		log.Panicf("Error starting exec command : %s", err)
	}

	// Kick off the stdout/stderr capturing
	go readOutputs(ctx, cancel, stdout, logout, toBlue("STDOUT"))
	go readOutputs(ctx, cancel, stderr, logerr, toRed("STDERR"))

	return cancel
}

func main() {
	// Get go file to run from cli arg
	fn, err := getFileToRun()
	if err != nil {
		log.Fatalln(err)
	}

	// Loggers used to print output from stderr/stdout
	logout := log.New(os.Stdout, "", 0)
	logerr := log.New(os.Stderr, "", 0)

	// Log the start
	lastReload := time.Now()
	log.Printf("Starting file \"%s\"", fn)

	// Create the context and cancel function
	cancel := runAndGetCancel(fn, logout, logerr)
	defer cancel()

	for {
		// Check for file updates
		if anyFilesUpdatedSince(lastReload) {
			// Tell the user reload is starting...
			log.Println("Update detected. Reloading...")

			// Cancel the current running command
			cancel()

			// Checkpoint the last reload time
			lastReload = time.Now()

			// Rerun the command
			cancel = runAndGetCancel(fn, logout, logerr)
		}

		// Wait 100ms before checking again
		time.Sleep(time.Millisecond * 100)
	}
}
