package main

import (
	"context"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

type RealCounterService struct {
	UnimplementedCounterServiceServer
}

func (r *RealCounterService) ExternalMessage(ctx context.Context, request *ExternalRequest) (*ExternalResponse, error) {
	return &ExternalResponse{
		Result: request.Content + "!!",
	}, nil
}

func (r *RealCounterService) Increment(c context.Context, req *UnaryRequest) (*UnaryResponse, error) {
	return &UnaryResponse{
		Result: req.Counter + 1,
	}, nil
}

func (r *RealCounterService) FailingIncrement(c context.Context, req *UnaryRequest) (*UnaryResponse, error) {
	return nil, status.Errorf(codes.Unavailable, "this increment does not work")
}

func (r *RealCounterService) EchoBinary(c context.Context, req *BinaryRequest) (*BinaryResponse, error) {
	return &BinaryResponse{
		Data: req.Data,
	}, nil
}

func (r *RealCounterService) StreamingIncrements(req *StreamingRequest, service CounterService_StreamingIncrementsServer) error {
	times := 5
	counter := req.Counter

	for i := 0; i < times; i++ {
		counter++
		err := service.Send(&StreamingResponse{
			Result: counter,
		})
		if err != nil {
			return err
		}

		time.Sleep(200 * time.Millisecond)
	}

	return nil
}

func (r *RealCounterService) HTTPGet(ctx context.Context, req *HttpGetRequest) (*HttpGetResponse, error) {
	return &HttpGetResponse{
		Result: req.NumToIncrease + 1,
	}, nil
}

func (r *RealCounterService) HTTPPostWithNestedBodyPath(ctx context.Context, in *HttpPostRequest) (*HttpPostResponse, error) {
	return &HttpPostResponse{
		PostResult: in.A + in.Req.B,
	}, nil
}

func (r *RealCounterService) HTTPPostWithStarBodyPath(ctx context.Context, in *HttpPostRequest) (*HttpPostResponse, error) {
	return &HttpPostResponse{
		PostResult: in.A + in.Req.B + in.C,
	}, nil
}

func (r *RealCounterService) HTTPPatch(ctx context.Context, in *HttpPatchRequest) (*HttpPatchResponse, error) {
	return &HttpPatchResponse{
		PatchResult: in.A + in.C,
	}, nil
}

func (r *RealCounterService) HTTPDelete(ctx context.Context, req *HttpDeleteRequest) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, nil
}

func (r *RealCounterService) HTTPDeleteWithParams(ctx context.Context, req *HttpDeleteWithParamsRequest) (*HttpDeleteWithParamsResponse, error) {
	return &HttpDeleteWithParamsResponse{Reason: req.Reason}, nil
}

func (r *RealCounterService) HTTPGetWithURLSearchParams(ctx context.Context, in *HTTPGetWithURLSearchParamsRequest) (*HTTPGetWithURLSearchParamsResponse, error) {
	totalC := 0
	for _, c := range in.GetC() {
		totalC += int(c)
	}
	return &HTTPGetWithURLSearchParamsResponse{
		UrlSearchParamsResult: in.GetA() + in.GetB().GetB() + in.GetD().GetD() + int32(totalC),
	}, nil
}

func (r *RealCounterService) HTTPGetWithZeroValueURLSearchParams(ctx context.Context, in *HTTPGetWithZeroValueURLSearchParamsRequest) (*HTTPGetWithZeroValueURLSearchParamsResponse, error) {
	var incrementedD []int32
	for _, d := range in.GetC().GetD() {
		incrementedD = append(incrementedD, d+1)
	}
	return &HTTPGetWithZeroValueURLSearchParamsResponse{
		A: in.GetA(),
		B: in.GetB() + "hello",
		ZeroValueMsg: &ZeroValueMsg{
			C: in.GetC().GetC() + 1,
			D: incrementedD,
			E: !in.GetC().GetE(),
		},
	}, nil
}

func (r *RealCounterService) HTTPGetWithOptionalFields(ctx context.Context, in *OptionalFieldsRequest) (*OptionalFieldsResponse, error) {
	return &OptionalFieldsResponse{
		// Empty fields aren't explicity enumerated at all.

		ZeroStr:       "",
		ZeroNumber:    int32(0),
		ZeroMsg:       &OptionalFieldsSubMsg{},
		ZeroOptStr:    p(""),
		ZeroOptNumber: p(int32(0)),
		ZeroOptMsg:    &OptionalFieldsSubMsg{},

		DefinedStr:       "hello",
		DefinedNumber:    int32(123),
		DefinedMsg:       &OptionalFieldsSubMsg{Str: "hello", OptStr: p("hello")},
		DefinedOptStr:    p("hello"),
		DefinedOptNumber: p(int32(123)),
		DefinedOptMsg:    &OptionalFieldsSubMsg{Str: "hello", OptStr: p("hello")},
	}, nil
}

func p[T any](v T) *T {
	return &v
}
