package namenode

type NameNodeConfig struct {
	Replication int
	Tolerance   int
}

type ConfigFunc func(*NameNodeConfig)

func defaultNameNodeConfig() *NameNodeConfig {
	return &NameNodeConfig{
		Replication: 2,
		Tolerance:   0,
	}
}

func NewNameNodeConfig(opts ...ConfigFunc) *NameNodeConfig {
	def := defaultNameNodeConfig()
	for _, fn := range opts {
		fn(def)
	}
	return def
}

func WithReplication(n int) ConfigFunc {
	return func(cfg *NameNodeConfig) {
		cfg.Replication = n
	}
}

func WithTolerance(n int) ConfigFunc {
	return func(cfg *NameNodeConfig) {
		cfg.Tolerance = n
	}
}
