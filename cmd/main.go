package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/pprof"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/hashicorp/hcl"
	gitlab "github.com/xanzy/go-gitlab"

	"github.com/cosminilie/gitbot"

	"github.com/cosminilie/gitbot/plugins"
	_ "github.com/cosminilie/gitbot/plugins/droprights"
	_ "github.com/cosminilie/gitbot/plugins/lgtm"
)

var (
	//Vars that need to be set at build time using :
	//go build  -ldflags "-X main.majorVersion=1 -X main.minorVersion=0 -X main.gitVersion=c553786277bf05c0aa0320b7c7fc8249c73a27c0
	// -X main.buildDate=`date -u +%Y-%m-%d_%H:%M:%S`" -o BackupService
	majorVersion = "MajorVersion not set"
	minorVersion = "MinorVersion not set"
	gitVersion   = "GitVersion not set"
	buildDate    = "BuildDate not set"
)

//Config struct in loading HCL configuration
type Config struct {
	Token            string         `hcl:"token"`
	GitURL           string         `hcl:"git-api-URL"`
	DefaultApprovers []string       `hcl:"default-approvers"`
	Repos            []plugins.Repo `hcl:"repo,expand"`
}

func main() {

	var (
		debugAddr   = flag.String("debug.addr", ":9090", "Debug and metrics listen address")
		showVersion = flag.Bool("version", false, "Display build version")
		configFile  = flag.String("config", "/etc/githook.conf", "GitLab Hook config file")
	)
	flag.Parse()

	if *showVersion {
		fmt.Printf("Major Version: %s\t Minor Version: %s\nGitVersion: %s\t,BuildDate: %s\n", majorVersion, minorVersion, gitVersion, buildDate)
		os.Exit(0)
	}

	// Logging domain.
	var logger log.Logger
	{
		logger = log.NewLogfmtLogger(os.Stdout)
		logger = log.NewContext(logger).With("ts", log.DefaultTimestampUTC)
		logger = log.NewContext(logger).With("caller", log.DefaultCaller)
	}
	logger.Log("msg", "hello")
	defer logger.Log("msg", "goodbye")

	//global error chan
	errc := make(chan error)

	// Interrupt handler.
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		errc <- fmt.Errorf("%s", <-c)
	}()

	// Business domain.

	//create config object
	content, err := ioutil.ReadFile(*configFile)
	if err != nil {
		fmt.Printf("Error loading configuration file %s. Failed with: %s\n", *configFile, err)
		os.Exit(1)

	}
	conf := &Config{}
	hclParseTree, err := hcl.ParseBytes(content)
	if err != nil {
		fmt.Printf("Failed to parse configuration data: %s\n", err)
		os.Exit(1)
	}
	if err := hcl.DecodeObject(&conf, hclParseTree); err != nil {
		fmt.Printf("Failed to decode configuration data: %s\n", err)
		os.Exit(1)

	}

	//fmt.Println("Config file is: ", *configFile)

	//create gilabclient
	var client *gitlab.Client
	httpclient := &http.Client{
		Timeout: 10 * time.Second,
	}

	client = gitlab.NewClient(httpclient, conf.Token)
	client.SetBaseURL(conf.GitURL)

	//create service
	var service gitbot.Service
	svclogger := log.NewContext(logger).With("service", "basicservice")
	service = gitbot.NewBasicService(svclogger, client, conf.Repos, conf.DefaultApprovers)
	go func() {
		for e := range service.GetErrors() {
			errc <- e
		}
	}()

	//business domain
	httplogger := log.NewContext(logger).With("transport", "HTTP")
	httpserver := &gitbot.Server{
		Logger:  httplogger,
		Service: service,
	}

	// Debug listener.
	go func() {
		m := http.NewServeMux()
		m.Handle("/debug/pprof/", http.HandlerFunc(pprof.Index))
		m.Handle("/debug/pprof/cmdline", http.HandlerFunc(pprof.Cmdline))
		m.Handle("/debug/pprof/profile", http.HandlerFunc(pprof.Profile))
		m.Handle("/debug/pprof/symbol", http.HandlerFunc(pprof.Symbol))
		m.Handle("/debug/pprof/trace", http.HandlerFunc(pprof.Trace))
		logger.Log("addr", *debugAddr)
		errc <- http.ListenAndServe(*debugAddr, m)

	}()

	// HTTP transport.
	go func() {
		logger.Log("addr", ":9091")

		m := http.NewServeMux()
		m.Handle("/hook", httpserver)
		errc <- http.ListenAndServe(":9091", m)
	}()
	fmt.Println(<-errc)
}
