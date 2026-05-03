/*
comment_service_test.go 覆盖评论服务的核心业务测试。*/
package service

import (
	"blog/model"
	"database/sql"
	"testing"
	"time"
)

type fakeCommentRepo struct {
	comments       []model.Comment
	created        *model.Comment
	createPostID   int64
	createUsername string
	createContent  string
	author         string
	deleteID       int64
}

// ListByPostID 模拟读取评论列表。
func (f *fakeCommentRepo) ListByPostID(postID int64) ([]model.Comment, error) {
	return f.comments, nil
}

// Create 模拟创建评论。
func (f *fakeCommentRepo) Create(postID int64, username, content string) (*model.Comment, error) {
	f.createPostID = postID
	f.createUsername = username
	f.createContent = content
	if f.created != nil {
		return f.created, nil
	}
	return &model.Comment{ID: 1, PostID: postID, Username: username, Content: content, CreatedAt: time.Now()}, nil
}

// GetAuthorByID 模拟读取评论作者。
func (f *fakeCommentRepo) GetAuthorByID(commentID int64) (string, error) {
	if f.author == "" {
		return "", sql.ErrNoRows
	}
	return f.author, nil
}

// Delete 模拟删除评论。
func (f *fakeCommentRepo) Delete(commentID int64) error {
	f.deleteID = commentID
	if commentID == 404 {
		return sql.ErrNoRows
	}
	return nil
}

type fakeCommentBlogRepo struct {
	blog *model.Blog
}

// GetByID 模拟读取博客详情。
func (f *fakeCommentBlogRepo) GetByID(blogID int64) (*model.Blog, error) {
	if f.blog == nil {
		return nil, sql.ErrNoRows
	}
	return f.blog, nil
}

// TestCommentServiceCreateCommentRequiresLogin 验证发表评论需要登录。
func TestCommentServiceCreateCommentRequiresLogin(t *testing.T) {
	service := NewCommentService(&fakeCommentRepo{}, &fakeCommentBlogRepo{
		blog: &model.Blog{ID: 1, Status: "published", AuthorUsername: "alice"},
	})

	_, err := service.CreateComment(1, "hello", "", model.PermissionUser)
	if err == nil || err.Error() != "unauthorized" {
		t.Fatalf("expected unauthorized error, got %v", err)
	}
}

// TestCommentServiceCreateCommentRejectsEmptyContent 验证空评论会被拒绝。
func TestCommentServiceCreateCommentRejectsEmptyContent(t *testing.T) {
	service := NewCommentService(&fakeCommentRepo{}, &fakeCommentBlogRepo{
		blog: &model.Blog{ID: 1, Status: "published", AuthorUsername: "alice"},
	})

	_, err := service.CreateComment(1, "   ", "bob", model.PermissionUser)
	if err == nil || err.Error() != "content is required" {
		t.Fatalf("expected content required error, got %v", err)
	}
}

// TestCommentServiceCreateCommentRequiresAccessibleBlog 验证无权访问的博客不能评论。
func TestCommentServiceCreateCommentRequiresAccessibleBlog(t *testing.T) {
	service := NewCommentService(&fakeCommentRepo{}, &fakeCommentBlogRepo{
		blog: &model.Blog{ID: 1, Status: "draft", AuthorUsername: "alice"},
	})

	_, err := service.CreateComment(1, "hello", "bob", model.PermissionUser)
	if err == nil || err.Error() != "forbidden" {
		t.Fatalf("expected forbidden error, got %v", err)
	}
}

// TestCommentServiceDeleteCommentAllowsAuthor 验证评论作者可以删除自己的评论。
func TestCommentServiceDeleteCommentAllowsAuthor(t *testing.T) {
	repo := &fakeCommentRepo{author: "alice"}
	service := NewCommentService(repo, &fakeCommentBlogRepo{})

	if err := service.DeleteComment(3, "alice", model.PermissionUser); err != nil {
		t.Fatalf("DeleteComment returned error: %v", err)
	}
	if repo.deleteID != 3 {
		t.Fatalf("expected delete id 3, got %d", repo.deleteID)
	}
}

// TestCommentServiceDeleteCommentAllowsManager 。。֤。。。。Ա。。。。ɾ。。。。。ۡ。
func TestCommentServiceDeleteCommentAllowsManager(t *testing.T) {
	repo := &fakeCommentRepo{author: "alice"}
	service := NewCommentService(repo, &fakeCommentBlogRepo{})

	if err := service.DeleteComment(5, "manager", model.PermissionAdmin); err != nil {
		t.Fatalf("DeleteComment returned error: %v", err)
	}
	if repo.deleteID != 5 {
		t.Fatalf("expected delete id 5, got %d", repo.deleteID)
	}
}

// TestCommentServiceDeleteCommentRejectsOtherUser 验证普通用户不能删除他人评论。
func TestCommentServiceDeleteCommentRejectsOtherUser(t *testing.T) {
	repo := &fakeCommentRepo{author: "alice"}
	service := NewCommentService(repo, &fakeCommentBlogRepo{})

	err := service.DeleteComment(7, "bob", model.PermissionUser)
	if err == nil || err.Error() != "forbidden" {
		t.Fatalf("expected forbidden error, got %v", err)
	}
}

