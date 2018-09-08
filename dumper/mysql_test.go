package dumper

import (
	"bytes"
	"io"
	"strings"
	"testing"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func TestAnalyzeUsernameAndDatabase(t *testing.T) {
	out := new(bytes.Buffer)
	dumper := &MysqlDumper{
		logger: NewTestLogger(out),
	}
	// https://dev.mysql.com/doc/internals/en/connection-phase-packets.html
	in := []byte{
		0x54, 0x00, 0x00, 0x01, 0x8d, 0xa6, 0x0f, 0x00, 0x00, 0x00, 0x00, 0x01, 0x08, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x70, 0x61, 0x6d, 0x00, 0x14, 0xab, 0x09, 0xee, 0xf6, 0xbc, 0xb1, 0x32,
		0x3e, 0x61, 0x14, 0x38, 0x65, 0xc0, 0x99, 0x1d, 0x95, 0x7d, 0x75, 0xd4, 0x47, 0x74, 0x65, 0x73,
		0x74, 0x00, 0x6d, 0x79, 0x73, 0x71, 0x6c, 0x5f, 0x6e, 0x61, 0x74, 0x69, 0x76, 0x65, 0x5f, 0x70,
		0x61, 0x73, 0x73, 0x77, 0x6f, 0x72, 0x64, 0x00,
	}
	direction := ClientToRemote
	persistent := &DumpValues{}
	additional := []DumpValue{}

	err := dumper.Dump(in, direction, persistent, additional)
	if err != nil {
		t.Errorf("%v", err)
	}

	expected := []DumpValue{
		DumpValue{
			Key:   "username",
			Value: "pam",
		},
		DumpValue{
			Key:   "database",
			Value: "test",
		},
	}

	actual := persistent.Values
	if len(actual) != len(expected) {
		t.Errorf("actual %v\nwant %v", actual, expected)
	}
	if actual[0] != expected[0] {
		t.Errorf("actual %v\nwant %v", actual, expected)
	}
	if actual[1] != expected[1] {
		t.Errorf("actual %v\nwant %v", actual, expected)
	}

	log := out.String()

	if !strings.Contains(log, "") {
		t.Errorf("%v not be %v", log, "")
	}
}

// NewTestLogger ...
func NewTestLogger(out io.Writer) *zap.Logger {
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "ts",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.StringDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	logger := zap.New(zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderConfig),
		zapcore.AddSync(out),
		zapcore.DebugLevel,
	))

	return logger
}
