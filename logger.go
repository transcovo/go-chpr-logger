/*

This package implements our standard Logrus + Logentries + Sentry configuration.

The package manages a singleton instance of logrus.Logger, initialized from environment variables.

You will use it exactly as you would use logrus.


Examples

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

IMPORTANT: when logging an error or a warning, don't put an error you're not sure about the message as the message.
The reason for that is that sentry uses the message to regroup similar events. If your library returns different
messages for different occurrences of the same problem, you will flood your team with a lot of alerts for the same
problem. Instead, put the error in the data.


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
Alias for the logrus.Fields

This will allow most of our code to not directly depend on logrus, making it much easier if we have to switch
to another logger later.
 */
type Fields logrus.Fields

var logger *logrus.Logger

/*
Creates a sentry hook catching message of level warning and worse and sending them to sentry
 */
func createSentryHook(sentryDns string) logrus.Hook {
	hook, err := logrus_sentry.NewSentryHook(sentryDns, []logrus.Level{
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
	hook.StacktraceConfiguration.Context = 12

	// 4 is the magic number to use so the stack starts where logger.Error(... was used
	hook.StacktraceConfiguration.Skip = 4

	return hook
}

/*
Returns the already configured logrus.Logger singletton.

Note: this instance is cached, so if the environment changes, you will need to call ReloadConfiguration() before
calling this method.
 */
func CreateLogger() *logrus.Logger {
	newLogger := &logrus.Logger{
		Out:       os.Stdout,
		Formatter: new(logrus.TextFormatter),
		Hooks:     make(logrus.LevelHooks),
		Level:     logrus.DebugLevel,
	}

	sentryDns := os.Getenv("SENTRY_DSN")
	if sentryDns != "" {
		hook := createSentryHook(sentryDns)
		newLogger.Hooks.Add(hook)
	}
	return newLogger
}

/*
Returns the already configured logrus.Logger singletton.

Note: this instance is cached, so if the environment changes, you will need to call ReloadConfiguration() before
calling this method.
 */
func GetLogger() *logrus.Logger {
	if logger == nil {
		logger = CreateLogger()
	}
	return logger
}

/*
Reloads configuration from the environment. Mostly useful for tests.
 */
func ReloadConfiguration() {
	logger = nil
}

/*
Shorthand for GetLogger().WithFields(fields). Use instead of logrus.WithFields.
 */
func WithFields(fields Fields) *logrus.Entry {
	return GetLogger().WithFields(logrus.Fields(fields))
}
/*
Shorthand for GetLogger().WithField(fields). Use instead of logrus.WithField.
 */
func WithField(key string, value interface{}) *logrus.Entry {
	return GetLogger().WithField(key, value)
}
