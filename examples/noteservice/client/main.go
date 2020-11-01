package main

import (
	"context"
	"log"

	pb "github.com/shayanh/grpc-go-contracts/examples/noteservice/proto"
	"google.golang.org/grpc"
)

const (
	noteStoreAddress = "localhost:8000"
	authAddress      = "localhost:8001"
)

func main() {
	token := "some-token-0"

	func() {
		conn, err := grpc.Dial(noteStoreAddress, grpc.WithInsecure())
		if err != nil {
			log.Fatal(err)
		}
		defer conn.Close()

		c := pb.NewNoteStoreClient(conn)
		resp, err := c.GetNote(context.TODO(), &pb.GetNoteRequest{NoteId: 1, Token: token})
		if err != nil {
			log.Fatal(err)
		}
		log.Print(resp.GetText())
	}()

	func() {
		conn, err := grpc.Dial(authAddress, grpc.WithInsecure())
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
