// Copyright 2016 Google Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
////////////////////////////////////////////////////////////////////////////////

package query

import (
	"fmt"
	"os"
	"time"

	"github.com/golang/protobuf/proto"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

	log "github.com/golang/glog"
	ocpb "github.com/openconfig/reference/rpc/openconfig"
)

const (
	// TimeFormat is the default time format for the query.
	TimeFormat = "2000-01-01-01:01:01.00000000"
	// Delimiter is the default delimiter for the query.
	Delimiter = "/"
	// DialTimeout is the default dial timeout for the query.
	DialTimeout = 10 * time.Second
)

var (
	// defaultDisplay is the default implementation for displaying output from the
	// query.
	defaultDisplay = func(b []byte) {
		os.Stdout.Write(b)
	}
)

// Query contains the target and query for a request.
type Query struct {
	Address     string
	Target      string
	DialOptions []grpc.DialOption
	// Queries is a list queries made of query elements.
	Queries [][]string
	// Update is a single SetRequest to be made on the target.
	Update *ocpb.SetRequest
}

// Config contains the configuration for displaying a query.
type Config struct {
	Count       uint64
	Once        bool
	Delimiter   string
	DialTimeout time.Duration
	Display     func([]byte)
	TimeFormat  string
}

// Display creates a gRPC connection to the query target and makes a Get call
// for the queried paths and displays the response via cfg.Display.
func Display(ctx context.Context, query Query, cfg *Config) error {
	c, err := createClient(ctx, query, cfg)
	if err != nil {
		return err
	}
	request, err := createRequest(query)
	if err != nil {
		return err
	}
	log.Infof("Get with:\n%s\n", request)
	resp, err := c.Get(ctx, request)
	if err != nil {
		return err
	}
	for _, n := range resp.GetNotification() {
		cfg.Display([]byte(n.String()))
	}
	return nil
}

// DisplayStream creates a gRPC connection to the query target and makes a
// Subscribe call for the queried paths and streams the response via
// cfg.Display.
func DisplayStream(ctx context.Context, query Query, cfg *Config) error {
	c, err := createClient(ctx, query, cfg)
	if err != nil {
		return err
	}
	request, err := createSubscribeRequest(query)
	if err != nil {
		return err
	}
	stream, err := c.Subscribe(ctx)
	if err != nil {
		return err
	}
	if err := stream.Send(request); err != nil {
		return err
	}
	log.Infof("Subscribed with:\n%s", proto.MarshalTextString(request))
	for {
		resp, err := stream.Recv()
		log.V(2).Info(proto.MarshalTextString(resp))
		if err != nil {
			// TODO(hines): This should be io.EOF but for some reason the server
			// currently sends this code.
			if grpc.Code(err) != codes.OutOfRange {
				return err
			}
			return nil
		}
		switch resp.Response.(type) {
		default:
			log.Infof("Unknown response:\n%s\n", resp.String())
		case *ocpb.SubscribeResponse_Heartbeat:
			log.Infof("Heartbeat:%s\n", resp.String())
		case *ocpb.SubscribeResponse_Update:
			cfg.Display([]byte(proto.MarshalTextString(resp)))
		case *ocpb.SubscribeResponse_SyncResponse:
			log.Infof("Sync Response: %s", resp.String())
			if cfg.Once {
				stream.CloseSend()
				return nil
			}
		}
	}
}

// Update sends a SetRequest to the target.  If the Set fails an error will be
// returned.  The response will be displayed by the configure cfg.Display.
func Update(ctx context.Context, query Query, cfg *Config) error {
	c, err := createClient(ctx, query, cfg)
	_ = c
	if err != nil {
		return err
	}
	if query.Update == nil {
		return fmt.Errorf("query.Updates must be defined for Update")
	}
	resp, err := c.Set(ctx, query.Update)
	if err != nil {
		return fmt.Errorf("failed to set %s: %s", proto.MarshalTextString(query.Update), err)
	}
	cfg.Display([]byte(resp.String()))
	return nil
}

func createClient(ctx context.Context, query Query, cfg *Config) (ocpb.OpenConfigClient, error) {
	switch {
	case ctx == nil:
		return nil, fmt.Errorf("ctx must not be nil")
	case cfg == nil:
		return nil, fmt.Errorf("cfg must not be nil")
	case query.Target == "":
		return nil, fmt.Errorf("query target must be specified")
	}
	if cfg.Display == nil {
		cfg.Display = defaultDisplay
	}
	log.Infof("Creating connection: %+v", query)
	conn, err := grpc.Dial(query.Target, query.DialOptions...)
	if err != nil {
		return nil, err
	}
	return ocpb.NewOpenConfigClient(conn), nil
}

func createRequest(q Query) (*ocpb.GetRequest, error) {
	r := &ocpb.GetRequest{}
	for _, qItem := range q.Queries {
		p := &ocpb.Path{
			Element: qItem,
		}
		r.Path = append(r.Path, p)
	}
	return r, nil
}

func createSubscribeRequest(q Query) (*ocpb.SubscribeRequest, error) {
	// TODO(hines): Re-add once bug is resolved for lack of mode support.
	//subList := &ocpb.SubscriptionList{
	//	 Mode: &ocpb.SubscriptionList_Once{
	//	 Once: true,
	//	 },
	// }
	subList := &ocpb.SubscriptionList{}
	for _, qItem := range q.Queries {
		subList.Subscription = append(subList.Subscription, &ocpb.Subscription{
			Path: &ocpb.Path{
				Element: qItem,
			},
		})
	}
	return &ocpb.SubscribeRequest{
		Request: &ocpb.SubscribeRequest_Subscribe{
			Subscribe: subList,
		}}, nil
}
