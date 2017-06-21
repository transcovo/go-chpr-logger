package logger

import (
	"testing"
	"os"
	"io/ioutil"
	"github.com/stretchr/testify/assert"
	"bytes"
	"net/http/httptest"
	"net/http"
	"strings"
	"io"
	"encoding/base64"
	"compress/zlib"
	"encoding/json"
	"github.com/getsentry/raven-go"
	"github.com/sirupsen/logrus"
)

/*
This helper function allows testing that certain things appear on the standard output
 */
func captureStdout(task func()) []byte {

	in, out, err := os.Pipe()
	if err != nil {
		panic(err)
	}

	os.Stdout.Sync()
	tmp := os.Stdout
	os.Stdout = out
	ReloadConfiguration()
	defer (func() {
		os.Stdout = tmp
		ReloadConfiguration()
	})()

	task()

	out.Close()

	stdout, err := ioutil.ReadAll(in)
	if err != nil {
		panic(err)
	}
	return stdout
}

/*
withEnvVariable helper function executes a task in a context when the value of an environment variable is temporarily
changed

key is the name of the environment variable to change
value is the temporary value
task specifies the task to execute
 */
func withEnvVariable(key string, value string, task func()) {
	oldValue := os.Getenv(key)
	defer os.Setenv(key, oldValue)

	os.Setenv(key, value)
	task()
}

/*
Tests that with the default setup, debug messages and params appear on stdout
 */
func TestDebug_Local(t *testing.T) {
	stdout := captureStdout(func() {
		WithFields(logrus.Fields{
			"name": "str param",
			"count": 1,
		}).Debug("test debug")
		WithField("count", 2).Debug("test debug 2")
	})

	lines := bytes.Split(stdout, []byte{'\n'})
	assert.Len(t, lines, 3)

	line1 := string(lines[0])
	assert.Contains(t, line1, `level=debug`)
	assert.Contains(t, line1, "test debug")
	assert.Contains(t, line1, `name="str param"`)
	assert.Contains(t, line1, `count=1`)

	line2 := string(lines[1])
	assert.Contains(t, line2, `level=debug`)
	assert.Contains(t, line2, "test debug 2")
	assert.Contains(t, line2, `count=2`)
}

/*
Tests that when level is set to info, debug messages are not printed
*/
func TestDebug_Local_Filtered(t *testing.T) {
	withEnvVariable("LOGGER_LEVEL", "info", func() {
		stdout := captureStdout(func() {
			WithFields(logrus.Fields{}).Debug("test info")
		})

		assert.Len(t, stdout, 0)
	})
}

/*
Tests that when level to an invalid value, it falls back to info after logging a warning
*/
func TestDebug_Local_InvalidLevel(t *testing.T) {
	withEnvVariable("LOGGER_LEVEL", "qdzeazaz", func() {
		var stdout []byte

		stdout = captureStdout(func() {
			WithFields(logrus.Fields{}).Debug("debug log")
		})

		assert.Contains(t, string(stdout), "Invalid LOGGER_LEVEL value, please use debug, info, warning or error")
		assert.Contains(t, string(stdout), "qdzeazaz")
		assert.NotContains(t, string(stdout), "debug log")

		stdout = captureStdout(func() {
			WithFields(logrus.Fields{}).Info("an info")
		})

		assert.Contains(t, string(stdout), "an info")
	})
}
/*
Tests that getLevelFromEnv return the expected value with correct input
*/
func TestGetLevelFromEnv(t *testing.T) {
	withEnvVariable("LOGGER_LEVEL", "debug", func() {
		level, err := getLevelFromEnv()
		assert.Nil(t, err)
		assert.Equal(t, logrus.DebugLevel, level)
	})
	withEnvVariable("LOGGER_LEVEL", "info", func() {
		level, err := getLevelFromEnv()
		assert.Nil(t, err)
		assert.Equal(t, logrus.InfoLevel, level)
	})
	withEnvVariable("LOGGER_LEVEL", "warning", func() {
		level, err := getLevelFromEnv()
		assert.Nil(t, err)
		assert.Equal(t, logrus.WarnLevel, level)
	})
	withEnvVariable("LOGGER_LEVEL", "error", func() {
		level, err := getLevelFromEnv()
		assert.Nil(t, err)
		assert.Equal(t, logrus.ErrorLevel, level)
	})
}
/*
Tests that when level to an invalid value that is a logrus level, it falls back to info after logging a warning
*/
func TestGetLevelFromEnv_InvalidLogrus(t *testing.T) {
	withEnvVariable("LOGGER_LEVEL", "warn", func() {
		level, err := getLevelFromEnv()
		assert.NotNil(t, err)
		assert.Equal(t, logrus.InfoLevel, level)
	})
	withEnvVariable("LOGGER_LEVEL", "fatal", func() {
		level, err := getLevelFromEnv()
		assert.NotNil(t, err)
		assert.Equal(t, logrus.InfoLevel, level)
	})
	withEnvVariable("LOGGER_LEVEL", "panic", func() {
		level, err := getLevelFromEnv()
		assert.NotNil(t, err)
		assert.Equal(t, logrus.InfoLevel, level)
	})
}
/*
Tests that with the default setup, info messages and params appear on stdout
 */
func TestInfo_Local(t *testing.T) {
	stdout := captureStdout(func() {
		WithFields(logrus.Fields{
			"name": "str param",
			"count": 1,
		}).Info("test info")
		WithField("count", 2).Info("test info 2")
	})
	lines := bytes.Split(stdout, []byte{'\n'})
	assert.Len(t, lines, 3)

	line1 := string(lines[0])
	assert.Contains(t, line1, `level=info`)
	assert.Contains(t, line1, "test info")
	assert.Contains(t, line1, `name="str param"`)
	assert.Contains(t, line1, `count=1`)

	line2 := string(lines[1])
	assert.Contains(t, line2, `level=info`)
	assert.Contains(t, line2, "test info 2")
	assert.Contains(t, line2, `count=2`)
}

/*
Tests that with the default setup, error messages and params appear on stdout
 */
func TestError_Local(t *testing.T) {
	stdout := captureStdout(func() {
		WithFields(logrus.Fields{
			"name": "str param",
			"count": 1,
		}).Error("test error")
		WithField("count", 2).Error("test error 2")

	})

	lines := bytes.Split(stdout, []byte{'\n'})
	assert.Len(t, lines, 3)

	line1 := string(lines[0])
	assert.Contains(t, line1, `level=error`)
	assert.Contains(t, line1, "test error")
	assert.Contains(t, line1, `name="str param"`)
	assert.Contains(t, line1, `count=1`)

	line2 := string(lines[1])
	assert.Contains(t, line2, `level=error`)
	assert.Contains(t, line2, "test error 2")
	assert.Contains(t, line2, `count=2`)
}

/*
Tests that with SENTRY_DSN set, info messages are *not sent* to the sentry server
 */
func TestError_Info(t *testing.T) {
	handle := func(res http.ResponseWriter, req *http.Request) {
		assert.Fail(t, "Sentry server was called for an info, which is not the intended behavior")
	}

	handler := http.HandlerFunc(handle)

	ts := httptest.NewServer(handler)
	defer ts.Close()

	testServerHost := strings.Split(ts.URL, "http://")[1]

	withEnvVariable("SENTRY_DSN", "http://aaa:bbb@" + testServerHost + "/123", func() {
		WithFields(logrus.Fields{
			"name": "str param",
			"count": 1,
		}).Info("test sentry error")
	})
	ReloadConfiguration()
}

/*
Groups all the stuff required to make a local mock sentry server
 */
type TestSentryServer struct {
	PacketChannel chan *raven.Packet
	Server        *httptest.Server
	Host          string
}

/*
Starts a mock sentry server that will accept all messages, parse then and send them to a channel.

Use the channel to receive messages sent to the server.
 */
func startMockSentryServer(t *testing.T) *TestSentryServer {
	packetChannel := make(chan *raven.Packet, 1)
	// inspired from
	// https://github.com/evalphobia/logrus_sentry/blob/162fd93cca6f0b1170f1b12c980a004744297f13/sentry_test.go#L45
	handle := func(res http.ResponseWriter, req *http.Request) {
		defer req.Body.Close()
		contentType := req.Header.Get("Content-Type")
		var bodyReader io.Reader = req.Body
		// underlying client will compress and encode payload above certain size
		if contentType == "application/octet-stream" {
			bodyReader = base64.NewDecoder(base64.StdEncoding, bodyReader)
			var err error
			bodyReader, err = zlib.NewReader(bodyReader)
			if err != nil {
				t.Fatal(err.Error())
			}
		}

		d := json.NewDecoder(bodyReader)
		p := &raven.Packet{}
		err := d.Decode(p)
		if err != nil {
			t.Fatal(err.Error())
		}
		packetChannel <- p
	}

	handler := http.HandlerFunc(handle)
	ts := httptest.NewServer(handler)
	testServerHost := strings.Split(ts.URL, "http://")[1]
	return &TestSentryServer{
		PacketChannel:packetChannel,
		Server:ts,
		Host:testServerHost,
	}
}
/*
Tests that with SENTRY_DSN set, error messages and params are sent to the sentry server
 */
func TestError_Sentry(t *testing.T) {
	ts := startMockSentryServer(t)
	defer ts.Server.Close()

	withEnvVariable("SENTRY_DSN", "http://aaa:bbb@" + ts.Host + "/123", func() {
		ReloadConfiguration()

		WithFields(logrus.Fields{
			"name": "str param",
			"count": 1,
		}).Error("test sentry error")
	})
	packet := <-ts.PacketChannel

	assert.Equal(t, raven.Severity("error"), packet.Level)
	assert.Equal(t, "test sentry error", packet.Message)
	assert.Equal(t, "str param", packet.Extra["name"])
	assert.EqualValues(t, 1, packet.Extra["count"])
	assert.Equal(t, "go", packet.Platform)
}

/*
Test that createSentryHook panics when an invalid sentry url is provided
 */
func TestCreateSentryHookInvalidUrl(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("Panic did not occur")
		}
	}()
	createSentryHook("Not a valid sentry URL")
}
