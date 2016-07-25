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
	"net"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/golang/protobuf/proto"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

	log "github.com/golang/glog"
	ocpb "github.com/openconfig/reference/rpc/openconfig"
)

type generator interface {
	Path() ocpb.Path
	Next() (*ocpb.Notification, *Event)
	Last() *ocpb.Notification
}

type FixedGenerator struct {
	mu sync.Mutex
	v  ocpb.Value
	p  ocpb.Path
	l  *ocpb.Notification
}

func NewFixedGenerator(p ocpb.Path, v ocpb.Value) *FixedGenerator {
	return &FixedGenerator{
		v: v,
		p: p,
	}
}

func (g *FixedGenerator) Path() ocpb.Path {
	return g.p
}

func (g *FixedGenerator) Next() (*ocpb.Notification, *Event) {
	g.mu.Lock()
	defer g.mu.Unlock()
	n := &ocpb.Notification{
		Timestamp: time.Now().UnixNano(),
		Update: []*ocpb.Update{{
			Path:  &g.p,
			Value: &g.v,
		}},
	}
	g.l = n
	return n, nil
}

func (g *FixedGenerator) Last() *ocpb.Notification {
	g.mu.Lock()
	defer g.mu.Unlock()
	return g.l
}

type fakeTarget struct {
	mu     sync.Mutex
	mapper map[string]generator
	t      *testing.T
}

func key(p []string) string {
	if len(p) == 0 {
		return ""
	}
	return strings.Join(p, "/")
}

func hash(prefix, path []string) string {
	h := append(append([]string{}, prefix...), path...)
	return strings.Join(h, "/")
}

func newFakeGRPCTarget() *fakeTarget {
	return &fakeTarget{
		mapper: map[string]generator{},
	}
}

type Event struct{}

func (ft *fakeTarget) next(k string) (*ocpb.Notification, *Event, error) {
	v, ok := ft.mapper[k]
	if !ok {
		return nil, nil, fmt.Errorf("key not found: %s", k)
	}
	if v == nil {
		return nil, nil, fmt.Errorf("func not defined for key: %s", k)
	}
	n, event := v.Next()
	return n, event, nil
}

func (ft *fakeTarget) Subscribe(stream ocpb.OpenConfig_SubscribeServer) error {
	return nil
}

func (ft *fakeTarget) Get(ctx context.Context, in *ocpb.GetRequest) (*ocpb.GetResponse, error) {
	var prefix string
	if in.GetPrefix() == nil {
		prefix = ""
	} else {
		prefix = key(in.GetPrefix().Element)
	}
	r := &ocpb.GetResponse{}
	for _, p := range in.GetPath() {
		path := key(p.Element)
		if prefix != "" {
			path = fmt.Sprintf("%s/%s", prefix, path)
		}
		n, event, err := ft.next(path)
		if err != nil {
			log.Error(err)
			panic("internal fake server error")
		}
		if event != nil {
			return nil, grpc.Errorf(codes.Internal, "internal event")
		}
		r.Notification = append(r.Notification, n)
	}
	return r, nil
}

func (ft *fakeTarget) RegisterGenerator(g generator) error {
	ft.mu.Lock()
	defer ft.mu.Unlock()
	p := g.Path()
	ft.mapper[key(p.Element)] = g
	return nil
}

func (ft *fakeTarget) GetGenerator(path []string) generator {
	ft.mu.Lock()
	defer ft.mu.Unlock()
	if g, ok := ft.mapper[key(path)]; ok {
		return g
	}
	return nil
}

func (ft *fakeTarget) GetModels(ctx context.Context, in *ocpb.GetModelsRequest) (*ocpb.GetModelsResponse, error) {
	return nil, nil
}

func (ft *fakeTarget) Set(ctx context.Context, in *ocpb.SetRequest) (*ocpb.SetResponse, error) {
	return nil, nil
}

func startLocalServer() (*grpc.Server, *fakeTarget, string, error) {
	lis, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return nil, nil, "", fmt.Errorf("create GRPC server listener: %v", err)
	}
	sAddr := lis.Addr().String()
	ft := newFakeGRPCTarget()
	s := grpc.NewServer()
	ocpb.RegisterOpenConfigServer(s, ft)
	go s.Serve(lis)
	return s, ft, sAddr, nil
}

func fakeDisplayer(ft *fakeTarget, t *testing.T) func([]byte) {
	return func(b []byte) {
		var p ocpb.Notification
		err := proto.UnmarshalText(string(b), &p)
		if err != nil {
			t.Errorf("failed to unmarshal %s: %s", b, err)
			return
		}
		var prefix []string
		if p.GetPrefix() != nil {
			prefix = append(prefix, p.GetPrefix().Element...)
		}
		for _, u := range p.GetUpdate() {
			path := make([]string, len(prefix)+len(u.GetPath().Element))
			copy(path, prefix)
			copy(path[copy(path, prefix):], u.GetPath().Element)
			lastP := ft.GetGenerator(path).Last()
			if !proto.Equal(&p, lastP) {
				t.Errorf("proto.Equal(%s, %s) failed", p.String(), lastP.String())
			}
		}
	}
}

func TestDisplay(t *testing.T) {
	var tCtx context.Context
	tests := []struct {
		ctx       context.Context
		query     Query
		cfg       *Config
		generator generator
		err       error
	}{{
		ctx: tCtx,
		query: Query{
			Target: "test",
		},
		cfg: &Config{},
		err: fmt.Errorf("grpc: no transport security set (use grpc.WithInsecure() explicitly or set credentials)"),
	}, {
		ctx: tCtx,
		query: Query{
			Target: "test",
		},
		cfg: nil,
		err: fmt.Errorf("cfg must not be nil"),
	}, {
		ctx: tCtx,
		query: Query{
			Target: "localhost",
			DialOptions: []grpc.DialOption{
				grpc.WithInsecure(),
			},
		},
		cfg: &Config{},
		err: fmt.Errorf("query target must be specified"),
	}, {
		ctx: tCtx,
		query: Query{
			Target: "localhost",
			DialOptions: []grpc.DialOption{
				grpc.WithInsecure(),
			},
			Queries: [][]string{
				[]string{"foo"},
			},
		},
		cfg: &Config{},
		generator: NewFixedGenerator(ocpb.Path{
			Element: []string{"foo"},
		}, ocpb.Value{
			Value: []byte("42"),
		}),
	}}

	for _, tt := range tests {
		var s *grpc.Server
		var ft *fakeTarget
		if tt.query.Target == "localhost" {
			var err error
			s, ft, tt.query.Target, err = startLocalServer()
			if tt.generator != nil {
				ft.RegisterGenerator(tt.generator)
			}
			if err != nil {
				t.Fatal("failed to start server")
			}
			defer s.Stop()
		}
		t.Logf("test Display(%+v, %+v, %+v)", tt.ctx, tt.query, tt.cfg)
		if tt.cfg != nil {
			tt.cfg.Display = fakeDisplayer(ft, t)
		}
		err := Display(context.Background(), tt.query, tt.cfg)
		if tt.err != nil && err != nil && tt.err.Error() != err.Error() {
			t.Errorf("failed Display(%+v, %+v, %+v): got %s, want %s", tt.ctx, tt.query, tt.cfg, err, tt.err)
		}
	}
}
