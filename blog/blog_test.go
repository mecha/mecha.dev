package blog

import (
	"log/slog"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestInitDb(t *testing.T) {
	slog.SetLogLoggerLevel(slog.LevelError.Level())
	err := InitDB()
	defer DestroyDB()
	assert.Nil(t, err, "should be able to init db without error")
}

func TestInsertPost(t *testing.T) {
	err := InitDB()
	defer DestroyDB()
	assert.Nil(t, err, "should be able to init db without error")

	post := &Post{
		Slug:      "test",
		Title:     "Test Post",
		Excerpt:   "This is a test post.",
		Body:      "This is a test post.",
		Date:      time.Now(),
		Public:    true,
	}

	err = InsertPost(post)
	assert.Nil(t, err, "should insert post without error")
}

func TestGetPost(t *testing.T) {
	err := InitDB()
	defer DestroyDB()
	assert.Nil(t, err, "should be able to init db without error")

	date := time.Date(2025, 06, 29, 10, 15, 30, 0, time.UTC)
	insertPost := &Post{
		Slug:      "test",
		Title:     "Test Post",
		Excerpt:   "This is a test post.",
		Body:      "This is a test post.",
		Date:      date,
		Public:    true,
	}

	err = InsertPost(insertPost)
	assert.Nil(t, err, "should insert post without error")

	post, err := GetPost("test")
	assert.Nil(t, err, "should get post without error")
	assert.Equal(t, insertPost, post)
}

func TestDeletePost(t *testing.T) {
	err := InitDB()
	defer DestroyDB()
	assert.Nil(t, err, "should be able to init db without error")

	insertPost := &Post{
		Slug:      "test",
		Title:     "Test Post",
		Excerpt:   "This is a test post.",
		Body:      "This is a test post.",
		Date:      time.Now(),
		Public:    true,
	}

	err = InsertPost(insertPost)
	assert.Nil(t, err, "should insert post without error")

	_, err = GetPost("test")
	assert.Nil(t, err, "should get post without error")

	deleted, err := DeletePost("test")
	assert.Nil(t, err, "should delete post without error")
	assert.True(t, deleted)
}

func TestNumPosts(t *testing.T) {
	err := InitDB()
	defer DestroyDB()
	assert.Nil(t, err, "should be able to init db without error")

	posts := []*Post{
		{Slug: "post1", Public: true},
		{Slug: "post2", Public: true},
		{Slug: "post3", Public: false},
		{Slug: "post4", Public: true},
	}

	for _, post := range posts {
		err = InsertPost(post)
		assert.Nil(t, err, "should insert post without error")
	}

	num, err := NumPosts()
	assert.Nil(t, err, "should count posts without error")
	assert.Equal(t, 3, num)
}

func TestGetPosts(t *testing.T) {
	err := InitDB()
	defer DestroyDB()
	assert.Nil(t, err, "should be able to init db without error")

	p1 := &Post{Slug: "post1", Public: true}
	p2 := &Post{Slug: "post2", Public: true}
	p3 := &Post{Slug: "post3", Public: false}
	p5 := &Post{Slug: "post5", Public: true}
	p4 := &Post{Slug: "post4", Public: true}

	insertPosts := []*Post{p1, p2, p3, p4, p5}

	for _, post := range insertPosts {
		err = InsertPost(post)
		assert.Nil(t, err, "should insert post without error")
	}

	posts, err := GetPosts(3, 0)
	assert.Nil(t, err, "should get posts without error")
	assert.Equal(t, []*Post{p1, p2, p4}, posts)

	posts, err = GetPosts(3, 3)
	assert.Nil(t, err, "should get posts without error")
	assert.Equal(t, []*Post{p5}, posts)

	posts, err = GetPosts(3, 6)
	assert.Nil(t, err, "should get posts without error")
	assert.Equal(t, []*Post{}, posts)
}

func TestSearchPosts(t *testing.T) {
	err := InitDB()
	defer DestroyDB()
	assert.Nil(t, err, "should be able to init db without error")

	p1 := &Post{Slug: "post1", Public: true, Body: "cats and dogs"}
	p2 := &Post{Slug: "post2", Public: true, Body: "dogs and cats"}
	p3 := &Post{Slug: "post3", Public: false, Body: "catdog"}
	p4 := &Post{Slug: "post4", Public: true, Body: "you want a kitty?"}
	p5 := &Post{Slug: "post5", Public: true, Body: "you want a doggo?"}

	insertPosts := []*Post{p1, p2, p3, p4, p5}

	for _, post := range insertPosts {
		err = InsertPost(post)
		assert.Nil(t, err, "should insert post without error")
	}

	posts, err := SearchPosts("cats", 5, 0)
	assert.Nil(t, err, "should search posts without error")
	assert.Equal(t, []*Post{p1, p2}, posts)

	posts, err = SearchPosts("dog", 5, 0)
	assert.Nil(t, err, "should search posts without error")
	assert.Equal(t, []*Post{p1, p2, p5}, posts)

	posts, err = SearchPosts("dog", 2, 2)
	assert.Nil(t, err, "should search posts without error")
	assert.Equal(t, []*Post{p5}, posts)
}
