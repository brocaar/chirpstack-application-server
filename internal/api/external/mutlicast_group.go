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
	"github.com/brocaar/chirpstack-application-server/internal/backend/networkserver"
	"github.com/brocaar/chirpstack-application-server/internal/multicast"
	"github.com/brocaar/chirpstack-application-server/internal/storage"
	"github.com/brocaar/lorawan"
)

// MulticastGroupAPI implements the multicast-group api.
type MulticastGroupAPI struct {
	validator        auth.Validator
	routingProfileID uuid.UUID
}

// NewMulticastGroupAPI creates a new multicast-group API.
func NewMulticastGroupAPI(validator auth.Validator, routingProfileID uuid.UUID) *MulticastGroupAPI {
	return &MulticastGroupAPI{
		validator:        validator,
		routingProfileID: routingProfileID,
	}
}

// Create creates the given multicast-group.
func (a *MulticastGroupAPI) Create(ctx context.Context, req *pb.CreateMulticastGroupRequest) (*pb.CreateMulticastGroupResponse, error) {
	if req.MulticastGroup == nil {
		return nil, grpc.Errorf(codes.InvalidArgument, "multicast_group must not be nil")
	}

	if err := a.validator.Validate(ctx,
		auth.ValidateMulticastGroupsAccess(auth.Create, req.MulticastGroup.ApplicationId)); err != nil {
		return nil, grpc.Errorf(codes.Unauthenticated, "authentication failed: %s", err)
	}

	app, err := storage.GetApplication(ctx, storage.DB(), req.MulticastGroup.ApplicationId)
	if err != nil {
		return nil, helpers.ErrToRPCError(err)
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
		Name:          req.MulticastGroup.Name,
		ApplicationID: req.MulticastGroup.ApplicationId,
		MulticastGroup: ns.MulticastGroup{
			McAddr:           mcAddr[:],
			McNwkSKey:        mcNwkSKey[:],
			GroupType:        ns.MulticastGroupType(req.MulticastGroup.GroupType),
			Dr:               req.MulticastGroup.Dr,
			Frequency:        req.MulticastGroup.Frequency,
			PingSlotPeriod:   req.MulticastGroup.PingSlotPeriod,
			ServiceProfileId: app.ServiceProfileID.Bytes(),
			RoutingProfileId: a.routingProfileID.Bytes(),
			FCnt:             req.MulticastGroup.FCnt,
		},
	}

	if err = mg.MCAppSKey.UnmarshalText([]byte(req.MulticastGroup.McAppSKey)); err != nil {
		return nil, grpc.Errorf(codes.InvalidArgument, "mc_app_s_key: %s", err)
	}

	if err = storage.Transaction(func(tx sqlx.Ext) error {
		if err := storage.CreateMulticastGroup(ctx, tx, &mg); err != nil {
			return helpers.ErrToRPCError(err)
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

	mg, err := storage.GetMulticastGroup(ctx, storage.DB(), mgID, false, false)
	if err != nil {
		return nil, helpers.ErrToRPCError(err)
	}

	var mcAddr lorawan.DevAddr
	var mcNwkSKey lorawan.AES128Key
	copy(mcAddr[:], mg.MulticastGroup.McAddr)
	copy(mcNwkSKey[:], mg.MulticastGroup.McNwkSKey)

	out := pb.GetMulticastGroupResponse{
		MulticastGroup: &pb.MulticastGroup{
			Id:             mgID.String(),
			Name:           mg.Name,
			McAddr:         mcAddr.String(),
			McNwkSKey:      mcNwkSKey.String(),
			McAppSKey:      mg.MCAppSKey.String(),
			FCnt:           mg.MulticastGroup.FCnt,
			GroupType:      pb.MulticastGroupType(mg.MulticastGroup.GroupType),
			Dr:             mg.MulticastGroup.Dr,
			Frequency:      mg.MulticastGroup.Frequency,
			PingSlotPeriod: mg.MulticastGroup.PingSlotPeriod,
			ApplicationId:  mg.ApplicationID,
		},
	}

	out.CreatedAt, err = ptypes.TimestampProto(mg.CreatedAt)
	if err != nil {
		return nil, helpers.ErrToRPCError(err)
	}

	out.UpdatedAt, err = ptypes.TimestampProto(mg.UpdatedAt)
	if err != nil {
		return nil, helpers.ErrToRPCError(err)
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

	mg, err := storage.GetMulticastGroup(ctx, storage.DB(), mgID, false, false)
	if err != nil {
		return nil, helpers.ErrToRPCError(err)
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
		GroupType:        ns.MulticastGroupType(req.MulticastGroup.GroupType),
		Dr:               req.MulticastGroup.Dr,
		Frequency:        req.MulticastGroup.Frequency,
		PingSlotPeriod:   req.MulticastGroup.PingSlotPeriod,
		ServiceProfileId: mg.MulticastGroup.ServiceProfileId,
		RoutingProfileId: mg.MulticastGroup.RoutingProfileId,
		FCnt:             req.MulticastGroup.FCnt,
	}

	if err = mg.MCAppSKey.UnmarshalText([]byte(req.MulticastGroup.McAppSKey)); err != nil {
		return nil, grpc.Errorf(codes.InvalidArgument, "mc_app_s_key: %s", err)
	}

	if err = storage.Transaction(func(tx sqlx.Ext) error {
		if err := storage.UpdateMulticastGroup(ctx, tx, &mg); err != nil {
			return helpers.ErrToRPCError(err)
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

	if err = storage.Transaction(func(tx sqlx.Ext) error {
		if err := storage.DeleteMulticastGroup(ctx, tx, mgID); err != nil {
			return helpers.ErrToRPCError(err)
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
		ApplicationID:  req.ApplicationId,
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

	// if app filter has been set, validate the client has access to the given application
	if req.ApplicationId != 0 {
		idFilter = true

		if err = a.validator.Validate(ctx,
			auth.ValidateApplicationAccess(req.ApplicationId, auth.Read)); err != nil {
			return nil, grpc.Errorf(codes.Unauthenticated, "authentication failed: %s", err)
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
		user, err := a.validator.GetUser(ctx)
		if err != nil {
			return nil, helpers.ErrToRPCError(err)
		}

		if !user.IsAdmin {
			return nil, grpc.Errorf(codes.Unauthenticated, "client must be global admin for unfiltered request")
		}
	}

	count, err := storage.GetMulticastGroupCount(ctx, storage.DB(), filters)
	if err != nil {
		return nil, helpers.ErrToRPCError(err)
	}

	items, err := storage.GetMulticastGroups(ctx, storage.DB(), filters)
	if err != nil {
		return nil, helpers.ErrToRPCError(err)
	}

	out := pb.ListMulticastGroupResponse{
		TotalCount: int64(count),
	}

	for _, item := range items {
		out.Result = append(out.Result, &pb.MulticastGroupListItem{
			Id:              item.ID.String(),
			Name:            item.Name,
			ApplicationName: item.ApplicationName,
			ApplicationId:   item.ApplicationID,
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

	// validate that the device is under the same app as the multicast-group
	dev, err := storage.GetDevice(ctx, storage.DB(), devEUI, false, true)
	if err != nil {
		return nil, helpers.ErrToRPCError(err)
	}

	mg, err := storage.GetMulticastGroup(ctx, storage.DB(), mgID, false, true)
	if err != nil {
		return nil, helpers.ErrToRPCError(err)
	}

	if dev.ApplicationID != mg.ApplicationID {
		return nil, grpc.Errorf(codes.FailedPrecondition, "device and multicast must be under the same application")
	}

	if err = storage.Transaction(func(tx sqlx.Ext) error {
		if err := storage.AddDeviceToMulticastGroup(ctx, tx, mgID, devEUI); err != nil {
			return helpers.ErrToRPCError(err)
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

	if err = storage.Transaction(func(tx sqlx.Ext) error {
		if err := storage.RemoveDeviceFromMulticastGroup(ctx, tx, mgID, devEUI); err != nil {
			return helpers.ErrToRPCError(err)
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

	if err = storage.Transaction(func(tx sqlx.Ext) error {
		var err error
		fCnt, err = multicast.Enqueue(ctx, tx, mgID, uint8(req.MulticastQueueItem.FPort), req.MulticastQueueItem.Data)
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

	n, err := storage.GetNetworkServerForMulticastGroupID(ctx, storage.DB(), mgID)
	if err != nil {
		return nil, helpers.ErrToRPCError(err)
	}

	nsClient, err := networkserver.GetPool().Get(n.Server, []byte(n.CACert), []byte(n.TLSCert), []byte(n.TLSKey))
	if err != nil {
		return nil, helpers.ErrToRPCError(err)
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

	queueItems, err := multicast.ListQueue(ctx, storage.DB(), mgID)
	if err != nil {
		return nil, helpers.ErrToRPCError(err)
	}

	var resp pb.ListMulticastGroupQueueItemsResponse
	for i := range queueItems {
		resp.MulticastQueueItems = append(resp.MulticastQueueItems, &queueItems[i])
	}

	return &resp, nil
}
