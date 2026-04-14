/*
这个文件定义会话系统的常量和抽象接口
*/
package session

import "time"

const (
	// CookieName 是浏览器中保存会话标识的 Cookie 名称
	CookieName = "sessionID"

	// SessionPrefix 是 Redis 中会话键的统一前缀
	SessionPrefix = "session:"

	// Expire 是默认的会话过期时间
	Expire = time.Hour
)

// Store 定义会话存储需要提供的能力
type Store interface {
	Create(userID string) (string, error)
	Get(sessionID string) (string, error)
	Update(sessionID, userID string) error
	Delete(sessionID string) error
}
