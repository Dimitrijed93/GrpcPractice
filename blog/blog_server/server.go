package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"

	greetpb "github.com/dimitrijed93/demo"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type server struct{}

type blogItem struct {
	ID       primitive.ObjectID `bson:"id,omitempty"`
	AuthorID string             `bson:"author_id"`
	Title    string             `bson:"title"`
	Content  string             `bson:"content"`
}

var collection *mongo.Collection

func (*server) CreateBlog(ctx context.Context, req *greetpb.CreateBlogRequest) (*greetpb.CreateBlogResponse, error) {

	data := blogItem{
		AuthorID: req.Blog.Id,
		Title:    req.Blog.Title,
		Content:  req.Blog.Content,
	}

	res, err := collection.InsertOne(context.Background(), data)
	if err != nil {
		return nil, status.Errorf(
			codes.Internal,
			fmt.Sprintf("Internal error %v", err),
		)
	}

	oid, ok := res.InsertedID.(primitive.ObjectID)
	if !ok {
		return nil, status.Errorf(
			codes.Internal,
			fmt.Sprintf("Failed to convert  %v", err),
		)
	}

	return &greetpb.CreateBlogResponse{
		Blog: &greetpb.Blog{
			Id:       oid.Hex(),
			AuthorId: data.AuthorID,
			Title:    data.Title,
			Content:  data.Content,
		},
	}, nil

}

func (*server) ReadBlog(ctx context.Context, req *greetpb.ReadBlogRequest) (*greetpb.ReadBlogResponse, error) {
	blogId := req.BlogId

	oid, err := primitive.ObjectIDFromHex(blogId)
	if err != nil {
		return nil, status.Errorf(
			codes.InvalidArgument,
			fmt.Sprint("bad id"),
		)
	}

	data := &blogItem{}

	filter := bson.M{"_id": oid}

	res := collection.FindOne(context.Background(), filter)
	if decodeError := res.Decode(data); decodeError != nil {
		return nil, status.Errorf(
			codes.NotFound,
			fmt.Sprintf("not found"),
		)
	}

	return &greetpb.ReadBlogResponse{
		Blog: &greetpb.Blog{
			Id:       data.ID.Hex(),
			AuthorId: data.AuthorID,
			Content:  data.Content,
		},
	}, nil

}

func dataToBlogPb(data *blogItem) *greetpb.Blog {
	return dataToBlogPb(data)
}

func (*server) DeleteBlog(ctx context.Context, req *greetpb.DeleteBlogRequest) (*greetpb.DeleteBlogResponse, error) {
	blogId := req.BlogId

	oid, err := primitive.ObjectIDFromHex(blogId)
	if err != nil {
		return nil, status.Errorf(
			codes.InvalidArgument,
			fmt.Sprint("bad id"),
		)
	}
	filter := bson.M{"_id": oid}

	res, err1 := collection.DeleteOne(context.Background(), filter)

	if err1 != nil {
		return nil, status.Errorf(
			codes.Internal,
			fmt.Sprint("Cannot delete"),
		)
	}

	if res.DeletedCount == 0 {
		return nil, status.Errorf(
			codes.Internal,
			fmt.Sprint("Cannot delete"),
		)
	}

	return &greetpb.DeleteBlogResponse{
		BlogId: oid.Hex(),
	}, nil
}

func (*server) UpdateBlog(ctx context.Context, req *greetpb.UpdateBlogRequest) (*greetpb.UpdateBlogResponse, error) {

	blog := req.Blog

	oid, err := primitive.ObjectIDFromHex(blog.Id)
	if err != nil {
		return nil, status.Errorf(
			codes.InvalidArgument,
			fmt.Sprint("bad id"),
		)
	}

	data := &blogItem{}

	filter := bson.M{"_id": oid}

	res := collection.FindOne(context.Background(), filter)
	if decodeError := res.Decode(data); decodeError != nil {
		return nil, status.Errorf(
			codes.NotFound,
			fmt.Sprintf("not found"),
		)
	}

	data.AuthorID = blog.AuthorId
	data.Content = blog.Content
	data.Title = blog.Title

	_, er2r := collection.ReplaceOne(context.Background(), filter, data)

	if er2r != nil {
		log.Fatalf("could not update")
	}

	return &greetpb.UpdateBlogResponse{
		Blog: dataToBlogPb(data),
	}, nil

}

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	fmt.Printf("Blog Server")

	client, mongoErr := mongo.NewClient(options.Client().ApplyURI("mongodb://localhost:27017"))
	if mongoErr != nil {
		log.Fatalf("Failed to create client to MongoDB")
	}

	mongoErr = client.Connect(context.TODO())

	if mongoErr != nil {
		log.Fatal("Failed to connect to MongoDB")
	}

	collection = client.Database("blogs").Collection("blog")

	lis, err := net.Listen("tcp", "0.0.0.0:50051")

	if err != nil {
		log.Fatalf("Failed to listen")
	}

	s := grpc.NewServer()
	greetpb.RegisterBlogServiceServer(s, &server{})

	go func() {
		fmt.Println("Starting server...")
		if err := s.Serve(lis); err != nil {
			log.Fatalf("failed to serve")
		}
	}()

	ch := make(chan os.Signal, 1)

	signal.Notify(ch, os.Interrupt)

	<-ch
	fmt.Println("Stopping the server")
	s.Stop()
	fmt.Println("Stoping the listener")
	lis.Close()
	client.Disconnect(context.TODO())
	fmt.Println("End of program")

}
