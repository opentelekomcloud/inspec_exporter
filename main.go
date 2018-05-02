package main

import (
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"time"

	"gopkg.in/alecthomas/kingpin.v2"

	"github.com/fsnotify/fsnotify"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/log"
	"github.com/prometheus/common/version"
	"github.com/spf13/viper"
	"os/exec"
)

var (
	configFile    = kingpin.Flag("config.file", "Path to configuration file.").Default("inspec").String()
	listenAddress = kingpin.Flag("web.listen-address", "Address to listen on for web interface and telemetry.").Default(":9124").String()

	// Metrics about the inspec exporter itself.
	inspecDuration = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Name: "inspec_collection_duration_seconds",
			Help: "Duration of collections by the inspec exporter",
		},
		[]string{"module"},
	)
	inspecRequestErrors = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "inspec_request_errors_total",
			Help: "Errors in requests to the inspec exporter",
		},
	)
)

func init() {
	prometheus.MustRegister(inspecDuration)
	prometheus.MustRegister(inspecRequestErrors)
	prometheus.MustRegister(version.NewCollector("inspec_exporter"))
}

func handler(w http.ResponseWriter, r *http.Request) {
	target := r.URL.Query().Get("target")
	if target == "" {
		http.Error(w, "'target' parameter must be specified", 400)
		inspecRequestErrors.Inc()
		return
	}
	moduleName := r.URL.Query().Get("module")
	module := viper.GetStringMap(moduleName)
	if module == nil {
		http.Error(w, fmt.Sprintf("Unkown module '%s'", moduleName), 400)
		inspecRequestErrors.Inc()
		return
	}
	log.Debugf("Scraping target '%s' with module '%s'", target, moduleName)

	start := time.Now()
	registry := prometheus.NewRegistry()

	m := Module{
		path:            viper.GetString(fmt.Sprintf("%v.path", moduleName)),
		needSudo:        viper.GetBool(fmt.Sprintf("%v.need_sudo", moduleName)),
		prefix:          viper.GetString(fmt.Sprintf("%v.prefix", moduleName)),
		sshIdentityFile: viper.GetString(fmt.Sprintf("%v.ssh_identity_file", moduleName)),
		sshPort:         viper.GetInt(fmt.Sprintf("%v.ssh_port", moduleName)),
		sshUser:         viper.GetString(fmt.Sprintf("%v.ssh_user", moduleName)),
	}
	collector := collector{target: target, module: &m}
	registry.MustRegister(collector)
	// Delegate http serving to Promethues client library, which will call collector.Collect.
	h := promhttp.HandlerFor(registry, promhttp.HandlerOpts{})
	h.ServeHTTP(w, r)
	duration := float64(time.Since(start).Seconds())
	inspecDuration.WithLabelValues(moduleName).Observe(duration)
	log.Debugf("Scrape of target '%s' with module '%s' took %f seconds", target, moduleName, duration)
}

func main() {
	log.AddFlags(kingpin.CommandLine)
	kingpin.Version(version.Print("inspec_exporter"))
	kingpin.HelpFlag.Short('h')
	kingpin.Parse()

	log.Infoln("Starting inspec exporter", version.Info())
	log.Infoln("Build context", version.BuildContext())

	viper.AddConfigPath(".")
	viper.SetConfigName(*configFile)              // name of config file (without extension)
	viper.AddConfigPath("/etc/inspec_exporter/")  // path to look for the config file in
	viper.AddConfigPath("$HOME/.inspec_exporter") // call multiple times to add many search paths
	err := viper.ReadInConfig()                   // Find and read the config file
	if err != nil {                               // Handle errors reading the config file
		panic(fmt.Errorf("Fatal error config file: %s \n", err))
	}
	_, inspecLookErr := exec.LookPath(viper.GetString("inspec_path"))
	if inspecLookErr != nil {
		panic(inspecLookErr)
	}
	viper.WatchConfig()
	viper.OnConfigChange(func(e fsnotify.Event) {
		fmt.Println("Config file changed:", e.Name)
	})

	http.Handle("/metrics", promhttp.Handler()) // Normal metrics endpoint for inspec exporter itself.
	http.HandleFunc("/inspec", handler)         // Endpoint to do inspec scrapes.

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<html>
            <head>
            <title>inspec Exporter</title>
            <style>
            label{
            display:inline-block;
            width:75px;
            }
            form label {
            margin: 10px;
            }
            form input {
            margin: 10px;
            }
            </style>
            </head>
            <body>
            	<h1>inspec Exporter</h1>
            	<form action="/inspec">
            		<label>Target:</label> <input type="text" name="target" placeholder="X.X.X.X" value="1.2.3.4"><br>
            		<label>Module:</label> <input type="text" name="module" placeholder="module" value="sudoers"><br>
            		<input type="submit" value="Submit">
            	</form>
				<p><a href="/metrics">Metrics</a></p>
            </body>
            </html>`))
	})

	log.Infof("Listening on %s", *listenAddress)
	log.Fatal(http.ListenAndServe(*listenAddress, nil))
}
