package external

import (
	"github.com/gofrs/uuid"
	"github.com/golang/protobuf/ptypes"

	"github.com/brocaar/chirpstack-api/go/v3/ns"

	pb "github.com/brocaar/chirpstack-api/go/v3/as/external/api"
	"github.com/brocaar/chirpstack-application-server/internal/api/external/auth"
	"github.com/brocaar/chirpstack-application-server/internal/api/helpers"
	"github.com/brocaar/chirpstack-application-server/internal/storage"
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
			Channels:      req.GatewayProfile.Channels,
			StatsInterval: req.GatewayProfile.StatsInterval,
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

	err := storage.Transaction(func(tx sqlx.Ext) error {
		return storage.CreateGatewayProfile(ctx, tx, &gp)
	})
	if err != nil {
		return nil, helpers.ErrToRPCError(err)
	}

	gpID, err := uuid.FromBytes(gp.GatewayProfile.Id)
	if err != nil {
		return nil, helpers.ErrToRPCError(err)
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

	gp, err := storage.GetGatewayProfile(ctx, storage.DB(), gpID)
	if err != nil {
		return nil, helpers.ErrToRPCError(err)
	}

	out := pb.GetGatewayProfileResponse{
		GatewayProfile: &pb.GatewayProfile{
			Id:              gpID.String(),
			Name:            gp.Name,
			NetworkServerId: gp.NetworkServerID,
			Channels:        gp.GatewayProfile.Channels,
			StatsInterval:   gp.GatewayProfile.StatsInterval,
		},
	}

	out.CreatedAt, err = ptypes.TimestampProto(gp.CreatedAt)
	if err != nil {
		return nil, helpers.ErrToRPCError(err)
	}
	out.UpdatedAt, err = ptypes.TimestampProto(gp.UpdatedAt)
	if err != nil {
		return nil, helpers.ErrToRPCError(err)
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

	gp, err := storage.GetGatewayProfile(ctx, storage.DB(), gpID)
	if err != nil {
		return nil, helpers.ErrToRPCError(err)
	}

	gp.Name = req.GatewayProfile.Name
	gp.GatewayProfile.Channels = req.GatewayProfile.Channels
	gp.GatewayProfile.ExtraChannels = []*ns.GatewayProfileExtraChannel{}
	gp.GatewayProfile.StatsInterval = req.GatewayProfile.StatsInterval

	for _, ec := range req.GatewayProfile.ExtraChannels {
		gp.GatewayProfile.ExtraChannels = append(gp.GatewayProfile.ExtraChannels, &ns.GatewayProfileExtraChannel{
			Frequency:        ec.Frequency,
			Bandwidth:        ec.Bandwidth,
			Bitrate:          ec.Bitrate,
			SpreadingFactors: ec.SpreadingFactors,
			Modulation:       ec.Modulation,
		})
	}

	err = storage.Transaction(func(tx sqlx.Ext) error {
		return storage.UpdateGatewayProfile(ctx, tx, &gp)
	})
	if err != nil {
		return nil, helpers.ErrToRPCError(err)
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

	err = storage.DeleteGatewayProfile(ctx, storage.DB(), gpID)
	if err != nil {
		return nil, helpers.ErrToRPCError(err)
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
		count, err = storage.GetGatewayProfileCount(ctx, storage.DB())
		if err != nil {
			return nil, helpers.ErrToRPCError(err)
		}

		gps, err = storage.GetGatewayProfiles(ctx, storage.DB(), int(req.Limit), int(req.Offset))
		if err != nil {
			return nil, helpers.ErrToRPCError(err)
		}
	} else {
		count, err = storage.GetGatewayProfileCountForNetworkServerID(ctx, storage.DB(), req.NetworkServerId)
		if err != nil {
			return nil, helpers.ErrToRPCError(err)
		}

		gps, err = storage.GetGatewayProfilesForNetworkServerID(ctx, storage.DB(), req.NetworkServerId, int(req.Limit), int(req.Offset))
		if err != nil {
			return nil, helpers.ErrToRPCError(err)
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
			return nil, helpers.ErrToRPCError(err)
		}
		row.UpdatedAt, err = ptypes.TimestampProto(gp.UpdatedAt)
		if err != nil {
			return nil, helpers.ErrToRPCError(err)
		}

		out.Result = append(out.Result, &row)
	}

	return &out, nil
}
