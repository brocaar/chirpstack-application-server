package api

import (
	"time"

	"github.com/jmoiron/sqlx"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

	pb "github.com/gusseleet/lora-app-server/api"
	"github.com/gusseleet/lora-app-server/internal/api/auth"
	"github.com/gusseleet/lora-app-server/internal/config"
	"github.com/gusseleet/lora-app-server/internal/storage"
	"github.com/brocaar/lorawan/backend"
)

// ServiceProfileServiceAPI export the ServiceProfile related functions.
type ServiceProfileServiceAPI struct {
	validator auth.Validator
}

// NewServiceProfileServiceAPI creates a new ServiceProfileServiceAPI.
func NewServiceProfileServiceAPI(validator auth.Validator) *ServiceProfileServiceAPI {
	return &ServiceProfileServiceAPI{
		validator: validator,
	}
}

// Create creates the given service-profile.
func (a *ServiceProfileServiceAPI) Create(ctx context.Context, req *pb.CreateServiceProfileRequest) (*pb.CreateServiceProfileResponse, error) {
	if req.ServiceProfile == nil {
		return nil, grpc.Errorf(codes.InvalidArgument, "serviceProfile expected")
	}

	if err := a.validator.Validate(ctx,
		auth.ValidateServiceProfilesAccess(auth.Create, req.OrganizationID),
	); err != nil {
		return nil, grpc.Errorf(codes.Unauthenticated, "authentication failed: %s", err)
	}

	sp := storage.ServiceProfile{
		OrganizationID:  req.OrganizationID,
		NetworkServerID: req.NetworkServerID,
		Name:            req.Name,
		ServiceProfile: backend.ServiceProfile{
			ULRate:                 int(req.ServiceProfile.UlRate),
			ULBucketSize:           int(req.ServiceProfile.UlBucketSize),
			DLRate:                 int(req.ServiceProfile.DlRate),
			DLBucketSize:           int(req.ServiceProfile.DlBucketSize),
			AddGWMetadata:          req.ServiceProfile.AddGWMetadata,
			DevStatusReqFreq:       int(req.ServiceProfile.DevStatusReqFreq),
			ReportDevStatusBattery: req.ServiceProfile.ReportDevStatusBattery,
			ReportDevStatusMargin:  req.ServiceProfile.ReportDevStatusMargin,
			DRMin:          int(req.ServiceProfile.DrMin),
			DRMax:          int(req.ServiceProfile.DrMax),
			ChannelMask:    backend.HEXBytes(req.ServiceProfile.ChannelMask),
			PRAllowed:      req.ServiceProfile.PrAllowed,
			HRAllowed:      req.ServiceProfile.HrAllowed,
			RAAllowed:      req.ServiceProfile.RaAllowed,
			NwkGeoLoc:      req.ServiceProfile.NwkGeoLoc,
			TargetPER:      backend.Percentage(req.ServiceProfile.TargetPER),
			MinGWDiversity: int(req.ServiceProfile.MinGWDiversity),
		},
	}

	switch req.ServiceProfile.UlRatePolicy {
	case pb.RatePolicy_MARK:
		sp.ServiceProfile.ULRatePolicy = backend.Mark
	case pb.RatePolicy_DROP:
		sp.ServiceProfile.ULRatePolicy = backend.Drop
	}

	switch req.ServiceProfile.DlRatePolicy {
	case pb.RatePolicy_MARK:
		sp.ServiceProfile.DLRatePolicy = backend.Mark
	case pb.RatePolicy_DROP:
		sp.ServiceProfile.DLRatePolicy = backend.Drop
	}

	// as this also performs a remote call to create the service-profile
	// on the network-server, wrap it in a transaction
	err := storage.Transaction(config.C.PostgreSQL.DB, func(tx sqlx.Ext) error {
		return storage.CreateServiceProfile(tx, &sp)
	})
	if err != nil {
		return nil, errToRPCError(err)
	}

	return &pb.CreateServiceProfileResponse{
		ServiceProfileID: sp.ServiceProfile.ServiceProfileID,
	}, nil
}

// Get returns the service-profile matching the given id.
func (a *ServiceProfileServiceAPI) Get(ctx context.Context, req *pb.GetServiceProfileRequest) (*pb.GetServiceProfileResponse, error) {
	if err := a.validator.Validate(ctx,
		auth.ValidateServiceProfileAccess(auth.Read, req.ServiceProfileID),
	); err != nil {
		return nil, grpc.Errorf(codes.Unauthenticated, "authentication failed: %s", err)
	}

	sp, err := storage.GetServiceProfile(config.C.PostgreSQL.DB, req.ServiceProfileID)
	if err != nil {
		return nil, errToRPCError(err)
	}

	resp := pb.GetServiceProfileResponse{
		CreatedAt:       sp.CreatedAt.Format(time.RFC3339Nano),
		UpdatedAt:       sp.UpdatedAt.Format(time.RFC3339Nano),
		OrganizationID:  sp.OrganizationID,
		NetworkServerID: sp.NetworkServerID,
		Name:            sp.Name,
		ServiceProfile: &pb.ServiceProfile{
			ServiceProfileID:       sp.ServiceProfile.ServiceProfileID,
			UlRate:                 uint32(sp.ServiceProfile.ULRate),
			UlBucketSize:           uint32(sp.ServiceProfile.ULBucketSize),
			DlRate:                 uint32(sp.ServiceProfile.DLRate),
			DlBucketSize:           uint32(sp.ServiceProfile.DLBucketSize),
			AddGWMetadata:          sp.ServiceProfile.AddGWMetadata,
			DevStatusReqFreq:       uint32(sp.ServiceProfile.DevStatusReqFreq),
			ReportDevStatusBattery: sp.ServiceProfile.ReportDevStatusBattery,
			ReportDevStatusMargin:  sp.ServiceProfile.ReportDevStatusMargin,
			DrMin:          uint32(sp.ServiceProfile.DRMin),
			DrMax:          uint32(sp.ServiceProfile.DRMax),
			ChannelMask:    []byte(sp.ServiceProfile.ChannelMask),
			PrAllowed:      sp.ServiceProfile.PRAllowed,
			HrAllowed:      sp.ServiceProfile.HRAllowed,
			RaAllowed:      sp.ServiceProfile.RAAllowed,
			NwkGeoLoc:      sp.ServiceProfile.NwkGeoLoc,
			TargetPER:      uint32(sp.ServiceProfile.TargetPER),
			MinGWDiversity: uint32(sp.ServiceProfile.MinGWDiversity),
		},
	}

	switch sp.ServiceProfile.ULRatePolicy {
	case backend.Mark:
		resp.ServiceProfile.UlRatePolicy = pb.RatePolicy_MARK
	case backend.Drop:
		resp.ServiceProfile.UlRatePolicy = pb.RatePolicy_DROP
	}

	switch sp.ServiceProfile.DLRatePolicy {
	case backend.Mark:
		resp.ServiceProfile.DlRatePolicy = pb.RatePolicy_MARK
	case backend.Drop:
		resp.ServiceProfile.DlRatePolicy = pb.RatePolicy_DROP
	}

	return &resp, nil
}

// Update updates the given serviceprofile.
func (a *ServiceProfileServiceAPI) Update(ctx context.Context, req *pb.UpdateServiceProfileRequest) (*pb.UpdateServiceProfileResponse, error) {
	if req.ServiceProfile == nil {
		return nil, grpc.Errorf(codes.InvalidArgument, "serviceProfile expected")
	}

	if err := a.validator.Validate(ctx,
		auth.ValidateServiceProfileAccess(auth.Update, req.ServiceProfile.ServiceProfileID),
	); err != nil {
		return nil, grpc.Errorf(codes.Unauthenticated, "authentication failed: %s", err)
	}

	sp, err := storage.GetServiceProfile(config.C.PostgreSQL.DB, req.ServiceProfile.ServiceProfileID)
	if err != nil {
		return nil, errToRPCError(err)
	}

	sp.Name = req.Name
	sp.ServiceProfile = backend.ServiceProfile{
		ServiceProfileID:       req.ServiceProfile.ServiceProfileID,
		ULRate:                 int(req.ServiceProfile.UlRate),
		ULBucketSize:           int(req.ServiceProfile.UlBucketSize),
		DLRate:                 int(req.ServiceProfile.DlRate),
		DLBucketSize:           int(req.ServiceProfile.DlBucketSize),
		AddGWMetadata:          req.ServiceProfile.AddGWMetadata,
		DevStatusReqFreq:       int(req.ServiceProfile.DevStatusReqFreq),
		ReportDevStatusBattery: req.ServiceProfile.ReportDevStatusBattery,
		ReportDevStatusMargin:  req.ServiceProfile.ReportDevStatusMargin,
		DRMin:          int(req.ServiceProfile.DrMin),
		DRMax:          int(req.ServiceProfile.DrMax),
		ChannelMask:    backend.HEXBytes(req.ServiceProfile.ChannelMask),
		PRAllowed:      req.ServiceProfile.PrAllowed,
		HRAllowed:      req.ServiceProfile.HrAllowed,
		RAAllowed:      req.ServiceProfile.RaAllowed,
		NwkGeoLoc:      req.ServiceProfile.NwkGeoLoc,
		TargetPER:      backend.Percentage(req.ServiceProfile.TargetPER),
		MinGWDiversity: int(req.ServiceProfile.MinGWDiversity),
	}

	switch req.ServiceProfile.UlRatePolicy {
	case pb.RatePolicy_MARK:
		sp.ServiceProfile.ULRatePolicy = backend.Mark
	case pb.RatePolicy_DROP:
		sp.ServiceProfile.ULRatePolicy = backend.Drop
	}

	switch req.ServiceProfile.DlRatePolicy {
	case pb.RatePolicy_MARK:
		sp.ServiceProfile.DLRatePolicy = backend.Mark
	case pb.RatePolicy_DROP:
		sp.ServiceProfile.DLRatePolicy = backend.Drop
	}

	// as this also performs a remote call to create the service-profile
	// on the network-server, wrap it in a transaction
	err = storage.Transaction(config.C.PostgreSQL.DB, func(tx sqlx.Ext) error {
		return storage.UpdateServiceProfile(tx, &sp)
	})
	if err != nil {
		return nil, errToRPCError(err)
	}

	return &pb.UpdateServiceProfileResponse{}, nil
}

// Delete deletes the service-profile matching the given id.
func (a *ServiceProfileServiceAPI) Delete(ctx context.Context, req *pb.DeleteServiceProfileRequest) (*pb.DeleteServiceProfileResponse, error) {
	if err := a.validator.Validate(ctx,
		auth.ValidateServiceProfileAccess(auth.Delete, req.ServiceProfileID),
	); err != nil {
		return nil, grpc.Errorf(codes.Unauthenticated, "authentication failed: %s", err)
	}

	err := storage.Transaction(config.C.PostgreSQL.DB, func(tx sqlx.Ext) error {
		return storage.DeleteServiceProfile(tx, req.ServiceProfileID)
	})
	if err != nil {
		return nil, errToRPCError(err)
	}

	return &pb.DeleteServiceProfileResponse{}, nil
}

// List lists the available service-profiles.
func (a *ServiceProfileServiceAPI) List(ctx context.Context, req *pb.ListServiceProfileRequest) (*pb.ListServiceProfileResponse, error) {
	if err := a.validator.Validate(ctx,
		auth.ValidateServiceProfilesAccess(auth.List, req.OrganizationID),
	); err != nil {
		return nil, grpc.Errorf(codes.Unauthenticated, "authentication failed: %s", err)
	}

	isAdmin, err := a.validator.GetIsAdmin(ctx)
	if err != nil {
		return nil, errToRPCError(err)
	}

	username, err := a.validator.GetUsername(ctx)
	if err != nil {
		return nil, errToRPCError(err)
	}

	var count int
	var sps []storage.ServiceProfileMeta

	if req.OrganizationID == 0 {
		if isAdmin {
			sps, err = storage.GetServiceProfiles(config.C.PostgreSQL.DB, int(req.Limit), int(req.Offset))
			if err != nil {
				return nil, errToRPCError(err)
			}

			count, err = storage.GetServiceProfileCount(config.C.PostgreSQL.DB)
			if err != nil {
				return nil, errToRPCError(err)
			}
		} else {
			sps, err = storage.GetServiceProfilesForUser(config.C.PostgreSQL.DB, username, int(req.Limit), int(req.Offset))
			if err != nil {
				return nil, errToRPCError(err)
			}

			count, err = storage.GetServiceProfileCountForUser(config.C.PostgreSQL.DB, username)
			if err != nil {
				return nil, errToRPCError(err)
			}
		}
	} else {
		sps, err = storage.GetServiceProfilesForOrganizationID(config.C.PostgreSQL.DB, req.OrganizationID, int(req.Limit), int(req.Offset))
		if err != nil {
			return nil, errToRPCError(err)
		}

		count, err = storage.GetServiceProfileCountForOrganizationID(config.C.PostgreSQL.DB, req.OrganizationID)
		if err != nil {
			return nil, errToRPCError(err)
		}
	}

	resp := pb.ListServiceProfileResponse{
		TotalCount: int64(count),
	}
	for _, sp := range sps {
		resp.Result = append(resp.Result, &pb.ServiceProfileMeta{
			ServiceProfileID: sp.ServiceProfileID,
			Name:             sp.Name,
			OrganizationID:   sp.OrganizationID,
			NetworkServerID:  sp.NetworkServerID,
			CreatedAt:        sp.CreatedAt.Format(time.RFC3339Nano),
			UpdatedAt:        sp.UpdatedAt.Format(time.RFC3339Nano),
		})
	}

	return &resp, nil
}
