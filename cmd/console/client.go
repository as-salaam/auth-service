package main

import (
	"context"
	"encoding/json"
	"flag"
	"log"
	"time"

	pb "github.com/as-salaam/auth-service/internal/auth"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/examples/data"
)

var (
	tls                = flag.Bool("tls", false, "Connection uses TLS if true, else plain TCP")
	caFile             = flag.String("ca_file", "", "The file containing the CA root cert file")
	serverAddr         = flag.String("addr", "localhost:4001", "The server address in the format of host:port")
	serverHostOverride = flag.String("server_host_override", "x.test.example.com", "The server name used to verify the hostname returned by the TLS handshake")
)

func runAuthenticateRoute(client pb.AuthClient) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	in := &pb.AuthenticateRequest{Token: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJVc2VySUQiOiJhNDhkZDRlNi0zYTNiLTRhMjgtYTYwMS02ODVlOTcxNmRhYmYiLCJleHAiOjE2ODM3ODk4OTZ9.k5PQmLghLDR-n03pozhaMWElybGHL1KGqKRl8BbXapk"}
	authResponse, err := client.Authenticate(ctx, in)
	if err != nil {
		log.Fatalf("client.Authenticate failed: %v", err)
	}

	a, _ := json.MarshalIndent(authResponse, "", "\t")
	log.Println(authResponse.Authenticated, authResponse.User)
	log.Println(string(a))
}

func main() {
	flag.Parse()
	var opts []grpc.DialOption
	if *tls {
		if *caFile == "" {
			*caFile = data.Path("x509/ca_cert.pem")
		}
		creds, err := credentials.NewClientTLSFromFile(*caFile, *serverHostOverride)
		if err != nil {
			log.Fatalf("Failed to create TLS credentials: %v", err)
		}
		opts = append(opts, grpc.WithTransportCredentials(creds))
	} else {
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}

	conn, err := grpc.Dial(*serverAddr, opts...)
	if err != nil {
		log.Fatalf("fail to dial: %v", err)
	}
	defer conn.Close()
	client := pb.NewAuthClient(conn)

	runAuthenticateRoute(client)
}
