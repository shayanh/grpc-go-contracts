package main

import (
	"context"
	"log"
	"net"
	"sync"

	pb "github.com/shayanh/grpc-go-contracts/examples/noteservice/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	port = ":8000"
)

type noteStoreServer struct {
	pb.UnimplementedNoteStoreServer

	authServerAddress string

	mutex sync.Mutex
	notes []*pb.Note
}

var noteStore noteStoreServer

func init() {
	noteStore.mutex.Lock()
	defer noteStore.mutex.Unlock()

	noteStore.authServerAddress = "localhost:8001"
	noteStore.notes = []*pb.Note{
		{NoteId: 0, Text: "blah blah blah"},
		{NoteId: 1, Text: "very important note"},
	}
}

func main() {
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatal(err)
	}

	s := grpc.NewServer()
	pb.RegisterNoteStoreServer(s, &noteStore)
	if err := s.Serve(lis); err != nil {
		log.Fatal(err)
	}
}

func (ns *noteStoreServer) authenticate(ctx context.Context, token string) (int, error) {
	conn, err := grpc.Dial(ns.authServerAddress, grpc.WithInsecure())
	if err != nil {
		return -1, err
	}
	defer conn.Close()

	c := pb.NewAuthServiceClient(conn)
	resp, err := c.Authenticate(ctx, &pb.AuthenticateRequest{Token: token})
	if err != nil {
		return -1, err
	}
	return int(resp.UserId), nil
}

func (ns *noteStoreServer) GetNote(ctx context.Context, in *pb.GetNoteRequest) (*pb.Note, error) {
	_, err := ns.authenticate(ctx, in.Token)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "authentication failed")
	}

	for _, note := range ns.notes {
		if note.NoteId == in.NoteId {
			return note, nil
		}
	}
	return nil, status.Errorf(codes.NotFound, "no note with ID %d", in.NoteId)
}
