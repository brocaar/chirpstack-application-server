package external

import (
	"github.com/gofrs/uuid"
	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/empty"

	"github.com/brocaar/loraserver/api/ns"

	"github.com/jmoiron/sqlx"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

	pb "github.com/brocaar/chirpstack-application-server/api"
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

	dp := storage.DeviceProfile{
		OrganizationID:       req.DeviceProfile.OrganizationId,
		NetworkServerID:      req.DeviceProfile.NetworkServerId,
		Name:                 req.DeviceProfile.Name,
		PayloadCodec:         codec.Type(req.DeviceProfile.PayloadCodec),
		PayloadEncoderScript: req.DeviceProfile.PayloadEncoderScript,
		PayloadDecoderScript: req.DeviceProfile.PayloadDecoderScript,
		DeviceProfile: ns.DeviceProfile{
			SupportsClassB:      req.DeviceProfile.SupportsClassB,
			ClassBTimeout:       req.DeviceProfile.ClassBTimeout,
			PingSlotPeriod:      req.DeviceProfile.PingSlotPeriod,
			PingSlotDr:          req.DeviceProfile.PingSlotDr,
			PingSlotFreq:        req.DeviceProfile.PingSlotFreq,
			SupportsClassC:      req.DeviceProfile.SupportsClassC,
			ClassCTimeout:       req.DeviceProfile.ClassCTimeout,
			MacVersion:          req.DeviceProfile.MacVersion,
			RegParamsRevision:   req.DeviceProfile.RegParamsRevision,
			RxDelay_1:           req.DeviceProfile.RxDelay_1,
			RxDrOffset_1:        req.DeviceProfile.RxDrOffset_1,
			RxDatarate_2:        req.DeviceProfile.RxDatarate_2,
			RxFreq_2:            req.DeviceProfile.RxFreq_2,
			MaxEirp:             req.DeviceProfile.MaxEirp,
			MaxDutyCycle:        req.DeviceProfile.MaxDutyCycle,
			SupportsJoin:        req.DeviceProfile.SupportsJoin,
			RfRegion:            req.DeviceProfile.RfRegion,
			Supports_32BitFCnt:  req.DeviceProfile.Supports_32BitFCnt,
			FactoryPresetFreqs:  req.DeviceProfile.FactoryPresetFreqs,
			GeolocBufferTtl:     req.DeviceProfile.GeolocBufferTtl,
			GeolocMinBufferSize: req.DeviceProfile.GeolocMinBufferSize,
		},
	}

	// as this also performs a remote call to create the device-profile
	// on the network-server, wrap it in a transaction
	err := storage.Transaction(func(tx sqlx.Ext) error {
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
			GeolocBufferTtl:      dp.DeviceProfile.GeolocBufferTtl,
			GeolocMinBufferSize:  dp.DeviceProfile.GeolocMinBufferSize,
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

		dp.Name = req.DeviceProfile.Name
		dp.PayloadCodec = codec.Type(req.DeviceProfile.PayloadCodec)
		dp.PayloadEncoderScript = req.DeviceProfile.PayloadEncoderScript
		dp.PayloadDecoderScript = req.DeviceProfile.PayloadDecoderScript
		dp.DeviceProfile = ns.DeviceProfile{
			Id:                  dpID.Bytes(),
			SupportsClassB:      req.DeviceProfile.SupportsClassB,
			ClassBTimeout:       req.DeviceProfile.ClassBTimeout,
			PingSlotPeriod:      req.DeviceProfile.PingSlotPeriod,
			PingSlotDr:          req.DeviceProfile.PingSlotDr,
			PingSlotFreq:        req.DeviceProfile.PingSlotFreq,
			SupportsClassC:      req.DeviceProfile.SupportsClassC,
			ClassCTimeout:       req.DeviceProfile.ClassCTimeout,
			MacVersion:          req.DeviceProfile.MacVersion,
			RegParamsRevision:   req.DeviceProfile.RegParamsRevision,
			RxDelay_1:           req.DeviceProfile.RxDelay_1,
			RxDrOffset_1:        req.DeviceProfile.RxDrOffset_1,
			RxDatarate_2:        req.DeviceProfile.RxDatarate_2,
			RxFreq_2:            req.DeviceProfile.RxFreq_2,
			MaxEirp:             req.DeviceProfile.MaxEirp,
			MaxDutyCycle:        req.DeviceProfile.MaxDutyCycle,
			SupportsJoin:        req.DeviceProfile.SupportsJoin,
			RfRegion:            req.DeviceProfile.RfRegion,
			Supports_32BitFCnt:  req.DeviceProfile.Supports_32BitFCnt,
			FactoryPresetFreqs:  req.DeviceProfile.FactoryPresetFreqs,
			GeolocBufferTtl:     req.DeviceProfile.GeolocBufferTtl,
			GeolocMinBufferSize: req.DeviceProfile.GeolocMinBufferSize,
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

	isAdmin, err := a.validator.GetIsAdmin(ctx)
	if err != nil {
		return nil, helpers.ErrToRPCError(err)
	}

	username, err := a.validator.GetUsername(ctx)
	if err != nil {
		return nil, helpers.ErrToRPCError(err)
	}

	var count int
	var dps []storage.DeviceProfileMeta

	if req.ApplicationId != 0 {
		dps, err = storage.GetDeviceProfilesForApplicationID(ctx, storage.DB(), req.ApplicationId, int(req.Limit), int(req.Offset))
		if err != nil {
			return nil, helpers.ErrToRPCError(err)
		}

		count, err = storage.GetDeviceProfileCountForApplicationID(ctx, storage.DB(), req.ApplicationId)
		if err != nil {
			return nil, helpers.ErrToRPCError(err)
		}
	} else if req.OrganizationId != 0 {
		dps, err = storage.GetDeviceProfilesForOrganizationID(ctx, storage.DB(), req.OrganizationId, int(req.Limit), int(req.Offset))
		if err != nil {
			return nil, helpers.ErrToRPCError(err)
		}

		count, err = storage.GetDeviceProfileCountForOrganizationID(ctx, storage.DB(), req.OrganizationId)
		if err != nil {
			return nil, helpers.ErrToRPCError(err)
		}
	} else {
		if isAdmin {
			dps, err = storage.GetDeviceProfiles(ctx, storage.DB(), int(req.Limit), int(req.Offset))
			if err != nil {
				return nil, helpers.ErrToRPCError(err)
			}

			count, err = storage.GetDeviceProfileCount(ctx, storage.DB())
			if err != nil {
				return nil, helpers.ErrToRPCError(err)
			}
		} else {
			dps, err = storage.GetDeviceProfilesForUser(ctx, storage.DB(), username, int(req.Limit), int(req.Offset))
			if err != nil {
				return nil, helpers.ErrToRPCError(err)
			}

			count, err = storage.GetDeviceProfileCountForUser(ctx, storage.DB(), username)
			if err != nil {
				return nil, helpers.ErrToRPCError(err)
			}
		}
	}

	resp := pb.ListDeviceProfileResponse{
		TotalCount: int64(count),
	}

	for _, dp := range dps {
		row := pb.DeviceProfileListItem{
			Id:              dp.DeviceProfileID.String(),
			Name:            dp.Name,
			OrganizationId:  dp.OrganizationID,
			NetworkServerId: dp.NetworkServerID,
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
