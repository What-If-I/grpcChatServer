//go:generate protoc -I ./rpc --go_out=plugins=grpc:./rpc ./rpc/chat.proto

package main

import (
	"log"
	"net"

	pb "./protobufs"
	"fmt"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"os"
)

const defaultPort = ":50051"

func getServerPort() string {
	port := os.Getenv("SERVER_PORT")
	if port == "" {
		return defaultPort
	}
	return port
}

type chat struct{}

var subscribers = map[pb.Chat_SubscribeServer]*pb.User{}

func (s *chat) Subscribe(subscriber *pb.User, stream pb.Chat_SubscribeServer) error {
	stream.Send(&pb.Reply{fmt.Sprintf("Greetings %s! Feel free to chat here.", subscriber.Name)})
	notifyAllSubs(pb.Reply{fmt.Sprintf("%s joined the chat!", subscriber.Name)})

	subscribers[stream] = subscriber
	<-stream.Context().Done() // Waiting channel to be closed by client
	delete(subscribers, stream)
	notifyAllSubs(pb.Reply{fmt.Sprintf("%s left the chat.", subscriber.Name)})

	log.Printf("Closing %s's channel.", subscriber.Name)
	return nil
}

func (s *chat) SendMessage(ctx context.Context, message *pb.Message) (*pb.Reply, error) {
	log.Printf("Got message from %s: %s", message.User.Name, message.Text)

	msg := fmt.Sprintf("%s: %s", message.User.Name, message.Text)
	notifyAllSubs(pb.Reply{Message: msg})

	msg = fmt.Sprintf("Received message from %s \"%s\"", message.User.Name, message.Text)
	return &pb.Reply{Message: msg}, nil
}

func notifyAllSubs(msg pb.Reply) {
	for stream, user := range subscribers {
		err := stream.Send(&msg)
		if err != nil {
			log.Printf("Failed send message to %s. Reason: %s", user.Name, err)
		}
	}
}

func main() {
	port := getServerPort()
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	pb.RegisterChatServer(s, &chat{})
	reflection.Register(s)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
