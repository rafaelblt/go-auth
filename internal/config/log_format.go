package config

type LogFormat string

const (
	LogFormatJSON LogFormat = "json"
	LogFormatText LogFormat = "text"
)

func (f LogFormat) valid() bool {
	return f == LogFormatJSON || f == LogFormatText
}
