// (c) 2012 Alexander Solovyov
// under terms of ISC license

package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	flags "github.com/jessevdk/go-flags"
	gostatic "github.com/piranha/gostatic/lib"
	"github.com/piranha/gostatic/processors"
)

const (
	// ExitCodeOk is used when the application exits without error.
	ExitCodeOk = 0
	// ExitCodeInvalidFlags is used when invalid flags are passed.
	ExitCodeInvalidFlags = 1
	// ExitCodeInvalidConfig is used when an invalid configuration file is given.
	ExitCodeInvalidConfig = 2
	// ExitCodeOther is used in all other situations.
	ExitCodeOther = 127
)

// Opts contains the flags which have been parsed by go-flags.
type Opts struct {
	ShowProcessors bool    `long:"processors" description:"show page processors"`
	ShowConfig     bool    `long:"show-config" description:"print config as JSON"`
	ShowSummary    bool    `long:"summary" description:"print all pages on stdout"`
	InitExample    *string `short:"i" long:"init" description:"create example site"`
	DumpPage       string  `short:"d" long:"dump" description:"print page metadata as JSON"`

	// checked in Page.Changed()
	Force bool `short:"f" long:"force" description:"force building all pages"`

	Watch bool   `short:"w" long:"watch" description:"serve site on HTTP and rebuild on changes"`
	Port  string `short:"p" long:"port" default:"8000" description:"port to serve on"`

	Verbose bool `short:"v" long:"verbose" description:"enable verbose output"`
	Version bool `short:"V" long:"version" description:"show version and exit"`
}

var opts Opts

func main() {
	argparser := flags.NewParser(&opts,
		flags.PrintErrors|flags.PassDoubleDash|flags.HelpFlag)
	argparser.Usage = "[OPTIONS] path/to/config\n\nBuild a site."

	args, err := argparser.Parse()
	if err != nil {
		errhandle(fmt.Errorf("cannot parse flags: %v", err))
		os.Exit(ExitCodeOther)
	}

	if opts.ShowSummary && opts.Watch {
		errhandle(fmt.Errorf("--summary and --watch do not mix together well"))
		os.Exit(ExitCodeOther)
	}

	if opts.Verbose {
		gostatic.DEBUG = true
	}

	if opts.Version {
		out("libgostatic %s\n", gostatic.VERSION)
		return
	}

	if opts.InitExample != nil {
		target, _ := os.Getwd()
		if len(*opts.InitExample) > 0 {
			target = filepath.Join(target, *opts.InitExample)
		}
		gostatic.WriteExample(target)
		return
	}

	gostatic.TemplateFuncMap["paginator"] = processors.CurrentPaginator

	procs := gostatic.ProcessorMap{
		"template":               processors.NewTemplateProcessor(),
		"inner-template":         processors.NewInnerTemplateProcessor(),
		"config":                 processors.NewConfigProcessor(),
		"markdown":               processors.NewMarkdownProcessor(),
		"ext":                    processors.NewExtProcessor(),
		"directorify":            processors.NewDirectorifyProcessor(),
		"tags":                   processors.NewTagsProcessor(),
		"paginate":               processors.NewPaginateProcessor(),
		"paginate-collect-pages": processors.NewPaginateCollectPagesProcessor(),
		"relativize":             processors.NewRelativizeProcessor(),
		"rename":                 processors.NewRenameProcessor(),
		"external":               processors.NewExternalProcessor(),
		"ignore":                 processors.NewIgnoreProcessor(),
	}

	if opts.ShowProcessors {
		procs.ProcessorSummary()
		return
	}

	if len(args) == 0 {
		argparser.WriteHelp(os.Stderr)
		os.Exit(ExitCodeInvalidFlags)
		return
	}

	config, err := gostatic.NewSiteConfig(args[0])
	if err != nil {
		errhandle(fmt.Errorf("invalid config file '%s': %v", args[0], err))
		os.Exit(ExitCodeInvalidConfig)
	}

	site := gostatic.NewSite(config, procs)

	if opts.Force {
		site.ForceRefresh = true
	}

	if opts.ShowConfig {
		x, err := json.MarshalIndent(config, "", "  ")
		errhandle(err)
		fmt.Fprintln(os.Stderr, string(x))
		return
	}

	if len(opts.DumpPage) > 0 {
		page := site.PageBySomePath(opts.DumpPage)
		if page == nil {
			out("Page '%s' not found (supply source or destination path)\n",
				opts.DumpPage)
			return
		}
		dump, err := json.MarshalIndent(page, "", "  ")
		errhandle(err)
		out("%s\n", dump)
		return
	}

	if opts.ShowSummary {
		site.Summary()
	} else {
		site.Render()
	}

	if opts.Watch {
		go gostatic.Watch(site)
		//StartWatcher(config, procs)
		out("Starting server at *:%s...\n", opts.Port)

		fs := http.FileServer(http.Dir(config.Output))
		http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Cache-Control", "no-store")
			fs.ServeHTTP(w, r)
		})

		err := http.ListenAndServe(":"+opts.Port, nil)
		errhandle(err)
	}
}
