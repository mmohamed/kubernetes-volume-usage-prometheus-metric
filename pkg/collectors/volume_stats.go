package collectors

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/golang/glog"
	"github.com/prometheus/client_golang/prometheus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/kustomize/kyaml/sets"
)

const (
	volumeStatsCapacityBytesKey  = "kubelet_volume_stats_capacity_bytes"
	volumeStatsAvailableBytesKey = "kubelet_volume_stats_available_bytes"
	volumeStatsUsedBytesKey      = "kubelet_volume_stats_used_bytes"
	volumeStatsInodesKey         = "kubelet_volume_stats_inodes"
	volumeStatsInodesFreeKey     = "kubelet_volume_stats_inodes_free"
	volumeStatsInodesUsedKey     = "kubelet_volume_stats_inodes_used"
)

var (
	volumeStatsCapacityBytes = prometheus.NewDesc(
		volumeStatsCapacityBytesKey,
		"Capacity in bytes of the volume",
		[]string{"namespace", "persistentvolumeclaim"}, nil,
	)
	volumeStatsAvailableBytes = prometheus.NewDesc(
		volumeStatsAvailableBytesKey,
		"Number of available bytes in the volume",
		[]string{"namespace", "persistentvolumeclaim"}, nil,
	)
	volumeStatsUsedBytes = prometheus.NewDesc(
		volumeStatsUsedBytesKey,
		"Number of used bytes in the volume",
		[]string{"namespace", "persistentvolumeclaim"}, nil,
	)
	volumeStatsInodes = prometheus.NewDesc(
		volumeStatsInodesKey,
		"Maximum number of inodes in the volume",
		[]string{"namespace", "persistentvolumeclaim"}, nil,
	)
	volumeStatsInodesFree = prometheus.NewDesc(
		volumeStatsInodesFreeKey,
		"Number of free inodes in the volume",
		[]string{"namespace", "persistentvolumeclaim"}, nil,
	)
	volumeStatsInodesUsed = prometheus.NewDesc(
		volumeStatsInodesUsedKey,
		"Number of used inodes in the volume",
		[]string{"namespace", "persistentvolumeclaim"}, nil,
	)
)

type volumeStatsCollector struct {
	clientset *kubernetes.Clientset
}

// NewVolumeStatsCollector creates a new volume stats prometheus collector.
func NewVolumeStatsCollector(config *rest.Config) prometheus.Collector {
	// create the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	return &volumeStatsCollector{clientset: clientset}
}

// Describe implements the prometheus.Collector interface.
func (collector *volumeStatsCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- volumeStatsCapacityBytes
	ch <- volumeStatsAvailableBytes
	ch <- volumeStatsUsedBytes
	ch <- volumeStatsInodes
	ch <- volumeStatsInodesFree
	ch <- volumeStatsInodesUsed
}

// Collect implements the prometheus.Collector interface.
func (collector *volumeStatsCollector) Collect(ch chan<- prometheus.Metric) {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	nodes, err := collector.clientset.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		panic(err.Error())
	}

	allPVCs := sets.String{}

	addGauge := func(desc *prometheus.Desc, parsedPVCRefData map[string]interface{}, v float64, lv ...string) {
		lv = append([]string{parsedPVCRefData["namespace"].(string), parsedPVCRefData["name"].(string)}, lv...)
		ch <- prometheus.MustNewConstMetric(desc, prometheus.GaugeValue, v, lv...)
	}

	for _, node := range nodes.Items {
		//fmt.Printf("Node %s \n", node.GetName())

		raw, err := collector.clientset.RESTClient().Get().RequestURI(fmt.Sprintf("/api/v1/nodes/%s/proxy/stats/summary", node.GetName())).DoRaw(ctx)
		if err != nil {
			panic(err.Error())
		}

		var parsedNodeData map[string]interface{}

		json.Unmarshal([]byte(raw), &parsedNodeData)

		parsedPodsData := parsedNodeData["pods"].([]interface{})

		for _, podData := range parsedPodsData {

			pod := podData.(map[string]interface{})

			//parsedPodRefData := pod["podRef"].(map[string]interface{})
			//fmt.Printf("Pod %s in %s \n", parsedPodRefData["name"], parsedPodRefData["namespace"])

			if pod["volume"] != nil {

				parsedVolumesData := pod["volume"].([]interface{})

				for _, volumeData := range parsedVolumesData {

					volume := volumeData.(map[string]interface{})

					if volume["pvcRef"] != nil {

						//fmt.Printf("Volume %s \n", volume["name"])
						parsedPVCRefData := volume["pvcRef"].(map[string]interface{})
						pvcUniqStr := fmt.Sprintf("%s/%s", parsedPVCRefData["namespace"], parsedPVCRefData["name"])

						if !allPVCs.Has(pvcUniqStr) {

							addGauge(volumeStatsCapacityBytes, parsedPVCRefData, volume["capacityBytes"].(float64))
							addGauge(volumeStatsAvailableBytes, parsedPVCRefData, volume["availableBytes"].(float64))
							addGauge(volumeStatsUsedBytes, parsedPVCRefData, volume["usedBytes"].(float64))
							addGauge(volumeStatsInodes, parsedPVCRefData, volume["inodes"].(float64))
							addGauge(volumeStatsInodesFree, parsedPVCRefData, volume["inodesFree"].(float64))
							addGauge(volumeStatsInodesUsed, parsedPVCRefData, volume["inodesUsed"].(float64))

							allPVCs.Insert(pvcUniqStr)
						}
					}
				}
			}
		}
	}
	glog.Infof("Found metrics for %d volume attached to %d nodes", len(allPVCs), len(nodes.Items))
}
