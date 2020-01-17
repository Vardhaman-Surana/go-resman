package logger

import (
	"context"
	"github.com/sirupsen/logrus"
	"os"
	"runtime"
	"strings"
)

type Logger struct {
	*logrus.Logger
}

var logger *Logger

func InitLogger(level logrus.Level) {
	logger = new(Logger)
	logger.Logger = logrus.New()
	logger.SetLevel(level)
	logger.SetOutput(os.Stdout)
	logger.SetFormatter(&logrus.JSONFormatter{
		FieldMap: logrus.FieldMap{
			logrus.FieldKeyTime: "@server_timestamp",
		},
	})
}

func LogInfo(requestId string, requestUrl string, msg string, status int) {
	fields := getFields(requestId, requestUrl, status)
	logger.WithFields(fields).Info(msg)
}

func LogWarn(requestId string, requestUrl string, msg string, status int) {
	fields := getFields(requestId, requestUrl, status)
	logger.WithFields(fields).Warn(msg)
}

func LogError(requestId string, requestUrl string, msg string, status int) {
	fields := getFields(requestId, requestUrl, status)
	logger.WithFields(fields).Error(msg)
}

func LogFatal(msg string){
	pc, _, _, ok := runtime.Caller(2)
	var fname string
	if ok {
		funcName := strings.Split(runtime.FuncForPC(pc).Name(), "/")
		fname = funcName[len(funcName)-1]
	}
	logger.WithFields(logrus.Fields{
		"function": fname,
	}).Fatal(msg)
}

func LogDebug(requestId string, requestUrl string, msg string) {
	fields := getFields(requestId, requestUrl, 0)
	logger.WithFields(fields).Debug(msg)
}

func getFields(requestId string, requestUrl string, status int) logrus.Fields {
	pc, file, lineNo, ok := runtime.Caller(2)
	var fname string
	if ok {
		funcName := strings.Split(runtime.FuncForPC(pc).Name(), "/")
		fname = funcName[len(funcName)-1]
	}
	if status == 0 {
		return logrus.Fields{
			"requestId":  requestId,
			"requestUrl": requestUrl,
			"filename":   file,
			"lineNo":     lineNo,
			"funcName":   fname,
		}
	}
	return logrus.Fields{
		"requestId":  requestId,
		"requestUrl": requestUrl,
		"status":     status,
		"filename":   file,
		"lineNo":     lineNo,
		"funcName":   fname,
	}
}

func GetRequestFieldsFromContext(ctx context.Context)(string,string){
	reqId:= ctx.Value("reqId")
	reqUrl:= ctx.Value("reqUrl")
	return reqId.(string),reqUrl.(string)
}
