package domain

import (
	"fmt"
	"gRPCServer/pkg/logger"
	"io"
	"os"
	"path"
	"path/filepath"
)

type CompositeLogger struct {
	RequestResponseLogger logger.Logger
	HttpTrafficLogger     logger.Logger
	ApplicationLogger     ApplicationLogger
}

type ApplicationLogger struct {
	errorWarnLogger logger.Logger
	debugLogger     logger.Logger
}

func (e ApplicationLogger) Debug(msg string, params map[string]interface{}) {
	e.debugLogger.Debug(msg, params)
}

func (e ApplicationLogger) Warn(msg string, params map[string]interface{}) {
	e.errorWarnLogger.Warn(msg, params)
}

func (e ApplicationLogger) Error(msg string, params map[string]interface{}) {
	e.errorWarnLogger.Error(msg, params)
}

type loggerWriters struct {
	GrpcTraffic io.Writer
	HttpTraffic io.Writer
	ErrorWarn   io.Writer
	Debug       io.Writer
}

type LoggerWritersPaths struct {
	GrpcTrafficFilePath string
	HttpTrafficFilePath string
	ErrorWarnFilePath   string
	DebugFilePath       string
}

func NewCompositeLogger(outDir string, loglvl string, paths LoggerWritersPaths) (CompositeLogger, error) {
	p := filepath.Clean(filepath.Join(outDir, paths.GrpcTrafficFilePath))
	grpcTrafficFile, err := os.OpenFile(p, os.O_APPEND, os.ModeDir)
	if err != nil {
		return CompositeLogger{}, fmt.Errorf("GrpcTrafficFilePath is incorrect or not enough rights."+
			" Details: %w", err)
	}
	p = filepath.Clean(filepath.Join(outDir, paths.HttpTrafficFilePath))
	httpTrafficFile, err := os.OpenFile(p, os.O_APPEND, os.ModeAppend)
	if err != nil {
		return CompositeLogger{}, fmt.Errorf("HttpTrafficFilePath is incorrect or not enough rights."+
			" Details: %w", err)
	}
	p = filepath.Clean(filepath.Join(outDir, paths.ErrorWarnFilePath))
	errorWarnFile, err := os.OpenFile(p, os.O_APPEND, os.ModeAppend)
	if err != nil {
		return CompositeLogger{}, fmt.Errorf("ErrorWarnFilePath is incorrect or not enough rights."+
			" Details: %w", err)
	}
	p = filepath.Clean(path.Join(outDir, paths.DebugFilePath))
	debugFile, err := os.OpenFile(p, os.O_APPEND, os.ModeAppend)
	if err != nil {
		return CompositeLogger{}, fmt.Errorf("DebugFilePath is incorrect or not enough rights."+
			" Details: %w", err)
	}
	writers := loggerWriters{
		GrpcTraffic: grpcTrafficFile,
		HttpTraffic: httpTrafficFile,
		ErrorWarn:   errorWarnFile,
		Debug:       debugFile,
	}
	return createCompositeLogger(writers, loglvl)
}

func createCompositeLogger(writers loggerWriters, loglvl string) (CompositeLogger, error) {
	reqRespLog, err := logger.NewLogrusLogger(writers.GrpcTraffic, loglvl)
	if err != nil {
		return CompositeLogger{}, fmt.Errorf("error in GrpcTrafficLogger. Details: %w", err)
	}
	httpLog, err := logger.NewLogrusLogger(writers.HttpTraffic, loglvl)
	if err != nil {
		return CompositeLogger{}, fmt.Errorf("error in HttpTrafficLogger. Details: %w", err)
	}
	errWarnLog, err := logger.NewLogrusLogger(writers.ErrorWarn, loglvl)
	if err != nil {
		return CompositeLogger{}, fmt.Errorf("error in ErrorWarnLogger. Details: %w", err)
	}
	debugLog, err := logger.NewLogrusLogger(writers.Debug, loglvl)
	if err != nil {
		return CompositeLogger{}, fmt.Errorf("error in DebugLogger. Details: %w", err)
	}
	return CompositeLogger{
		RequestResponseLogger: reqRespLog,
		HttpTrafficLogger:     httpLog,
		ApplicationLogger: ApplicationLogger{
			errorWarnLogger: errWarnLog,
			debugLogger:     debugLog,
		},
	}, nil
}
