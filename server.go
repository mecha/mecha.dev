package main

import (
	"database/sql"
	"errors"
	"log"
	"log/slog"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/caddyserver/certmagic"
	"github.com/mecha/mecha.dev/blog"
	"github.com/mecha/mecha.dev/projects"
	"github.com/mecha/mecha.dev/views"
)

const NumPostsPerPage = 20

func runHttpServer() {
	handler := createHttpHandler()

	var err error
	if Flags.TLSEnabled {
		certmagic.DefaultACME.Agreed = true
		certmagic.DefaultACME.Email = "mail@miguelmuscat.me"
		if Flags.UseStagingCA {
			slog.Info("Using staging CA")
			certmagic.DefaultACME.CA = certmagic.LetsEncryptStagingCA
		}
		slog.Info("Serving HTTPS")
		err = certmagic.HTTPS([]string{"mecha.dev"}, handler)
	} else {
		addr := ":" + strconv.Itoa(Flags.PortNum)
		slog.Info("Serving HTTP")
		err = http.ListenAndServe(addr, handler)
	}

	if !errors.Is(err, http.ErrServerClosed) {
		slog.Error(err.Error())
	}
}

func createHttpHandler() http.Handler {
	mux := http.NewServeMux()

	assetsFs := http.StripPrefix("/assets", http.FileServer(http.Dir("public/assets")))
	mux.HandleFunc("/assets/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/assets/" {
			w.WriteHeader(404) // prevent listing contents of assets dir
		} else {
			assetsFs.ServeHTTP(w, r)
		}
	})

	projectsFs := http.StripPrefix("/projects", http.FileServer(http.Dir("public/projects")))
	mux.HandleFunc("/projects/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/projects/" {
			views.Write("templates/projects.gotmpl", w, projects.GetAll())
		} else {
			projectsFs.ServeHTTP(w, r)
		}
	})

	mux.HandleFunc("/blog", func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()
		search := strings.TrimSpace(query.Get("q"))

		page, err := strconv.Atoi(query.Get("page"))
		if page < 1 || err != nil {
			page = 1
		}

		pageSize, err := strconv.Atoi(query.Get("num"))
		if pageSize < 1 || err != nil {
			pageSize = NumPostsPerPage
		}

		posts, err := blog.SearchPosts(search, pageSize, pageSize*(page-1))
		if err != nil {
			log.Fatal(err)
		}

		total, err := blog.NumPosts()
		if err != nil {
			log.Fatal(err)
		}

		numPages := int(math.Ceil(float64(total) / float64(pageSize)))

		views.Write("templates/blog.gotmpl", w, map[string]any{
			"Posts":    posts,
			"Search":   search,
			"Page":     page,
			"NumPages": numPages,
		})
	})

	mux.HandleFunc("/blog/{id}", func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		post, err := blog.GetPost(id)
		if err == nil {
			views.Write("templates/blog-post.gotmpl", w, post)
		} else if errors.Is(err, sql.ErrNoRows) {
			w.WriteHeader(404)
			views.Write("templates/404.gotmpl", w, nil)
		} else {
			w.WriteHeader(500)
			views.Write("templates/500.gotmpl", w, err)
		}
	})

	mux.HandleFunc("/blog/feed", func(w http.ResponseWriter, r *http.Request) {
		format := r.URL.Query().Get("format")
		if format == "" {
			format = "rss"
		}

		page, err := strconv.Atoi(r.URL.Query().Get("page"))
		if page < 1 || err != nil {
			page = 1
		}

		err = blog.WriteFeed(w, NumPostsPerPage, page, format)
		if err != nil {
			w.WriteHeader(500)
			views.Write("templates/500.gotmpl", w, err)
		}
	})

	mux.HandleFunc("/about", func(w http.ResponseWriter, r *http.Request) {
		data := map[string]any{
			"ProYears":       time.Now().Year() - 2013,
			"HobbyYearsMore": 2013 - 2006,
		}
		views.Write("templates/about.gotmpl", w, data)
	})

	mux.HandleFunc("/robots.txt", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("User-Agent: *\n"))
		w.Write([]byte("Allow: /"))
	})

	mux.HandleFunc("/500", func(w http.ResponseWriter, r *http.Request) {
		views.Write("templates/500.gotmpl", w, errors.New("something is about to blow"))
	})

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			views.Write("templates/home.gotmpl", w, nil)
		} else {
			w.WriteHeader(404)
			views.Write("templates/404.gotmpl", w, nil)
		}
	})

	return mux
}
