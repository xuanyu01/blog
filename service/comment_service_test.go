/*
覆盖评论服务的核心业务测试。
*/
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
	createParentID int64
	createUsername string
	createContent  string
	createErr      error
	author         string
	deleteID       int64
}

func (f *fakeCommentRepo) ListByPostID(postID int64) ([]model.Comment, error) {
	return f.comments, nil
}

func (f *fakeCommentRepo) Create(postID int64, parentID int64, username, content string) (*model.Comment, error) {
	f.createPostID = postID
	f.createParentID = parentID
	f.createUsername = username
	f.createContent = content
	if f.createErr != nil {
		return nil, f.createErr
	}
	if f.created != nil {
		return f.created, nil
	}
	return &model.Comment{ID: 1, PostID: postID, Username: username, Content: content, CreatedAt: time.Now()}, nil
}

func (f *fakeCommentRepo) GetAuthorByID(commentID int64) (string, error) {
	if f.author == "" {
		return "", sql.ErrNoRows
	}
	return f.author, nil
}

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

func (f *fakeCommentBlogRepo) GetByID(blogID int64) (*model.Blog, error) {
	if f.blog == nil {
		return nil, sql.ErrNoRows
	}
	return f.blog, nil
}

func TestCommentServiceCreateCommentRequiresLogin(t *testing.T) {
	service := NewCommentService(&fakeCommentRepo{}, &fakeCommentBlogRepo{
		blog: &model.Blog{ID: 1, Status: "published", AuthorUsername: "alice"},
	})

	_, err := service.CreateComment(1, 0, "hello", "", model.PermissionUser)
	if err == nil || err.Error() != "unauthorized" {
		t.Fatalf("expected unauthorized error, got %v", err)
	}
}

func TestCommentServiceCreateCommentRejectsEmptyContent(t *testing.T) {
	service := NewCommentService(&fakeCommentRepo{}, &fakeCommentBlogRepo{
		blog: &model.Blog{ID: 1, Status: "published", AuthorUsername: "alice"},
	})

	_, err := service.CreateComment(1, 0, "   ", "bob", model.PermissionUser)
	if err == nil || err.Error() != "content is required" {
		t.Fatalf("expected content required error, got %v", err)
	}
}

func TestCommentServiceCreateCommentRequiresAccessibleBlog(t *testing.T) {
	service := NewCommentService(&fakeCommentRepo{}, &fakeCommentBlogRepo{
		blog: &model.Blog{ID: 1, Status: "draft", AuthorUsername: "alice"},
	})

	_, err := service.CreateComment(1, 0, "hello", "bob", model.PermissionUser)
	if err == nil || err.Error() != "forbidden" {
		t.Fatalf("expected forbidden error, got %v", err)
	}
}

func TestCommentServiceCreateCommentPassesParentID(t *testing.T) {
	repo := &fakeCommentRepo{}
	service := NewCommentService(repo, &fakeCommentBlogRepo{
		blog: &model.Blog{ID: 1, Status: "published", AuthorUsername: "alice"},
	})

	_, err := service.CreateComment(1, 9, "reply", "bob", model.PermissionUser)
	if err != nil {
		t.Fatalf("CreateComment returned error: %v", err)
	}
	if repo.createParentID != 9 {
		t.Fatalf("expected parent id 9, got %d", repo.createParentID)
	}
}

func TestCommentServiceCreateCommentRejectsNegativeParentID(t *testing.T) {
	service := NewCommentService(&fakeCommentRepo{}, &fakeCommentBlogRepo{
		blog: &model.Blog{ID: 1, Status: "published", AuthorUsername: "alice"},
	})

	_, err := service.CreateComment(1, -1, "reply", "bob", model.PermissionUser)
	if err == nil || err.Error() != "invalid parent comment id" {
		t.Fatalf("expected invalid parent comment id error, got %v", err)
	}
}

func TestCommentServiceCreateCommentNormalizesMissingParent(t *testing.T) {
	service := NewCommentService(&fakeCommentRepo{createErr: sql.ErrNoRows}, &fakeCommentBlogRepo{
		blog: &model.Blog{ID: 1, Status: "published", AuthorUsername: "alice"},
	})

	_, err := service.CreateComment(1, 9, "reply", "bob", model.PermissionUser)
	if err == nil || err.Error() != "parent comment not found" {
		t.Fatalf("expected parent comment not found error, got %v", err)
	}
}

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

func TestCommentServiceDeleteCommentRejectsOtherUser(t *testing.T) {
	repo := &fakeCommentRepo{author: "alice"}
	service := NewCommentService(repo, &fakeCommentBlogRepo{})

	err := service.DeleteComment(7, "bob", model.PermissionUser)
	if err == nil || err.Error() != "forbidden" {
		t.Fatalf("expected forbidden error, got %v", err)
	}
}
