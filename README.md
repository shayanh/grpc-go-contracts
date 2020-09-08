# gRPC Go Contracts

Define preconditions and postcondtions for your gRPC APIs. For more information see [Design by Contract](https://en.wikipedia.org/wiki/Design_by_contract).


## Usage

Define contract for an RPC:

```go
rpcContract := &contracts.UnaryRPCContract{
    Method: pb.CheckoutServiceServer.PlaceOrder,
    PreConditions: []contracts.Condition{
        func(req *pb.PlaceOrderRequest) error {
            if req.GetCreditCard() == nil {
                return errors.New("credit card cannot be nil")
            }
            year := int32(time.Now().Year())
            if req.GetCreditCard().GetCreditCardExpirationYear() < year || req.CreditCard.GetCreditCardExpirationYear() > year+4 {
                log.Info("year = ", year, " req.year = ", req.GetCreditCard().GetCreditCardExpirationYear())
                return errors.New("credit card year is in invalid range")
            }
            if req.CreditCard.GetCreditCardExpirationMonth() > 12 {
                return errors.New("credit card month is in invalid range")
            }
            return nil
        },
        func(req *pb.PlaceOrderRequest) error {
            if req.GetUserId() == "" {
                return errors.New("user must be authenticated")
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
            shippingCalls := calls.Filter(pb.ShippingServiceClient.ShipOrder)
            if len(shippingCalls) < 1 {
                return errors.New("no call to shipping service")
            }
            shippingCall := shippingCalls[0]
            if shippingCall.Error != nil || shippingCall.Response.(*pb.ShipOrderResponse).GetTrackingId() == "" {
                return errors.New("invalid response from shipping service")
            }
            return nil
        },
    },
}
```


Define contract for the server and register related RPC contracts:

```go
var log *logrus.Logger
serverContract = contracts.NewServerContract(log)
serverContract.RegisterUnaryRPCContract(rpcContract)
```