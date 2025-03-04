package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/fsnotify/fsnotify"
	"github.com/mecha/mecha.dev/blog"
	"github.com/mecha/mecha.dev/md"
	"github.com/mecha/mecha.dev/projects"
	"github.com/mecha/mecha.dev/views"
)

const (
	PostsDir    = "content/posts"
	ProjectsDir = "content/projects"
)

var (
	Flags   = FlagsObj{}
	Version = "<unknown>"
)

type FlagsObj struct {
	Verbose      bool
	Quiet        bool
	Watch        bool
	PortNum      int
}

func main() {
	parseFlags()

	initLogger()
	initBlog()
	initProjects()
	initVersion()

	if Flags.Watch {
		startPostFileWatcher()
		startProjectFileWatcher()
		startViewTemplateFileWatcher()
	}

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
			fmt.Printf("  -%s\t\t%s\n", f.Name, f.Usage)
		})
	}
	Flags = FlagsObj{}
	flag.BoolVar(&Flags.Verbose, "v", false, "Enables debug logging.")
	flag.BoolVar(&Flags.Quiet, "q", false, "Disables all non-error logging.")
	flag.BoolVar(&Flags.Watch, "w", false, "Watch blog post and view template files for changes.")
	flag.IntVar(&Flags.PortNum, "p", 8080, "The HTTP port to serve through.")
	flag.Parse()
}

func initLogger() {
	if Flags.Quiet {
		slog.SetLogLoggerLevel(slog.LevelError.Level())
	} else if Flags.Verbose {
		slog.SetLogLoggerLevel(slog.LevelDebug.Level())
	} else {
		slog.SetLogLoggerLevel(slog.LevelInfo.Level())
	}
}

func initBlog() {
	err := blog.InitDB()
	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}

	err = blog.LoadPostsFromDir(PostsDir)
	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}
}

func initProjects() {
	err := projects.LoadFromDir(ProjectsDir)
	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}
}

func initVersion() {
	raw, err := os.ReadFile(".git/refs/heads/main")
	if err == nil {
		Version = string(raw[:8])
	} else {
		slog.Warn(err.Error())
	}
}

func startPostFileWatcher() {
	slog.Debug("main: starting post file watcher")

	postWatcher := NewDirWatcher(PostsDir, func(event fsnotify.Event) {
		if event.Has(fsnotify.Write) {
			postFile := event.Name
			slog.Debug("main: reloading post", "file", postFile)

			err := blog.LoadPostFromFile(postFile)
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
	tmplWatcher := NewDirWatcher("./templates", func(event fsnotify.Event) {
		if event.Has(fsnotify.Write) {
			tmplFile := event.Name
			slog.Debug("main: invalidating cached view template", "file", tmplFile)
			views.ClearCache(tmplFile)
		}
	})

	err := tmplWatcher.Start()
	if err != nil {
		slog.Error("main: error starting view template file watcher: " + err.Error())
	}
}
