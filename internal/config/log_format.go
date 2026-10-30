package config

type LogFormat string

const (
	LogFormatJSON LogFormat = "json"
	LogFormatText LogFormat = "text"
)

var checkLogFormat = oneOf(LogFormatJSON, LogFormatText)
