package auth

import "context"

type ActorType string

const (
	ActorTypeUser   ActorType = "user"
	ActorTypeAPIKey ActorType = "apikey"
)

// Actor represents an authenticated principal.
type Actor struct {
	Type  ActorType
	ID    string
	Email string
}

type contextKey struct{}

// WithActor stores the actor on the request context.
func WithActor(ctx context.Context, actor Actor) context.Context {
	return context.WithValue(ctx, contextKey{}, actor)
}

// ActorFromContext returns the authenticated actor, if any.
func ActorFromContext(ctx context.Context) (Actor, bool) {
	v := ctx.Value(contextKey{})
	if v == nil {
		return Actor{}, false
	}
	actor, ok := v.(Actor)
	return actor, ok
}

// CreatedByLabel returns the value stored in config_versions.created_by.
func (a Actor) CreatedByLabel() string {
	switch a.Type {
	case ActorTypeUser:
		return a.Email
	case ActorTypeAPIKey:
		return "apikey:" + a.ID
	default:
		return a.ID
	}
}
