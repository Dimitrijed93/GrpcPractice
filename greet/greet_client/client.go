package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"time"

	greetpb "github.com/dimitrijed93/demo"
	"google.golang.org/grpc"
)

func main() {
	fmt.Printf("client")

	conn, err := grpc.Dial("localhost:50051", grpc.WithInsecure())

	if err != nil {
		log.Fatalf("could not connect")
	}

	defer conn.Close()

	c := greetpb.NewGreetServiceClient(conn)

	// doServerStreaming(c)
	// doClientStreaming(c)
	doBiDiStreaming(c)

}

func doBiDiStreaming(c greetpb.GreetServiceClient) {

	req := []*greetpb.GreetEveryoneRequest{
		{
			Greeting: &greetpb.Greeting{
				FirstName: "Dimitrije",
				LastName:  "Dragicevic",
			},
		},
		{
			Greeting: &greetpb.Greeting{
				FirstName: "Mirko",
				LastName:  "Mirkovic",
			},
		}, {
			Greeting: &greetpb.Greeting{
				FirstName: "Marko",
				LastName:  "Markovic",
			},
		},
	}

	stream, err := c.GreetEveryone(context.Background())

	if err != nil {
		log.Fatal("Error creating stream")
	}

	wait := make(chan struct{})

	go func() {
		for _, r := range req {
			fmt.Printf("Sending req %v", r)
			stream.Send(r)
			time.Sleep(1000 * time.Millisecond)
		}
		stream.CloseSend()
	}()

	go func() {

		for {
			res, err := stream.Recv()
			if err == io.EOF {
				break
			}
			if err != nil {
				log.Fatalf("Error while rec mess")
				close(wait)
			}
			fmt.Printf("Received %v", res.GetResult())
		}
		close(wait)

	}()

	<-wait
}

func doClientStreaming(c greetpb.GreetServiceClient) {
	req := []*greetpb.LongGreetRequest{
		{
			Greeting: &greetpb.Greeting{
				FirstName: "Dimitrije",
				LastName:  "Dragicevic",
			},
		},
		{
			Greeting: &greetpb.Greeting{
				FirstName: "Mirko",
				LastName:  "Mirkovic",
			},
		}, {
			Greeting: &greetpb.Greeting{
				FirstName: "Marko",
				LastName:  "Markovic",
			},
		},
	}
	stream, err := c.LongGreet(context.Background())
	if err != nil {
		log.Fatalf("Unable to read stream")
	}

	for _, item := range req {
		stream.Send(item)
		time.Sleep(100 * time.Millisecond)

	}

	res, err := stream.CloseAndRecv()
	if err != nil {
		log.Fatal("Error resposne")
	}

	log.Printf("Response %v", res)

}

func doServerStreaming(c greetpb.GreetServiceClient) {
	req := &greetpb.GreetManyTimesRequest{
		Greeting: &greetpb.Greeting{
			FirstName: "Dimitrije",
			LastName:  "Dragicevic",
		},
	}
	res, err := c.GreetManyTimes(context.Background(), req)
	if err != nil {
		log.Fatalf("ERORR STREAM")
	}

	for {
		msg, err := res.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalf("error while reading stream")
		}

		log.Printf("Result %v", msg.GetResult())

	}
}
