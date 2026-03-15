package auth

import "context"

type contextKey string

const UserIDContextKey contextKey = "auth_user_id"

const ClaimsContextKey contextKey = "auth_claims"

func WithUserID(ctx context.Context, userID int) context.Context {
	return context.WithValue(ctx, UserIDContextKey, userID)
}

func UserIDFromContext(ctx context.Context) (int, bool) {
	id, ok := ctx.Value(UserIDContextKey).(int)
	return id, ok
}

func WithClaims(ctx context.Context, claims Claims) context.Context {
	ctx = context.WithValue(ctx, UserIDContextKey, claims.UserID)
	return context.WithValue(ctx, ClaimsContextKey, claims)
}

func ClaimsFromContext(ctx context.Context) (Claims, bool) {
	claims, ok := ctx.Value(ClaimsContextKey).(Claims)
	return claims, ok
}
