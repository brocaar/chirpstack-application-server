package api

import (
	"time"

	pb "github.com/brocaar/lora-app-server/api"
	"github.com/brocaar/lora-app-server/internal/api/auth"
	"github.com/brocaar/lora-app-server/internal/config"
	"github.com/brocaar/lora-app-server/internal/storage"
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
	if err := a.validator.Validate(ctx,
		auth.ValidateGatewayProfileAccess(auth.Create),
	); err != nil {
		return nil, grpc.Errorf(codes.Unauthenticated, "authentication failed: %s", err)
	}

	if req.GatewayProfile == nil {
		return nil, grpc.Errorf(codes.InvalidArgument, "empty gateway-profile")
	}

	gp := storage.GatewayProfile{
		NetworkServerID: req.NetworkServerID,
		Name:            req.Name,
	}

	for _, c := range req.GatewayProfile.Channels {
		gp.Channels = append(gp.Channels, int(c))
	}

	for _, ec := range req.GatewayProfile.ExtraChannels {
		c := storage.ExtraChannel{
			Frequency: int(ec.Frequency),
			Bandwidth: int(ec.Bandwidth),
			Bitrate:   int(ec.Bitrate),
		}

		switch ec.Modulation {
		case pb.Modulation_FSK:
			c.Modulation = storage.ModulationFSK
		default:
			c.Modulation = storage.ModulationLoRa
		}

		for _, sf := range ec.SpreadingFactors {
			c.SpreadingFactors = append(c.SpreadingFactors, int(sf))
		}

		gp.ExtraChannels = append(gp.ExtraChannels, c)
	}

	err := storage.Transaction(config.C.PostgreSQL.DB, func(tx sqlx.Ext) error {
		return storage.CreateGatewayProfile(tx, &gp)
	})
	if err != nil {
		return nil, errToRPCError(err)
	}

	return &pb.CreateGatewayProfileResponse{
		GatewayProfileID: gp.GatewayProfileID,
	}, nil
}

// Get returns the gateway-profile matching the given id.
func (a *GatewayProfileAPI) Get(ctx context.Context, req *pb.GetGatewayProfileRequest) (*pb.GetGatewayProfileResponse, error) {
	if err := a.validator.Validate(ctx,
		auth.ValidateGatewayProfileAccess(auth.Read),
	); err != nil {
		return nil, grpc.Errorf(codes.Unauthenticated, "authentication failed: %s", err)
	}

	gp, err := storage.GetGatewayProfile(config.C.PostgreSQL.DB, req.GatewayProfileID)
	if err != nil {
		return nil, errToRPCError(err)
	}

	out := pb.GetGatewayProfileResponse{
		Name:            gp.Name,
		NetworkServerID: gp.NetworkServerID,
		CreatedAt:       gp.CreatedAt.Format(time.RFC3339Nano),
		UpdatedAt:       gp.UpdatedAt.Format(time.RFC3339Nano),
		GatewayProfile: &pb.GatewayProfile{
			GatewayProfileID: gp.GatewayProfileID,
		},
	}

	for _, c := range gp.Channels {
		out.GatewayProfile.Channels = append(out.GatewayProfile.Channels, uint32(c))
	}

	for _, ec := range gp.ExtraChannels {
		c := pb.GatewayProfileExtraChannel{
			Frequency: uint32(ec.Frequency),
			Bandwidth: uint32(ec.Bandwidth),
			Bitrate:   uint32(ec.Bitrate),
		}

		switch ec.Modulation {
		case storage.ModulationFSK:
			c.Modulation = pb.Modulation_FSK
		default:
			c.Modulation = pb.Modulation_LORA
		}

		for _, sf := range ec.SpreadingFactors {
			c.SpreadingFactors = append(c.SpreadingFactors, uint32(sf))
		}

		out.GatewayProfile.ExtraChannels = append(out.GatewayProfile.ExtraChannels, &c)
	}

	return &out, nil
}

// Update updates the given gateway-profile.
func (a *GatewayProfileAPI) Update(ctx context.Context, req *pb.UpdateGatewayProfileRequest) (*pb.UpdateGatewayProfileResponse, error) {
	if err := a.validator.Validate(ctx,
		auth.ValidateGatewayProfileAccess(auth.Update),
	); err != nil {
		return nil, grpc.Errorf(codes.Unauthenticated, "authentication failed: %s", err)
	}

	if req.GatewayProfile == nil {
		return nil, grpc.Errorf(codes.InvalidArgument, "empty gateway-profile")
	}

	gp, err := storage.GetGatewayProfile(config.C.PostgreSQL.DB, req.GatewayProfile.GatewayProfileID)
	if err != nil {
		return nil, errToRPCError(err)
	}

	gp.Name = req.Name
	gp.Channels = []int{}
	gp.ExtraChannels = []storage.ExtraChannel{}

	for _, c := range req.GatewayProfile.Channels {
		gp.Channels = append(gp.Channels, int(c))
	}

	for _, ec := range req.GatewayProfile.ExtraChannels {
		c := storage.ExtraChannel{
			Frequency: int(ec.Frequency),
			Bandwidth: int(ec.Bandwidth),
			Bitrate:   int(ec.Bitrate),
		}

		switch ec.Modulation {
		case pb.Modulation_FSK:
			c.Modulation = storage.ModulationFSK
		default:
			c.Modulation = storage.ModulationLoRa
		}

		for _, sf := range ec.SpreadingFactors {
			c.SpreadingFactors = append(c.SpreadingFactors, int(sf))
		}

		gp.ExtraChannels = append(gp.ExtraChannels, c)
	}

	err = storage.Transaction(config.C.PostgreSQL.DB, func(tx sqlx.Ext) error {
		return storage.UpdateGatewayProfile(tx, &gp)
	})
	if err != nil {
		return nil, errToRPCError(err)
	}

	return &pb.UpdateGatewayProfileResponse{}, nil
}

// Delete deletes the gateway-profile matching the given id.
func (a *GatewayProfileAPI) Delete(ctx context.Context, req *pb.DeleteGatewayProfileRequest) (*pb.DeleteGatewayProfileResponse, error) {
	if err := a.validator.Validate(ctx,
		auth.ValidateGatewayProfileAccess(auth.Delete),
	); err != nil {
		return nil, grpc.Errorf(codes.Unauthenticated, "authentication failed: %s", err)
	}

	err := storage.DeleteGatewayProfile(config.C.PostgreSQL.DB, req.GatewayProfileID)
	if err != nil {
		return nil, errToRPCError(err)
	}

	return &pb.DeleteGatewayProfileResponse{}, nil
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

	if req.NetworkServerID == 0 {
		count, err = storage.GetGatewayProfileCount(config.C.PostgreSQL.DB)
		if err != nil {
			return nil, errToRPCError(err)
		}

		gps, err = storage.GetGatewayProfiles(config.C.PostgreSQL.DB, int(req.Limit), int(req.Offset))
		if err != nil {
			return nil, errToRPCError(err)
		}
	} else {
		count, err = storage.GetGatewayProfileCountForNetworkServerID(config.C.PostgreSQL.DB, req.NetworkServerID)
		if err != nil {
			return nil, errToRPCError(err)
		}

		gps, err = storage.GetGatewayProfilesForNetworkServerID(config.C.PostgreSQL.DB, req.NetworkServerID, int(req.Limit), int(req.Offset))
		if err != nil {
			return nil, errToRPCError(err)
		}
	}

	out := pb.ListGatewayProfilesResponse{
		TotalCount: int64(count),
	}

	for _, gp := range gps {
		out.Result = append(out.Result, &pb.GatewayProfileMeta{
			GatewayProfileID: gp.GatewayProfileID,
			Name:             gp.Name,
			NetworkServerID:  gp.NetworkServerID,
			CreatedAt:        gp.CreatedAt.Format(time.RFC3339Nano),
			UpdatedAt:        gp.UpdatedAt.Format(time.RFC3339Nano),
		})
	}

	return &out, nil
}
