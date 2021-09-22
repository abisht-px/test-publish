# Logging
[![Documentation](http://godoc-uns.dreamworks.net/github.anim.dreamworks.com/golang/logging.git?status.svg)](http://godoc-uns.dreamworks.net/github.anim.dreamworks.com/golang/logging.git)
[![Go Report Card](http://goreportcard-uns.dreamworks.net/badge/github.anim.dreamworks.com/golang/logging.git)](http://goreportcard-uns.dreamworks.net/report/github.anim.dreamworks.com/golang/logging.git)

This package is a wrapper for Logrus that provides global logging settings for DWA Golang projects.

## Initializing the Logger

The logger should be initialized in your `main.go`. While you can include this code elsewhere it is not recommended as you could overwrite downstream configurations.

### Configuration

To set up the logger with the default configuration:
```go
// Assume a boolean debug exists that represents whether the current program is in debug or release mode.
logging.DefaultConfig(nil, debug)
```

By default the "debug" formatter displays pretty formatted text. The production formatter is in easy-to-parse JSON.

Note: If you already have a Logrus logger you want to re-purpose for global logging, pass it in as the first argument in the `DefaultConfig` function, otherwise `nil` is expected.

To change some basic configurations the following helpers are available:
```go
// NoColors ensures that logs do not have any color, good for unicode-sensitive log ingestion.
logging.NoColors(nil)

// NoTimestamp removes the timestamp from the log output completely.
logging.NoTimestamp(nil)

// FullTimestamp ensures that the full complete datetime is shown in the log output.
// This only has an effect on the debug formatter, the prod logs always contain the full timestamp.
logging.FullTimestamp(nil)
```

For the above options, the configuration will be applied to the global logger. If you want to apply these options to another Logrus logger, simply pass that logger as the argument in the function, otherwise `nil` is expected.


## Using the Logger

### Getting the current Logger

To get the current Logrus logger, use the `Log` method:
```go
current_logger := logging.Log()
```

### Log Levels

By default there are 7 log levels available:
```go
// If in debug mode, Trace-level and higher messages are displayed.
logging.Log().Trace("This is a Trace-level message!")
logging.Log().Debug("This is a Debug-level message!")
logging.Log().Info("This is an Info-level message!")

// If not in debug mode, Warn-level and higher messages are displayed.
logging.Log().Warn("This is a Warn-level message!")
logging.Log().Error("This is an Error-level message!")

// In addition to logging, Fatal will call `os.Exit(1)` to exit the program.
logging.Log().Fatal("This is a Fatal-level message!")

// In addition to logging, Panic will also trigger a panic in the program.
logging.Log().Panic("This is a Panic-level message!")
```

Note on Printing: Like the `fmt` library, each of these levels includes a "f" and "ln" variant.

For example, for Info-level logs you could use:

* `logging.Log().Info("my log")` or
* `logging.Log().Infof("my %v", "log")` or
* `logging.Log().Infoln("my log")`

For your convenience, all common log-level printers are available directly from the package:

For example, the above Info-level logs could be made like this:

* `logging.Info("my log")` or
* `logging.Infof("my %v, "log")` or
* `logging.Infoln("my log")`

Levels available directly from the package: `Trace`, `Debug`, `Info`, `Warn`, `Error`, `Fatal`, `Panic`, and `Print`.

### Setting arbitrary Log Levels

To change the log level that the logger is logging at, use the Logrus built-in `SetLevel` function.

```go
// This will only display Error, Fatal, and Panic messages
logging.Log().SetLevel(logging.ErrorLevel)
```
