package soil

type logger interface {
	WriteErr(string, ...interface{})
	WriteMsg(string, ...interface{})
	WriteDebug(string, ...interface{})
}
