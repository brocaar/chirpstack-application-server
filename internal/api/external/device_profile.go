package external

import (
	"database/sql"
	"time"

	"github.com/gofrs/uuid"
	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/lib/pq/hstore"

	"github.com/brocaar/chirpstack-api/go/v3/ns"

	"github.com/jmoiron/sqlx"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

	pb "github.com/brocaar/chirpstack-api/go/v3/as/external/api"
	"github.com/brocaar/chirpstack-application-server/internal/api/external/auth"
	"github.com/brocaar/chirpstack-application-server/internal/api/helpers"
	"github.com/brocaar/chirpstack-application-server/internal/codec"
	"github.com/brocaar/chirpstack-application-server/internal/storage"
)

// DeviceProfileServiceAPI exports the ServiceProfile related functions.
type DeviceProfileServiceAPI struct {
	validator auth.Validator
}

// NewDeviceProfileServiceAPI creates a new DeviceProfileServiceAPI.
func NewDeviceProfileServiceAPI(validator auth.Validator) *DeviceProfileServiceAPI {
	return &DeviceProfileServiceAPI{
		validator: validator,
	}
}

// Create creates the given device-profile.
func (a *DeviceProfileServiceAPI) Create(ctx context.Context, req *pb.CreateDeviceProfileRequest) (*pb.CreateDeviceProfileResponse, error) {
	if req.DeviceProfile == nil {
		return nil, grpc.Errorf(codes.InvalidArgument, "deviceProfile expected")
	}

	if err := a.validator.Validate(ctx,
		auth.ValidateDeviceProfilesAccess(auth.Create, req.DeviceProfile.OrganizationId, 0),
	); err != nil {
		return nil, grpc.Errorf(codes.Unauthenticated, "authentication failed: %s", err)
	}

	var err error
	var uplinkInterval time.Duration
	if req.DeviceProfile.UplinkInterval != nil {
		uplinkInterval, err = ptypes.Duration(req.DeviceProfile.UplinkInterval)
		if err != nil {
			return nil, helpers.ErrToRPCError(err)
		}
	}

	dp := storage.DeviceProfile{
		OrganizationID:       req.DeviceProfile.OrganizationId,
		NetworkServerID:      req.DeviceProfile.NetworkServerId,
		Name:                 req.DeviceProfile.Name,
		PayloadCodec:         codec.Type(req.DeviceProfile.PayloadCodec),
		PayloadEncoderScript: req.DeviceProfile.PayloadEncoderScript,
		PayloadDecoderScript: req.DeviceProfile.PayloadDecoderScript,
		Tags: hstore.Hstore{
			Map: make(map[string]sql.NullString),
		},
		UplinkInterval: uplinkInterval,
		DeviceProfile: ns.DeviceProfile{
			SupportsClassB:     req.DeviceProfile.SupportsClassB,
			ClassBTimeout:      req.DeviceProfile.ClassBTimeout,
			PingSlotPeriod:     req.DeviceProfile.PingSlotPeriod,
			PingSlotDr:         req.DeviceProfile.PingSlotDr,
			PingSlotFreq:       req.DeviceProfile.PingSlotFreq,
			SupportsClassC:     req.DeviceProfile.SupportsClassC,
			ClassCTimeout:      req.DeviceProfile.ClassCTimeout,
			MacVersion:         req.DeviceProfile.MacVersion,
			RegParamsRevision:  req.DeviceProfile.RegParamsRevision,
			RxDelay_1:          req.DeviceProfile.RxDelay_1,
			RxDrOffset_1:       req.DeviceProfile.RxDrOffset_1,
			RxDatarate_2:       req.DeviceProfile.RxDatarate_2,
			RxFreq_2:           req.DeviceProfile.RxFreq_2,
			MaxEirp:            req.DeviceProfile.MaxEirp,
			MaxDutyCycle:       req.DeviceProfile.MaxDutyCycle,
			SupportsJoin:       req.DeviceProfile.SupportsJoin,
			RfRegion:           req.DeviceProfile.RfRegion,
			Supports_32BitFCnt: req.DeviceProfile.Supports_32BitFCnt,
			FactoryPresetFreqs: req.DeviceProfile.FactoryPresetFreqs,
			AdrAlgorithmId:     req.DeviceProfile.AdrAlgorithmId,
		},
	}

	for k, v := range req.DeviceProfile.Tags {
		dp.Tags.Map[k] = sql.NullString{Valid: true, String: v}
	}

	// as this also performs a remote call to create the device-profile
	// on the network-server, wrap it in a transaction
	err = storage.Transaction(func(tx sqlx.Ext) error {
		return storage.CreateDeviceProfile(ctx, tx, &dp)
	})
	if err != nil {
		return nil, helpers.ErrToRPCError(err)
	}

	dpID, err := uuid.FromBytes(dp.DeviceProfile.Id)
	if err != nil {
		return nil, helpers.ErrToRPCError(err)
	}

	return &pb.CreateDeviceProfileResponse{
		Id: dpID.String(),
	}, nil
}

// Get returns the device-profile matching the given id.
func (a *DeviceProfileServiceAPI) Get(ctx context.Context, req *pb.GetDeviceProfileRequest) (*pb.GetDeviceProfileResponse, error) {
	dpID, err := uuid.FromString(req.Id)
	if err != nil {
		return nil, grpc.Errorf(codes.InvalidArgument, "uuid error: %s", err)
	}

	if err := a.validator.Validate(ctx,
		auth.ValidateDeviceProfileAccess(auth.Read, dpID),
	); err != nil {
		return nil, grpc.Errorf(codes.Unauthenticated, "authentication failed: %s", err)
	}

	dp, err := storage.GetDeviceProfile(ctx, storage.DB(), dpID, false, false)
	if err != nil {
		return nil, helpers.ErrToRPCError(err)
	}

	resp := pb.GetDeviceProfileResponse{
		DeviceProfile: &pb.DeviceProfile{
			Id:                   dpID.String(),
			Name:                 dp.Name,
			OrganizationId:       dp.OrganizationID,
			NetworkServerId:      dp.NetworkServerID,
			PayloadCodec:         string(dp.PayloadCodec),
			PayloadEncoderScript: dp.PayloadEncoderScript,
			PayloadDecoderScript: dp.PayloadDecoderScript,
			SupportsClassB:       dp.DeviceProfile.SupportsClassB,
			ClassBTimeout:        dp.DeviceProfile.ClassBTimeout,
			PingSlotPeriod:       dp.DeviceProfile.PingSlotPeriod,
			PingSlotDr:           dp.DeviceProfile.PingSlotDr,
			PingSlotFreq:         dp.DeviceProfile.PingSlotFreq,
			SupportsClassC:       dp.DeviceProfile.SupportsClassC,
			ClassCTimeout:        dp.DeviceProfile.ClassCTimeout,
			MacVersion:           dp.DeviceProfile.MacVersion,
			RegParamsRevision:    dp.DeviceProfile.RegParamsRevision,
			RxDelay_1:            dp.DeviceProfile.RxDelay_1,
			RxDrOffset_1:         dp.DeviceProfile.RxDrOffset_1,
			RxDatarate_2:         dp.DeviceProfile.RxDatarate_2,
			RxFreq_2:             dp.DeviceProfile.RxFreq_2,
			MaxEirp:              dp.DeviceProfile.MaxEirp,
			MaxDutyCycle:         dp.DeviceProfile.MaxDutyCycle,
			SupportsJoin:         dp.DeviceProfile.SupportsJoin,
			RfRegion:             dp.DeviceProfile.RfRegion,
			Supports_32BitFCnt:   dp.DeviceProfile.Supports_32BitFCnt,
			FactoryPresetFreqs:   dp.DeviceProfile.FactoryPresetFreqs,
			Tags:                 make(map[string]string),
			UplinkInterval:       ptypes.DurationProto(dp.UplinkInterval),
			AdrAlgorithmId:       dp.DeviceProfile.AdrAlgorithmId,
		},
	}

	resp.CreatedAt, err = ptypes.TimestampProto(dp.CreatedAt)
	if err != nil {
		return nil, helpers.ErrToRPCError(err)
	}
	resp.UpdatedAt, err = ptypes.TimestampProto(dp.UpdatedAt)
	if err != nil {
		return nil, helpers.ErrToRPCError(err)
	}

	for k, v := range dp.Tags.Map {
		resp.DeviceProfile.Tags[k] = v.String
	}

	return &resp, nil
}

// Update updates the given device-profile.
func (a *DeviceProfileServiceAPI) Update(ctx context.Context, req *pb.UpdateDeviceProfileRequest) (*empty.Empty, error) {
	if req.DeviceProfile == nil {
		return nil, grpc.Errorf(codes.InvalidArgument, "deviceProfile expected")
	}

	dpID, err := uuid.FromString(req.DeviceProfile.Id)
	if err != nil {
		return nil, grpc.Errorf(codes.InvalidArgument, "uuid error: %s", err)
	}

	if err := a.validator.Validate(ctx,
		auth.ValidateDeviceProfileAccess(auth.Update, dpID),
	); err != nil {
		return nil, grpc.Errorf(codes.Unauthenticated, "authentication failed: %s", err)
	}

	// As this also performs a remote call to update the device-profile
	// on the network-server, wrap it in a transaction.
	// This also locks the local device-profile record in the database.
	err = storage.Transaction(func(tx sqlx.Ext) error {
		dp, err := storage.GetDeviceProfile(ctx, tx, dpID, true, false)
		if err != nil {
			return err
		}

		var uplinkInterval time.Duration
		if req.DeviceProfile.UplinkInterval != nil {
			uplinkInterval, err = ptypes.Duration(req.DeviceProfile.UplinkInterval)
			if err != nil {
				return err
			}
		}

		dp.Name = req.DeviceProfile.Name
		dp.PayloadCodec = codec.Type(req.DeviceProfile.PayloadCodec)
		dp.PayloadEncoderScript = req.DeviceProfile.PayloadEncoderScript
		dp.PayloadDecoderScript = req.DeviceProfile.PayloadDecoderScript
		dp.Tags = hstore.Hstore{
			Map: make(map[string]sql.NullString),
		}
		dp.UplinkInterval = uplinkInterval
		dp.DeviceProfile = ns.DeviceProfile{
			Id:                 dpID.Bytes(),
			SupportsClassB:     req.DeviceProfile.SupportsClassB,
			ClassBTimeout:      req.DeviceProfile.ClassBTimeout,
			PingSlotPeriod:     req.DeviceProfile.PingSlotPeriod,
			PingSlotDr:         req.DeviceProfile.PingSlotDr,
			PingSlotFreq:       req.DeviceProfile.PingSlotFreq,
			SupportsClassC:     req.DeviceProfile.SupportsClassC,
			ClassCTimeout:      req.DeviceProfile.ClassCTimeout,
			MacVersion:         req.DeviceProfile.MacVersion,
			RegParamsRevision:  req.DeviceProfile.RegParamsRevision,
			RxDelay_1:          req.DeviceProfile.RxDelay_1,
			RxDrOffset_1:       req.DeviceProfile.RxDrOffset_1,
			RxDatarate_2:       req.DeviceProfile.RxDatarate_2,
			RxFreq_2:           req.DeviceProfile.RxFreq_2,
			MaxEirp:            req.DeviceProfile.MaxEirp,
			MaxDutyCycle:       req.DeviceProfile.MaxDutyCycle,
			SupportsJoin:       req.DeviceProfile.SupportsJoin,
			RfRegion:           req.DeviceProfile.RfRegion,
			Supports_32BitFCnt: req.DeviceProfile.Supports_32BitFCnt,
			FactoryPresetFreqs: req.DeviceProfile.FactoryPresetFreqs,
			AdrAlgorithmId:     req.DeviceProfile.AdrAlgorithmId,
		}

		for k, v := range req.DeviceProfile.Tags {
			dp.Tags.Map[k] = sql.NullString{Valid: true, String: v}
		}

		return storage.UpdateDeviceProfile(ctx, tx, &dp)
	})
	if err != nil {
		return nil, helpers.ErrToRPCError(err)
	}

	return &empty.Empty{}, nil
}

// Delete deletes the device-profile matching the given id.
func (a *DeviceProfileServiceAPI) Delete(ctx context.Context, req *pb.DeleteDeviceProfileRequest) (*empty.Empty, error) {
	dpID, err := uuid.FromString(req.Id)
	if err != nil {
		return nil, grpc.Errorf(codes.InvalidArgument, "uuid error: %s", err)
	}

	if err := a.validator.Validate(ctx,
		auth.ValidateDeviceProfileAccess(auth.Delete, dpID),
	); err != nil {
		return nil, grpc.Errorf(codes.Unauthenticated, "authentication failed: %s", err)
	}

	// as this also performs a remote call to delete the device-profile
	// on the network-server, wrap it in a transaction
	err = storage.Transaction(func(tx sqlx.Ext) error {
		return storage.DeleteDeviceProfile(ctx, tx, dpID)
	})
	if err != nil {
		return nil, helpers.ErrToRPCError(err)
	}

	return &empty.Empty{}, nil
}

// List lists the available device-profiles.
func (a *DeviceProfileServiceAPI) List(ctx context.Context, req *pb.ListDeviceProfileRequest) (*pb.ListDeviceProfileResponse, error) {
	if req.ApplicationId != 0 {
		if err := a.validator.Validate(ctx,
			auth.ValidateDeviceProfilesAccess(auth.List, 0, req.ApplicationId),
		); err != nil {
			return nil, grpc.Errorf(codes.Unauthenticated, "authentication failed: %s", err)
		}
	} else {
		if err := a.validator.Validate(ctx,
			auth.ValidateDeviceProfilesAccess(auth.List, req.OrganizationId, 0),
		); err != nil {
			return nil, grpc.Errorf(codes.Unauthenticated, "authentication failed: %s", err)
		}
	}

	filters := storage.DeviceProfileFilters{
		Limit:          int(req.Limit),
		Offset:         int(req.Offset),
		OrganizationID: req.OrganizationId,
		ApplicationID:  req.ApplicationId,
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

		// Filter on user ID when org and app ID are not set and user is not
		// global admin.
		if !user.IsAdmin && filters.OrganizationID == 0 && filters.ApplicationID == 0 {
			filters.UserID = user.ID
		}
	case auth.SubjectAPIKey:
		// Nothing to do as the validator function already validated that the
		// API key is either of type admin, org (for the req.OrganizationId) or
		// app (for the req.ApplicationId).
	default:
		return nil, grpc.Errorf(codes.Unauthenticated, "invalid token subject: %s", sub)
	}

	count, err := storage.GetDeviceProfileCount(ctx, storage.DB(), filters)
	if err != nil {
		return nil, helpers.ErrToRPCError(err)
	}

	dps, err := storage.GetDeviceProfiles(ctx, storage.DB(), filters)
	if err != nil {
		return nil, helpers.ErrToRPCError(err)
	}

	resp := pb.ListDeviceProfileResponse{
		TotalCount: int64(count),
	}

	for _, dp := range dps {
		row := pb.DeviceProfileListItem{
			Id:                dp.DeviceProfileID.String(),
			Name:              dp.Name,
			OrganizationId:    dp.OrganizationID,
			NetworkServerId:   dp.NetworkServerID,
			NetworkServerName: dp.NetworkServerName,
		}

		row.CreatedAt, err = ptypes.TimestampProto(dp.CreatedAt)
		if err != nil {
			return nil, helpers.ErrToRPCError(err)
		}
		row.UpdatedAt, err = ptypes.TimestampProto(dp.UpdatedAt)
		if err != nil {
			return nil, helpers.ErrToRPCError(err)
		}

		resp.Result = append(resp.Result, &row)
	}

	return &resp, nil
}
