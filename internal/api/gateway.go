package api

import (
	"strings"
	"time"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

	pb "github.com/brocaar/lora-app-server/api"
	"github.com/brocaar/lora-app-server/internal/api/auth"
	"github.com/brocaar/lora-app-server/internal/common"
	"github.com/brocaar/lora-app-server/internal/storage"
	"github.com/brocaar/loraserver/api/ns"
	"github.com/brocaar/lorawan"
	"github.com/jmoiron/sqlx"
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
	err := a.validator.Validate(ctx, auth.ValidateGatewaysAccess(auth.Create, req.OrganizationID))
	if err != nil {
		return nil, grpc.Errorf(codes.Unauthenticated, "authentication failed: %s", err)
	}

	var mac lorawan.EUI64
	if err := mac.UnmarshalText([]byte(req.Mac)); err != nil {
		return nil, grpc.Errorf(codes.InvalidArgument, "bad gateway mac: %s", err)
	}

	createReq := ns.CreateGatewayRequest{
		Mac:         mac[:],
		Name:        req.Name,
		Description: req.Description,
		Latitude:    req.Latitude,
		Longitude:   req.Longitude,
		Altitude:    req.Altitude,
	}

	err = storage.Transaction(a.ctx.DB, func(tx *sqlx.Tx) error {
		err = storage.CreateGateway(tx, &storage.Gateway{
			MAC:            mac,
			Name:           req.Name,
			Description:    req.Description,
			OrganizationID: req.OrganizationID,
		})
		if err != nil {
			return errToRPCError(err)
		}

		_, err = a.ctx.NetworkServer.CreateGateway(ctx, &createReq)
		if err != nil && grpc.Code(err) != codes.AlreadyExists {
			return err
		}

		return nil
	})
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

	getResp, err := a.ctx.NetworkServer.GetGateway(ctx, &ns.GetGatewayRequest{
		Mac: mac[:],
	})
	if err != nil {
		return nil, err
	}

	gw, err := storage.GetGateway(a.ctx.DB, mac)
	if err != nil {
		return nil, errToRPCError(err)
	}

	ret := &pb.GetGatewayResponse{
		Mac:         mac.String(),
		Name:        gw.Name,
		Description: gw.Description,
		Latitude:    getResp.Latitude,
		Longitude:   getResp.Longitude,
		Altitude:    getResp.Altitude,
		CreatedAt:   gw.CreatedAt.Format(time.RFC3339Nano),
		UpdatedAt:   gw.UpdatedAt.Format(time.RFC3339Nano),
		FirstSeenAt: getResp.FirstSeenAt,
		LastSeenAt:  getResp.LastSeenAt,
	}
	return ret, err
}

// List lists the gateways.
func (a *GatewayAPI) List(ctx context.Context, req *pb.ListGatewayRequest) (*pb.ListGatewayResponse, error) {
	err := a.validator.Validate(ctx, auth.ValidateGatewaysAccess(auth.List, req.OrganizationID))
	if err != nil {
		return nil, grpc.Errorf(codes.Unauthenticated, "authentication failed: %s", err)
	}

	var count int
	var gws []storage.Gateway

	if req.OrganizationID == 0 {
		isAdmin, err := a.validator.GetIsAdmin(ctx)
		if err != nil {
			return nil, errToRPCError(err)
		}

		if isAdmin {
			// in case of admin user list all gateways
			count, err = storage.GetGatewayCount(a.ctx.DB)
			if err != nil {
				return nil, errToRPCError(err)
			}

			gws, err = storage.GetGateways(a.ctx.DB, int(req.Limit), int(req.Offset))
			if err != nil {
				return nil, errToRPCError(err)
			}
		} else {
			// filter result based on user
			username, err := a.validator.GetUsername(ctx)
			if err != nil {
				return nil, errToRPCError(err)
			}
			count, err = storage.GetGatewayCountForUser(a.ctx.DB, username)
			if err != nil {
				return nil, errToRPCError(err)
			}
			gws, err = storage.GetGatewaysForUser(a.ctx.DB, username, int(req.Limit), int(req.Offset))
			if err != nil {
				return nil, errToRPCError(err)
			}
		}
	} else {
		count, err = storage.GetGatewayCountForOrganizationID(a.ctx.DB, req.OrganizationID)
		if err != nil {
			return nil, errToRPCError(err)
		}
		gws, err = storage.GetGatewaysForOrganizationDB(a.ctx.DB, req.OrganizationID, int(req.Limit), int(req.Offset))
		if err != nil {
			return nil, errToRPCError(err)
		}
	}

	result := make([]*pb.ListGatewayItem, 0, len(gws))
	for i := range gws {
		result = append(result, &pb.ListGatewayItem{
			Mac:            gws[i].MAC.String(),
			Name:           gws[i].Name,
			Description:    gws[i].Description,
			CreatedAt:      gws[i].CreatedAt.Format(time.RFC3339Nano),
			UpdatedAt:      gws[i].UpdatedAt.Format(time.RFC3339Nano),
			OrganizationID: gws[i].OrganizationID,
		})
	}

	return &pb.ListGatewayResponse{
		TotalCount: int32(count),
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

	isAdmin, err := a.validator.GetIsAdmin(ctx)
	if err != nil {
		return nil, errToRPCError(err)
	}

	err = storage.Transaction(a.ctx.DB, func(tx *sqlx.Tx) error {
		gw, err := storage.GetGateway(a.ctx.DB, mac)
		if err != nil {
			return errToRPCError(err)
		}

		gw.Name = req.Name
		gw.Description = req.Description
		if isAdmin {
			gw.OrganizationID = req.OrganizationID
		}

		err = storage.UpdateGateway(tx, &gw)
		if err != nil {
			return errToRPCError(err)
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
			return err
		}
		return nil
	})
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

	err = storage.Transaction(a.ctx.DB, func(tx *sqlx.Tx) error {
		err = storage.DeleteGateway(tx, mac)
		if err != nil {
			return errToRPCError(err)
		}

		_, err = a.ctx.NetworkServer.DeleteGateway(ctx, &ns.DeleteGatewayRequest{
			Mac: mac[:],
		})
		if err != nil && grpc.Code(err) != codes.NotFound {
			return err
		}

		return nil
	})
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
			Timestamp:           stat.Timestamp,
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
