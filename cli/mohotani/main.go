package main

import (
	"io/ioutil"
	"log"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/docker/docker/client"
	"github.com/docopt/docopt-go"
	"github.com/tjamet/mohotani/dns/lister"
	"github.com/tjamet/mohotani/dns/lister/docker"
	"github.com/tjamet/mohotani/dns/provider"
	"github.com/tjamet/mohotani/dns/provider/gandi"
	"github.com/tjamet/mohotani/dns/provider/log_provider"
	"github.com/tjamet/mohotani/dns/updater"
	"github.com/tjamet/mohotani/ip"
	"github.com/tjamet/mohotani/listener"
	"github.com/tjamet/mohotani/logger"
)

func stripAlign(in string) string {
	pattern, err := regexp.Compile("\n(?:[\t ]*[|]|[\t ]*)")
	if err != nil {
		log.Fatalf("Internal error, please report bug, error: %s", err.Error())
	}
	return pattern.ReplaceAllString(in, "\n")
}

func oneOf(args map[string]interface{}, keys ...string) string {
	provided := []string{}
	for _, key := range keys {
		value, ok := args[key]
		if ok && value != nil {
			b, ok := value.(bool)
			if !ok || b {
				provided = append(provided, key)
			}
		}
	}
	if len(provided) > 1 {
		log.Fatalf("Only one of %s should be provided", strings.Join(provided, ", "))
	}
	if len(provided) == 0 {
		log.Fatalf("One of %s must be provided", strings.Join(keys, ", "))
	}
	return provided[0]
}

func newIPListener(args map[string]interface{}, ticker <-chan time.Time, method string, logger logger.Logger) listener.Listener {
	switch method {
	case "static":
		ips := args["--ips.static.values"]
		if ips == nil {
			log.Fatal("static ips resolution requires IPs provided on the command line with --ips.static.values option, separated by comas")
		}
		return &listener.PollListener{
			Ticker: ticker,
			Logger: logger,
			Poll:   (&ip.Static{IPs: strings.Split(ips.(string), ",")}).Resolve,
		}
	case "ipify":
		ipify := ip.NewIPify()
		if url := args["--ips.ipify.url"]; url != nil {
			ipify.URL = url.(string)
		}
		return &listener.PollListener{
			Ticker: ticker,
			Logger: logger,
			Poll:   ipify.Resolve,
		}
	default:
		log.Fatalf("Unknown IP listener %s", method)
	}
	return nil
}

func newDNSUpdater(args map[string]interface{}, method string, logger logger.Logger) provider.Updater {
	switch method {
	case "log":
		return &logProvider.Log{
			Logger: logger,
		}
	case "gandi":
		keyVal := args["--gandi.key"]
		key := ""
		if keyVal != nil {
			key = keyVal.(string)
		} else {
			keyPath := args["--gandi.key-file"]
			if keyPath != nil {
				f, err := os.Open(keyPath.(string))
				if err != nil {
					log.Fatalf("Failed to open gandi key path %s : %s", keyPath.(string), err.Error())
				}
				b, err := ioutil.ReadAll(f)
				if err != nil {
					log.Fatalf("Failed to read gandi key path %s : %s", keyPath.(string), err.Error())
				}
				key = strings.Trim(string(b), " ")
			}
		}
		if key == "" {
			log.Fatalf("Missing gandi api key, please provide it through --gandi.key or --gandi.key-file")
		}
		return gandi.New(key)
	default:
		log.Fatalf("Unknown IP listener %s", method)
	}
	return nil
}

func newDomainListener(args map[string]interface{}, ticker <-chan time.Time, method string, logger logger.Logger) listener.Listener {
	switch method {
	case "static":
		ips := args["--domains.static.values"]
		if ips == nil {
			log.Fatal("static ips resolution requires IPs provided on the command line with --domains.static.values option, separated by comas")
		}
		return &listener.PollListener{
			Ticker: ticker,
			Logger: logger,
			Poll:   (&lister.Static{Domains: strings.Split(ips.(string), ",")}).List,
		}
	case "docker":
		cl, err := client.NewEnvClient()
		if err != nil {
			log.Fatalf("Failed to create docker client: %s", err.Error())
		}
		d := &docker.Lister{
			Client: cl,
			Logger: logger,
		}
		if args["--domains.docker.watch"].(bool) {
			ticker = d.EventTicker(ticker)
		}
		return &listener.PollListener{
			Ticker: ticker,
			Logger: logger,
			Poll:   d.List,
		}
	default:
		log.Fatalf("Unknown IP listener %s", method)
	}
	return nil
}

func main() {
	usage := `mohotani keeps your DNS records up to date
	|Usage: mohotani [options]

	|Options:
	|   --gandi                           Use gandi live DNS API to update DNS records
	|   --gandi.key=<key>                 The API key to connect to gandi
	|   --gandi.key-file=<path>           The path of a file containing the API key to connect to gandi
	|   --log                             Log domain changes only
	|   --domains.static                  Use a static list of domains to be updated, with domains provided on the command line
	|   --domains.static.values=<domains> The list of domains to be updated, coma separated values
	|   --domains.docker                  Use the docker domain lister. The list of domains will be retrieved from containers and services 
	|                                     using the Host matcher from traefik: https://docs.traefik.io/basics/#matchers
	|                                     The host connection must be specified by environment variables (DOCKER_*) defined here:
	|                                     https://docs.docker.com/engine/reference/commandline/cli/#environment-variables
	|   --domains.docker.watch            Refresh domain list everytime a container or service is deployed
	|   --ips.static                      Use the static IP resolver, with IPs given on the command line
	|   --ips.static.values=<ips>         The list of domains to be updated, coma separated valuse
	|   --ips.ipify                       Use ipify resolver to resolve the public IP address
	|   --ips.ipify.url=<url>             Use a different URL than the default one to reach the IPIFY API
	|   --watch.delay=<delay>             The interval at which IP or Domain list polling should occur (go ParseDuration format) [default: 5s]
	`
	args, err := docopt.Parse(stripAlign(usage), os.Args[1:], true, "0.0.0", false, true)
	if err != nil {
		log.Fatal(err)
	}
	duration, err := time.ParseDuration(args["--watch.delay"].(string))
	if err != nil {
		log.Fatalf(stripAlign(`Failed to parse duration: %s.
			|
			|A duration string is a possibly signed sequence of decimal numbers, each with optional fraction and a unit suffix,
			|such as "300ms", "-1.5h" or "2h45m". Valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h".`), err.Error())
	}
	logger := log.New(os.Stdout, "Mohotani: ", log.LstdFlags|log.Llongfile)
	providerMethod := strings.Replace(oneOf(args, "--gandi", "--log"), "--", "", 1)
	IPListenerMethod := strings.Replace(oneOf(args, "--ips.static", "--ips.ipify"), "--ips.", "", 1)
	DomainListenerMethod := strings.Replace(oneOf(args, "--domains.static", "--domains.docker"), "--domains.", "", 1)

	u := &updater.Updater{
		Updater:        newDNSUpdater(args, providerMethod, logger),
		IPListener:     newIPListener(args, time.NewTicker(duration).C, IPListenerMethod, logger),
		DomainListener: newDomainListener(args, time.NewTicker(duration).C, DomainListenerMethod, logger),
		Logger:         logger,
	}
	u.Start()
}
