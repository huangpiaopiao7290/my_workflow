package constant

const (
	EnvDev  = "dev"
	EnvTest = "test"
)

const (
	HttpSuccess = 0
	HttpFailure = 1
)

func GetMessage(code int) string {
	switch code {
	case HttpSuccess:
		return "success"
	case HttpFailure:
		return "failed"
	default:
		return "unknown"
	}
}
