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
	port = ":8001"
)

type user struct {
	userID int
	token  string
}

type authServiceServer struct {
	pb.UnimplementedAuthServiceServer

	mutex sync.Mutex
	users []*user
}

var authService authServiceServer

func init() {
	authService.mutex.Lock()
	defer authService.mutex.Unlock()

	authService.users = []*user{
		{userID: 0, token: "some-token-0"},
		{userID: 1, token: "some-token-1"},
	}
}

func main() {
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatal(err)
	}

	s := grpc.NewServer()
	pb.RegisterAuthServiceServer(s, &authService)
	if err := s.Serve(lis); err != nil {
		log.Fatal(err)
	}
}

func (as *authServiceServer) Authenticate(ctx context.Context, in *pb.AuthenticateRequest) (*pb.AuthenticateResponse, error) {
	for _, user := range as.users {
		if user.token == in.Token {
			return &pb.AuthenticateResponse{UserId: int32(user.userID)}, nil
		}
	}
	return nil, status.Error(codes.Unauthenticated, "invalid token")
}
