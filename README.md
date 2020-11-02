# gRPC Go Contracts

[![PkgGoDev](https://pkg.go.dev/badge/github.com/shayanh/grpc-go-contracts/contracts)](https://pkg.go.dev/github.com/shayanh/grpc-go-contracts/contracts)

Verify the communication of your microservices by writing contracts for your RPCs.

gRPC Go Contracts implements contract programming (aka Design by Contract) for gRPC methods written in go. It supports: 

* **Preconditions**: Preconditions are conditions that must always be true just before the execution of the RPC. In a precondition, you can access RPC's input values.
* **Postconditions**: Postconditions are conditions that must always be true just after the execution of the RPC. In a postcondition, you can access the RPC's input and return values. Moreover, you will be able to access RPC calls made by the requested RPC during the request lifetime. This allows you to verify the execution order of RPC calls, which is amazing! For more details please see the [example](#usage-and-example) below.

In the case of contract violation, gRPC Go Contracts logs the contract error message and related parameters. At this time, just unary RPCs are supported. 

For more information please see: https://en.wikipedia.org/wiki/Design_by_contract

## Installation

```bash
$ go get github.com/shayanh/grpc-go-contracts/contracts
```

## Usage and Example

Let's consider a very simple note-taking application named NoteService. NoteService consists of two microservices:

* [**NoteStore**](examples/noteservice/notestore/main.go): NoteStore simply stores notes. Its only API is `GetNote(note_id, token)`. `GetNote` first authenticates the input `token` by calling AuthServices. If authentication was successful, it returns the related note.
* [**AuthService**](examples/noteservice/authservice/main.go): AuthService is responsible for authentication. Its only API is `Authenticate(token)`. `Authenticate` gets a token, and if the token was valid, it returns the related user ID.

<p align="center">
    <img src="img/NoteService.png?raw=true" alt="NoteService diagram">
</p>

Protocol buffers definition of these services:

```protobuf
package noteservice;

service NoteStore {
    rpc GetNote(GetNoteRequest) returns (Note) {}
}

message GetNoteRequest {
    int32 note_id = 1;
    string token = 2;
}

message Note {
    int32 note_id = 1;
    string text = 2;
}

service AuthService {
    rpc Authenticate(AuthenticateRequest) returns (AuthenticateResponse) {}
}

message AuthenticateRequest {
    string token = 1;
}

message AuthenticateResponse {
    int32 user_id = 1;
}
```

Now we want to write the following precondition for `GetNote` RPC:

1. `note_id` must be non-negative.

And we want to have the following postconditions for `GetNote` RPC:

1. If `GetNote` return value has no error, then `GetNote` must successfully have called `Authenticate` RPC on AuthService. 
2. If `GetNote` return value has no error, then output note ID must be equal to input `note_id`.

First, we define an `RPCContract` for `GetNote`:

```go
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
```

Next, we define a `ServiceContract` for the NoteStore service and a `ServerContract` for the gRPC server:

```go
noteStoreContract := &contracts.ServiceContract{
    ServiceName: "noteservice.NoteStore",
    RPCContracts: []*contracts.UnaryRPCContract{
        getNoteContract,
    },
}
serverContract := contracts.NewServerContract(log.Println)
serverContract.RegisterServiceContract(noteStoreContract)
```

Finally, we use `serverContract`'s interceptors in the gRPC server and clients:

```go
// server
s := grpc.NewServer(grpc.UnaryInterceptor(contract.UnaryServerInterceptor()))

// client
conn, err := grpc.Dial(authServerAddress, grpc.WithUnaryInterceptor(ns.contract.UnaryClientInterceptor()))
```

A complete version of the NoteService example containing all of the source codes is available [here](examples/noteservice/).


## API Documentation

See complete API documentation [here](https://pkg.go.dev/github.com/shayanh/grpc-go-contracts/contracts).


## TODO

- [ ] Write tests!
- [ ] Support streaming RPCs.
- [ ] Add terminate option on contract violation.
- [ ] Native support of popular logging libraries.
- [ ] Add asynchronous contract checking option.