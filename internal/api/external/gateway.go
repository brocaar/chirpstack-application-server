package external

import (
	"database/sql"
	"strconv"
	"strings"

	"github.com/gofrs/uuid"
	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq/hstore"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

	pb "github.com/brocaar/chirpstack-api/go/v3/as/external/api"
	"github.com/brocaar/chirpstack-api/go/v3/common"
	"github.com/brocaar/chirpstack-api/go/v3/ns"
	"github.com/brocaar/chirpstack-application-server/internal/api/external/auth"
	"github.com/brocaar/chirpstack-application-server/internal/api/helpers"
	"github.com/brocaar/chirpstack-application-server/internal/backend/networkserver"
	"github.com/brocaar/chirpstack-application-server/internal/storage"
	"github.com/brocaar/lorawan"
)

// GatewayAPI exports the Gateway related functions.
type GatewayAPI struct {
	validator auth.Validator
}

// NewGatewayAPI creates a new GatewayAPI.
func NewGatewayAPI(validator auth.Validator) *GatewayAPI {
	return &GatewayAPI{
		validator: validator,
	}
}

// Create creates the given gateway.
func (a *GatewayAPI) Create(ctx context.Context, req *pb.CreateGatewayRequest) (*empty.Empty, error) {
	if req.Gateway == nil {
		return nil, grpc.Errorf(codes.InvalidArgument, "gateway must not be nil")
	}

	if req.Gateway.Location == nil {
		return nil, grpc.Errorf(codes.InvalidArgument, "gateway.location must not be nil")
	}

	// validate that user has access to organization
	err := a.validator.Validate(ctx, auth.ValidateGatewaysAccess(auth.Create, req.Gateway.OrganizationId))
	if err != nil {
		return nil, grpc.Errorf(codes.Unauthenticated, "authentication failed: %s", err)
	}

	// also validate that the network-server is accessible for the given organization
	err = a.validator.Validate(ctx, auth.ValidateOrganizationNetworkServerAccess(auth.Read, req.Gateway.OrganizationId, req.Gateway.NetworkServerId))
	if err != nil {
		return nil, grpc.Errorf(codes.Unauthenticated, "authentication failed: %s", err)
	}

	// gateway id
	var mac lorawan.EUI64
	if err := mac.UnmarshalText([]byte(req.Gateway.Id)); err != nil {
		return nil, grpc.Errorf(codes.InvalidArgument, "bad gateway mac: %s", err)
	}

	// gateway profile id (if any)
	var gpID uuid.UUID
	if req.Gateway.GatewayProfileId != "" {
		gpID, err = uuid.FromString(req.Gateway.GatewayProfileId)
		if err != nil {
			return nil, grpc.Errorf(codes.InvalidArgument, err.Error())
		}
	}

	// service-profile id (if any)
	var spID uuid.UUID
	if req.Gateway.ServiceProfileId != "" {
		spID, err = uuid.FromString(req.Gateway.ServiceProfileId)
		if err != nil {
			return nil, grpc.Errorf(codes.InvalidArgument, err.Error())
		}

		// validate that the service-profile has the same organization id
		sp, err := storage.GetServiceProfile(ctx, storage.DB(), spID, true)
		if err != nil {
			return nil, helpers.ErrToRPCError(err)
		}

		if sp.OrganizationID != req.Gateway.OrganizationId {
			return nil, grpc.Errorf(codes.InvalidArgument, "service-profile must be under the same organization")
		}
	}

	// Set CreateGatewayRequest struct.
	createReq := ns.CreateGatewayRequest{
		Gateway: &ns.Gateway{
			Id:               mac[:],
			Location:         req.Gateway.Location,
			RoutingProfileId: applicationServerID.Bytes(),
		},
	}
	if gpID != uuid.Nil {
		createReq.Gateway.GatewayProfileId = gpID.Bytes()
	}
	if spID != uuid.Nil {
		createReq.Gateway.ServiceProfileId = spID.Bytes()
	}

	for _, board := range req.Gateway.Boards {
		var gwBoard ns.GatewayBoard

		if board.FpgaId != "" {
			var fpgaID lorawan.EUI64
			if err := fpgaID.UnmarshalText([]byte(board.FpgaId)); err != nil {
				return nil, grpc.Errorf(codes.InvalidArgument, "fpga_id: %s", err)
			}
			gwBoard.FpgaId = fpgaID[:]
		}

		if board.FineTimestampKey != "" {
			var key lorawan.AES128Key
			if err := key.UnmarshalText([]byte(board.FineTimestampKey)); err != nil {
				return nil, grpc.Errorf(codes.InvalidArgument, "fine_timestamp_key: %s", err)
			}
			gwBoard.FineTimestampKey = key[:]
		}

		createReq.Gateway.Boards = append(createReq.Gateway.Boards, &gwBoard)
	}
	tags := hstore.Hstore{
		Map: make(map[string]sql.NullString),
	}
	for k, v := range req.Gateway.Tags {
		tags.Map[k] = sql.NullString{Valid: true, String: v}
	}

	// A transaction is needed as:
	//  * A remote gRPC call is performed and in case of error, we want to
	//    rollback the transaction.
	//  * We want to lock the organization so that we can validate the
	//    max gateway count.
	err = storage.Transaction(func(tx sqlx.Ext) error {
		org, err := storage.GetOrganization(ctx, tx, req.Gateway.OrganizationId, true)
		if err != nil {
			return helpers.ErrToRPCError(err)
		}

		// Validate max. gateway count when != 0.
		if org.MaxGatewayCount != 0 {
			count, err := storage.GetGatewayCount(ctx, tx, storage.GatewayFilters{OrganizationID: req.Gateway.OrganizationId})
			if err != nil {
				return helpers.ErrToRPCError(err)
			}

			if count >= org.MaxGatewayCount {
				return helpers.ErrToRPCError(storage.ErrOrganizationMaxGatewayCount)
			}
		}

		gw := storage.Gateway{
			MAC:             mac,
			Name:            req.Gateway.Name,
			Description:     req.Gateway.Description,
			OrganizationID:  req.Gateway.OrganizationId,
			Ping:            req.Gateway.DiscoveryEnabled,
			NetworkServerID: req.Gateway.NetworkServerId,
			Latitude:        req.Gateway.Location.Latitude,
			Longitude:       req.Gateway.Location.Longitude,
			Altitude:        req.Gateway.Location.Altitude,
			Tags:            tags,
		}

		if gpID != uuid.Nil {
			gw.GatewayProfileID = &gpID
		}

		if spID != uuid.Nil {
			gw.ServiceProfileID = &spID
		}

		err = storage.CreateGateway(ctx, tx, &gw)
		if err != nil {
			return helpers.ErrToRPCError(err)
		}

		n, err := storage.GetNetworkServer(ctx, tx, req.Gateway.NetworkServerId)
		if err != nil {
			return helpers.ErrToRPCError(err)
		}

		nsClient, err := networkserver.GetPool().Get(n.Server, []byte(n.CACert), []byte(n.TLSCert), []byte(n.TLSKey))
		if err != nil {
			return helpers.ErrToRPCError(err)
		}

		_, err = nsClient.CreateGateway(ctx, &createReq)
		if err != nil && grpc.Code(err) != codes.AlreadyExists {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return &empty.Empty{}, nil
}

// Get returns the gateway matching the given Mac.
func (a *GatewayAPI) Get(ctx context.Context, req *pb.GetGatewayRequest) (*pb.GetGatewayResponse, error) {
	var mac lorawan.EUI64
	if err := mac.UnmarshalText([]byte(req.Id)); err != nil {
		return nil, grpc.Errorf(codes.InvalidArgument, "bad gateway mac: %s", err)
	}

	err := a.validator.Validate(ctx, auth.ValidateGatewayAccess(auth.Read, mac))
	if err != nil {
		return nil, grpc.Errorf(codes.Unauthenticated, "authentication failed: %s", err)
	}

	gw, err := storage.GetGateway(ctx, storage.DB(), mac, false)
	if err != nil {
		return nil, helpers.ErrToRPCError(err)
	}

	n, err := storage.GetNetworkServer(ctx, storage.DB(), gw.NetworkServerID)
	if err != nil {
		return nil, helpers.ErrToRPCError(err)
	}

	nsClient, err := networkserver.GetPool().Get(n.Server, []byte(n.CACert), []byte(n.TLSCert), []byte(n.TLSKey))
	if err != nil {
		return nil, helpers.ErrToRPCError(err)
	}

	getResp, err := nsClient.GetGateway(ctx, &ns.GetGatewayRequest{
		Id: mac[:],
	})
	if err != nil {
		return nil, err
	}

	resp := pb.GetGatewayResponse{
		Gateway: &pb.Gateway{
			Id:               mac.String(),
			Name:             gw.Name,
			Description:      gw.Description,
			OrganizationId:   gw.OrganizationID,
			DiscoveryEnabled: gw.Ping,
			Location: &common.Location{
				Latitude:  gw.Latitude,
				Longitude: gw.Longitude,
				Altitude:  gw.Altitude,
			},
			NetworkServerId: gw.NetworkServerID,
			Tags:            make(map[string]string),
			Metadata:        make(map[string]string),
		},
	}

	resp.CreatedAt, err = ptypes.TimestampProto(gw.CreatedAt)
	if err != nil {
		return nil, helpers.ErrToRPCError(err)
	}
	resp.UpdatedAt, err = ptypes.TimestampProto(gw.UpdatedAt)
	if err != nil {
		return nil, helpers.ErrToRPCError(err)
	}

	if gw.FirstSeenAt != nil {
		resp.FirstSeenAt, err = ptypes.TimestampProto(*gw.FirstSeenAt)
		if err != nil {
			return nil, helpers.ErrToRPCError(err)
		}
	}

	if gw.LastSeenAt != nil {
		resp.LastSeenAt, err = ptypes.TimestampProto(*gw.LastSeenAt)
		if err != nil {
			return nil, helpers.ErrToRPCError(err)
		}
	}

	if len(getResp.Gateway.GatewayProfileId) != 0 {
		gpID, err := uuid.FromBytes(getResp.Gateway.GatewayProfileId)
		if err != nil {
			return nil, helpers.ErrToRPCError(err)
		}
		resp.Gateway.GatewayProfileId = gpID.String()
	}

	if len(getResp.Gateway.ServiceProfileId) != 0 {
		spID, err := uuid.FromBytes(getResp.Gateway.ServiceProfileId)
		if err != nil {
			return nil, helpers.ErrToRPCError(err)
		}
		resp.Gateway.ServiceProfileId = spID.String()
	}

	for i := range getResp.Gateway.Boards {
		var gwBoard pb.GatewayBoard

		if len(getResp.Gateway.Boards[i].FpgaId) != 0 {
			var fpgaID lorawan.EUI64
			copy(fpgaID[:], getResp.Gateway.Boards[i].FpgaId)
			gwBoard.FpgaId = fpgaID.String()
		}

		if len(getResp.Gateway.Boards[i].FineTimestampKey) != 0 {
			var key lorawan.AES128Key
			copy(key[:], getResp.Gateway.Boards[i].FineTimestampKey)
			gwBoard.FineTimestampKey = key.String()
		}

		resp.Gateway.Boards = append(resp.Gateway.Boards, &gwBoard)
	}

	for k, v := range gw.Tags.Map {
		resp.Gateway.Tags[k] = v.String
	}
	for k, v := range gw.Metadata.Map {
		resp.Gateway.Metadata[k] = v.String
	}

	return &resp, err
}

// List lists the gateways.
func (a *GatewayAPI) List(ctx context.Context, req *pb.ListGatewayRequest) (*pb.ListGatewayResponse, error) {
	err := a.validator.Validate(ctx, auth.ValidateGatewaysAccess(auth.List, req.OrganizationId))
	if err != nil {
		return nil, grpc.Errorf(codes.Unauthenticated, "authentication failed: %s", err)
	}

	filters := storage.GatewayFilters{
		Search:         req.Search,
		Limit:          int(req.Limit),
		Offset:         int(req.Offset),
		OrganizationID: req.OrganizationId,
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

		// Filter on username when OrganizationID is not set and the user is
		// not a global admin.
		if !user.IsAdmin && filters.OrganizationID == 0 {
			filters.UserID = user.ID
		}

	case auth.SubjectAPIKey:
		// Nothing to do as the validator function already validated that the
		// API Key has access to the given OrganizationID.
	default:
		return nil, grpc.Errorf(codes.Unauthenticated, "invalid token subject: %s", err)
	}

	count, err := storage.GetGatewayCount(ctx, storage.DB(), filters)
	if err != nil {
		return nil, helpers.ErrToRPCError(err)
	}

	gws, err := storage.GetGateways(ctx, storage.DB(), filters)
	if err != nil {
		return nil, helpers.ErrToRPCError(err)
	}

	resp := pb.ListGatewayResponse{
		TotalCount: int64(count),
	}

	for _, gw := range gws {
		row := pb.GatewayListItem{
			Id:                gw.MAC.String(),
			Name:              gw.Name,
			Description:       gw.Description,
			OrganizationId:    gw.OrganizationID,
			NetworkServerId:   gw.NetworkServerID,
			NetworkServerName: gw.NetworkServerName,
			Location: &common.Location{
				Latitude:  gw.Latitude,
				Longitude: gw.Longitude,
				Altitude:  gw.Altitude,
			},
		}

		row.CreatedAt, err = ptypes.TimestampProto(gw.CreatedAt)
		if err != nil {
			return nil, helpers.ErrToRPCError(err)
		}
		row.UpdatedAt, err = ptypes.TimestampProto(gw.UpdatedAt)
		if err != nil {
			return nil, helpers.ErrToRPCError(err)
		}

		if gw.FirstSeenAt != nil {
			row.FirstSeenAt, err = ptypes.TimestampProto(*gw.FirstSeenAt)
			if err != nil {
				return nil, helpers.ErrToRPCError(err)
			}
		}
		if gw.LastSeenAt != nil {
			row.LastSeenAt, err = ptypes.TimestampProto(*gw.LastSeenAt)
			if err != nil {
				return nil, helpers.ErrToRPCError(err)
			}
		}

		resp.Result = append(resp.Result, &row)
	}

	return &resp, nil
}

// Update updates the given gateway.
func (a *GatewayAPI) Update(ctx context.Context, req *pb.UpdateGatewayRequest) (*empty.Empty, error) {
	if req.Gateway == nil {
		return nil, grpc.Errorf(codes.InvalidArgument, "gateway must not be nil")
	}

	if req.Gateway.Location == nil {
		return nil, grpc.Errorf(codes.InvalidArgument, "gateway.location must not be nil")
	}

	var mac lorawan.EUI64
	if err := mac.UnmarshalText([]byte(req.Gateway.Id)); err != nil {
		return nil, grpc.Errorf(codes.InvalidArgument, "bad gateway mac: %s", err)
	}

	err := a.validator.Validate(ctx, auth.ValidateGatewayAccess(auth.Update, mac))
	if err != nil {
		return nil, grpc.Errorf(codes.Unauthenticated, "authentication failed: %s", err)
	}

	tags := hstore.Hstore{
		Map: make(map[string]sql.NullString),
	}
	for k, v := range req.Gateway.Tags {
		tags.Map[k] = sql.NullString{Valid: true, String: v}
	}

	var gpID uuid.UUID
	if req.Gateway.GatewayProfileId != "" {
		gpID, err = uuid.FromString(req.Gateway.GatewayProfileId)
		if err != nil {
			return nil, grpc.Errorf(codes.InvalidArgument, err.Error())
		}
	}

	var spID uuid.UUID
	if req.Gateway.ServiceProfileId != "" {
		spID, err = uuid.FromString(req.Gateway.ServiceProfileId)
		if err != nil {
			return nil, grpc.Errorf(codes.InvalidArgument, err.Error())
		}
	}

	err = storage.Transaction(func(tx sqlx.Ext) error {
		gw, err := storage.GetGateway(ctx, tx, mac, true)
		if err != nil {
			return helpers.ErrToRPCError(err)
		}

		// validate that the service-profile has the same organization id
		if spID != uuid.Nil {
			sp, err := storage.GetServiceProfile(ctx, storage.DB(), spID, true)
			if err != nil {
				return helpers.ErrToRPCError(err)
			}

			if sp.OrganizationID != gw.OrganizationID {
				return grpc.Errorf(codes.InvalidArgument, "service-profile must be under the same organization")
			}
		}

		gw.Name = req.Gateway.Name
		gw.Description = req.Gateway.Description
		gw.Ping = req.Gateway.DiscoveryEnabled
		gw.Latitude = req.Gateway.Location.Latitude
		gw.Longitude = req.Gateway.Location.Longitude
		gw.Altitude = req.Gateway.Location.Altitude
		gw.Tags = tags
		gw.GatewayProfileID = nil
		gw.ServiceProfileID = nil
		if gpID != uuid.Nil {
			gw.GatewayProfileID = &gpID
		}
		if spID != uuid.Nil {
			gw.ServiceProfileID = &spID
		}

		err = storage.UpdateGateway(ctx, tx, &gw)
		if err != nil {
			return helpers.ErrToRPCError(err)
		}

		updateReq := ns.UpdateGatewayRequest{
			Gateway: &ns.Gateway{
				Id:       mac[:],
				Location: req.Gateway.Location,
			},
		}

		if gpID != uuid.Nil {
			updateReq.Gateway.GatewayProfileId = gpID.Bytes()
		}

		if spID != uuid.Nil {
			updateReq.Gateway.ServiceProfileId = spID.Bytes()
		}

		for _, board := range req.Gateway.Boards {
			var gwBoard ns.GatewayBoard

			if board.FpgaId != "" {
				var fpgaID lorawan.EUI64
				if err := fpgaID.UnmarshalText([]byte(board.FpgaId)); err != nil {
					return grpc.Errorf(codes.InvalidArgument, "fpga_id: %s", err)
				}
				gwBoard.FpgaId = fpgaID[:]
			}

			if board.FineTimestampKey != "" {
				var key lorawan.AES128Key
				if err := key.UnmarshalText([]byte(board.FineTimestampKey)); err != nil {
					return grpc.Errorf(codes.InvalidArgument, "fine_timestamp_key: %s", err)
				}
				gwBoard.FineTimestampKey = key[:]
			}

			updateReq.Gateway.Boards = append(updateReq.Gateway.Boards, &gwBoard)
		}

		n, err := storage.GetNetworkServer(ctx, tx, gw.NetworkServerID)
		if err != nil {
			return helpers.ErrToRPCError(err)
		}

		nsClient, err := networkserver.GetPool().Get(n.Server, []byte(n.CACert), []byte(n.TLSCert), []byte(n.TLSKey))
		if err != nil {
			return helpers.ErrToRPCError(err)
		}

		_, err = nsClient.UpdateGateway(ctx, &updateReq)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return &empty.Empty{}, nil
}

// Delete deletes the gateway matching the given ID.
func (a *GatewayAPI) Delete(ctx context.Context, req *pb.DeleteGatewayRequest) (*empty.Empty, error) {
	var mac lorawan.EUI64
	if err := mac.UnmarshalText([]byte(req.Id)); err != nil {
		return nil, grpc.Errorf(codes.InvalidArgument, "bad gateway mac: %s", err)
	}

	err := a.validator.Validate(ctx, auth.ValidateGatewayAccess(auth.Delete, mac))
	if err != nil {
		return nil, grpc.Errorf(codes.Unauthenticated, "authentication failed: %s", err)
	}

	err = storage.Transaction(func(tx sqlx.Ext) error {
		err = storage.DeleteGateway(ctx, tx, mac)
		if err != nil {
			return helpers.ErrToRPCError(err)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return &empty.Empty{}, nil
}

// GenerateGatewayClientCertificate returns a TLS certificate for gateway authentication / authorization.
func (a *GatewayAPI) GenerateGatewayClientCertificate(ctx context.Context, req *pb.GenerateGatewayClientCertificateRequest) (*pb.GenerateGatewayClientCertificateResponse, error) {
	var id lorawan.EUI64
	if err := id.UnmarshalText([]byte(req.GatewayId)); err != nil {
		return nil, grpc.Errorf(codes.InvalidArgument, "gateway id: %s", err)
	}

	err := a.validator.Validate(ctx, auth.ValidateGatewayAccess(auth.Update, id))
	if err != nil {
		return nil, grpc.Errorf(codes.Unauthenticated, "authentication failed: %s", err)
	}

	gw, err := storage.GetGateway(ctx, storage.DB(), id, false)
	if err != nil {
		return nil, helpers.ErrToRPCError(err)
	}

	n, err := storage.GetNetworkServer(ctx, storage.DB(), gw.NetworkServerID)
	if err != nil {
		return nil, helpers.ErrToRPCError(err)
	}

	nsClient, err := networkserver.GetPool().Get(n.Server, []byte(n.CACert), []byte(n.TLSCert), []byte(n.TLSKey))
	if err != nil {
		return nil, helpers.ErrToRPCError(err)
	}

	resp, err := nsClient.GenerateGatewayClientCertificate(ctx, &ns.GenerateGatewayClientCertificateRequest{
		Id: id[:],
	})
	if err != nil {
		return nil, helpers.ErrToRPCError(err)
	}

	return &pb.GenerateGatewayClientCertificateResponse{
		ExpiresAt: resp.ExpiresAt,
		TlsCert:   string(resp.TlsCert),
		TlsKey:    string(resp.TlsKey),
		CaCert:    string(resp.CaCert),
	}, nil
}

// GetStats gets the gateway statistics for the gateway with the given Mac.
func (a *GatewayAPI) GetStats(ctx context.Context, req *pb.GetGatewayStatsRequest) (*pb.GetGatewayStatsResponse, error) {
	var gatewayID lorawan.EUI64
	if err := gatewayID.UnmarshalText([]byte(req.GatewayId)); err != nil {
		return nil, grpc.Errorf(codes.InvalidArgument, "bad gateway mac: %s", err)
	}

	err := a.validator.Validate(ctx, auth.ValidateGatewayAccess(auth.Read, gatewayID))
	if err != nil {
		return nil, grpc.Errorf(codes.Unauthenticated, "authentication failed: %s", err)
	}

	start, err := ptypes.Timestamp(req.StartTimestamp)
	if err != nil {
		return nil, grpc.Errorf(codes.InvalidArgument, err.Error())
	}

	end, err := ptypes.Timestamp(req.EndTimestamp)
	if err != nil {
		return nil, grpc.Errorf(codes.InvalidArgument, err.Error())
	}

	_, ok := ns.AggregationInterval_value[strings.ToUpper(req.Interval)]
	if !ok {
		return nil, grpc.Errorf(codes.InvalidArgument, "bad interval: %s", req.Interval)
	}

	metrics, err := storage.GetMetrics(ctx, storage.AggregationInterval(strings.ToUpper(req.Interval)), "gw:"+gatewayID.String(), start, end)
	if err != nil {
		return nil, helpers.ErrToRPCError(err)
	}

	result := make([]*pb.GatewayStats, len(metrics))
	for i, m := range metrics {
		result[i] = &pb.GatewayStats{
			RxPacketsReceived:     int32(m.Metrics["rx_count"]),
			RxPacketsReceivedOk:   int32(m.Metrics["rx_ok_count"]),
			TxPacketsReceived:     int32(m.Metrics["tx_count"]),
			TxPacketsEmitted:      int32(m.Metrics["tx_ok_count"]),
			TxPacketsPerDr:        make(map[uint32]uint32),
			RxPacketsPerDr:        make(map[uint32]uint32),
			TxPacketsPerFrequency: make(map[uint32]uint32),
			RxPacketsPerFrequency: make(map[uint32]uint32),
			TxPacketsPerStatus:    make(map[string]uint32),
		}

		result[i].Timestamp, err = ptypes.TimestampProto(m.Time)
		if err != nil {
			return nil, helpers.ErrToRPCError(err)
		}

		for k, v := range m.Metrics {
			if strings.HasPrefix(k, "tx_freq_") {
				if freq, err := strconv.ParseUint(strings.TrimPrefix(k, "tx_freq_"), 10, 32); err == nil {
					result[i].TxPacketsPerFrequency[uint32(freq)] = uint32(v)
				}
			}

			if strings.HasPrefix(k, "rx_freq_") {
				if freq, err := strconv.ParseUint(strings.TrimPrefix(k, "rx_freq_"), 10, 32); err == nil {
					result[i].RxPacketsPerFrequency[uint32(freq)] = uint32(v)
				}
			}

			if strings.HasPrefix(k, "tx_dr_") {
				if freq, err := strconv.ParseUint(strings.TrimPrefix(k, "tx_dr_"), 10, 32); err == nil {
					result[i].TxPacketsPerDr[uint32(freq)] = uint32(v)
				}
			}

			if strings.HasPrefix(k, "rx_dr_") {
				if freq, err := strconv.ParseUint(strings.TrimPrefix(k, "rx_dr_"), 10, 32); err == nil {
					result[i].RxPacketsPerDr[uint32(freq)] = uint32(v)
				}
			}

			if strings.HasPrefix(k, "tx_status_") {
				status := strings.TrimPrefix(k, "tx_status_")
				result[i].TxPacketsPerStatus[status] = uint32(v)
			}
		}
	}

	return &pb.GetGatewayStatsResponse{
		Result: result,
	}, nil
}

// GetLastPing returns the last emitted ping and gateways receiving this ping.
func (a *GatewayAPI) GetLastPing(ctx context.Context, req *pb.GetLastPingRequest) (*pb.GetLastPingResponse, error) {
	var mac lorawan.EUI64
	if err := mac.UnmarshalText([]byte(req.GatewayId)); err != nil {
		return nil, grpc.Errorf(codes.InvalidArgument, "bad gateway mac: %s", err)
	}

	err := a.validator.Validate(ctx, auth.ValidateGatewayAccess(auth.Read, mac))
	if err != nil {
		return nil, grpc.Errorf(codes.Unauthenticated, "authentication failed: %s", err)
	}

	ping, pingRX, err := storage.GetLastGatewayPingAndRX(ctx, storage.DB(), mac)
	if err != nil {
		return nil, helpers.ErrToRPCError(err)
	}

	resp := pb.GetLastPingResponse{
		Frequency: uint32(ping.Frequency),
		Dr:        uint32(ping.DR),
	}

	resp.CreatedAt, err = ptypes.TimestampProto(ping.CreatedAt)
	if err != nil {
		return nil, helpers.ErrToRPCError(err)
	}

	for _, rx := range pingRX {
		resp.PingRx = append(resp.PingRx, &pb.PingRX{
			GatewayId: rx.GatewayMAC.String(),
			Rssi:      int32(rx.RSSI),
			LoraSnr:   rx.LoRaSNR,
			Latitude:  rx.Location.Latitude,
			Longitude: rx.Location.Longitude,
			Altitude:  rx.Altitude,
		})
	}

	return &resp, nil
}

// StreamFrameLogs streams the uplink and downlink frame-logs for the given mac.
// Note: these are the raw LoRaWAN frames and this endpoint is intended for debugging.
func (a *GatewayAPI) StreamFrameLogs(req *pb.StreamGatewayFrameLogsRequest, srv pb.GatewayService_StreamFrameLogsServer) error {
	var mac lorawan.EUI64

	if err := mac.UnmarshalText([]byte(req.GatewayId)); err != nil {
		return grpc.Errorf(codes.InvalidArgument, "mac: %s", err)
	}

	err := a.validator.Validate(srv.Context(), auth.ValidateGatewayAccess(auth.Read, mac))
	if err != nil {
		return grpc.Errorf(codes.Unauthenticated, "authentication failed: %s", err)
	}

	n, err := storage.GetNetworkServerForGatewayMAC(srv.Context(), storage.DB(), mac)
	if err != nil {
		return helpers.ErrToRPCError(err)
	}

	nsClient, err := networkserver.GetPool().Get(n.Server, []byte(n.CACert), []byte(n.TLSCert), []byte(n.TLSKey))
	if err != nil {
		return helpers.ErrToRPCError(err)
	}

	streamClient, err := nsClient.StreamFrameLogsForGateway(srv.Context(), &ns.StreamFrameLogsForGatewayRequest{
		GatewayId: mac[:],
	})
	if err != nil {
		return err
	}

	for {
		resp, err := streamClient.Recv()
		if err != nil {
			return err
		}

		up, down, err := convertUplinkAndDownlinkFrames(resp.GetUplinkFrameSet(), resp.GetDownlinkFrame(), false)
		if err != nil {
			return helpers.ErrToRPCError(err)
		}

		var frameResp pb.StreamGatewayFrameLogsResponse
		if up != nil {
			up.PublishedAt = resp.GetUplinkFrameSet().GetPublishedAt()
			frameResp.Frame = &pb.StreamGatewayFrameLogsResponse_UplinkFrame{
				UplinkFrame: up,
			}
		}

		if down != nil {
			down.PublishedAt = resp.GetDownlinkFrame().GetPublishedAt()
			frameResp.Frame = &pb.StreamGatewayFrameLogsResponse_DownlinkFrame{
				DownlinkFrame: down,
			}
		}

		err = srv.Send(&frameResp)
		if err != nil {
			return err
		}
	}
}
