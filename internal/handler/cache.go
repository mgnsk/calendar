package handler

// Cache specifies cache interface.
type Cache[K comparable, V any] interface {
	Fetch(K, func() (V, error)) (V, error)
}

// NoopCache is a no-op cache.
type NoopCache[K comparable, V any] struct{}

// Fetch an item from the cache.
func (NoopCache[K, V]) Fetch(_ K, f func() (V, error)) (V, error) {
	return f()
}
