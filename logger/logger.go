package logger

import(
	"github.com/rifflock/lfshook"
	"github.com/sirupsen/logrus"
)

//Create logger
var Log *logrus.Logger

//Initialize Logging
func InitLogger() (*logrus.Logger) {
	if Log != nil {
		return Log
	}

	//initialize logger
	Log = logrus.New()

	//Create log routing depending on severity
	pathMap := lfshook.PathMap{
		logrus.InfoLevel:  "/Users/nkunkel/Programming/Go/logs/info.log",
		logrus.ErrorLevel: "/Users/nkunkel/Programming/Go/logs/error.log",
		logrus.FatalLevel: "/Users/nkunkel/Programming/Go/logs/error.log",
		logrus.DebugLevel: "/Users/nkunkel/Programming/Go/logs/debug.log",
	}

	//Create formatter
	formatter := &logrus.TextFormatter{
		TimestampFormat: "02-01-2006 15:04:05",
		FullTimestamp: true,
	}

	//Create hook
	hook := lfshook.NewHook(pathMap, formatter)

	//Add hook
	Log.AddHook(hook)

	//Set minimum threshold for logging to debug
	Log.SetLevel(logrus.DebugLevel)

	return Log
}

func Debug(args ...interface{}) {
	Log.Debug(args...)
}

func Info(args ...interface{}) {
	Log.Info(args...)
}

func Error(args ...interface{}) {
	Log.Error(args...)
}