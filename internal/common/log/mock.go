package log

type MockLogger struct {
	*SugaredLogger
}

func (m MockLogger) Debug(i ...interface{}) {
}

func (m MockLogger) Debugf(s string, i ...interface{}) {
}

func (m MockLogger) Info(i ...interface{}) {
}

func (m MockLogger) Infof(s string, i ...interface{}) {
}

func (m MockLogger) Warn(i ...interface{}) {
}

func (m MockLogger) Warnf(s string, i ...interface{}) {
}

func (m MockLogger) Error(i ...interface{}) {
}

func (m MockLogger) Errorf(s string, i ...interface{}) {
}

func (m MockLogger) Fatal(i ...interface{}) {
}

func (m MockLogger) Fatalf(s string, i ...interface{}) {
}

func (m MockLogger) Panic(i ...interface{}) {
}

func (m MockLogger) Panicf(s string, i ...interface{}) {
}

func (m MockLogger) With(i ...interface{}) Logger {
	return m
}

func (m MockLogger) Sync() error {
	return nil
}

func (m MockLogger) Sugar() *SugaredLogger {
	return m.SugaredLogger
}
