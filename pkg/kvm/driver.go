package kvm

import (
	"fmt"
	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/kubernetes-csi/csi-lib-utils/protosanitizer"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"k8s.io/klog"
	"net"
	"os"
	"strings"
)

type Driver struct {
	nodeID   string
	endpoint string
}

const (
	version       = "1.0.0"
	driverName    = "kvm.csi.dianduidian.com"
	TopologyKey   = "topology." + "dianduidian.com" + "/zone"
	DevicePathKey = "devicePath"
)

var (
	volCaps = []csi.VolumeCapability_AccessMode{
		{
			Mode: csi.VolumeCapability_AccessMode_SINGLE_NODE_WRITER,
		},
	}

	// controllerCaps represents the capability of controller service
	// https://github.com/container-storage-interface/spec/blob/master/lib/go/csi/csi.pb.go
	ctlCaps = []csi.ControllerServiceCapability_RPC_Type{
		csi.ControllerServiceCapability_RPC_CREATE_DELETE_VOLUME,
		csi.ControllerServiceCapability_RPC_PUBLISH_UNPUBLISH_VOLUME,
		csi.ControllerServiceCapability_RPC_GET_CAPACITY,
	}
)

func NewDriver(nodeID, endpoint string) *Driver {
	klog.V(4).Infof("Driver: %v version: %v", driverName, version)

	n := &Driver{
		nodeID:   nodeID,
		endpoint: endpoint,
	}

	return n
}

// func (n *kvmDriver) Run() {
// 	n.ns = NewNodeServer(n)

// 	// n.ns = NewNodeServer(n)
// 	s := NewNonBlockingGRPCServer()
// 	s.Start(n.endpoint,
// 		NewDefaultIdentityServer(n),
// 		// KVM plugin has not implemented ControllerServer
// 		// using default controllerserver.
// 		NewControllerServer(n),
// 		n.ns)
// 	s.Wait()
// }

func (d *Driver) Run() {

	ctl := NewControllerServer()
	identity := NewIdentityServer()
	node := NewNodeServer(d.nodeID)

	opts := []grpc.ServerOption{
		grpc.UnaryInterceptor(logGRPC),
	}

	srv := grpc.NewServer(opts...)

	csi.RegisterControllerServer(srv, ctl)
	csi.RegisterIdentityServer(srv, identity)
	csi.RegisterNodeServer(srv, node)

	proto, addr, err := ParseEndpoint(d.endpoint)
	if err != nil {
		klog.Fatal(err.Error())
	}

	if proto == "unix" {
		addr = "/" + addr
		if err := os.Remove(addr); err != nil && !os.IsNotExist(err) {
			klog.Fatalf("Failed to remove %s, error: %s", addr, err.Error())
		}
	}

	listener, err := net.Listen(proto, addr)
	if err != nil {
		klog.Fatalf("Failed to listen: %v", err)
	}

	srv.Serve(listener)
}

func logGRPC(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	klog.V(3).Infof("GRPC call: %s", info.FullMethod)
	klog.V(5).Infof("GRPC request: %s", protosanitizer.StripSecrets(req))
	resp, err := handler(ctx, req)
	if err != nil {
		klog.Errorf("GRPC error: %v", err)
	} else {
		klog.V(5).Infof("GRPC response: %s", protosanitizer.StripSecrets(resp))
	}
	return resp, err
}

func ParseEndpoint(ep string) (string, string, error) {
	if strings.HasPrefix(strings.ToLower(ep), "unix://") || strings.HasPrefix(strings.ToLower(ep), "tcp://") {
		s := strings.SplitN(ep, "://", 2)
		if s[1] != "" {
			return s[0], s[1], nil
		}
	}
	return "", "", fmt.Errorf("Invalid endpoint: %v", ep)
}
