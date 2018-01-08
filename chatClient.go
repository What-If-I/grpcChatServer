package main

import (
	"log"
	"os"

	pb "./protobufs"
	"bufio"
	"fmt"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"io"
	"strings"
	"time"
)

const (
	address = "localhost:50051"
)

func input(prompt string) string {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print(prompt)
	text, _ := reader.ReadString('\n')
	return strings.Replace(text, "\n", "", -1)
}

func main() {
	// Set up a connection to the server.
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	client := pb.NewChatClient(conn)

	name := input("Please enter your name: ")
	user := pb.User{Name: name}

	go func() {
		ctx, cancel := context.WithCancel(context.Background())
		stream, err := client.Subscribe(ctx, &user)
		if err != nil {
			log.Fatalf("%v.RouteChat(_) = _, %v", client, err)
		}
		for {
			in, err := stream.Recv()
			if err == io.EOF {
				time.Sleep(1 * time.Second)
				//stream.CloseSend()
				//break
			} else if err != nil {
				log.Fatalf("Failed to receive a message : %v", err)
			} else {
				log.Println(in.Message)
			}
		}
		defer cancel()
	}()

	for {
		message := input("")
		_, err := client.SendMessage(context.Background(), &pb.Message{Text: message, User: &user})
		if err != nil {
			log.Fatalf("Error: %v", err)
		}
	}
}
