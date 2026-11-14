package logger

import (
	"database/sql"
	"os"

	zaplogfmt "github.com/jsternberg/zap-logfmt"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/DeRuina/timberjack"
	"github.com/google/uuid"
)

var (
	Log      *zap.Logger
	logError *zap.Logger
)

func Init(logsPath string) error {
	writer := &timberjack.Logger{
		Filename:         logsPath,
		MaxSize:          100,
		MaxBackups:       30,
		MaxAge:           30,
		Compression:      "gzip",
		LocalTime:        true,
		RotateAt:         []string{"00:00"},
		BackupTimeFormat: "2006-01-02-15-04-05",
		FileMode:         0o600,
	}

	fileWriter := zapcore.AddSync(writer)
	consoleWriter := zapcore.AddSync(os.Stdout)

	fileCfg := zap.NewProductionEncoderConfig()
	fileCfg.EncodeTime = zapcore.TimeEncoderOfLayout("15:04:05.000")
	fileCfg.EncodeLevel = zapcore.CapitalLevelEncoder
	fileCfg.EncodeCaller = zapcore.ShortCallerEncoder

	consoleCfg := zap.NewProductionEncoderConfig()
	consoleCfg.EncodeTime = zapcore.RFC3339NanoTimeEncoder
	consoleCfg.EncodeCaller = zapcore.ShortCallerEncoder

	core := zapcore.NewTee(
		zapcore.NewCore(zaplogfmt.NewEncoder(fileCfg), fileWriter, zapcore.DebugLevel),
		zapcore.NewCore(zapcore.NewJSONEncoder(consoleCfg), consoleWriter, zapcore.DebugLevel),
	)

	l := zap.New(core, zap.AddCaller())

	Log = l
	logError = l.WithOptions(zap.AddCallerSkip(1))
	return nil
}

func LogError(msg string, service string) {
	if msg == sql.ErrNoRows.Error() {
		return
	}

	requestID := uuid.New().String()

	logError.Error(
		msg,
		zap.String("service", service),
		zap.String("request_id", requestID),
	)
}
