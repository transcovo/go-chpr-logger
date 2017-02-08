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
Tests that with the default setup, debug messages and params appear on stdout
 */
func TestDebug_Local(t *testing.T) {
	stdout := captureStdout(func() {
		WithFields(Fields{
			"name": "str param",
			"count": 1,
		}).Debug("My message")
		WithField("count", 2).Debug("My message 2")
	})

	lines := bytes.Split(stdout, []byte{'\n'})
	assert.Len(t, lines, 3)

	line1 := string(lines[0])
	assert.Contains(t, line1, `level=debug`)
	assert.Contains(t, line1, "My message")
	assert.Contains(t, line1, `name="str param"`)
	assert.Contains(t, line1, `count=1`)

	line2 := string(lines[1])
	assert.Contains(t, line1, `level=debug`)
	assert.Contains(t, line2, "My message 2")
	assert.Contains(t, line2, `count=2`)
}

/*
Tests that with the default setup, info messages and params appear on stdout
 */
func TestInfo_Local(t *testing.T) {
	stdout := captureStdout(func() {
		WithFields(Fields{
			"name": "str param",
			"count": 1,
		}).Info("My message")
		WithField("count", 2).Info("My message 2")
	})
	lines := bytes.Split(stdout, []byte{'\n'})
	assert.Len(t, lines, 3)

	line1 := string(lines[0])
	assert.Contains(t, line1, `level=info`)
	assert.Contains(t, line1, "My message")
	assert.Contains(t, line1, `name="str param"`)
	assert.Contains(t, line1, `count=1`)

	line2 := string(lines[1])
	assert.Contains(t, line1, `level=info`)
	assert.Contains(t, line2, "My message 2")
	assert.Contains(t, line2, `count=2`)
}

/*
Tests that with the default setup, error messages and params appear on stdout
 */
func TestError_Local(t *testing.T) {
	stdout := captureStdout(func() {
		WithFields(Fields{
			"name": "str param",
			"count": 1,
		}).Error("My message")
		WithField("count", 2).Error("My message 2")

	})

	lines := bytes.Split(stdout, []byte{'\n'})
	assert.Len(t, lines, 3)

	line1 := string(lines[0])
	assert.Contains(t, line1, `level=error`)
	assert.Contains(t, line1, "My message")
	assert.Contains(t, line1, `name="str param"`)
	assert.Contains(t, line1, `count=1`)

	line2 := string(lines[1])
	assert.Contains(t, line1, `level=error`)
	assert.Contains(t, line2, "My message 2")
	assert.Contains(t, line2, `count=2`)
}

/*
Tests that with SENTRY_DSN set, info messages are *not sent* to the sentry server
 */
func TestError_Info(t *testing.T) {
	oldSentryDns := os.Getenv("SENTRY_DSN")
	defer os.Setenv("SENTRY_DSN", oldSentryDns)

	handle := func(res http.ResponseWriter, req *http.Request) {
		assert.Fail(t, "Sentry server was called for an info, which is not the intended behavior")
	}

	handler := http.HandlerFunc(handle)

	ts := httptest.NewServer(handler)
	defer ts.Close()

	testServerHost := strings.Split(ts.URL, "http://")[1]

	os.Setenv("SENTRY_DSN", "http://aaa:bbb@" + testServerHost + "/123")
	ReloadConfiguration()

	WithFields(Fields{
		"name": "str param",
		"count": 1,
	}).Info("My message")
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
	oldSentryDns := os.Getenv("SENTRY_DSN")
	defer os.Setenv("SENTRY_DSN", oldSentryDns)

	ts := startMockSentryServer(t)
	defer ts.Server.Close()

	os.Setenv("SENTRY_DSN", "http://aaa:bbb@" + ts.Host + "/123")
	ReloadConfiguration()

	WithFields(Fields{
		"name": "str param",
		"count": 1,
	}).Error("My message")

	packet := <-ts.PacketChannel

	assert.Equal(t, raven.Severity("error"), packet.Level)
	assert.Equal(t, "My message", packet.Message)
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
