/*
auth_test.go 覆盖登录守卫和权限守卫中间件的核心行为。
*/
package middleware

import (
	"blog/model"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

type fakeSessionStore struct {
	values map[string]string
}

func (f *fakeSessionStore) Create(userID string) (string, error)  { return "", nil }
func (f *fakeSessionStore) Update(sessionID, userID string) error { return nil }
func (f *fakeSessionStore) Delete(sessionID string) error         { return nil }
func (f *fakeSessionStore) Get(sessionID string) (string, error) {
	value, ok := f.values[sessionID]
	if !ok {
		return "", errors.New("not found")
	}
	return value, nil
}

type fakeCurrentUserProvider struct {
	user model.UserView
	err  error
}

func (f *fakeCurrentUserProvider) CurrentUser(sessionID string) (model.UserView, error) {
	return f.user, f.err
}

// TestRequireLoginRejectsMissingCookie 验证未登录请求会被登录守卫拦截。
func TestRequireLoginRejectsMissingCookie(t *testing.T) {
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	ctx, engine := gin.CreateTestContext(recorder)
	engine.Use(RequireLogin(&fakeSessionStore{}))
	engine.GET("/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	request := httptest.NewRequest(http.MethodGet, "/test", nil)
	ctx.Request = request
	engine.HandleContext(ctx)

	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", recorder.Code)
	}
}

// TestRequireManagerRejectsUser 验证普通用户不能通过管理权限守卫。
func TestRequireManagerRejectsUser(t *testing.T) {
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	engine := gin.New()
	engine.Use(RequireManager(
		&fakeSessionStore{values: map[string]string{"session-1": "alice"}},
		&fakeCurrentUserProvider{
			user: model.UserView{
				UserName:   "alice",
				Permission: model.PermissionUser,
				IsLogin:    true,
			},
		},
	))
	engine.GET("/admin", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	request := httptest.NewRequest(http.MethodGet, "/admin", nil)
	request.AddCookie(&http.Cookie{Name: "sessionID", Value: "session-1"})
	engine.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", recorder.Code)
	}
}

// TestRequireAdminAllowsAdmin 验证管理员可以通过管理员守卫。
func TestRequireAdminAllowsAdmin(t *testing.T) {
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	engine := gin.New()
	engine.Use(RequireAdmin(
		&fakeSessionStore{values: map[string]string{"session-1": "root"}},
		&fakeCurrentUserProvider{
			user: model.UserView{
				UserName:   "root",
				Permission: model.PermissionAdmin,
				IsLogin:    true,
			},
		},
	))
	engine.GET("/admin", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	request := httptest.NewRequest(http.MethodGet, "/admin", nil)
	request.AddCookie(&http.Cookie{Name: "sessionID", Value: "session-1"})
	engine.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", recorder.Code)
	}
}
