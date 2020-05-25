package gnmi

import (
	"context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"log"
	"time"
)

type Server struct {

}

func (c *Server) Capabilities(ctx context.Context, in *CapabilityRequest) (*CapabilityResponse, error) {
	log.Printf("Receive message %s", in)
	return nil, status.Errorf(codes.Unimplemented, "method Capabilities not implemented")
}
func (s *Server) Get(ctx context.Context, req *GetRequest) (*GetResponse, error) {
	log.Printf("func (s *Server) Get(): Receive message %s", req)
	if req.GetType() != GetRequest_ALL {
		return nil, status.Errorf(codes.Unimplemented, "unsupported request type: %s", GetRequest_DataType_name[int32(req.GetType())])
	}
	prefix := req.GetPrefix()
	paths := req.GetPath()
	notifications := make([]*Notification, len(paths))
	for i, path := range paths {
		val := &TypedValue{
			Value: &TypedValue_IntVal{
				IntVal: int64((i + 1) * 30),
			},
		}
		ts := time.Now().UnixNano()
		update := &Update{Path: path, Val: val}
		notifications[i] = &Notification{
			Timestamp: ts,
			Prefix:    prefix,
			Update:    []*Update{update},
		}	
	}
	return &GetResponse{Notification: notifications}, nil
}
func (c *Server) Set(ctx context.Context, in *SetRequest) (*SetResponse, error) {
	log.Printf("Receive message %s", in)
	return nil, status.Errorf(codes.Unimplemented, "method Set not implemented")
}
func (c *Server) Subscribe(GNMI_SubscribeServer) error {
	// log.Printf("Receive message %s", in)
	return status.Errorf(codes.Unimplemented, "method Subscribe not implemented")
}
func (c *Server) SubscribeWeb(in *SubscribeRequest, server GNMI_SubscribeWebServer) error {
	log.Printf("Receive message %s", in)
	return status.Errorf(codes.Unimplemented, "method SubscribeWeb not implemented")
}
