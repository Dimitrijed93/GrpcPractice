package main

import (
	"context"
	"fmt"
	"log"

	greetpb "github.com/dimitrijed93/demo"
	"google.golang.org/grpc"
)

func main() {

	opts := grpc.WithInsecure()

	client, err := grpc.Dial("localhost:50051", opts)
	must(err)
	defer client.Close()

	c := greetpb.NewBlogServiceClient(client)

	res, err := c.CreateBlog(context.Background(), &greetpb.CreateBlogRequest{
		Blog: &greetpb.Blog{
			AuthorId: "Dimitrije",
			Title:    "Demo",
			Content:  "Demo request",
		},
	})
	must(err)
	fmt.Printf("Created Blog %v", res.Blog)

	// _, err2 := c.ReadBlog(context.Background(), &greetpb.ReadBlogRequest{
	// 	BlogId: "1231312313",
	// })
	// must(err2)

	b, err3 := c.ReadBlog(context.Background(), &greetpb.ReadBlogRequest{
		BlogId: res.Blog.Id,
	})
	must(err3)

	fmt.Printf("Read: %v", b)

}

func must(err error) {
	if err != nil {
		log.Fatalf("Error: %v", err)
	}
}
