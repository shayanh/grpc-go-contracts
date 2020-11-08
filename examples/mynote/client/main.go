package main

import (
	"context"
	"log"

	pb "github.com/shayanh/grpc-go-contracts/examples/mynote/proto"
	"google.golang.org/grpc"
)

const (
	noteServiceAddr = "localhost:8000"
	authServiceAddr = "localhost:8001"
)

func main() {
	token := "some-token-0"

	func() {
		conn, err := grpc.Dial(noteServiceAddr, grpc.WithInsecure())
		if err != nil {
			log.Fatal(err)
		}
		defer conn.Close()

		c := pb.NewNoteServiceClient(conn)
		resp, err := c.GetNote(context.TODO(), &pb.GetNoteRequest{NoteId: 1, Token: token})
		if err != nil {
			log.Fatal(err)
		}
		log.Print(resp.GetText())
	}()

	func() {
		conn, err := grpc.Dial(authServiceAddr, grpc.WithInsecure())
		if err != nil {
			log.Fatal(err)
		}
		defer conn.Close()

		c := pb.NewAuthServiceClient(conn)
		resp, err := c.Authenticate(context.TODO(), &pb.AuthenticateRequest{Token: token})
		if err != nil {
			log.Fatal(err)
		}
		log.Print(resp.GetUserId())
	}()
}
