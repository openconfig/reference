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
//
// Cli connects to a provided target using a query and will either perform a
// Get or Subscribe (-use_subscribe) and return the results.  If -tls is not
// specified the connection will be unsecure.
// The basic per RPC authentication is provided by user/password authentication.
// If -outfile is not provided the notifications will be output to stdout.
//
// Examples:
// Get:
// ./cli -target=127.0.0.1:10162 -tls -user=test \
// -password=test -query=/bgp/neighbors -outfile=/tmp/foo -subscribe_once
//
// Subscribe:
// ./cli -target=127.0.0.1:10162 -tls -user=test \
// -password=test -query=/interfaces/interface[name=*]/state/counters -use_subscribe -outfile=/tmp/foo \
// -subscribe_once
//
package main

import (
	"crypto/tls"
	"flag"
	"os"
	"strings"
	"sync"

	"github.com/openconfig/reference/telemetry/auth/credential"
	"github.com/openconfig/reference/telemetry/collector/cli/query"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	log "github.com/golang/glog"
)

var (
	q   query.Query
	mu  sync.Mutex
	cfg = query.Config{
		Display: func(b []byte) {
			mu.Lock()
			defer mu.Unlock()
			os.Stdout.Write(append(b, '\n'))
		}}

	delimiter          = flag.String("delimiter", query.Delimiter, "Default seperator for query.")
	queryFlag          = flag.String("query", "", "List of comma seperated queries to make")
	targetFlag         = flag.String("target", "", "Target of the query (agent to connect to)")
	tlsFlag            = flag.Bool("tls", false, "Connection uses TLS if true, else plain TCP")
	caFile             = flag.String("ca_file", "", "The file containning the CA root cert file")
	user               = flag.String("user", "", "Username for the RPC auth")
	passwd             = flag.String("password", "", "Password for the RPC auth")
	serverHostOverride = flag.String("server_host_override", "", "The server name use to verify the hostname returned by TLS handshake")
	subscribe          = flag.Bool("use_subscribe", false, "Use subscription rather than get on request to target")
	subscribeOnce      = flag.Bool("subscribe_once", false, "Disconnect once subscription is synced.")
	outfile            = flag.String("outfile", "", "File to output received notifications")
)

// ParseQuery converts s to a list of Queries.
func ParseQuery(s, delimiter string) query.Query {
	queries := strings.Split(s, ",")
	q := query.Query{}
	for _, qItem := range queries {
		// Remove leading and trailing delimiters
		qItem := strings.Trim(qItem, delimiter)
		q.Queries = append(q.Queries, strings.Split(qItem, delimiter))
	}
	return q
}

func main() {
	flag.Parse()
	q := ParseQuery(*queryFlag, *delimiter)
	if len(q.Queries) == 0 {
		log.Fatal("--query must be set")
	}
	q.Target = *targetFlag
	if *tlsFlag {
		var sn string
		if *serverHostOverride != "" {
			sn = *serverHostOverride
		} else {
			sn = q.Target
		}
		var creds credentials.TransportCredentials
		if *caFile != "" {
			var err error
			creds, err = credentials.NewClientTLSFromFile(*caFile, sn)
			if err != nil {
				log.Fatalf("Failed to create TLS credentials %v\n", err)
			}
		} else {
			creds = credentials.NewTLS(&tls.Config{
				ServerName:         sn,
				InsecureSkipVerify: true,
			})
		}
		q.DialOptions = append(q.DialOptions, grpc.WithTransportCredentials(creds))
	} else {
		q.DialOptions = append(q.DialOptions, grpc.WithInsecure())
	}
	if *user != "" {
		pc := credential.NewPassCred(*user, *passwd, true)
		q.DialOptions = append(q.DialOptions, grpc.WithPerRPCCredentials(pc))
	}
	if *outfile != "" {
		f, err := os.Create(*outfile)
		var fMu sync.Mutex
		if err != nil {
			log.Fatalf("Failed to open file %s: %s", *outfile, err)
		}
		cfg.Display = func(b []byte) {
			fMu.Lock()
			defer fMu.Unlock()
			n := make([]byte, len(b)+1)
			n[copy(n, b)] = byte('\n')
			f.Write(n)
		}
	}
	switch {
	case q.Update != nil:
		if err := query.Update(context.Background(), q, &cfg); err != nil {
			log.Infof("query.Update failed for query %v %v: %s\n", cfg, q, err)
		}
	case len(q.Queries) > 0 && *subscribe:
		cfg.Once = *subscribeOnce
		if err := query.DisplayStream(context.Background(), q, &cfg); err != nil {
			log.Infof("query.DisplayStream failed for query %v %v: %s\n", cfg, q, err)
		}
	default:
		if err := query.Display(context.Background(), q, &cfg); err != nil {
			log.Infof("query.Display failed for query %v %v: %s\n", cfg, q, err)
		}
	}
}
