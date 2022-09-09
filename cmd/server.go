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
	"encoding/json"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/openziti/sdk-golang/ziti"
	"github.com/openziti/sdk-golang/ziti/config"
	"github.com/openziti/sdk-golang/ziti/edge"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	pb "google.golang.org/grpc/examples/helloworld/helloworld"
)

// service type
type HostedService struct {
	Name              string
	AddressByIdentity bool
}

// serverCmd represents the server command
var (
	//addressByIdentity bool
	hostedServices   []string
	hostedService    HostedService
	serverEchoString string
	serverCmd        = &cobra.Command{
		Use:   "server",
		Short: "grpc-echo app: server mode",
		Long:  `This option enables the server mode of this app. It would be run on the server side, i.e. at the destination side`,
		Run: func(cmd *cobra.Command, args []string) {
			server()
		},
	}
)

const (
	defaultServerEchoName = "world"
)

func init() {
	rootCmd.AddCommand(serverCmd)
	serverCmd.Flags().StringVar(&serverEchoString, "serverEchoString", defaultServerEchoName, "Name to greet")
	serverCmd.Flags().StringArrayVar(&hostedServices, "hostedServices", []string{}, "service json config to be hosted by this identity")
	//serverCmd.Flags().BoolVar(&addressByIdentity, "addressByIdentity", false, "Enable addressable identity")
}

// server is used to implement helloworld.GreeterServer.
type grpcServer struct {
	pb.UnimplementedGreeterServer
	serverIdentity *edge.CurrentIdentity
	serviceName    string
}

// SayHello implements helloworld.GreeterServer
func (s *grpcServer) SayHello(_ context.Context, in *pb.HelloRequest) (*pb.HelloReply, error) {
	log.Printf("Received: %v on service name: %v", in.GetName(), s.serviceName)
	log.Printf("server identity is %s", s.serverIdentity.Name)
	if in.GetName() == "ping" {
		return &pb.HelloReply{Message: s.serviceName + " " + s.serverIdentity.Name + "-->pong"}, nil
	}
	return &pb.HelloReply{Message: s.serviceName + " " + s.serverIdentity.Name + " " + in.GetName()}, nil
}

// start the ziti listener
func serve(listener edge.Listener, identity *edge.CurrentIdentity, serviceName string) {
	s := grpc.NewServer()
	pb.RegisterGreeterServer(s, &grpcServer{serverIdentity: identity, serviceName: serviceName})
	log.Printf("server listening for addressable service: %v", listener.Addr())

	if err := s.Serve(listener); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

// main function
func server() {

	cfg, err := config.NewFromFile(*cfgFile)
	if err != nil {
		log.Fatalf("failed to load ziti identity{%v}: %v", cfgFile, err)
	}

	ztx := ziti.NewContextWithConfig(cfg)
	err = ztx.Authenticate()
	if err != nil {
		log.Fatalf("failed to authenticate: %v", err)
	}

	/* Get Ziti Identity of the Server hosted the service */
	serverIdentity, err := ztx.GetCurrentIdentity()
	if err != nil {
		return
	}

	var lis edge.Listener

	for _, hs := range hostedServices {

		/* convert service string configuration to jason format */
		json.Unmarshal([]byte(hs), &hostedService)
		log.Print(hostedService)

		/* If addressable terminator service is requested, then enable it */
		if hostedService.AddressByIdentity {

			// set up ziti server identity to dial
			listenOptions := ziti.ListenOptions{
				ConnectTimeout:        10 * time.Second,
				MaxConnections:        10,
				BindUsingEdgeIdentity: hostedService.AddressByIdentity,
			}

			lis, err = ztx.ListenWithOptions(hostedService.Name, &listenOptions)
			if err != nil {
				log.Fatalf("failed to listen: %v", err)
			}

		} else {
			lis, err = ztx.Listen(hostedService.Name)
			if err != nil {
				log.Fatalf("failed to listen: %v", err)
			}

		}

		go serve(lis, serverIdentity, hostedService.Name)
	}

	/* start the main loop and watch for os signal interrupts to clean up resources and exit gracefully */
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGTERM, syscall.SIGINT)

	for sig := range sigCh {

		if sig != syscall.SIGHUP {
			os.Exit(1)
		}

	}
}
