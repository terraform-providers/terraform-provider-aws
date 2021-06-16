package elasticache

const (
	EngineMemcached = "memcached"
	EngineRedis     = "redis"
)

// Engine_Values returns all elements of the Engine enum
func Engine_Values() []string {
	return []string{
		EngineMemcached,
		EngineRedis,
	}
}
