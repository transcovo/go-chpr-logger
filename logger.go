/*

Package logger implements our standard Logrus + Logentries + Sentry configuration.

The package manages a singleton instance of logrus.Logger, initialized from environment variables.

You will use it exactly as you would use logrus.


Examples

Log debug info: use the Debug method instead of the Info method. Everything else works the same way.

Log an info:

	logger.WithFields(logger.Fields{
		"driverId": driverId,
		"rideId": rideId,
	}).Info("Ride accepted")

Log an info with a single field:

	logger.WithField("userId", userId).Info("User logged in")

Log an error:

	logger.WithFields(logger.Fields{
		"event": event,
		"err": err
	}).Error("Could not store ride end event")

IMPORTANT NOTE ABOUT SENTRY: Sentry groups similar occurrences of the same problem by taking into account the message
and the stack trace. So it's important to have an error message that's always the same for different occurrences of the
same problem. Keep things that can change in the metadata. A common practice in go for functions that return errors
is to handle what they can handle, and forward the rest as is. So the most likely scenario is that your err object can
be anything. In this situation, you have no choice but to put the err object in the metadata, and put the actual
consequence of the problem you are reporting in the error message.

In the long run, you will identify different type of errors you didn't think about thanks to the err in the metadata.
You will fix the most frequent, mark the sentry as resolved, and eventually it will come back with a less frequent cause
you also didn't think you had to handle, etc...


Configuration

SENTRY_DSN: If provided, warning and error logs will be sent to sentry.

LOGGER_NAME - not yet implemented - The name of the logger.

LOGGER_LEVEL - not yet implemented - The minimum level of the message to be actually logged.
Possible values: "debug", "info", "warning", "error".

LOGENTRIES_TOKEN - not yet implemented - If provided, logs will be sent to logentries.

LOGSTASH_HOST - not yet implemented - If provided, logs will be sent to logstash.

LOGSTASH_PORT - not yet implemented - Mandatory if LOGSTASH_HOST is provided.


Notes

Methods that allow logging without context are not provided, in order to discourage logging without context.

 */
package logger

import (
	"github.com/Sirupsen/logrus"
	"os"
	"github.com/evalphobia/logrus_sentry"
	"time"
)

/*
Fields type is an alias for the logrus.Fields type

This will allow most of our code to not directly depend on logrus, making it much easier if we have to switch
to another logger later.
 */
type Fields logrus.Fields

var logger *logrus.Logger

/*
Creates a sentry hook catching message of level warning or worse and sending them to sentry
 */
func createSentryHook(sentryDsn string) logrus.Hook {
	hook, err := logrus_sentry.NewSentryHook(sentryDsn, []logrus.Level{
		logrus.PanicLevel,
		logrus.FatalLevel,
		logrus.ErrorLevel,
		logrus.WarnLevel,
	})
	if err != nil {
		panic(err)
	}
	hook.Timeout = time.Second
	hook.StacktraceConfiguration.Enable = true
	hook.StacktraceConfiguration.Level = logrus.ErrorLevel

	// Number of lines of context code displayed around each line of the stack trace. 12 is a comfortable
	// amount, and there is no need to make this configurable for now. We can change it later.
	hook.StacktraceConfiguration.Context = 12

	// 4 is the magic number to use so the stack starts where logger.Error(... was used
	hook.StacktraceConfiguration.Skip = 4

	return hook
}

/*
CreateLogger creates a new instance of logrus.Logger, which is configured from the environment variables according to cp
conventions (see package overview)
 */
func CreateLogger() *logrus.Logger {
	newLogger := &logrus.Logger{
		Out:       os.Stdout,
		Formatter: new(logrus.TextFormatter),
		Hooks:     make(logrus.LevelHooks),
		Level:     logrus.DebugLevel,
	}

	sentryDsn := os.Getenv("SENTRY_DSN")
	if sentryDsn != "" {
		hook := createSentryHook(sentryDsn)
		newLogger.Hooks.Add(hook)
	}
	return newLogger
}

/*
GetLogger returns logrus.Logger singleton, already configured and ready to use.

This instance is cached, so if the environment changes, you will need to call ReloadConfiguration() to take changes
into account.
 */
func GetLogger() *logrus.Logger {
	if logger == nil {
		logger = CreateLogger()
	}
	return logger
}

/*
ReloadConfiguration reloads configuration from the environment. Mostly useful for tests.
 */
func ReloadConfiguration() {
	logger = nil
}

/*
WithFields is a shorthand for GetLogger().WithFields(fields). Use instead of logrus.WithFields.
 */
func WithFields(fields Fields) *logrus.Entry {
	return GetLogger().WithFields(logrus.Fields(fields))
}
/*
WithField is a shorthand for GetLogger().WithField(fields). Use instead of logrus.WithField.
 */
func WithField(key string, value interface{}) *logrus.Entry {
	return GetLogger().WithField(key, value)
}