package main

import (
	"flag"
	"k8s.io/klog"
	"kvm-csi-driver/pkg/kvm"
)

var (
	endpoint string
	nodeID   string
)

func init() {
	flag.Set("logtostderr", "true")
}

func main() {
	flag.StringVar(&endpoint, "endpoint", "", "CSI Endpoint")
	flag.StringVar(&nodeID, "nodeid", "", "node id")

	klog.InitFlags(nil)
	flag.Parse()

	d := kvm.NewDriver(nodeID, endpoint)
	d.Run()
}
