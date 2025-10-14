package log

import (
	"os"
	"path/filepath"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

var Logger *zap.SugaredLogger

// Init creates a zap logger that writes only to a rotating log file.
// verbose=true → sets log level to Debug, otherwise Info.
func Init(verbose bool) {
	// Determine log directory
	logDir, err := os.UserCacheDir()
	if err != nil {
		logDir = "."
	}
	logDir = filepath.Join(logDir, "goweather", "logs")
	_ = os.MkdirAll(logDir, 0755)

	logFile := filepath.Join(logDir, time.Now().Format("2006-01-02")+".log")

	// Configure rotation
	rotate := &lumberjack.Logger{
		Filename:   logFile,
		MaxSize:    10, // MB
		MaxBackups: 5,
		MaxAge:     7, // days
		Compress:   true,
	}

	// Encoder for JSON structured logs
	fileEncCfg := zap.NewProductionEncoderConfig()
	fileEncCfg.EncodeTime = zapcore.ISO8601TimeEncoder
	fileEncCfg.EncodeLevel = zapcore.CapitalLevelEncoder

	level := zap.InfoLevel
	if verbose {
		level = zap.DebugLevel
	}

	fileCore := zapcore.NewCore(
		zapcore.NewJSONEncoder(fileEncCfg),
		zapcore.AddSync(rotate),
		level,
	)

	// Build zap logger
	zapLogger := zap.New(fileCore, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))
	Logger = zapLogger.Sugar()

	Logger.Infow("File logging initialized",
		"file", logFile,
		"level", level.String(),
	)
}

// Sync flushes any buffered entries to disk.
func Sync() {
	if Logger != nil {
		_ = Logger.Sync()
	}
}

// Convenience wrappers
func Info(msg string, args ...interface{})  { Logger.Infof(msg, args...) }
func Warn(msg string, args ...interface{})  { Logger.Warnf(msg, args...) }
func Error(msg string, args ...interface{}) { Logger.Errorf(msg, args...) }
func Debug(msg string, args ...interface{}) { Logger.Debugf(msg, args...) }
func Fatal(msg string, args ...interface{}) { Logger.Fatalf(msg, args...) }
