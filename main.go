package main

import (
	"embed"
	"flag"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"os/signal"
	"path"
	"strconv"
	"strings"
	"syscall"

	"github.com/fsnotify/fsnotify"
	"github.com/mecha/mecha.dev/blog"
	"github.com/mecha/mecha.dev/md"
	"github.com/mecha/mecha.dev/projects"
	"github.com/mecha/mecha.dev/views"
)

var (
	Flags   = FlagsObj{}
	Version = "dev"
	//go:embed embed
	embedFS embed.FS
)

type FlagsObj struct {
	Verbose bool
	Quiet   bool
	Watch   bool
	PortNum int
	NoEmbed bool
}

const (
	PostsDir     = "embed/content/posts"
	ProjectsDir  = "embed/content/projects"
	TemplatesDir = "embed/templates"
)

func main() {
	parseFlags()

	if Flags.Quiet {
		slog.SetLogLoggerLevel(slog.LevelError.Level())
	} else if Flags.Verbose {
		slog.SetLogLoggerLevel(slog.LevelDebug.Level())
	} else {
		slog.SetLogLoggerLevel(slog.LevelInfo.Level())
	}

	if err := loadBlog(); err != nil {
		slog.Error("error loading blog: " + err.Error())
		os.Exit(1)
	}
	if err := loadProjects(); err != nil {
		slog.Error("error loading projects: " + err.Error())
		os.Exit(1)
	}

	if Flags.NoEmbed && Flags.Watch {
		startPostFileWatcher()
		startProjectFileWatcher()
		startViewTemplateFileWatcher()
	}

	views.TemplateFS = getFS(TemplatesDir)
	go runHttpServer()

	intSig := make(chan os.Signal, 1)
	signal.Notify(intSig, os.Interrupt, syscall.SIGTERM)
	<-intSig

	slog.Debug("Shutting down...")
	blog.DestroyDB()
}

func parseFlags() {
	flag.Usage = func() {
		fmt.Println("mecha.dev server")
		fmt.Println("https://github.com/mecha/mecha.dev")
		fmt.Println()
		fmt.Println("FLAGS:")
		fmt.Println("  -h, --help\tShow this help message")
		flag.VisitAll(func(f *flag.Flag) {
			fmt.Printf("  -%s\t\t%s Default: %s\n", f.Name, f.Usage, f.DefValue)
		})
	}
	Flags = FlagsObj{}
	flag.BoolVar(&Flags.Verbose, "verbose", false, "Enables debug logging.")
	flag.BoolVar(&Flags.Quiet, "quiet", false, "Disables all non-error logging.")
	flag.BoolVar(&Flags.Watch, "watch", false, "Watch blog post and view template files for changes.")
	flag.IntVar(&Flags.PortNum, "port", 8080, "The HTTP port to serve through.")
	flag.BoolVar(&Flags.NoEmbed, "noembed", false, "Reads files from the OS filesystem instead of the embedded filesystem.")
	flag.Parse()
}

func getFS(path string) fs.FS {
	if Flags.NoEmbed {
		return os.DirFS(path)
	} else {
		fsys, err := fs.Sub(embedFS, path)
		if err != nil {
			panic(err)
		}
		return fsys
	}
}

func loadBlog() error {
	err := blog.InitDB()
	if err != nil {
		return err
	}

	fsys := getFS(PostsDir)
	entries, err := fs.ReadDir(fsys, ".")

	if os.IsNotExist(err) {
		entries = []os.DirEntry{}
	} else if err != nil {
		return err
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

		file, err := fsys.Open(name)
		if err != nil {
			return err
		}

		post, err := blog.ParsePost(file)
		if err != nil {
			return err
		}

		err = blog.InsertPost(post)
		if err != nil {
			return err
		}

		num++
	}

	slog.Info("blog: loaded " + strconv.Itoa(num) + " blog posts")

	return nil
}

func loadProjects() error {
	fsys := getFS(ProjectsDir)
	entries, err := fs.ReadDir(fsys, ".")

	if os.IsNotExist(err) {
		entries = []os.DirEntry{}
	} else if err != nil {
		return err
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

		file, err := fsys.Open(name)
		if err != nil {
			return err
		}

		project, err := projects.Parse(file)
		if err != nil {
			return err
		}

		projects.LoadIntoCache(name, project)
		num++
	}

	slog.Info("projects: loaded " + strconv.Itoa(num) + " projects")
	return nil
}

func startPostFileWatcher() {
	slog.Debug("main: starting post file watcher")

	postWatcher := NewDirWatcher(PostsDir, func(event fsnotify.Event) {
		if event.Has(fsnotify.Write) {
			postFile := event.Name
			slog.Debug("main: reloading post", "file", postFile)

			post, err := blog.ParsePostFile(postFile)
			if err != nil {
				slog.Error(err.Error())
				return
			}

			err = blog.InsertPost(post)
			if err != nil {
				slog.Error(err.Error())
			}
		}
	})

	err := postWatcher.Start()
	if err != nil {
		slog.Error("main: error starting post file watcher: " + err.Error())
	}
}

func startProjectFileWatcher() {
	slog.Debug("main: starting project file watcher")
	mdWatcher := NewDirWatcher(ProjectsDir, func(event fsnotify.Event) {
		if event.Has(fsnotify.Write) {
			projFile := event.Name
			slog.Debug("main: reloading project file", "file", projFile)
			md.ClearCache(projFile)
			_, err := projects.LoadFromFile(projFile)
			if err != nil {
				slog.Error(err.Error())
			}
		}
	})

	err := mdWatcher.Start()
	if err != nil {
		slog.Error("main: error starting projects markdown file watcher: " + err.Error())
	}
}

func startViewTemplateFileWatcher() {
	slog.Debug("main: starting view template file watcher")
	tmplWatcher := NewDirWatcher(TemplatesDir, func(event fsnotify.Event) {
		if event.Has(fsnotify.Write) {
			tmplFile := path.Base(event.Name)
			slog.Debug("main: invalidating cached view template", "file", tmplFile)
			views.ClearCache(tmplFile)
		}
	})

	err := tmplWatcher.Start()
	if err != nil {
		slog.Error("main: error starting view template file watcher: " + err.Error())
	}
}
