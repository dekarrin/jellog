# Jellog

[![Go Reference](https://pkg.go.dev/badge/github.com/dekarrin/rosed.svg)](https://pkg.go.dev/github.com/dekarrin/rosed)

Custom and simplish Go logging library that supports logging of different types
via generics.

Jellog uses an architecture similar to python logs. It provides a `Logger` type
which has methods for logging messages at various pre-configured or custom
severity levels. The Logger routes the message to one or more `Handler` which
handles either further routing or writing the message to an output such as
files or std streams.

It also includes top-level functions intended for drop-in or near drop-in
replacement of use of the default logger in the built-in Go `log` library.

## Installing

```bash
go get github.com/dekarrin/jellog
```

## Use

Once imported, you can call package functions to invoke the default logger,
which will simply print to stderr with a pre-configured header similar to the
default built-in Go logger.

```golang
package main

import (
    "os"

    "github.com/dekarrin/jellog"
)

func main() {
    jellog.Printf("The program has started")

    f, err := os.Open("somefile.txt")
    if err != nil {
        jellog.Fatalf("Can't open somefile.txt: %v", err)
    }
    defer f.Close()

    jellog.Printf("The program is done!")
}

```

Running the above would result in the following being printed to stderr:

```
2009/11/10 23:00:00 INFO  The program has started
2009/11/10 23:00:00 FATAL ERROR Can't open somefile.txt: open somefile.txt: no such file or directory
```

If you need custom logging with source component information, create a Logger
and use that.

```golang
package main

import (
    "fmt"
    "io"
    "os"
    "http"

    "github.com/dekarrin/jellog"
)

var log jellog.Logger[string]

func main() {
    log = jellog.New(&jellog.LoggerOptions[string]{
        Options: jellog.Options[string]{
            Component: "server",
        },
    })

    stderrHandler := jellog.NewStderrHandler(nil)
    fileHandler, err := jellog.OpenFile("server.log", nil)
    if err != nil {
        jellog.Fatalf("could not open log file server.log: %v", err)
    }

    log.AddHandler(jellog.LvInfo, stderrHandler) // only show INFO-level or higher in stderr
    log.AddHandler(jellog.LvAll, fileHandler)    // show all levels in the file output

    // and then start using it!
    log.Info("Initialize server...")

    log.Debug("Starting config load...")
    conf, err := loadConfig("server.yml")
    if err != nil {
        if !conf.defaultConf {
            log.Fatalf("Problem loading config: %w", err)
        }
    }
    log.Debug("Config load complete")

    server := createHTTPServer(conf)
    log.Debug("Server created")

    log.Infof("Server is now listening for new connections")
    log.Fatalf("%v", server.ListenAndServe())
}

func loadConfig(filename string) (Config, error) {
    log.Tracef("Open file...")
    f, err := os.Open(filename)
    if err != nil {
        log.Warnf("config unreadable; using default config: can't open: %v", err)
        return Config{defaultConf: true}, err
    }
    defer f.Close()

    log.Tracef("About to start reading config...")
    confData, err := io.ReadAll(f)
    if err != nil {
        log.Errorf("%v", err)
        return Config{}, err
    }
    log.Tracef("Finished reading config")

    log.Tracef("About to start parsing config...")
    conf, err := parseConfig(confData)
    if err != nil {
        log.Errorf("%v", err)
        return Config{}, err
    }
    log.Tracef("Finished parsing config")
    return conf, nil
}

func parseConfig(data []byte) (Config, error) {
    // ...
    //
    // do some stuff with config, eventually end up with a:

    return Config{Port: 20, Address: "localhost"}, nil
}

func createHTTPServer(conf Config) *http.Server {
    return &http.Server{
        Addr: fmt.Sprintf("%s:%d", conf.Address, conf.Port),
    }
}

type Config struct {
    defaultConf bool
    Port int
    Address string
}
```

Assuming all went well, running the above might result in something such as the
following being printed to stderr:

```
2009/11/10 23:00:00 INFO  (server) Initialize server...
2009/11/10 23:00:04 INFO  (server) Server is is now listening for new connections
```

And even more details would be in server.log:

```
2009/11/10 23:00:00 INFO  (server) Initialize server...
2009/11/10 23:00:00 DEBUG (server) Starting config load...
2009/11/10 23:00:00 TRACE (server) Open file...
2009/11/10 23:00:01 TRACE (server) About to start reading config...
2009/11/10 23:00:03 TRACE (server) Finished reading config
2009/11/10 23:00:03 TRACE (server) About to start parsing config...
2009/11/10 23:00:03 TRACE (server) Finished parsing config
2009/11/10 23:00:04 DEBUG (server) Config load complete
2009/11/10 23:00:04 DEBUG (server) Server created
2009/11/10 23:00:04 INFO  (server) Server is is now listening for new connections
```
