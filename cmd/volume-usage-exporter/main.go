package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"path/filepath"

	"github.com/golang/glog"
	"github.com/mmohamed/kubernetes-volume-usage-prometheus-metric/pkg/collectors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

const (
	metricsPath = "/metrics"
	healthzPath = "/healthz"
)

func metricsServer(registry prometheus.Gatherer, port int) {
	// Address to listen on for web interface and telemetry
	listenAddress := fmt.Sprintf(":%d", port)

	glog.Infof("Starting metrics server: %s", listenAddress)
	// Add metricsPath
	http.Handle(metricsPath, promhttp.HandlerFor(registry, promhttp.HandlerOpts{}))
	// Add healthzPath
	http.HandleFunc(healthzPath, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte("ok"))
	})
	// Add index
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<html>
	<head>
		<title>Kube Metrics Server</title>
	</head>
	<body>
		<h1>Kube Metrics</h1>
		<ul>
			<li><a href='` + metricsPath + `'>metrics</a></li>
			<li><a href='` + healthzPath + `'>healthz</a></li>
		</ul>
	</body>
</html>`))
	})
	log.Fatal(http.ListenAndServe(listenAddress, nil))
}

var (
	optHelp           bool
	optPort           int
	optKubeconfigPath string
)

func init() {
	flag.BoolVar(&optHelp, "help", false, "print help info and exit")
	flag.IntVar(&optPort, "port", 9001, "port to expose metrics on")
	if home := homedir.HomeDir(); home != "" {
		flag.StringVar(&optKubeconfigPath, "kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		flag.StringVar(&optKubeconfigPath, "kubeconfig", "", "(optional) absolute path to the kubeconfig file")
	}
}

func main() {
	// We log to stderr because glog will default to logging to a file.
	flag.Set("logtostderr", "true")
	flag.Parse()

	if optHelp {
		flag.Usage()
		return
	}

	// try to create the in-cluster config
	config, err := rest.InClusterConfig()
	if err != nil {
		// use the current context in kubeconfig
		config, err = clientcmd.BuildConfigFromFlags("", optKubeconfigPath)
		if err != nil {
			panic(err.Error())
		}
	}

	registry := prometheus.NewRegistry()
	registry.MustRegister(collectors.NewVolumeStatsCollector(config))

	metricsServer(registry, optPort)
}
