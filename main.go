package main

import (
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"time"

	"gopkg.in/alecthomas/kingpin.v2"

	"io/ioutil"
	"os"
	"os/exec"

	"github.com/fsnotify/fsnotify"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/log"
	"github.com/prometheus/common/version"
	"github.com/spf13/viper"
)

var (
	configFile    = kingpin.Flag("config.file", "Filename to configuration file, without extention (DEFAULT: inspec)").Default("inspec").String()
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
	module := r.URL.Query().Get("module")

	globalPath := viper.GetString("profile_path")
	if _, err := os.Stat(globalPath); os.IsNotExist(err) {
		// path/to/whatever does not exist
		http.Error(w, "'profile_path' in config is empty or does not exists", 500)
		inspecRequestErrors.Inc()
		return
	}

	start := time.Now()
	registry := prometheus.NewRegistry()

	m := Module{}
	if module != "" {
		if _, err := os.Stat(globalPath + "/" + module); os.IsNotExist(err) {
			http.Error(w, fmt.Sprintf("Unkown module '%s'", module), 400)
			inspecRequestErrors.Inc()
			return
		}

		//TODO: use defaults
		if viper.GetStringMap(module) == nil {
			m = Module{
				path:            globalPath + "/" + module,
				needSudo:        false,
				prefix:          "inspec_" + module + "_",
				sshIdentityFile: "",
				sshPort:         0,
				sshUser:         "",
			}
		} else {
			m = Module{
				path:            viper.GetString(fmt.Sprintf("%v.path", module)),
				needSudo:        viper.GetBool(fmt.Sprintf("%v.need_sudo", module)),
				prefix:          "inspec_" + viper.GetString(fmt.Sprintf("%v.prefix", module)) + "_",
				sshIdentityFile: viper.GetString(fmt.Sprintf("%v.ssh_identity_file", module)),
				sshPort:         viper.GetInt(fmt.Sprintf("%v.ssh_port", module)),
				sshUser:         viper.GetString(fmt.Sprintf("%v.ssh_user", module)),
			}
		}
		collector := collector{target: target, module: &m}
		registry.MustRegister(collector)
	} else {
		if _, err := os.Stat(globalPath + "/"); os.IsNotExist(err) {
			http.Error(w, fmt.Sprintf("Profule path '%s'", module), 400)
			inspecRequestErrors.Inc()
			return
		}
		profiles, err := ioutil.ReadDir(globalPath + "/")
		if err != nil {
			http.Error(w, "'profile_path' is not readable", 500)
			inspecRequestErrors.Inc()
			return
		}
		for _, profile := range profiles {
			//TODO: use defaults
			if viper.GetStringMap(profile.Name()) == nil {
				m = Module{
					path:            globalPath + "/" + profile.Name(),
					needSudo:        false,
					prefix:          "inspec_" + profile.Name() + "_",
					sshIdentityFile: "",
					sshPort:         0,
					sshUser:         "",
				}
			} else {
				m = Module{
					path:            viper.GetString(fmt.Sprintf("%v.path", profile.Name())),
					needSudo:        viper.GetBool(fmt.Sprintf("%v.need_sudo", profile.Name())),
					prefix:          "inspec_" + viper.GetString(fmt.Sprintf("%v.prefix", profile.Name())) + "_",
					sshIdentityFile: viper.GetString(fmt.Sprintf("%v.ssh_identity_file", profile.Name())),
					sshPort:         viper.GetInt(fmt.Sprintf("%v.ssh_port", profile.Name())),
					sshUser:         viper.GetString(fmt.Sprintf("%v.ssh_user", profile.Name())),
				}
			}
			collector := collector{target: target, module: &m}
			registry.Register(collector)
		}
	}

	// Delegate http serving to Promethues client library, which will call collector.Collect.
	h := promhttp.HandlerFor(registry, promhttp.HandlerOpts{})
	h.ServeHTTP(w, r)
	duration := float64(time.Since(start).Seconds())
	inspecDuration.WithLabelValues(module).Observe(duration)
	log.Debugf("Scrape of target '%s' with module '%s' took %f seconds", target, module, duration)
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
		panic(fmt.Errorf("fatal error config file: %s", err))
	}
	_, inspecLookErr := exec.LookPath(viper.GetString("inspec_path"))
	if inspecLookErr != nil {
		panic(inspecLookErr)
	}
	viper.WatchConfig()
	viper.OnConfigChange(func(e fsnotify.Event) {
		fmt.Println("Config file changed:", e.Name)
	})

	http.HandleFunc("/metrics", handler)

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
            	<h1>Inspec Exporter</h1>
            	<form action="/metrics">
            		<label>Target:</label> <input type="text" name="target" placeholder="X.X.X.X" value=""><br>
            		<label>Module:</label> <input type="text" name="module" placeholder="module" value="linux-baseline"><br>
            		<input type="submit" value="Submit">
            	</form>
				<p><a href="/metrics">Metrics</a></p>
            </body>
            </html>`))
	})

	log.Infof("Listening on %s", *listenAddress)
	log.Fatal(http.ListenAndServe(*listenAddress, nil))
}
