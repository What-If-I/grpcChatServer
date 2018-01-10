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

	go listenToChat(client, &user)

	for {
		message := input("")
		_, err := client.SendMessage(context.Background(), &pb.Message{Text: message, User: &user})
		if err != nil {
			log.Fatalf("Error: %v", err)
		}
	}
}

// Ask for user input
func input(prompt string) string {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print(prompt)
	text, _ := reader.ReadString('\n')
	return strings.Replace(text, "\n", "", -1)
}

func listenToChat(client pb.ChatClient, user *pb.User) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	stream, err := client.Subscribe(ctx, user)
	if err != nil {
		log.Fatalf("%v.Subscribe(_) = _, %v", client, err)
		panic(err)
	}
	for {
		in, err := stream.Recv()
		if err == io.EOF {
			// continue to wait for other messages
			time.Sleep(1 * time.Second)
		} else if err != nil {
			log.Fatalf("Failed to receive a message : %v", err)
			panic(err)
		} else {
			log.Println(in.Message)
		}
	}
}
