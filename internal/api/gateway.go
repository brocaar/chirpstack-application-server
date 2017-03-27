package api

import (
	"encoding/hex"
	"strings"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

	pb "github.com/brocaar/lora-app-server/api"
	"github.com/brocaar/lora-app-server/internal/api/auth"
	"github.com/brocaar/lora-app-server/internal/common"
	"github.com/brocaar/loraserver/api/ns"
	"github.com/brocaar/lorawan"
)

// GatewayAPI exports the Gateway related functions.
type GatewayAPI struct {
	ctx       common.Context
	validator auth.Validator
}

// NewGatewayAPI creates a new GatewayAPI.
func NewGatewayAPI(ctx common.Context, validator auth.Validator) *GatewayAPI {
	return &GatewayAPI{
		ctx:       ctx,
		validator: validator,
	}
}

// Create creates the given gateway.
func (a *GatewayAPI) Create(ctx context.Context, req *pb.CreateGatewayRequest) (*pb.CreateGatewayResponse, error) {
	err := a.validator.Validate(ctx, auth.ValidateGatewaysAccess(auth.Create))
	if err != nil {
		return nil, grpc.Errorf(codes.Unauthenticated, "authentication failed: %s", err)
	}

	macbytes, err := hex.DecodeString(req.Mac)
	if err != nil {
		return nil, grpc.Errorf(codes.InvalidArgument, "bad gateway mac: %s", err)
	}

	createReq := ns.CreateGatewayRequest{
		Mac:         macbytes,
		Name:        req.Name,
		Description: req.Description,
		Latitude:    req.Latitude,
		Longitude:   req.Longitude,
		Altitude:    req.Altitude,
	}

	_, err = a.ctx.NetworkServer.CreateGateway(ctx, &createReq)
	if err != nil {
		return nil, err
	}

	return &pb.CreateGatewayResponse{}, nil
}

// Get returns the gateway matching the given Mac.
func (a *GatewayAPI) Get(ctx context.Context, req *pb.GetGatewayRequest) (*pb.GetGatewayResponse, error) {
	var mac lorawan.EUI64
	if err := mac.UnmarshalText([]byte(req.Mac)); err != nil {
		return nil, grpc.Errorf(codes.InvalidArgument, "bad gateway mac: %s", err)
	}

	err := a.validator.Validate(ctx, auth.ValidateGatewayAccess(auth.Read, mac))
	if err != nil {
		return nil, grpc.Errorf(codes.Unauthenticated, "authentication failed: %s", err)
	}

	getReq := ns.GetGatewayRequest{
		Mac: mac[:],
	}

	getResp, err := a.ctx.NetworkServer.GetGateway(ctx, &getReq)
	if err != nil {
		return nil, err
	}

	ret := &pb.GetGatewayResponse{
		Mac:         mac.String(),
		Name:        getResp.Name,
		Description: getResp.Description,
		Latitude:    getResp.Latitude,
		Longitude:   getResp.Longitude,
		Altitude:    getResp.Altitude,
		CreatedAt:   getResp.CreatedAt,
		UpdatedAt:   getResp.UpdatedAt,
		FirstSeenAt: getResp.FirstSeenAt,
		LastSeenAt:  getResp.LastSeenAt,
	}
	return ret, err
}

// List lists the gateways.
func (a *GatewayAPI) List(ctx context.Context, req *pb.ListGatewayRequest) (*pb.ListGatewayResponse, error) {
	err := a.validator.Validate(ctx, auth.ValidateGatewaysAccess(auth.List))
	if err != nil {
		return nil, grpc.Errorf(codes.Unauthenticated, "authentication failed: %s", err)
	}

	listReq := ns.ListGatewayRequest{
		Limit:  req.Limit,
		Offset: req.Offset,
	}
	gws, err := a.ctx.NetworkServer.ListGateways(ctx, &listReq)
	if err != nil {
		return nil, err
	}

	result := make([]*pb.GetGatewayResponse, len(gws.Result))
	for i, getResp := range gws.Result {
		result[i] = &pb.GetGatewayResponse{
			Mac:         hex.EncodeToString(getResp.Mac),
			Name:        getResp.Name,
			Description: getResp.Description,
			Latitude:    getResp.Latitude,
			Longitude:   getResp.Longitude,
			Altitude:    getResp.Altitude,
			CreatedAt:   getResp.CreatedAt,
			UpdatedAt:   getResp.UpdatedAt,
			FirstSeenAt: getResp.FirstSeenAt,
			LastSeenAt:  getResp.LastSeenAt,
		}
	}

	return &pb.ListGatewayResponse{
		TotalCount: gws.TotalCount,
		Result:     result,
	}, nil
}

// Update updates the given gateway.
func (a *GatewayAPI) Update(ctx context.Context, req *pb.UpdateGatewayRequest) (*pb.UpdateGatewayResponse, error) {
	var mac lorawan.EUI64
	if err := mac.UnmarshalText([]byte(req.Mac)); err != nil {
		return nil, grpc.Errorf(codes.InvalidArgument, "bad gateway mac: %s", err)
	}

	err := a.validator.Validate(ctx, auth.ValidateGatewayAccess(auth.Update, mac))
	if err != nil {
		return nil, grpc.Errorf(codes.Unauthenticated, "authentication failed: %s", err)
	}

	updateReq := ns.UpdateGatewayRequest{
		Mac:         mac[:],
		Name:        req.Name,
		Description: req.Description,
		Latitude:    req.Latitude,
		Longitude:   req.Longitude,
		Altitude:    req.Altitude,
	}
	_, err = a.ctx.NetworkServer.UpdateGateway(ctx, &updateReq)
	if err != nil {
		return nil, err
	}

	return &pb.UpdateGatewayResponse{}, nil
}

// Delete deletes the gateway matching the given ID.
func (a *GatewayAPI) Delete(ctx context.Context, req *pb.DeleteGatewayRequest) (*pb.DeleteGatewayResponse, error) {
	var mac lorawan.EUI64
	if err := mac.UnmarshalText([]byte(req.Mac)); err != nil {
		return nil, grpc.Errorf(codes.InvalidArgument, "bad gateway mac: %s", err)
	}

	err := a.validator.Validate(ctx, auth.ValidateGatewayAccess(auth.Delete, mac))
	if err != nil {
		return nil, grpc.Errorf(codes.Unauthenticated, "authentication failed: %s", err)
	}

	deleteReq := ns.DeleteGatewayRequest{
		Mac: mac[:],
	}
	_, err = a.ctx.NetworkServer.DeleteGateway(ctx, &deleteReq)
	if err != nil {
		return nil, err
	}

	return &pb.DeleteGatewayResponse{}, nil
}

// GetStats gets the gateway statistics for the gateway with the given Mac.
func (a *GatewayAPI) GetStats(ctx context.Context, req *pb.GetGatewayStatsRequest) (*pb.GetGatewayStatsResponse, error) {
	var mac lorawan.EUI64
	if err := mac.UnmarshalText([]byte(req.Mac)); err != nil {
		return nil, grpc.Errorf(codes.InvalidArgument, "bad gateway mac: %s", err)
	}

	err := a.validator.Validate(ctx, auth.ValidateGatewayAccess(auth.Read, mac))
	if err != nil {
		return nil, grpc.Errorf(codes.Unauthenticated, "authentication failed: %s", err)
	}

	interval, ok := ns.AggregationInterval_value[strings.ToUpper(req.Interval)]
	if !ok {
		return nil, grpc.Errorf(codes.InvalidArgument, "bad interval: %s", req.Interval)
	}

	statsReq := ns.GetGatewayStatsRequest{
		Mac:            mac[:],
		Interval:       ns.AggregationInterval(interval),
		StartTimestamp: req.StartTimestamp,
		EndTimestamp:   req.EndTimestamp,
	}
	stats, err := a.ctx.NetworkServer.GetGatewayStats(ctx, &statsReq)
	if err != nil {
		return nil, err
	}

	result := make([]*pb.GatewayStats, len(stats.Result))
	for i, stat := range stats.Result {
		result[i] = &pb.GatewayStats{
			StartTimestamp:      stat.StartTimestamp,
			RxPacketsReceived:   stat.RxPacketsReceived,
			RxPacketsReceivedOK: stat.RxPacketsReceivedOK,
			TxPacketsReceived:   stat.TxPacketsReceived,
			TxPacketsEmitted:    stat.TxPacketsEmitted,
		}
	}

	return &pb.GetGatewayStatsResponse{
		Result: result,
	}, nil
}
