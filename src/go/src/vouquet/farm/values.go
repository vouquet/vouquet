package farm

const (
	TYPE_SELL string = "SELL"
	TYPE_BUY  string = "BUY"

	SOIL_GMO string = "coinzcom"
)

var (
	SOIL_ALL []string

	DEFAULT_OpeOption *OpeOption
)

func init() {
	SOIL_ALL = []string{}
	SOIL_ALL = append(SOIL_ALL, SOIL_GMO)

	DEFAULT_OpeOption = &OpeOption{
		Stream: true,
	}
}
