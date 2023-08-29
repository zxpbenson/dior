package lg

type Appender interface {
	Output(maxdepth int, s string) error
}
