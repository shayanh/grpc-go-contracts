# gRPC Go Contracts

Define preconditions and postcondtions for your gRPC APIs. For more information see [Design by Contract](https://en.wikipedia.org/wiki/Design_by_contract).


## Installation

Using `go get`:

```bash
$ go get github.com/shayanh/grpc-go-contracts/contracts
```

## Usage

Define a contract for an RPC:

```go
placeOrderContract := &contracts.UnaryRPCContract{
    MethodName: "PlaceOrder",
    PreConditions: []contracts.Condition{
        // CreditCard number must be valid
        func(req *pb.PlaceOrderRequest) error {
            var creditCardNumberRegex = regexp.MustCompile("\\d{4}-\\d{4}-\\d{4}-\\d{4}")
            n := req.CreditCard.GetCreditCardNumber()
            if !creditCardNumberRegex.MatchString(n) {
                return errors.New("credit card number is not valid")
            }
            return nil
        },
    },
    PostConditions: []contracts.Condition{
        // Ensure a successful place order call, has a successful ship order call
        func(resp *pb.PlaceOrderResponse, respErr error, req *pb.PlaceOrderRequest, calls contracts.RPCCallHistory) error {
            if respErr != nil {
                return nil
            }
            if calls.Filter("hipstershop.ShippingService", "ShipOrder").Successful().Empty() {
                return errors.New("no successful call to shipping service")
            }
            return nil
        },
    },
}
```

Define contracts for a gRPC service and server:

```go
checkoutServiceContract := &contracts.ServiceContract{
    ServiceName: "hipstershop.CheckoutService",
    RPCContracts: []*contracts.UnaryRPCContract{
        placeOrderContract,
    },
}

var log *logrus.Logger
serverContract = contracts.NewServerContract(log)
serverContract.RegisterServiceContract(checkoutServiceContract)
```

And when using a gRPC client, remember to use `serverContract.UnaryClientInterceptor()`:

```go
conn, err := grpc.DialContext(ctx, shippingSvcAddr, grpc.WithUnaryInterceptor(serverContract.UnaryClientInterceptor()))
```