.PHONY: blog

grpc:
	 protoc  greet/greetpb/greet.proto --go_out=plugins=grpc:.
blog:
	protoc  blog/blogpb/blog.proto --go_out=plugins=grpc:.

all: grpc blog

server:
	go run blog/blog_server/server.go

client:
	go run blog/blog_client/client.go