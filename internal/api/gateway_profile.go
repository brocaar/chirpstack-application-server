package api

import (
	"github.com/golang/protobuf/ptypes"
	"github.com/gofrs/uuid"

	"github.com/brocaar/loraserver/api/ns"

	pb "github.com/brocaar/lora-app-server/api"
	"github.com/brocaar/lora-app-server/internal/api/auth"
	"github.com/brocaar/lora-app-server/internal/config"
	"github.com/brocaar/lora-app-server/internal/storage"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/jmoiron/sqlx"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
)

// GatewayProfileAPI exports the GatewayProfile related functions.
type GatewayProfileAPI struct {
	validator auth.Validator
}

// NewGatewayProfileAPI creates a new GatewayProfileAPI.
func NewGatewayProfileAPI(validator auth.Validator) *GatewayProfileAPI {
	return &GatewayProfileAPI{
		validator: validator,
	}
}

// Create creates the given gateway-profile.
func (a *GatewayProfileAPI) Create(ctx context.Context, req *pb.CreateGatewayProfileRequest) (*pb.CreateGatewayProfileResponse, error) {
	if req.GatewayProfile == nil {
		return nil, grpc.Errorf(codes.InvalidArgument, "gateway_profile must not be nil")
	}

	if err := a.validator.Validate(ctx,
		auth.ValidateGatewayProfileAccess(auth.Create),
	); err != nil {
		return nil, grpc.Errorf(codes.Unauthenticated, "authentication failed: %s", err)
	}

	gp := storage.GatewayProfile{
		NetworkServerID: req.GatewayProfile.NetworkServerId,
		Name:            req.GatewayProfile.Name,
		GatewayProfile: ns.GatewayProfile{
			Channels: req.GatewayProfile.Channels,
		},
	}

	for _, ec := range req.GatewayProfile.ExtraChannels {
		gp.GatewayProfile.ExtraChannels = append(gp.GatewayProfile.ExtraChannels, &ns.GatewayProfileExtraChannel{
			Frequency:        ec.Frequency,
			Bandwidth:        ec.Bandwidth,
			Bitrate:          ec.Bitrate,
			SpreadingFactors: ec.SpreadingFactors,
			Modulation:       ec.Modulation,
		})
	}

	err := storage.Transaction(config.C.PostgreSQL.DB, func(tx sqlx.Ext) error {
		return storage.CreateGatewayProfile(tx, &gp)
	})
	if err != nil {
		return nil, errToRPCError(err)
	}

	gpID, err := uuid.FromBytes(gp.GatewayProfile.Id)
	if err != nil {
		return nil, errToRPCError(err)
	}

	return &pb.CreateGatewayProfileResponse{
		Id: gpID.String(),
	}, nil
}

// Get returns the gateway-profile matching the given id.
func (a *GatewayProfileAPI) Get(ctx context.Context, req *pb.GetGatewayProfileRequest) (*pb.GetGatewayProfileResponse, error) {
	if err := a.validator.Validate(ctx,
		auth.ValidateGatewayProfileAccess(auth.Read),
	); err != nil {
		return nil, grpc.Errorf(codes.Unauthenticated, "authentication failed: %s", err)
	}

	gpID, err := uuid.FromString(req.Id)
	if err != nil {
		return nil, grpc.Errorf(codes.InvalidArgument, "uuid error: %s", err)
	}

	gp, err := storage.GetGatewayProfile(config.C.PostgreSQL.DB, gpID)
	if err != nil {
		return nil, errToRPCError(err)
	}

	out := pb.GetGatewayProfileResponse{
		GatewayProfile: &pb.GatewayProfile{
			Id:              gpID.String(),
			Name:            gp.Name,
			NetworkServerId: gp.NetworkServerID,
			Channels:        gp.GatewayProfile.Channels,
		},
	}

	out.CreatedAt, err = ptypes.TimestampProto(gp.CreatedAt)
	if err != nil {
		return nil, errToRPCError(err)
	}
	out.UpdatedAt, err = ptypes.TimestampProto(gp.UpdatedAt)
	if err != nil {
		return nil, errToRPCError(err)
	}

	for _, ec := range gp.GatewayProfile.ExtraChannels {
		out.GatewayProfile.ExtraChannels = append(out.GatewayProfile.ExtraChannels, &pb.GatewayProfileExtraChannel{
			Frequency:        ec.Frequency,
			Bandwidth:        ec.Bandwidth,
			Bitrate:          ec.Bitrate,
			SpreadingFactors: ec.SpreadingFactors,
			Modulation:       ec.Modulation,
		})
	}

	return &out, nil
}

// Update updates the given gateway-profile.
func (a *GatewayProfileAPI) Update(ctx context.Context, req *pb.UpdateGatewayProfileRequest) (*empty.Empty, error) {
	if req.GatewayProfile == nil {
		return nil, grpc.Errorf(codes.InvalidArgument, "gateway_profile must not be nil")
	}

	if err := a.validator.Validate(ctx,
		auth.ValidateGatewayProfileAccess(auth.Update),
	); err != nil {
		return nil, grpc.Errorf(codes.Unauthenticated, "authentication failed: %s", err)
	}

	gpID, err := uuid.FromString(req.GatewayProfile.Id)
	if err != nil {
		return nil, grpc.Errorf(codes.InvalidArgument, "uuid error: %s", err)
	}

	gp, err := storage.GetGatewayProfile(config.C.PostgreSQL.DB, gpID)
	if err != nil {
		return nil, errToRPCError(err)
	}

	gp.Name = req.GatewayProfile.Name
	gp.GatewayProfile.Channels = req.GatewayProfile.Channels
	gp.GatewayProfile.ExtraChannels = []*ns.GatewayProfileExtraChannel{}

	for _, ec := range req.GatewayProfile.ExtraChannels {
		gp.GatewayProfile.ExtraChannels = append(gp.GatewayProfile.ExtraChannels, &ns.GatewayProfileExtraChannel{
			Frequency:        ec.Frequency,
			Bandwidth:        ec.Bandwidth,
			Bitrate:          ec.Bitrate,
			SpreadingFactors: ec.SpreadingFactors,
			Modulation:       ec.Modulation,
		})
	}

	err = storage.Transaction(config.C.PostgreSQL.DB, func(tx sqlx.Ext) error {
		return storage.UpdateGatewayProfile(tx, &gp)
	})
	if err != nil {
		return nil, errToRPCError(err)
	}

	return &empty.Empty{}, nil
}

// Delete deletes the gateway-profile matching the given id.
func (a *GatewayProfileAPI) Delete(ctx context.Context, req *pb.DeleteGatewayProfileRequest) (*empty.Empty, error) {
	if err := a.validator.Validate(ctx,
		auth.ValidateGatewayProfileAccess(auth.Delete),
	); err != nil {
		return nil, grpc.Errorf(codes.Unauthenticated, "authentication failed: %s", err)
	}

	gpID, err := uuid.FromString(req.Id)
	if err != nil {
		return nil, grpc.Errorf(codes.InvalidArgument, "uuid error: %s", err)
	}

	err = storage.DeleteGatewayProfile(config.C.PostgreSQL.DB, gpID)
	if err != nil {
		return nil, errToRPCError(err)
	}

	return &empty.Empty{}, nil
}

// List returns the existing gateway-profiles.
func (a *GatewayProfileAPI) List(ctx context.Context, req *pb.ListGatewayProfilesRequest) (*pb.ListGatewayProfilesResponse, error) {
	if err := a.validator.Validate(ctx,
		auth.ValidateGatewayProfileAccess(auth.List),
	); err != nil {
		return nil, grpc.Errorf(codes.Unauthenticated, "authentication failed: %s", err)
	}

	var err error
	var count int
	var gps []storage.GatewayProfileMeta

	if req.NetworkServerId == 0 {
		count, err = storage.GetGatewayProfileCount(config.C.PostgreSQL.DB)
		if err != nil {
			return nil, errToRPCError(err)
		}

		gps, err = storage.GetGatewayProfiles(config.C.PostgreSQL.DB, int(req.Limit), int(req.Offset))
		if err != nil {
			return nil, errToRPCError(err)
		}
	} else {
		count, err = storage.GetGatewayProfileCountForNetworkServerID(config.C.PostgreSQL.DB, req.NetworkServerId)
		if err != nil {
			return nil, errToRPCError(err)
		}

		gps, err = storage.GetGatewayProfilesForNetworkServerID(config.C.PostgreSQL.DB, req.NetworkServerId, int(req.Limit), int(req.Offset))
		if err != nil {
			return nil, errToRPCError(err)
		}
	}

	out := pb.ListGatewayProfilesResponse{
		TotalCount: int64(count),
	}

	for _, gp := range gps {
		row := pb.GatewayProfileListItem{
			Id:                gp.GatewayProfileID.String(),
			Name:              gp.Name,
			NetworkServerName: gp.NetworkServerName,
			NetworkServerId:   gp.NetworkServerID,
		}

		row.CreatedAt, err = ptypes.TimestampProto(gp.CreatedAt)
		if err != nil {
			return nil, errToRPCError(err)
		}
		row.UpdatedAt, err = ptypes.TimestampProto(gp.UpdatedAt)
		if err != nil {
			return nil, errToRPCError(err)
		}

		out.Result = append(out.Result, &row)
	}

	return &out, nil
}
