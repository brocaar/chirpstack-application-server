package api

import (
	"github.com/golang/protobuf/ptypes"
	"github.com/gofrs/uuid"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/jmoiron/sqlx"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

	pb "github.com/brocaar/lora-app-server/api"
	"github.com/brocaar/lora-app-server/internal/api/auth"
	"github.com/brocaar/lora-app-server/internal/config"
	"github.com/brocaar/lora-app-server/internal/storage"
	"github.com/brocaar/loraserver/api/ns"
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
		return nil, grpc.Errorf(codes.InvalidArgument, "service_profile must not be nil")
	}

	if err := a.validator.Validate(ctx,
		auth.ValidateServiceProfilesAccess(auth.Create, req.ServiceProfile.OrganizationId),
	); err != nil {
		return nil, grpc.Errorf(codes.Unauthenticated, "authentication failed: %s", err)
	}

	sp := storage.ServiceProfile{
		OrganizationID:  req.ServiceProfile.OrganizationId,
		NetworkServerID: req.ServiceProfile.NetworkServerId,
		Name:            req.ServiceProfile.Name,
		ServiceProfile: ns.ServiceProfile{
			UlRate:                 req.ServiceProfile.UlRate,
			UlBucketSize:           req.ServiceProfile.UlBucketSize,
			DlRate:                 req.ServiceProfile.DlRate,
			DlBucketSize:           req.ServiceProfile.DlBucketSize,
			AddGwMetadata:          req.ServiceProfile.AddGwMetadata,
			DevStatusReqFreq:       req.ServiceProfile.DevStatusReqFreq,
			ReportDevStatusBattery: req.ServiceProfile.ReportDevStatusBattery,
			ReportDevStatusMargin:  req.ServiceProfile.ReportDevStatusMargin,
			DrMin:          req.ServiceProfile.DrMin,
			DrMax:          req.ServiceProfile.DrMax,
			ChannelMask:    req.ServiceProfile.ChannelMask,
			PrAllowed:      req.ServiceProfile.PrAllowed,
			HrAllowed:      req.ServiceProfile.HrAllowed,
			RaAllowed:      req.ServiceProfile.RaAllowed,
			NwkGeoLoc:      req.ServiceProfile.NwkGeoLoc,
			TargetPer:      req.ServiceProfile.TargetPer,
			MinGwDiversity: req.ServiceProfile.MinGwDiversity,
			UlRatePolicy:   ns.RatePolicy(req.ServiceProfile.UlRatePolicy),
			DlRatePolicy:   ns.RatePolicy(req.ServiceProfile.DlRatePolicy),
		},
	}

	// as this also performs a remote call to create the service-profile
	// on the network-server, wrap it in a transaction
	err := storage.Transaction(config.C.PostgreSQL.DB, func(tx sqlx.Ext) error {
		return storage.CreateServiceProfile(tx, &sp)
	})
	if err != nil {
		return nil, errToRPCError(err)
	}

	spID, err := uuid.FromBytes(sp.ServiceProfile.Id)
	if err != nil {
		return nil, errToRPCError(err)
	}

	return &pb.CreateServiceProfileResponse{
		Id: spID.String(),
	}, nil
}

// Get returns the service-profile matching the given id.
func (a *ServiceProfileServiceAPI) Get(ctx context.Context, req *pb.GetServiceProfileRequest) (*pb.GetServiceProfileResponse, error) {
	spID, err := uuid.FromString(req.Id)
	if err != nil {
		return nil, grpc.Errorf(codes.InvalidArgument, "uuid error: %s", err)
	}

	if err := a.validator.Validate(ctx,
		auth.ValidateServiceProfileAccess(auth.Read, spID),
	); err != nil {
		return nil, grpc.Errorf(codes.Unauthenticated, "authentication failed: %s", err)
	}

	sp, err := storage.GetServiceProfile(config.C.PostgreSQL.DB, spID)
	if err != nil {
		return nil, errToRPCError(err)
	}

	resp := pb.GetServiceProfileResponse{
		ServiceProfile: &pb.ServiceProfile{
			Id:                     spID.String(),
			Name:                   sp.Name,
			OrganizationId:         sp.OrganizationID,
			NetworkServerId:        sp.NetworkServerID,
			UlRate:                 sp.ServiceProfile.UlRate,
			UlBucketSize:           sp.ServiceProfile.UlBucketSize,
			DlRate:                 sp.ServiceProfile.DlRate,
			DlBucketSize:           sp.ServiceProfile.DlBucketSize,
			AddGwMetadata:          sp.ServiceProfile.AddGwMetadata,
			DevStatusReqFreq:       sp.ServiceProfile.DevStatusReqFreq,
			ReportDevStatusBattery: sp.ServiceProfile.ReportDevStatusBattery,
			ReportDevStatusMargin:  sp.ServiceProfile.ReportDevStatusMargin,
			DrMin:          sp.ServiceProfile.DrMin,
			DrMax:          sp.ServiceProfile.DrMax,
			ChannelMask:    sp.ServiceProfile.ChannelMask,
			PrAllowed:      sp.ServiceProfile.PrAllowed,
			HrAllowed:      sp.ServiceProfile.HrAllowed,
			RaAllowed:      sp.ServiceProfile.RaAllowed,
			NwkGeoLoc:      sp.ServiceProfile.NwkGeoLoc,
			TargetPer:      sp.ServiceProfile.TargetPer,
			MinGwDiversity: sp.ServiceProfile.MinGwDiversity,
			UlRatePolicy:   pb.RatePolicy(sp.ServiceProfile.UlRatePolicy),
			DlRatePolicy:   pb.RatePolicy(sp.ServiceProfile.DlRatePolicy),
		},
	}

	resp.CreatedAt, err = ptypes.TimestampProto(sp.CreatedAt)
	if err != nil {
		return nil, errToRPCError(err)
	}
	resp.UpdatedAt, err = ptypes.TimestampProto(sp.UpdatedAt)
	if err != nil {
		return nil, errToRPCError(err)
	}

	return &resp, nil
}

// Update updates the given serviceprofile.
func (a *ServiceProfileServiceAPI) Update(ctx context.Context, req *pb.UpdateServiceProfileRequest) (*empty.Empty, error) {
	if req.ServiceProfile == nil {
		return nil, grpc.Errorf(codes.InvalidArgument, "service_profile must not be nil")
	}

	spID, err := uuid.FromString(req.ServiceProfile.Id)
	if err != nil {
		return nil, grpc.Errorf(codes.InvalidArgument, "uuid error: %s", err)
	}

	if err := a.validator.Validate(ctx,
		auth.ValidateServiceProfileAccess(auth.Update, spID),
	); err != nil {
		return nil, grpc.Errorf(codes.Unauthenticated, "authentication failed: %s", err)
	}

	sp, err := storage.GetServiceProfile(config.C.PostgreSQL.DB, spID)
	if err != nil {
		return nil, errToRPCError(err)
	}

	sp.Name = req.ServiceProfile.Name
	sp.ServiceProfile = ns.ServiceProfile{
		Id:                     spID.Bytes(),
		UlRate:                 req.ServiceProfile.UlRate,
		UlBucketSize:           req.ServiceProfile.UlBucketSize,
		DlRate:                 req.ServiceProfile.DlRate,
		DlBucketSize:           req.ServiceProfile.DlBucketSize,
		AddGwMetadata:          req.ServiceProfile.AddGwMetadata,
		DevStatusReqFreq:       req.ServiceProfile.DevStatusReqFreq,
		ReportDevStatusBattery: req.ServiceProfile.ReportDevStatusBattery,
		ReportDevStatusMargin:  req.ServiceProfile.ReportDevStatusMargin,
		DrMin:          req.ServiceProfile.DrMin,
		DrMax:          req.ServiceProfile.DrMax,
		ChannelMask:    req.ServiceProfile.ChannelMask,
		PrAllowed:      req.ServiceProfile.PrAllowed,
		HrAllowed:      req.ServiceProfile.HrAllowed,
		RaAllowed:      req.ServiceProfile.RaAllowed,
		NwkGeoLoc:      req.ServiceProfile.NwkGeoLoc,
		TargetPer:      req.ServiceProfile.TargetPer,
		MinGwDiversity: req.ServiceProfile.MinGwDiversity,
		UlRatePolicy:   ns.RatePolicy(req.ServiceProfile.UlRatePolicy),
		DlRatePolicy:   ns.RatePolicy(req.ServiceProfile.DlRatePolicy),
	}

	// as this also performs a remote call to create the service-profile
	// on the network-server, wrap it in a transaction
	err = storage.Transaction(config.C.PostgreSQL.DB, func(tx sqlx.Ext) error {
		return storage.UpdateServiceProfile(tx, &sp)
	})
	if err != nil {
		return nil, errToRPCError(err)
	}

	return &empty.Empty{}, nil
}

// Delete deletes the service-profile matching the given id.
func (a *ServiceProfileServiceAPI) Delete(ctx context.Context, req *pb.DeleteServiceProfileRequest) (*empty.Empty, error) {
	spID, err := uuid.FromString(req.Id)
	if err != nil {
		return nil, grpc.Errorf(codes.InvalidArgument, "uuid error: %s", err)
	}

	if err := a.validator.Validate(ctx,
		auth.ValidateServiceProfileAccess(auth.Delete, spID),
	); err != nil {
		return nil, grpc.Errorf(codes.Unauthenticated, "authentication failed: %s", err)
	}

	err = storage.Transaction(config.C.PostgreSQL.DB, func(tx sqlx.Ext) error {
		return storage.DeleteServiceProfile(tx, spID)
	})
	if err != nil {
		return nil, errToRPCError(err)
	}

	return &empty.Empty{}, nil
}

// List lists the available service-profiles.
func (a *ServiceProfileServiceAPI) List(ctx context.Context, req *pb.ListServiceProfileRequest) (*pb.ListServiceProfileResponse, error) {
	if err := a.validator.Validate(ctx,
		auth.ValidateServiceProfilesAccess(auth.List, req.OrganizationId),
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

	if req.OrganizationId == 0 {
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
		sps, err = storage.GetServiceProfilesForOrganizationID(config.C.PostgreSQL.DB, req.OrganizationId, int(req.Limit), int(req.Offset))
		if err != nil {
			return nil, errToRPCError(err)
		}

		count, err = storage.GetServiceProfileCountForOrganizationID(config.C.PostgreSQL.DB, req.OrganizationId)
		if err != nil {
			return nil, errToRPCError(err)
		}
	}

	resp := pb.ListServiceProfileResponse{
		TotalCount: int64(count),
	}
	for _, sp := range sps {
		row := pb.ServiceProfileListItem{
			Id:              sp.ServiceProfileID.String(),
			Name:            sp.Name,
			OrganizationId:  sp.OrganizationID,
			NetworkServerId: sp.NetworkServerID,
		}

		row.CreatedAt, err = ptypes.TimestampProto(sp.CreatedAt)
		if err != nil {
			return nil, errToRPCError(err)
		}
		row.UpdatedAt, err = ptypes.TimestampProto(sp.UpdatedAt)
		if err != nil {
			return nil, errToRPCError(err)
		}

		resp.Result = append(resp.Result, &row)
	}

	return &resp, nil
}
