package main

import (
	"context"
	"errors"
	"log"
	"net"
	"sync"

	"github.com/shayanh/grpc-go-contracts/contracts"
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
	contract          *contracts.ServerContract

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

func createContract() *contracts.ServerContract {
	getNoteContract := &contracts.UnaryRPCContract{
		MethodName: "GetNote",
		PreConditions: []contracts.Condition{
			func(in *pb.GetNoteRequest) error {
				if in.NoteId < 0 {
					return errors.New("NoteId must be positive")
				}
				return nil
			},
		},
		PostConditions: []contracts.Condition{
			func(out *pb.Note, outErr error, in *pb.GetNoteRequest, calls contracts.RPCCallHistory) error {
				if outErr != nil {
					return nil
				}
				if calls.Filter("noteservice.AuthService", "Authenticate").Successful().Empty() {
					return errors.New("no successful call to auth service")
				}
				return nil
			},
			func(out *pb.Note, outErr error, in *pb.GetNoteRequest, calls contracts.RPCCallHistory) error {
				if outErr != nil {
					return nil
				}
				if in.NoteId != out.NoteId {
					return errors.New("wrong note id in response")
				}
				return nil
			},
		},
	}
	noteStoreContract := &contracts.ServiceContract{
		ServiceName: "noteservice.NoteStore",
		RPCContracts: []*contracts.UnaryRPCContract{
			getNoteContract,
		},
	}
	serverContract := contracts.NewServerContract(log.Println)
	serverContract.RegisterServiceContract(noteStoreContract)
	return serverContract
}

func main() {
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatal(err)
	}

	contract := createContract()
	noteStore.contract = contract

	s := grpc.NewServer(grpc.UnaryInterceptor(contract.UnaryServerInterceptor()))
	pb.RegisterNoteStoreServer(s, &noteStore)
	if err := s.Serve(lis); err != nil {
		log.Fatal(err)
	}
}

func (ns *noteStoreServer) authenticate(ctx context.Context, token string) (int, error) {
	conn, err := grpc.Dial(ns.authServerAddress, grpc.WithInsecure(),
		grpc.WithUnaryInterceptor(ns.contract.UnaryClientInterceptor()))
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
			// Wrong implementation:
			// return &pb.Note{NoteId: note.NoteId + 1, Text: note.Text}, nil
		}
	}
	return nil, status.Errorf(codes.NotFound, "no note with ID %d", in.NoteId)
}
