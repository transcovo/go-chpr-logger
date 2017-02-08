➤ Readme generated with godocdown

This package implements our standard Logrus + Logentries + Sentry configuration.

The package manages a singleton instance of logrus.Logger, initialized from
environment variables.

You will use it exactly as you would use logrus.


### Examples

Log an info:

    WithFields(logger.Fields{
    	"driverId": driverId,
    	"rideId": rideId,
    }).Info("Ride accepted")

Log an info with a single field:

    WithField("userId", userId).Info("User logged in")

Log an error:

    WithFields(logger.Fields{
    	"event": event,
    	"err": err
    }).Error("Could not store ride end event")

IMPORTANT: when logging an error or a warning, don't put an error you're not
sure about the message as the message. The reason for that is that sentry uses
the message to regroup similar events. If your library returns different
messages for different occurrences of the same problem, you will flood your team
with a lot of alerts for the same problem. Instead, put the error in the data.


### Configuration

SENTRY_DSN: If provided, warning and error logs will be sent to sentry.

LOGGER_NAME - not yet implemented - The name of the logger.

LOGGER_LEVEL - not yet implemented - The minimum level of the message to be
actually logged. Possible values: "debug", "info", "warning", "error".

LOGENTRIES_TOKEN - not yet implemented - If provided, logs will be sent to
logentries.

LOGSTASH_HOST - not yet implemented - If provided, logs will be sent to
logstash.

LOGSTASH_PORT - not yet implemented - Mandatory if LOGSTASH_HOST is provided.


### Notes

Methods that allow logging without context are not provided, in order to
discourage logging without context.

## Usage

#### func  CreateLogger

```go
func CreateLogger() *logrus.Logger
```
Returns the already configured logrus.Logger singletton.

Note: this instance is cached, so if the environment changes, you will need to
call ReloadConfiguration() before calling this method.

#### func  GetLogger

```go
func GetLogger() *logrus.Logger
```
Returns the already configured logrus.Logger singletton.

Note: this instance is cached, so if the environment changes, you will need to
call ReloadConfiguration() before calling this method.

#### func  ReloadConfiguration

```go
func ReloadConfiguration()
```
Reloads configuration from the environment. Mostly useful for tests.

#### func  WithField

```go
func WithField(key string, value interface{}) *logrus.Entry
```
Shorthand for GetLogger().WithField(fields). Use instead of logrus.WithField.

#### func  WithFields

```go
func WithFields(fields Fields) *logrus.Entry
```
Shorthand for GetLogger().WithFields(fields). Use instead of logrus.WithFields.

#### type Fields

```go
type Fields logrus.Fields
```

Alias for the logrus.Fields

This will allow most of our code to not directly depend on logrus, making it
much easier if we have to switch to another logger later.
