package logger

import(
	"github.com/rifflock/lfshook"
	"github.com/sirupsen/logrus"
)

//This still in early development and not in active use for logging yet.

//Create logger
var log *logrus.Logger

//Initialize Logging
func InitLogger() (*logrus.Logger) {
	if log != nil {
		return log
	}

	log = logrus.New()

	//Create log routing depending on severity
	pathMap := lfshook.PathMap{
		logrus.DebugLevel: "/Users/nkunkel/Programming/Go/logs/debug.log",
		logrus.InfoLevel:  "/Users/nkunkel/Programming/Go/logs/info.log",
		logrus.ErrorLevel: "/Users/nkunkel/Programming/Go/logs/error.log",
	}

	//Create formatter
	formatter := new(logrus.TextFormatter)

	//Set Timestamp format
	formatter.TimestampFormat = "02-01-2006 15:04:05"
	formatter.FullTimestamp = true

	log.Hooks.Add(lfshook.NewHook(
		pathMap,
		formatter,
	))

	log.SetFormatter(formatter)
	return log
}

func Debug(args ...interface{}) {

}

func Error(args ...interface{}) {

}

func Action(args ...interface{}) {

}

func Analytics(args ...interface{}) {

}