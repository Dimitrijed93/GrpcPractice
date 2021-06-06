package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"math"
	"net"
	"strconv"
	"time"

	greetpb "github.com/dimitrijed93/demo"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/status"
)

type server struct{}

func (*server) Greet(ctx context.Context, gr *greetpb.GreetRequest) (*greetpb.GreetResponse, error) {
	firstName := gr.GetGreeting().GetFirstName()
	result := "Hello " + firstName

	res := greetpb.GreetResponse{
		Result: result,
	}

	return &res, nil

}

func (*server) LongGreet(stream greetpb.GreetService_LongGreetServer) error {
	result := ""
	for {
		req, err := stream.Recv()
		if err == io.EOF {
			return stream.SendAndClose(&greetpb.LongGreetResponse{
				Result: result,
			})
		}
		if err != nil {
			log.Fatalf("Error  while reading stream %v", err)
		}

		firstName := req.GetGreeting().GetFirstName()
		result += "hello " + firstName + "! "
	}
}

func (*server) GreetManyTimes(req *greetpb.GreetManyTimesRequest,
	stream greetpb.GreetService_GreetManyTimesServer) error {
	firstName := req.GetGreeting().GetFirstName()

	for i := 0; i < 10; i++ {
		result := "hello " + firstName + "number " + strconv.Itoa(i)
		res := &greetpb.GreetManyTimesResponse{
			Result: result,
		}
		stream.Send(res)
		time.Sleep(1000 * time.Millisecond)
	}
	return nil
}

func (*server) GreetEveryone(stream greetpb.GreetService_GreetEveryoneServer) error {

	for {
		req, err := stream.Recv()
		if err == io.EOF {
			return nil
		}

		if err != nil {
			log.Fatal("Error while reading stream")
			return err
		}

		firstName := req.GetGreeting().GetFirstName()

		result := "Hello " + firstName
		sendErr := stream.Send(&greetpb.GreetEveryoneResponse{
			Result: result,
		})

		if sendErr != nil {
			log.Fatal("Error sending response %v", sendErr)
			return err
		}
	}

}

func (*server) SquareRoot(ctx context.Context, req *greetpb.SquareRootRequest) (*greetpb.SquareRootResponse, error) {
	number := req.GetNumber()
	if number < 0 {
		return nil, status.Errorf(
			codes.InvalidArgument, fmt.Sprintf("Negative number: %v", number))
	}
	return &greetpb.SquareRootResponse{
		NumberRoot: math.Sqrt(float64(number)),
	}, nil
}

func main() {
	fmt.Printf("server")

	certFile := "ssl/server.crt"
	keyFile := "ssl/server.pem"

	creds, sslErr := credentials.NewServerTLSFromFile(certFile, keyFile)

	opt := grpc.Creds(creds)

	if sslErr != nil {
		log.Fatalf("Fail loading cert")
		return
	}

	lis, err := net.Listen("tcp", "0.0.0.0:50051")

	if err != nil {
		log.Fatalf("Failed to listen")
	}

	s := grpc.NewServer(opt)
	greetpb.RegisterGreetServiceServer(s, &server{})

	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve")
	}
}
