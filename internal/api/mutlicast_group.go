package api

import (
	"github.com/gofrs/uuid"
	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/jmoiron/sqlx"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

	pb "github.com/brocaar/lora-app-server/api"
	"github.com/brocaar/lora-app-server/internal/api/auth"
	"github.com/brocaar/lora-app-server/internal/common"
	"github.com/brocaar/lora-app-server/internal/multicast"
	"github.com/brocaar/lora-app-server/internal/nsclient"
	"github.com/brocaar/lora-app-server/internal/storage"
	"github.com/brocaar/loraserver/api/ns"
	"github.com/brocaar/lorawan"
)

// MulticastGroupAPI implements the multicast-group api.
type MulticastGroupAPI struct {
	validator        auth.Validator
	db               *common.DBLogger
	routingProfileID uuid.UUID
	nsClientPool     nsclient.Pool
}

// NewMulticastGroupAPI creates a new multicast-group API.
func NewMulticastGroupAPI(validator auth.Validator, db *common.DBLogger, routingProfileID uuid.UUID, nsClientPool nsclient.Pool) *MulticastGroupAPI {
	return &MulticastGroupAPI{
		validator:        validator,
		db:               db,
		routingProfileID: routingProfileID,
		nsClientPool:     nsClientPool,
	}
}

// Create creates the given multicast-group.
func (a *MulticastGroupAPI) Create(ctx context.Context, req *pb.CreateMulticastGroupRequest) (*pb.CreateMulticastGroupResponse, error) {
	if req.MulticastGroup == nil {
		return nil, grpc.Errorf(codes.InvalidArgument, "multicast_group must not be nil")
	}

	spID, err := uuid.FromString(req.MulticastGroup.ServiceProfileId)
	if err != nil {
		return nil, grpc.Errorf(codes.InvalidArgument, err.Error())
	}

	sp, err := storage.GetServiceProfile(a.db, spID, true) // local-only, as we only want to fetch the org. id
	if err != nil {
		return nil, errToRPCError(err)
	}

	if err = a.validator.Validate(ctx,
		auth.ValidateMulticastGroupsAccess(auth.Create, sp.OrganizationID)); err != nil {
		return nil, grpc.Errorf(codes.Unauthenticated, "authentication failed: %s", err)
	}

	var mcAddr lorawan.DevAddr
	if err = mcAddr.UnmarshalText([]byte(req.MulticastGroup.McAddr)); err != nil {
		return nil, grpc.Errorf(codes.InvalidArgument, "mc_app_s_key: %s", err)
	}

	var mcNwkSKey lorawan.AES128Key
	if err = mcNwkSKey.UnmarshalText([]byte(req.MulticastGroup.McNwkSKey)); err != nil {
		return nil, grpc.Errorf(codes.InvalidArgument, "mc_net_s_key: %s", err)
	}

	mg := storage.MulticastGroup{
		Name:             req.MulticastGroup.Name,
		ServiceProfileID: spID,
		MulticastGroup: ns.MulticastGroup{
			McAddr:           mcAddr[:],
			McNwkSKey:        mcNwkSKey[:],
			FCnt:             req.MulticastGroup.FCnt,
			GroupType:        ns.MulticastGroupType(req.MulticastGroup.GroupType),
			Dr:               req.MulticastGroup.Dr,
			Frequency:        req.MulticastGroup.Frequency,
			PingSlotPeriod:   req.MulticastGroup.PingSlotPeriod,
			ServiceProfileId: spID.Bytes(),
			RoutingProfileId: a.routingProfileID.Bytes(),
		},
	}

	if err = mg.MCAppSKey.UnmarshalText([]byte(req.MulticastGroup.McAppSKey)); err != nil {
		return nil, grpc.Errorf(codes.InvalidArgument, "mc_app_s_key: %s", err)
	}

	if err = storage.Transaction(a.db, func(tx sqlx.Ext) error {
		if err := storage.CreateMulticastGroup(tx, &mg); err != nil {
			return errToRPCError(err)
		}

		return nil
	}); err != nil {
		return nil, err
	}

	var mgID uuid.UUID
	copy(mgID[:], mg.MulticastGroup.Id)

	return &pb.CreateMulticastGroupResponse{
		Id: mgID.String(),
	}, nil
}

// Get returns a multicast-group given an ID.
func (a *MulticastGroupAPI) Get(ctx context.Context, req *pb.GetMulticastGroupRequest) (*pb.GetMulticastGroupResponse, error) {
	mgID, err := uuid.FromString(req.Id)
	if err != nil {
		return nil, grpc.Errorf(codes.InvalidArgument, "id: %s", err)
	}

	if err = a.validator.Validate(ctx,
		auth.ValidateMulticastGroupAccess(auth.Read, mgID)); err != nil {
		return nil, grpc.Errorf(codes.Unauthenticated, "authentication failed: %s", err)
	}

	mg, err := storage.GetMulticastGroup(a.db, mgID, false, false)
	if err != nil {
		return nil, errToRPCError(err)
	}

	var mcAddr lorawan.DevAddr
	var mcNwkSKey lorawan.AES128Key
	copy(mcAddr[:], mg.MulticastGroup.McAddr)
	copy(mcNwkSKey[:], mg.MulticastGroup.McNwkSKey)

	out := pb.GetMulticastGroupResponse{
		MulticastGroup: &pb.MulticastGroup{
			Id:               mgID.String(),
			Name:             mg.Name,
			McAddr:           mcAddr.String(),
			McNwkSKey:        mcNwkSKey.String(),
			McAppSKey:        mg.MCAppSKey.String(),
			FCnt:             mg.MulticastGroup.FCnt,
			GroupType:        pb.MulticastGroupType(mg.MulticastGroup.GroupType),
			Dr:               mg.MulticastGroup.Dr,
			Frequency:        mg.MulticastGroup.Frequency,
			PingSlotPeriod:   mg.MulticastGroup.PingSlotPeriod,
			ServiceProfileId: mg.ServiceProfileID.String(),
		},
	}

	out.CreatedAt, err = ptypes.TimestampProto(mg.CreatedAt)
	if err != nil {
		return nil, errToRPCError(err)
	}

	out.UpdatedAt, err = ptypes.TimestampProto(mg.UpdatedAt)
	if err != nil {
		return nil, errToRPCError(err)
	}

	return &out, nil
}

// Update updates the given multicast-group.
func (a *MulticastGroupAPI) Update(ctx context.Context, req *pb.UpdateMulticastGroupRequest) (*empty.Empty, error) {
	if req.MulticastGroup == nil {
		return nil, grpc.Errorf(codes.InvalidArgument, "multicast_group must not be nil")
	}

	mgID, err := uuid.FromString(req.MulticastGroup.Id)
	if err != nil {
		return nil, grpc.Errorf(codes.InvalidArgument, "id: %s", err)
	}

	if err = a.validator.Validate(ctx,
		auth.ValidateMulticastGroupAccess(auth.Update, mgID)); err != nil {
		return nil, grpc.Errorf(codes.Unauthenticated, "authentication failed: %s", err)
	}

	mg, err := storage.GetMulticastGroup(a.db, mgID, false, false)
	if err != nil {
		return nil, errToRPCError(err)
	}

	var mcAddr lorawan.DevAddr
	if err = mcAddr.UnmarshalText([]byte(req.MulticastGroup.McAddr)); err != nil {
		return nil, grpc.Errorf(codes.InvalidArgument, "mc_app_s_key: %s", err)
	}

	var mcNwkSKey lorawan.AES128Key
	if err = mcNwkSKey.UnmarshalText([]byte(req.MulticastGroup.McNwkSKey)); err != nil {
		return nil, grpc.Errorf(codes.InvalidArgument, "mc_net_s_key: %s", err)
	}

	mg.Name = req.MulticastGroup.Name
	mg.MulticastGroup = ns.MulticastGroup{
		Id:               mg.MulticastGroup.Id,
		McAddr:           mcAddr[:],
		McNwkSKey:        mcNwkSKey[:],
		FCnt:             req.MulticastGroup.FCnt,
		GroupType:        ns.MulticastGroupType(req.MulticastGroup.GroupType),
		Dr:               req.MulticastGroup.Dr,
		Frequency:        req.MulticastGroup.Frequency,
		PingSlotPeriod:   req.MulticastGroup.PingSlotPeriod,
		ServiceProfileId: mg.MulticastGroup.ServiceProfileId,
		RoutingProfileId: mg.MulticastGroup.RoutingProfileId,
	}

	if err = mg.MCAppSKey.UnmarshalText([]byte(req.MulticastGroup.McAppSKey)); err != nil {
		return nil, grpc.Errorf(codes.InvalidArgument, "mc_app_s_key: %s", err)
	}

	if err = storage.Transaction(a.db, func(tx sqlx.Ext) error {
		if err := storage.UpdateMulticastGroup(tx, &mg); err != nil {
			return errToRPCError(err)
		}

		return nil
	}); err != nil {
		return nil, err
	}

	return &empty.Empty{}, nil
}

// Delete deletes a multicast-group given an ID.
func (a *MulticastGroupAPI) Delete(ctx context.Context, req *pb.DeleteMulticastGroupRequest) (*empty.Empty, error) {
	mgID, err := uuid.FromString(req.Id)
	if err != nil {
		return nil, grpc.Errorf(codes.InvalidArgument, "id: %s", err)
	}

	if err = storage.Transaction(a.db, func(tx sqlx.Ext) error {
		if err := storage.DeleteMulticastGroup(tx, mgID); err != nil {
			return errToRPCError(err)
		}
		return nil
	}); err != nil {
		return nil, err
	}

	return &empty.Empty{}, nil
}

// List lists the available multicast-groups.
func (a *MulticastGroupAPI) List(ctx context.Context, req *pb.ListMulticastGroupRequest) (*pb.ListMulticastGroupResponse, error) {
	var err error
	var idFilter bool

	filters := storage.MulticastGroupFilters{
		OrganizationID: req.OrganizationId,
		Search:         req.Search,
		Limit:          int(req.Limit),
		Offset:         int(req.Offset),
	}

	// if org. filter has been set, validate the client has access to the given org
	if filters.OrganizationID != 0 {
		idFilter = true

		if err = a.validator.Validate(ctx,
			auth.ValidateOrganizationAccess(auth.Read, req.OrganizationId)); err != nil {
			return nil, grpc.Errorf(codes.Unauthenticated, "authentication failed: %s", err)
		}
	}

	// if sp filter has been set, validate the client has access to the given sp
	if req.ServiceProfileId != "" {
		idFilter = true

		filters.ServiceProfileID, err = uuid.FromString(req.ServiceProfileId)
		if err != nil {
			return nil, grpc.Errorf(codes.InvalidArgument, "service_profile_id: %s", err)
		}

		if err = a.validator.Validate(ctx,
			auth.ValidateServiceProfileAccess(auth.Read, filters.ServiceProfileID)); err != nil {
			return nil, grpc.Errorf(codes.Unauthenticated, "authentication error: %s", err)
		}
	}

	// if devEUI has been set, validate the client has access to the given device
	if req.DevEui != "" {
		idFilter = true

		if err = filters.DevEUI.UnmarshalText([]byte(req.DevEui)); err != nil {
			return nil, grpc.Errorf(codes.InvalidArgument, "dev_eui: %s", err)
		}

		if err = a.validator.Validate(ctx,
			auth.ValidateNodeAccess(filters.DevEUI, auth.Read)); err != nil {
			return nil, grpc.Errorf(codes.Unauthenticated, "authentication error: %s", err)
		}
	}

	// listing all stored objects is for global admin only
	if !idFilter {
		isAdmin, err := a.validator.GetIsAdmin(ctx)
		if err != nil {
			return nil, errToRPCError(err)
		}

		if !isAdmin {
			return nil, grpc.Errorf(codes.Unauthenticated, "client must be global admin for unfiltered request")
		}
	}

	count, err := storage.GetMulticastGroupCount(a.db, filters)
	if err != nil {
		return nil, errToRPCError(err)
	}

	items, err := storage.GetMulticastGroups(a.db, filters)
	if err != nil {
		return nil, errToRPCError(err)
	}

	out := pb.ListMulticastGroupResponse{
		TotalCount: int64(count),
	}

	for _, item := range items {
		out.Result = append(out.Result, &pb.MulticastGroupListItem{
			Id:                 item.ID.String(),
			Name:               item.Name,
			ServiceProfileId:   item.ServiceProfileID.String(),
			ServiceProfileName: item.ServiceProfileName,
		})
	}

	return &out, nil
}

// AddDevice adds the given device to the multicast-group.
func (a *MulticastGroupAPI) AddDevice(ctx context.Context, req *pb.AddDeviceToMulticastGroupRequest) (*empty.Empty, error) {
	mgID, err := uuid.FromString(req.MulticastGroupId)
	if err != nil {
		return nil, grpc.Errorf(codes.InvalidArgument, "multicast_group_id: %s", err)
	}

	var devEUI lorawan.EUI64
	if err = devEUI.UnmarshalText([]byte(req.DevEui)); err != nil {
		return nil, grpc.Errorf(codes.InvalidArgument, "dev_eui: %s", err)
	}

	if err = a.validator.Validate(ctx,
		auth.ValidateMulticastGroupAccess(auth.Update, mgID)); err != nil {
		return nil, grpc.Errorf(codes.Unauthenticated, "authentication failed: %s", err)
	}

	// validate that the device is under the same service-profile as the multicast-group
	dev, err := storage.GetDevice(a.db, devEUI, false, true)
	if err != nil {
		return nil, errToRPCError(err)
	}

	app, err := storage.GetApplication(a.db, dev.ApplicationID)
	if err != nil {
		return nil, errToRPCError(err)
	}

	mg, err := storage.GetMulticastGroup(a.db, mgID, false, true)
	if err != nil {
		return nil, errToRPCError(err)
	}

	if app.ServiceProfileID != mg.ServiceProfileID {
		return nil, grpc.Errorf(codes.FailedPrecondition, "service-profile of device != service-profile of multicast-group")
	}

	if err = storage.Transaction(a.db, func(tx sqlx.Ext) error {
		if err := storage.AddDeviceToMulticastGroup(tx, mgID, devEUI); err != nil {
			return errToRPCError(err)
		}
		return nil
	}); err != nil {
		return nil, err
	}

	return &empty.Empty{}, nil
}

// RemoveDevice removes the given device from the multicast-group.
func (a *MulticastGroupAPI) RemoveDevice(ctx context.Context, req *pb.RemoveDeviceFromMulticastGroupRequest) (*empty.Empty, error) {
	mgID, err := uuid.FromString(req.MulticastGroupId)
	if err != nil {
		return nil, grpc.Errorf(codes.InvalidArgument, "multicast_group_id: %s", err)
	}

	var devEUI lorawan.EUI64
	if err = devEUI.UnmarshalText([]byte(req.DevEui)); err != nil {
		return nil, grpc.Errorf(codes.InvalidArgument, "dev_eui: %s", err)
	}

	if err = a.validator.Validate(ctx,
		auth.ValidateMulticastGroupAccess(auth.Update, mgID)); err != nil {
		return nil, grpc.Errorf(codes.Unauthenticated, "authentication failed: %s", err)
	}

	if err = storage.Transaction(a.db, func(tx sqlx.Ext) error {
		if err := storage.RemoveDeviceFromMulticastGroup(tx, mgID, devEUI); err != nil {
			return errToRPCError(err)
		}
		return nil
	}); err != nil {
		return nil, err
	}

	return &empty.Empty{}, nil
}

// Enqueue adds the given item to the multicast-queue.
func (a *MulticastGroupAPI) Enqueue(ctx context.Context, req *pb.EnqueueMulticastQueueItemRequest) (*pb.EnqueueMulticastQueueItemResponse, error) {
	var fCnt uint32

	if req.MulticastQueueItem == nil {
		return nil, grpc.Errorf(codes.InvalidArgument, "multicast_queue_item must not be nil")
	}

	if req.MulticastQueueItem.FPort == 0 {
		return nil, grpc.Errorf(codes.InvalidArgument, "f_port must be > 0")
	}

	mgID, err := uuid.FromString(req.MulticastQueueItem.MulticastGroupId)
	if err != nil {
		return nil, grpc.Errorf(codes.InvalidArgument, "multicast_group_id: %s", err)
	}

	if err = a.validator.Validate(ctx,
		auth.ValidateMulticastGroupQueueAccess(auth.Create, mgID)); err != nil {
		return nil, grpc.Errorf(codes.Unauthenticated, "authentication failed: %s", err)
	}

	if err = storage.Transaction(a.db, func(tx sqlx.Ext) error {
		var err error
		fCnt, err = multicast.Enqueue(tx, mgID, uint8(req.MulticastQueueItem.FPort), req.MulticastQueueItem.Data)
		if err != nil {
			return grpc.Errorf(codes.Internal, "enqueue multicast-group queue-item error: %s", err)
		}

		return nil
	}); err != nil {
		return nil, err
	}

	return &pb.EnqueueMulticastQueueItemResponse{
		FCnt: fCnt,
	}, nil
}

// FlushQueue flushes the multicast-group queue.
func (a *MulticastGroupAPI) FlushQueue(ctx context.Context, req *pb.FlushMulticastGroupQueueItemsRequest) (*empty.Empty, error) {
	mgID, err := uuid.FromString(req.MulticastGroupId)
	if err != nil {
		return nil, grpc.Errorf(codes.InvalidArgument, "multicast_group_id: %s", err)
	}

	if err = a.validator.Validate(ctx,
		auth.ValidateMulticastGroupQueueAccess(auth.Delete, mgID)); err != nil {
		return nil, grpc.Errorf(codes.Unauthenticated, "authentication failed: %s", err)
	}

	n, err := storage.GetNetworkServerForMulticastGroupID(a.db, mgID)
	if err != nil {
		return nil, errToRPCError(err)
	}

	nsClient, err := a.nsClientPool.Get(n.Server, []byte(n.CACert), []byte(n.TLSCert), []byte(n.TLSKey))
	if err != nil {
		return nil, errToRPCError(err)
	}

	_, err = nsClient.FlushMulticastQueueForMulticastGroup(ctx, &ns.FlushMulticastQueueForMulticastGroupRequest{
		MulticastGroupId: mgID.Bytes(),
	})
	if err != nil {
		return nil, err
	}

	return &empty.Empty{}, nil
}

// ListQueue lists the items in the multicast-group queue.
func (a *MulticastGroupAPI) ListQueue(ctx context.Context, req *pb.ListMulticastGroupQueueItemsRequest) (*pb.ListMulticastGroupQueueItemsResponse, error) {
	mgID, err := uuid.FromString(req.MulticastGroupId)
	if err != nil {
		return nil, grpc.Errorf(codes.InvalidArgument, "multicast_group_id: %s", err)
	}

	if err = a.validator.Validate(ctx,
		auth.ValidateMulticastGroupQueueAccess(auth.Read, mgID)); err != nil {
		return nil, grpc.Errorf(codes.Unauthenticated, "authentication failed: %s", err)
	}

	mg, err := storage.GetMulticastGroup(a.db, mgID, false, false)
	if err != nil {
		return nil, errToRPCError(err)
	}

	n, err := storage.GetNetworkServerForMulticastGroupID(a.db, mgID)
	if err != nil {
		return nil, errToRPCError(err)
	}

	nsClient, err := a.nsClientPool.Get(n.Server, []byte(n.CACert), []byte(n.TLSCert), []byte(n.TLSKey))
	if err != nil {
		return nil, errToRPCError(err)
	}

	queuItemsResp, err := nsClient.GetMulticastQueueItemsForMulticastGroup(ctx, &ns.GetMulticastQueueItemsForMulticastGroupRequest{
		MulticastGroupId: mgID.Bytes(),
	})
	if err != nil {
		return nil, err
	}

	var resp pb.ListMulticastGroupQueueItemsResponse
	var devAddr lorawan.DevAddr
	copy(devAddr[:], mg.MulticastGroup.McAddr)

	for _, qi := range queuItemsResp.MulticastQueueItems {
		b, err := lorawan.EncryptFRMPayload(mg.MCAppSKey, false, devAddr, qi.FCnt, qi.FrmPayload)
		if err != nil {
			return nil, errToRPCError(err)
		}

		resp.MulticastQueueItems = append(resp.MulticastQueueItems, &pb.MulticastQueueItem{
			MulticastGroupId: mgID.String(),
			FCnt:             qi.FCnt,
			FPort:            qi.FPort,
			Data:             b,
		})
	}

	return &resp, nil
}
