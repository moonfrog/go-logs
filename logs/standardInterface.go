package logs

type Logger interface {
	Printf(format string, v ...interface{})
}

type LoggerDef struct{}

func (this *LoggerDef) Printf(format string, v ...interface{}) {
	toPrint := []interface{}{format}
	toPrint = append(toPrint, v...)
	Infof(toPrint...)
}

var loggerInstance *LoggerDef = &LoggerDef{}

func StandardInterface() Logger {
	return loggerInstance
}
