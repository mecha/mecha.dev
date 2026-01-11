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

	if err := blog.InitDB(); err != nil {
		slog.Error("failed to initialize blog", slog.String("cause", err.Error()))
		os.Exit(1)
	}
	if _, err := blog.LoadFromFs(getFS(PostsDir)); err != nil {
		slog.Error("failed to load blog posts", slog.String("cause", err.Error()))
		os.Exit(1)
	}

	if _, err := projects.LoadFromFs(getFS(ProjectsDir)); err != nil {
		slog.Error("failed to load projects", slog.String("cause", err.Error()))
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

	slog.Info("Shutting down...")
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

func startPostFileWatcher() {
	slog.Debug("main: starting blog post file watcher")
	fsys := os.DirFS(".")

	postWatcher := NewDirWatcher(PostsDir, func(event fsnotify.Event) {
		if event.Has(fsnotify.Write) {
			postFile := event.Name
			slog.Debug("main: reloading blog post", "file", postFile)

			post, err := blog.ParsePostFile(fsys, postFile)
			if err != nil {
				slog.Error("failed to parse blog post file", slog.String("cause", err.Error()))
				return
			}

			err = blog.InsertPost(post)
			if err != nil {
				slog.Error("failed to insert/update blog post", slog.String("cause", err.Error()))
				return
			}
		}
	})

	err := postWatcher.Start()
	if err != nil {
		slog.Error("main: failed to start post file watcher", slog.String("cause", err.Error()))
	}
}

func startProjectFileWatcher() {
	slog.Debug("main: starting project file watcher")
	fsys := os.DirFS(".")

	mdWatcher := NewDirWatcher(ProjectsDir, func(event fsnotify.Event) {
		if event.Has(fsnotify.Write) {
			projFile := event.Name
			slog.Debug("main: reloading project file", "file", projFile)
			md.ClearCache(projFile)
			_, err := projects.LoadFromFile(fsys, projFile)
			if err != nil {
				slog.Error("main: failed to reload project file", slog.String("cause", err.Error()))
				return
			}
		}
	})

	err := mdWatcher.Start()
	if err != nil {
		slog.Error("main: failed to start projects markdown file watcher", slog.String("cause", err.Error()))
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
		slog.Error("main: failed to start view template file watcher", slog.String("cause", err.Error()))
	}
}
