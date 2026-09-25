package bitrise

import (
	"context"
	"fmt"
)

type ctxKey int

const (
	keyPAT ctxKey = iota
	keyEnabledGroups
)

func patFromCtx(ctx context.Context) (string, error) {
	v := ctx.Value(keyPAT)
	u, ok := v.(string)
	if !ok {
		return "", fmt.Errorf("unexpected type %T", v)
	}
	return u, nil
}

// PATFromCtx returns the caller's raw Bitrise token from context, or "" when
// none is attached. Callers must never log or persist the value.
func PATFromCtx(ctx context.Context) string {
	v, _ := ctx.Value(keyPAT).(string)
	return v
}

func ContextWithPAT(ctx context.Context, s string) context.Context {
	return context.WithValue(ctx, keyPAT, s)
}

func EnabledGroupsFromCtx(ctx context.Context) ([]string, error) {
	v := ctx.Value(keyEnabledGroups)
	u, ok := v.([]string)
	if !ok {
		return nil, fmt.Errorf("unexpected type %T", v)
	}
	return u, nil
}

func ContextWithEnabledGroups(ctx context.Context, a []string) context.Context {
	return context.WithValue(ctx, keyEnabledGroups, a)
}
