package logging

type ILogger interface {
	Infof(format string, args ...interface{})
	Infoln(args ...interface{})
	Errorf(format string, args ...interface{})
	Warningf(format string, args ...interface{})
}
