/*
定义会话系统使用的常量、过期时间和存储接口。
*/
package session

import "time"

const (
	// CookieName 是浏览器中保存会话标识的 Cookie 名称。
	CookieName = "sessionID"

	// SessionPrefix 是 Redis 中会话键的统一前缀。
	SessionPrefix = "session:"
)

// Expire 是当前生效的会话过期时间。
var Expire = time.Hour

// SetExpire 更新会话过期时间。
func SetExpire(expire time.Duration) {
	if expire > 0 {
		Expire = expire
	}
}

// Store 定义会话存储需要提供的能力。
type Store interface {
	Create(userID string) (string, error)
	Get(sessionID string) (string, error)
	Update(sessionID, userID string) error
	Delete(sessionID string) error
}
