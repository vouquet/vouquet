package soil

const (
	SOIL_GMO string = "coinzcom"
)

var (
	SOIL_ALL []string
)

func init() {
	SOIL_ALL = []string{}
	SOIL_ALL = append(SOIL_ALL, SOIL_GMO)
}
