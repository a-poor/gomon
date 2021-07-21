# gomon

_created by Austin Poor_

A tool, written in Go, for automatically reloading a go file when file changes are detected (like a simple version of [nodemon](https://nodemon.io/)).

It's meant to be a drop-in replacement for `go run`.

I created this tool to get some experiance writing Go and because it ended up being useful (I find it comes in handy when working with the `net/http` package).

_NOTE:_ This tool is currently in an early stage of development. It doesn't yet have unit tests and it doesn't have CLI options.

## Example

Say for example you have a `main.go` file that looks like this:

```go
package main

import (
    "fmt"
    "time"
)

func main() {
    i := 0
    for {
        fmt.Printf("[%5d]\n", i)
        i++
        time.Sleep(time.Second * 1)
    }
}
```

You could run `gomon main.go` and you'd see the output like this:

```bash
❯ ./gomon sample.go
2021/07/20 16:31:59 Starting file "main.go" at 16:31:59
STDOUT: [    0]
STDOUT: [    1]
STDOUT: [    2]
STDOUT: [    3]
STDOUT: [    4]
STDOUT: [    5]
STDOUT: [    6]
...
```

Then, say you make a change to the `main.go` file, you'll see the output like this:

```bash
❯ ./gomon sample.go
2021/07/20 16:37:40 Starting file "main.go" at 16:37:40
STDOUT: [    0]
STDOUT: [    1]
STDOUT: [    2]
STDOUT: [    3]
STDOUT: [    4]
STDOUT: [    5]
STDOUT: [    6]
...
2021/07/20 16:37:47 Update detected. Reloading...
STDOUT: [    0]
STDOUT: [    1]
STDOUT: [    2]
STDOUT: [    3]
```

And if that change we made to `main.go` causes an error, `gomon` will log the error and then continue to watch for changes.

```bash
❯ ./gomon sample.go
2021/07/20 16:42:03 Starting file "sample.go" at 16:42:03
STDOUT: [    0]
STDOUT: [    1]
STDOUT: [    2]
STDOUT: [    3]
STDOUT: [    4]
STDOUT: [    5]
STDOUT: [    6]
STDOUT: [    7]
STDOUT: [    8]
STDOUT: [    9]
STDOUT: [   10]
STDERR: panic: Oh no!
STDERR:
STDERR: goroutine 1 [running]:
STDERR: main.main()
STDERR: 	/Users/austinpoor/tmp/test-gomon/sample.go:15 +0xd7
STDERR: exit status 2
```

