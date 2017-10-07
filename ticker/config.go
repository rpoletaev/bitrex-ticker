package ticker

type Config struct {
	MaxQueryPerSecond int      `yaml:"max_query_per_second"`
	Markets           []string `yaml:"markets"`
}
