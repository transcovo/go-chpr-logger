[![CircleCI](https://circleci.com/gh/transcovo/go-chpr-logger.svg?style=shield)](https://circleci.com/gh/transcovo/go-chpr-logger)
[![codecov](https://codecov.io/gh/transcovo/go-chpr-logger/branch/master/graph/badge.svg)](https://codecov.io/gh/transcovo/go-chpr-logger)
[![GoDoc](https://godoc.org/github.com/transcovo/go-chpr-logger?status.svg)](https://godoc.org/github.com/transcovo/go-chpr-logger)

Doc below generated from godoc with godocdown (see dev-tools/test.sh)

--------------------
# logger
--
    import "github.com/transcovo/go-chpr-logger"

Package logger implements our standard Logrus + Logentries + Sentry
configuration.

The package manages a singleton instance of logrus.Logger, initialized from
environment variables.

You will use it exactly as you would use logrus.


### Examples

Log debug info: use the Debug method instead of the Info method. Everything else
works the same way.

Log an info:

    logger.WithFields(logrus.Fields{
    	"driverId": driverId,
    	"rideId": rideId,
    }).Info("Ride accepted")

Log an info with a single field:

    logger.WithField("userId", userId).Info("User logged in")

Log an error:

    logger.WithFields(logrus.Fields{
    	"event": event,
    	"err": err
    }).Error("Could not store ride end event")

IMPORTANT NOTE ABOUT SENTRY: Sentry groups similar occurrences of the same
problem by taking into account the message and the stack trace. So it's
important to have an error message that's always the same for different
occurrences of the same problem. Keep things that can change in the metadata. A
common practice in go for functions that return errors is to handle what they
can handle, and forward the rest as is. So the most likely scenario is that your
err object can be anything. In this situation, you have no choice but to put the
err object in the metadata, and put the actual consequence of the problem you
are reporting in the error message.

In the long run, you will identify different type of errors you didn't think
about thanks to the err in the metadata. You will fix the most frequent, mark
the sentry as resolved, and eventually it will come back with a less frequent
cause you also didn't think you had to handle, etc...


### Configuration

SENTRY_DSN: If provided, warning and error logs will be sent to sentry.

LOGGER_LEVEL: The minimum level of the message to be actually logged. Possible
values: "debug" (default, convenient for development), "info", "warning" or
"error". If an invalid value is provided, "info" will be used and a warning will
be logged.

LOGGER_NAME - not yet implemented - The name of the logger.

LOGENTRIES_TOKEN - not yet implemented - If provided, logs will be sent to
logentries.

LOGSTASH_HOST - not yet implemented - If provided, logs will be sent to
logstash.

LOGSTASH_PORT - not yet implemented - Mandatory if LOGSTASH_HOST is provided.


### Notes

Methods that allow logging without context are not provided, in order to
discourage logging without context.

## Usage

```go
var Debug = GetLogger().Debug
```
Debug is a shorthand to GetLogger().Debug

```go
var Error = GetLogger().Error
```
Error is a shorthand to GetLogger().Error

```go
var Info = GetLogger().Info
```
Info is a shorthand to GetLogger().Info

```go
var Warning = GetLogger().Warning
```
Warning is a shorthand to GetLogger().Warning

#### func  CreateLogger

```go
func CreateLogger() *logrus.Logger
```
CreateLogger creates a new instance of logrus.Logger, which is configured from
the environment variables according to cp conventions (see package overview)

#### func  GetLogger

```go
func GetLogger() *logrus.Logger
```
GetLogger returns logrus.Logger singleton, already configured and ready to use.

This instance is cached, so if the environment changes, you will need to call
ReloadConfiguration() to take changes into account.

#### func  ReloadConfiguration

```go
func ReloadConfiguration()
```
ReloadConfiguration reloads configuration from the environment. Mostly useful
for tests.

#### func  WithField

```go
func WithField(key string, value interface{}) *logrus.Entry
```
WithField is a shorthand for GetLogger().WithField(fields). Use instead of
logrus.WithField.

#### func  WithFields

```go
func WithFields(fields Fields) *logrus.Entry
```
WithFields is a shorthand for GetLogger().WithFields(fields). Use instead of
logrus.WithFields.

#### type Fields

```go
type Fields logrus.Fields
```

Fields type is an alias for the logrus.Fields type

This will allow most of our code to not directly depend on logrus, making it
much easier if we have to switch to another logger later.
