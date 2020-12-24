package external

import (
	"github.com/gofrs/uuid"
	"github.com/golang/protobuf/ptypes"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/jmoiron/sqlx"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

	pb "github.com/brocaar/chirpstack-api/go/v3/as/external/api"
	"github.com/brocaar/chirpstack-api/go/v3/ns"
	"github.com/brocaar/chirpstack-application-server/internal/api/external/auth"
	"github.com/brocaar/chirpstack-application-server/internal/api/helpers"
	"github.com/brocaar/chirpstack-application-server/internal/storage"
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
			DrMin:                  req.ServiceProfile.DrMin,
			DrMax:                  req.ServiceProfile.DrMax,
			ChannelMask:            req.ServiceProfile.ChannelMask,
			PrAllowed:              req.ServiceProfile.PrAllowed,
			HrAllowed:              req.ServiceProfile.HrAllowed,
			RaAllowed:              req.ServiceProfile.RaAllowed,
			NwkGeoLoc:              req.ServiceProfile.NwkGeoLoc,
			TargetPer:              req.ServiceProfile.TargetPer,
			MinGwDiversity:         req.ServiceProfile.MinGwDiversity,
			UlRatePolicy:           ns.RatePolicy(req.ServiceProfile.UlRatePolicy),
			DlRatePolicy:           ns.RatePolicy(req.ServiceProfile.DlRatePolicy),
			GwsPrivate:             req.ServiceProfile.GwsPrivate,
		},
	}

	// as this also performs a remote call to create the service-profile
	// on the network-server, wrap it in a transaction
	err := storage.Transaction(func(tx sqlx.Ext) error {
		return storage.CreateServiceProfile(ctx, tx, &sp)
	})
	if err != nil {
		return nil, helpers.ErrToRPCError(err)
	}

	spID, err := uuid.FromBytes(sp.ServiceProfile.Id)
	if err != nil {
		return nil, helpers.ErrToRPCError(err)
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

	sp, err := storage.GetServiceProfile(ctx, storage.DB(), spID, false)
	if err != nil {
		return nil, helpers.ErrToRPCError(err)
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
			DrMin:                  sp.ServiceProfile.DrMin,
			DrMax:                  sp.ServiceProfile.DrMax,
			ChannelMask:            sp.ServiceProfile.ChannelMask,
			PrAllowed:              sp.ServiceProfile.PrAllowed,
			HrAllowed:              sp.ServiceProfile.HrAllowed,
			RaAllowed:              sp.ServiceProfile.RaAllowed,
			NwkGeoLoc:              sp.ServiceProfile.NwkGeoLoc,
			TargetPer:              sp.ServiceProfile.TargetPer,
			MinGwDiversity:         sp.ServiceProfile.MinGwDiversity,
			UlRatePolicy:           pb.RatePolicy(sp.ServiceProfile.UlRatePolicy),
			DlRatePolicy:           pb.RatePolicy(sp.ServiceProfile.DlRatePolicy),
			GwsPrivate:             sp.ServiceProfile.GwsPrivate,
		},
	}

	resp.CreatedAt, err = ptypes.TimestampProto(sp.CreatedAt)
	if err != nil {
		return nil, helpers.ErrToRPCError(err)
	}
	resp.UpdatedAt, err = ptypes.TimestampProto(sp.UpdatedAt)
	if err != nil {
		return nil, helpers.ErrToRPCError(err)
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

	sp, err := storage.GetServiceProfile(ctx, storage.DB(), spID, false)
	if err != nil {
		return nil, helpers.ErrToRPCError(err)
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
		DrMin:                  req.ServiceProfile.DrMin,
		DrMax:                  req.ServiceProfile.DrMax,
		ChannelMask:            req.ServiceProfile.ChannelMask,
		PrAllowed:              req.ServiceProfile.PrAllowed,
		HrAllowed:              req.ServiceProfile.HrAllowed,
		RaAllowed:              req.ServiceProfile.RaAllowed,
		NwkGeoLoc:              req.ServiceProfile.NwkGeoLoc,
		TargetPer:              req.ServiceProfile.TargetPer,
		MinGwDiversity:         req.ServiceProfile.MinGwDiversity,
		UlRatePolicy:           ns.RatePolicy(req.ServiceProfile.UlRatePolicy),
		DlRatePolicy:           ns.RatePolicy(req.ServiceProfile.DlRatePolicy),
		GwsPrivate:             req.ServiceProfile.GwsPrivate,
	}

	// as this also performs a remote call to create the service-profile
	// on the network-server, wrap it in a transaction
	err = storage.Transaction(func(tx sqlx.Ext) error {
		return storage.UpdateServiceProfile(ctx, tx, &sp)
	})
	if err != nil {
		return nil, helpers.ErrToRPCError(err)
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

	err = storage.Transaction(func(tx sqlx.Ext) error {
		return storage.DeleteServiceProfile(ctx, tx, spID)
	})
	if err != nil {
		return nil, helpers.ErrToRPCError(err)
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

	filters := storage.ServiceProfileFilters{
		Limit:           int(req.Limit),
		Offset:          int(req.Offset),
		OrganizationID:  req.OrganizationId,
		NetworkServerID: req.NetworkServerId,
	}

	sub, err := a.validator.GetSubject(ctx)
	if err != nil {
		return nil, helpers.ErrToRPCError(err)
	}

	switch sub {
	case auth.SubjectUser:
		user, err := a.validator.GetUser(ctx)
		if err != nil {
			return nil, helpers.ErrToRPCError(err)
		}

		// Filter on user ID when OrganizationID is not set and the user is
		// not a global admin.
		if !user.IsAdmin && filters.OrganizationID == 0 {
			filters.UserID = user.ID
		}

	case auth.SubjectAPIKey:
		// Nothing to do as the validator function already validated that the
		// API Key has access to the given OrganizationID.
	default:
		return nil, grpc.Errorf(codes.Unauthenticated, "invalid token subject: %s", sub)
	}

	sps, err := storage.GetServiceProfiles(ctx, storage.DB(), filters)
	if err != nil {
		return nil, helpers.ErrToRPCError(err)
	}

	count, err := storage.GetServiceProfileCount(ctx, storage.DB(), filters)
	if err != nil {
		return nil, helpers.ErrToRPCError(err)
	}

	resp := pb.ListServiceProfileResponse{
		TotalCount: int64(count),
	}
	for _, sp := range sps {
		row := pb.ServiceProfileListItem{
			Id:                sp.ServiceProfileID.String(),
			Name:              sp.Name,
			OrganizationId:    sp.OrganizationID,
			NetworkServerId:   sp.NetworkServerID,
			NetworkServerName: sp.NetworkServerName,
		}

		row.CreatedAt, err = ptypes.TimestampProto(sp.CreatedAt)
		if err != nil {
			return nil, helpers.ErrToRPCError(err)
		}
		row.UpdatedAt, err = ptypes.TimestampProto(sp.UpdatedAt)
		if err != nil {
			return nil, helpers.ErrToRPCError(err)
		}

		resp.Result = append(resp.Result, &row)
	}

	return &resp, nil
}
