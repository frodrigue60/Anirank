package og

import "log"

const DefaultMaxConcurrent = 3

// TryGenerate runs fn only if a render slot is available.
// If the process is at capacity, acquired is false and fn is not called.
// Cache hits should never go through this path.
func (g *Generator) TryGenerate(fn func() error) (acquired bool, err error) {
	if g.sem == nil {
		return true, fn()
	}
	select {
	case g.sem <- struct{}{}:
		defer func() { <-g.sem }()
		return true, fn()
	default:
		log.Printf("[OG] concurrency limit reached (max=%d)", cap(g.sem))
		return false, nil
	}
}

func newSemaphore(maxConcurrent int) chan struct{} {
	if maxConcurrent <= 0 {
		maxConcurrent = DefaultMaxConcurrent
	}
	return make(chan struct{}, maxConcurrent)
}
