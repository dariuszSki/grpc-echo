package cmd

/*
Copyright © 2022 dariuszSki dsliwinski@aol.com

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

import (
	"context"
	"flag"
	"github.com/openziti/sdk-golang/ziti"
	"github.com/openziti/sdk-golang/ziti/config"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	pb "google.golang.org/grpc/examples/helloworld/helloworld"
	"log"
	"net"
	"time"
)

const (
	defaultClientEchoName = "world"
)

var (
	clientEchoString = flag.String("clientEchoString", defaultClientEchoName, "Name to greet")
	// clientCmd represents the client command
	clientCmd = &cobra.Command{
		Use:   "client",
		Short: "grpc-echo app: client mode",
		Long:  `This option enables the client mode of this app. It would be run on the client side, i.e. at the origination side`,
		Run: func(cmd *cobra.Command, args []string) {
			client()
		},
	}
)

func init() {
	rootCmd.AddCommand(clientCmd)
	clientCmd.Flags().StringVar(clientEchoString, "clientEchoString", defaultClientEchoName, "Name to greet")
	clientCmd.Flags().StringVar(sIdentity, "sIdentity", "", "Optional Ziti Server Identity if you require a specific destination")
}

func client() {

	cfg, err := config.NewFromFile(*cfgFile)
	if err != nil {
		log.Fatalf("failed to load ziti identity{%v}: %v", cfgFile, err)
	}

	ztx := ziti.NewContextWithConfig(cfg)
	err = ztx.Authenticate()
	if err != nil {
		log.Fatalf("failed to authenticate: %v", err)
	}

	if *sIdentity != "" {
		// set up ziti server identity to dial
		dialOptions := &ziti.DialOptions{
			Identity:       *sIdentity,
			ConnectTimeout: 1 * time.Minute,
		}
		// Set up a connection to the server.
		conn, err = grpc.Dial(*service,
			grpc.WithTransportCredentials(insecure.NewCredentials()),
			grpc.WithContextDialer(func(ctx context.Context, s string) (net.Conn, error) {
				return ztx.DialWithOptions(s, dialOptions)
			}),
		)
	} else {
		// Set up a connection to the server.
		conn, err = grpc.Dial(*service,
			grpc.WithTransportCredentials(insecure.NewCredentials()),
			grpc.WithContextDialer(func(ctx context.Context, s string) (net.Conn, error) {
				return ztx.Dial(s)
			}),
		)
	}

	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer func(conn *grpc.ClientConn) {
		err := conn.Close()
		if err != nil {

		}
	}(conn)
	c := pb.NewGreeterClient(conn)

	// Contact the server and print out its response.
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	r, err := c.SayHello(ctx, &pb.HelloRequest{Name: *clientEchoString})
	if err != nil {
		log.Fatalf("could not greet: %v", err)
	}
	log.Printf("Message: %s", r.GetMessage())
}
