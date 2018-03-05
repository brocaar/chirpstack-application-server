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
		auth.ValidateDeviceProfilesAccess(auth.Create, req.OrganizationID, 0),
	); err != nil {
		return nil, grpc.Errorf(codes.Unauthenticated, "authentication failed: %s", err)
	}

	dp := storage.DeviceProfile{
		OrganizationID:  req.OrganizationID,
		NetworkServerID: req.NetworkServerID,
		Name:            req.Name,
		DeviceProfile: backend.DeviceProfile{
			SupportsClassB:    req.DeviceProfile.SupportsClassB,
			ClassBTimeout:     int(req.DeviceProfile.ClassBTimeout),
			PingSlotPeriod:    int(req.DeviceProfile.PingSlotPeriod),
			PingSlotDR:        int(req.DeviceProfile.PingSlotDR),
			PingSlotFreq:      backend.Frequency(req.DeviceProfile.PingSlotFreq),
			SupportsClassC:    req.DeviceProfile.SupportsClassC,
			ClassCTimeout:     int(req.DeviceProfile.ClassCTimeout),
			MACVersion:        req.DeviceProfile.MacVersion,
			RegParamsRevision: req.DeviceProfile.RegParamsRevision,
			RXDelay1:          int(req.DeviceProfile.RxDelay1),
			RXDROffset1:       int(req.DeviceProfile.RxDROffset1),
			RXDataRate2:       int(req.DeviceProfile.RxDataRate2),
			RXFreq2:           backend.Frequency(req.DeviceProfile.RxFreq2),
			MaxEIRP:           int(req.DeviceProfile.MaxEIRP),
			MaxDutyCycle:      backend.Percentage(req.DeviceProfile.MaxDutyCycle),
			SupportsJoin:      req.DeviceProfile.SupportsJoin,
			RFRegion:          backend.RFRegion(req.DeviceProfile.RfRegion),
			Supports32bitFCnt: req.DeviceProfile.Supports32BitFCnt,
		},
	}

	for _, freq := range req.DeviceProfile.FactoryPresetFreqs {
		dp.DeviceProfile.FactoryPresetFreqs = append(dp.DeviceProfile.FactoryPresetFreqs, backend.Frequency(freq))
	}

	// as this also performs a remote call to create the device-profile
	// on the network-server, wrap it in a transaction
	err := storage.Transaction(config.C.PostgreSQL.DB, func(tx sqlx.Ext) error {
		return storage.CreateDeviceProfile(tx, &dp)
	})
	if err != nil {
		return nil, errToRPCError(err)
	}

	return &pb.CreateDeviceProfileResponse{
		DeviceProfileID: dp.DeviceProfile.DeviceProfileID,
	}, nil
}

// Get returns the device-profile matching the given id.
func (a *DeviceProfileServiceAPI) Get(ctx context.Context, req *pb.GetDeviceProfileRequest) (*pb.GetDeviceProfileResponse, error) {
	if err := a.validator.Validate(ctx,
		auth.ValidateDeviceProfileAccess(auth.Read, req.DeviceProfileID),
	); err != nil {
		return nil, grpc.Errorf(codes.Unauthenticated, "authentication failed: %s", err)
	}

	dp, err := storage.GetDeviceProfile(config.C.PostgreSQL.DB, req.DeviceProfileID)
	if err != nil {
		return nil, errToRPCError(err)
	}

	resp := pb.GetDeviceProfileResponse{
		Name:            dp.Name,
		OrganizationID:  dp.OrganizationID,
		NetworkServerID: dp.NetworkServerID,
		CreatedAt:       dp.CreatedAt.Format(time.RFC3339Nano),
		UpdatedAt:       dp.UpdatedAt.Format(time.RFC3339Nano),
		DeviceProfile: &pb.DeviceProfile{
			DeviceProfileID:   dp.DeviceProfile.DeviceProfileID,
			SupportsClassB:    dp.DeviceProfile.SupportsClassB,
			ClassBTimeout:     uint32(dp.DeviceProfile.ClassBTimeout),
			PingSlotPeriod:    uint32(dp.DeviceProfile.PingSlotPeriod),
			PingSlotDR:        uint32(dp.DeviceProfile.PingSlotDR),
			PingSlotFreq:      uint32(dp.DeviceProfile.PingSlotFreq),
			SupportsClassC:    dp.DeviceProfile.SupportsClassC,
			ClassCTimeout:     uint32(dp.DeviceProfile.ClassCTimeout),
			MacVersion:        dp.DeviceProfile.MACVersion,
			RegParamsRevision: dp.DeviceProfile.RegParamsRevision,
			RxDelay1:          uint32(dp.DeviceProfile.RXDelay1),
			RxDROffset1:       uint32(dp.DeviceProfile.RXDROffset1),
			RxDataRate2:       uint32(dp.DeviceProfile.RXDataRate2),
			RxFreq2:           uint32(dp.DeviceProfile.RXFreq2),
			MaxEIRP:           uint32(dp.DeviceProfile.MaxEIRP),
			MaxDutyCycle:      uint32(dp.DeviceProfile.MaxDutyCycle),
			SupportsJoin:      dp.DeviceProfile.SupportsJoin,
			RfRegion:          string(dp.DeviceProfile.RFRegion),
			Supports32BitFCnt: dp.DeviceProfile.Supports32bitFCnt,
		},
	}

	for _, freq := range dp.DeviceProfile.FactoryPresetFreqs {
		resp.DeviceProfile.FactoryPresetFreqs = append(resp.DeviceProfile.FactoryPresetFreqs, uint32(freq))
	}

	return &resp, nil
}

// Update updates the given device-profile.
func (a *DeviceProfileServiceAPI) Update(ctx context.Context, req *pb.UpdateDeviceProfileRequest) (*pb.UpdateDeviceProfileResponse, error) {
	if req.DeviceProfile == nil {
		return nil, grpc.Errorf(codes.InvalidArgument, "deviceProfile expected")
	}

	if err := a.validator.Validate(ctx,
		auth.ValidateDeviceProfileAccess(auth.Update, req.DeviceProfile.DeviceProfileID),
	); err != nil {
		return nil, grpc.Errorf(codes.Unauthenticated, "authentication failed: %s", err)
	}

	dp, err := storage.GetDeviceProfile(config.C.PostgreSQL.DB, req.DeviceProfile.DeviceProfileID)
	if err != nil {
		return nil, errToRPCError(err)
	}

	dp.Name = req.Name
	dp.DeviceProfile = backend.DeviceProfile{
		DeviceProfileID:   req.DeviceProfile.DeviceProfileID,
		SupportsClassB:    req.DeviceProfile.SupportsClassB,
		ClassBTimeout:     int(req.DeviceProfile.ClassBTimeout),
		PingSlotPeriod:    int(req.DeviceProfile.PingSlotPeriod),
		PingSlotDR:        int(req.DeviceProfile.PingSlotDR),
		PingSlotFreq:      backend.Frequency(req.DeviceProfile.PingSlotFreq),
		SupportsClassC:    req.DeviceProfile.SupportsClassC,
		ClassCTimeout:     int(req.DeviceProfile.ClassCTimeout),
		MACVersion:        req.DeviceProfile.MacVersion,
		RegParamsRevision: req.DeviceProfile.RegParamsRevision,
		RXDelay1:          int(req.DeviceProfile.RxDelay1),
		RXDROffset1:       int(req.DeviceProfile.RxDROffset1),
		RXDataRate2:       int(req.DeviceProfile.RxDataRate2),
		RXFreq2:           backend.Frequency(req.DeviceProfile.RxFreq2),
		MaxEIRP:           int(req.DeviceProfile.MaxEIRP),
		MaxDutyCycle:      backend.Percentage(req.DeviceProfile.MaxDutyCycle),
		SupportsJoin:      req.DeviceProfile.SupportsJoin,
		RFRegion:          backend.RFRegion(req.DeviceProfile.RfRegion),
		Supports32bitFCnt: req.DeviceProfile.Supports32BitFCnt,
	}

	for _, freq := range req.DeviceProfile.FactoryPresetFreqs {
		dp.DeviceProfile.FactoryPresetFreqs = append(dp.DeviceProfile.FactoryPresetFreqs, backend.Frequency(freq))
	}

	// as this also performs a remote call to update the device-profile
	// on the network-server, wrap it in a transaction
	err = storage.Transaction(config.C.PostgreSQL.DB, func(tx sqlx.Ext) error {
		return storage.UpdateDeviceProfile(tx, &dp)
	})
	if err != nil {
		return nil, errToRPCError(err)
	}

	return &pb.UpdateDeviceProfileResponse{}, nil
}

// Delete deletes the device-profile matching the given id.
func (a *DeviceProfileServiceAPI) Delete(ctx context.Context, req *pb.DeleteDeviceProfileRequest) (*pb.DeleteDeviceProfileResponse, error) {
	if err := a.validator.Validate(ctx,
		auth.ValidateDeviceProfileAccess(auth.Delete, req.DeviceProfileID),
	); err != nil {
		return nil, grpc.Errorf(codes.Unauthenticated, "authentication failed: %s", err)
	}

	// as this also performs a remote call to delete the device-profile
	// on the network-server, wrap it in a transaction
	err := storage.Transaction(config.C.PostgreSQL.DB, func(tx sqlx.Ext) error {
		return storage.DeleteDeviceProfile(tx, req.DeviceProfileID)
	})
	if err != nil {
		return nil, errToRPCError(err)
	}

	return &pb.DeleteDeviceProfileResponse{}, nil
}

// List lists the available device-profiles.
func (a *DeviceProfileServiceAPI) List(ctx context.Context, req *pb.ListDeviceProfileRequest) (*pb.ListDeviceProfileResponse, error) {
	if req.ApplicationID != 0 {
		if err := a.validator.Validate(ctx,
			auth.ValidateDeviceProfilesAccess(auth.List, 0, req.ApplicationID),
		); err != nil {
			return nil, grpc.Errorf(codes.Unauthenticated, "authentication failed: %s", err)
		}
	} else {
		if err := a.validator.Validate(ctx,
			auth.ValidateDeviceProfilesAccess(auth.List, req.OrganizationID, 0),
		); err != nil {
			return nil, grpc.Errorf(codes.Unauthenticated, "authentication failed: %s", err)
		}
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
	var dps []storage.DeviceProfileMeta

	if req.ApplicationID != 0 {
		dps, err = storage.GetDeviceProfilesForApplicationID(config.C.PostgreSQL.DB, req.ApplicationID, int(req.Limit), int(req.Offset))
		if err != nil {
			return nil, errToRPCError(err)
		}

		count, err = storage.GetDeviceProfileCountForApplicationID(config.C.PostgreSQL.DB, req.ApplicationID)
		if err != nil {
			return nil, errToRPCError(err)
		}
	} else if req.OrganizationID != 0 {
		dps, err = storage.GetDeviceProfilesForOrganizationID(config.C.PostgreSQL.DB, req.OrganizationID, int(req.Limit), int(req.Offset))
		if err != nil {
			return nil, errToRPCError(err)
		}

		count, err = storage.GetDeviceProfileCountForOrganizationID(config.C.PostgreSQL.DB, req.OrganizationID)
		if err != nil {
			return nil, errToRPCError(err)
		}
	} else {
		if isAdmin {
			dps, err = storage.GetDeviceProfiles(config.C.PostgreSQL.DB, int(req.Limit), int(req.Offset))
			if err != nil {
				return nil, errToRPCError(err)
			}

			count, err = storage.GetDeviceProfileCount(config.C.PostgreSQL.DB)
			if err != nil {
				return nil, errToRPCError(err)
			}
		} else {
			dps, err = storage.GetDeviceProfilesForUser(config.C.PostgreSQL.DB, username, int(req.Limit), int(req.Offset))
			if err != nil {
				return nil, errToRPCError(err)
			}

			count, err = storage.GetDeviceProfileCountForUser(config.C.PostgreSQL.DB, username)
			if err != nil {
				return nil, errToRPCError(err)
			}
		}
	}

	resp := pb.ListDeviceProfileResponse{
		TotalCount: int64(count),
	}
	for _, dp := range dps {
		resp.Result = append(resp.Result, &pb.DeviceProfileMeta{
			DeviceProfileID: dp.DeviceProfileID,
			Name:            dp.Name,
			OrganizationID:  dp.OrganizationID,
			NetworkServerID: dp.NetworkServerID,
			CreatedAt:       dp.CreatedAt.Format(time.RFC3339Nano),
			UpdatedAt:       dp.UpdatedAt.Format(time.RFC3339Nano),
		})
	}

	return &resp, nil
}
