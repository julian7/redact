package logger

// Generated by go generate, do not edit

// Debug writes a log entry using fmt.Sprint
func (l *Logger) Debug(attrs ...interface{}) {
	l.Log(DebugLevel, attrs...)
}

// Debugf writes a log entry using fmt.Sprintf
func (l *Logger) Debugf(attrs ...interface{}) {
	l.Logf(DebugLevel, attrs...)
}

// Info writes a log entry using fmt.Sprint
func (l *Logger) Info(attrs ...interface{}) {
	l.Log(InfoLevel, attrs...)
}

// Infof writes a log entry using fmt.Sprintf
func (l *Logger) Infof(attrs ...interface{}) {
	l.Logf(InfoLevel, attrs...)
}

// Warn writes a log entry using fmt.Sprint
func (l *Logger) Warn(attrs ...interface{}) {
	l.Log(WarnLevel, attrs...)
}

// Warnf writes a log entry using fmt.Sprintf
func (l *Logger) Warnf(attrs ...interface{}) {
	l.Logf(WarnLevel, attrs...)
}

// Error writes a log entry using fmt.Sprint
func (l *Logger) Error(attrs ...interface{}) {
	l.Log(ErrorLevel, attrs...)
}

// Errorf writes a log entry using fmt.Sprintf
func (l *Logger) Errorf(attrs ...interface{}) {
	l.Logf(ErrorLevel, attrs...)
}

// Fatal writes a log entry using fmt.Sprint
func (l *Logger) Fatal(attrs ...interface{}) {
	l.Log(FatalLevel, attrs...)
}

// Fatalf writes a log entry using fmt.Sprintf
func (l *Logger) Fatalf(attrs ...interface{}) {
	l.Logf(FatalLevel, attrs...)
}
