package gnmi

import (
	"cmf/proto/gnmi_ext"
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
	"log"
)

type GNMIServer struct {
}

func (c *GNMIServer) Capabilities(ctx context.Context, in *CapabilityRequest) (*CapabilityResponse, error) {
	log.Printf("Receive message %s", in)
	return nil, status.Errorf(codes.Unimplemented, "method Capabilities not implemented")
}
func (c *GNMIServer) Get(ctx context.Context, in *GetRequest) (*GetResponse, error) {
	log.Printf("Receive message %s", in)
	return nil, status.Errorf(codes.Unimplemented, "method Get not implemented")
}
func (c *GNMIServer) Set(ctx context.Context, in *SetRequest) (*SetResponse, error) {
	log.Printf("Receive message %s", in)
	return nil, status.Errorf(codes.Unimplemented, "method Set not implemented")
}
func (c *GNMIServer) Subscribe(GNMI_SubscribeServer) error {
	log.Printf("Receive message %s", in)
	return status.Errorf(codes.Unimplemented, "method Subscribe not implemented")
}
func (c *GNMIServer) SubscribeWeb(in *SubscribeRequest, server GNMI_SubscribeWebServer) error {
	log.Printf("Receive message %s", in)
	return status.Errorf(codes.Unimplemented, "method SubscribeWeb not implemented")
}
