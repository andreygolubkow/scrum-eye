package common

import "context"

// Facts is a shared cross-module scratchpad accumulated during the run.
// Modules may read and append to it using the context helpers below.

type Facts struct {
	FeatureIDs []int
}

type ctxFactsKey struct{}

// WithFacts returns a new context carrying the given Facts pointer.
func WithFacts(ctx context.Context, f *Facts) context.Context {
	return context.WithValue(ctx, ctxFactsKey{}, f)
}

// GetFacts fetches the Facts pointer from context, creating one if absent.
func GetFacts(ctx context.Context) *Facts {
	if v := ctx.Value(ctxFactsKey{}); v != nil {
		if f, ok := v.(*Facts); ok && f != nil {
			return f
		}
	}
	f := &Facts{}
	return f
}
