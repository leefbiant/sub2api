package service

import (
	"context"
	"strconv"
	"time"
)

// OpenAITokenCacheKey 生成 OpenAI OAuth 账号的缓存键
// 格式: "openai:account:{account_id}"
func OpenAITokenCacheKey(account *Account) string {
	return "openai:account:" + strconv.FormatInt(account.ID, 10)
}

// ClaudeTokenCacheKey 生成 Claude (Anthropic) OAuth 账号的缓存键
// 格式: "claude:account:{account_id}"
func ClaudeTokenCacheKey(account *Account) string {
	return "claude:account:" + strconv.FormatInt(account.ID, 10)
}


// TokenCache interface for token caching operations (extracted from removed geminicli)
type TokenCache interface {
	GetAccessToken(ctx context.Context, key string) (string, error)
	SetAccessToken(ctx context.Context, key string, token string, ttl time.Duration) error
	AcquireRefreshLock(ctx context.Context, key string, ttl time.Duration) (bool, error)
	ReleaseRefreshLock(ctx context.Context, key string) error
}

// GeminiTokenCacheKey stub - kept for compatibility with token_cache_invalidator
func GeminiTokenCacheKey(account *Account) string {
	return "gemini:account:" + strconv.FormatInt(account.ID, 10)
}