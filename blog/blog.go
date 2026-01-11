package blog

import (
	"database/sql"
	"errors"
	"io/fs"
	"log"
	"log/slog"
	"os"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

var db *sql.DB

func InitDB() error {
	if db != nil {
		return errors.New("blog is already initialized")
	}

	slog.Info("blog: initializing in-memory sqlite database")
	conn, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		return err
	}

	db = conn

	slog.Info("blog: creating posts table")
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS posts (
		slug TEXT PRIMARY KEY,
		title TEXT,
		excerpt TEXT,
		body TEXT,
		date TEXT,
		public INTEGER
	)`)
	if err != nil {
		return err
	}

	slog.Info("blog: creating fts virual table")
	_, err = db.Exec(`CREATE VIRTUAL TABLE IF NOT EXISTS posts_fts USING fts5(slug, title, body, tokenize = 'trigram')`)
	if err != nil {
		return err
	}

	slog.Info("blog: creating fts insert trigger")
	_, err = db.Exec(`CREATE TRIGGER IF NOT EXISTS posts_fts_insert AFTER INSERT ON posts
	BEGIN
		INSERT INTO posts_fts (slug, title, body) VALUES (NEW.slug, NEW.title, NEW.body);
	END`)
	if err != nil {
		return err
	}

	slog.Info("blog: creating fts delete trigger")
	_, err = db.Exec(`CREATE TRIGGER IF NOT EXISTS posts_fts_insert AFTER DELETE ON posts
	BEGIN
		DELETE FROM posts_fts WHERE slug = OLD.slug;
	END`)
	if err != nil {
		return err
	}

	return nil
}

func DestroyDB() {
	if db == nil {
		return
	}

	slog.Info("blog: tearing down sqlite database")
	err := db.Close()
	if err != nil {
		log.Fatal(err)
	}
	db = nil
}

func LoadFromFs(fsys fs.FS) (int, error) {
	entries, err := fs.ReadDir(fsys, ".")

	if os.IsNotExist(err) {
		entries = []os.DirEntry{}
	} else if err != nil {
		return 0, err
	}

	num := 0
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		if !strings.HasSuffix(name, ".md") {
			continue
		}

		post, err := ParsePostFile(fsys, name)
		if err != nil {
			return num, err
		}

		err = InsertPost(post)
		if err != nil {
			return num, err
		}

		num++
	}

	slog.Info("blog: loaded blog posts from filesystem", slog.Int("num", num))

	return num, nil
}

func GetPostBySlug(slug string) (*Post, error) {
	stmt, err := db.Prepare(`
		SELECT slug, title, excerpt, body, date, public
		FROM posts
		WHERE slug = ?
		LIMIT 1
	`)
	if err != nil {
		return nil, err
	}

	rows, err := stmt.Query(slug)
	if err != nil {
		return nil, err
	}

	found := rows.Next()
	if !found {
		return nil, sql.ErrNoRows
	}

	post, err := rowToPost(rows)
	rows.Next()

	return post, err
}

func NumPublicPosts() (int, error) {
	row := db.QueryRow("SELECT COUNT(slug) FROM posts WHERE public = true")
	count := 0
	err := row.Scan(&count)
	return count, err
}

func GetPosts(limit, offset int) ([]*Post, error) {
	stmt, err := db.Prepare(`
		SELECT slug, title, excerpt, body, date, public
		FROM posts
		WHERE public = true
		ORDER BY date(date) DESC
		LIMIT ? OFFSET ?
	`)
	if err != nil {
		return nil, err
	}

	rows, err := stmt.Query(limit, offset)
	if err != nil {
		return nil, err
	}

	return manyRowsToPosts(rows)
}

func SearchPosts(term string, limit, offset int) ([]*Post, error) {
	term = strings.TrimSpace(term)
	if len(term) < 3 {
		return GetPosts(limit, offset)
	}

	stmt, err := db.Prepare(`
		SELECT slug, title, excerpt, body, date, public
		FROM posts
		WHERE public = true AND slug IN (
			SELECT slug
			FROM posts_fts
			WHERE posts_fts MATCH ?
			ORDER BY rank
		)
		LIMIT ? OFFSET ?
	`)
	if err != nil {
		return nil, err
	}

	rows, err := stmt.Query(term, limit, offset)
	if err != nil {
		return nil, err
	}

	return manyRowsToPosts(rows)
}

func InsertPost(post *Post) error {
	stmt, err := db.Prepare(`
		REPLACE INTO posts (slug, title, excerpt, body, date, public) VALUES
		(?, ?, ?, ?, ?, ?)
	`)

	_, err = stmt.Exec(post.Slug, post.Title, post.Excerpt, post.Body, post.Date.Format(time.RFC3339), post.Public)
	return err
}

func DeletePost(slug string) (bool, error) {
	stmt, err := db.Prepare("DELETE FROM posts WHERE slug = ?")
	if err != nil {
		return false, err
	}
	res, err := stmt.Exec(slug)
	if err != nil {
		return false, err
	}
	num, err := res.RowsAffected()
	if err != nil {
		return false, err
	}
	return num > 0, nil
}

func DeleteAllPosts() error {
	_, err := db.Exec("DELETE FROM posts")
	if err != nil {
		return err
	}
	return nil
}

func manyRowsToPosts(rows *sql.Rows) ([]*Post, error) {
	posts := make([]*Post, 0)
	for rows.Next() {
		post, err := rowToPost(rows)
		if err != nil {
			return nil, err
		}
		posts = append(posts, post)
	}
	return posts, nil
}

func rowToPost(rows *sql.Rows) (*Post, error) {
	post := &Post{}
	dateStr, pubStr := "", 0

	err := rows.Scan(&post.Slug, &post.Title, &post.Excerpt, &post.Body, &dateStr, &pubStr)
	if err != nil {
		return nil, err
	}

	post.Public = pubStr != 0

	date, err := time.Parse(time.RFC3339, dateStr)
	if err != nil {
		return post, err
	}
	post.Date = date

	return post, nil
}
