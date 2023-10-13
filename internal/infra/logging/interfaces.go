package logging

type ILogger interface {
	Info(args ...interface{})
	Infof(format string, args ...interface{})
	Infoln(args ...interface{})
	Fatal(args ...interface{})
	Errorf(format string, args ...interface{})
	Warningf(format string, args ...interface{})
	Debugf(format string, args ...interface{})
}
