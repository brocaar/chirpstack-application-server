package api

import (
	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/jmoiron/sqlx"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

	pb "github.com/brocaar/lora-app-server/api"
	"github.com/brocaar/lora-app-server/internal/api/auth"
	"github.com/brocaar/lora-app-server/internal/config"
	"github.com/brocaar/lora-app-server/internal/storage"
)

// NetworkServerAPI exports the NetworkServer related functions.
type NetworkServerAPI struct {
	validator auth.Validator
}

// NewNetworkServerAPI creates a new NetworkServerAPI.
func NewNetworkServerAPI(validator auth.Validator) *NetworkServerAPI {
	return &NetworkServerAPI{
		validator: validator,
	}
}

// Create creates the given network-server.
func (a *NetworkServerAPI) Create(ctx context.Context, req *pb.CreateNetworkServerRequest) (*pb.CreateNetworkServerResponse, error) {
	if req.NetworkServer == nil {
		return nil, grpc.Errorf(codes.InvalidArgument, "network_server must not be nil")
	}

	if err := a.validator.Validate(ctx,
		auth.ValidateNetworkServersAccess(auth.Create, 0),
	); err != nil {
		return nil, grpc.Errorf(codes.Unauthenticated, "authentication failed: %s", err)
	}

	ns := storage.NetworkServer{
		Name:                        req.NetworkServer.Name,
		Server:                      req.NetworkServer.Server,
		CACert:                      req.NetworkServer.CaCert,
		TLSCert:                     req.NetworkServer.TlsCert,
		TLSKey:                      req.NetworkServer.TlsKey,
		RoutingProfileCACert:        req.NetworkServer.RoutingProfileCaCert,
		RoutingProfileTLSCert:       req.NetworkServer.RoutingProfileTlsCert,
		RoutingProfileTLSKey:        req.NetworkServer.RoutingProfileTlsKey,
		GatewayDiscoveryEnabled:     req.NetworkServer.GatewayDiscoveryEnabled,
		GatewayDiscoveryInterval:    int(req.NetworkServer.GatewayDiscoveryInterval),
		GatewayDiscoveryTXFrequency: int(req.NetworkServer.GatewayDiscoveryTxFrequency),
		GatewayDiscoveryDR:          int(req.NetworkServer.GatewayDiscoveryDr),
	}

	err := storage.Transaction(config.C.PostgreSQL.DB, func(tx sqlx.Ext) error {
		return storage.CreateNetworkServer(tx, &ns)
	})
	if err != nil {
		return nil, errToRPCError(err)
	}

	return &pb.CreateNetworkServerResponse{
		Id: ns.ID,
	}, nil
}

// Get returns the network-server matching the given id.
func (a *NetworkServerAPI) Get(ctx context.Context, req *pb.GetNetworkServerRequest) (*pb.GetNetworkServerResponse, error) {
	if err := a.validator.Validate(ctx,
		auth.ValidateNetworkServerAccess(auth.Read, req.Id),
	); err != nil {
		return nil, grpc.Errorf(codes.Unauthenticated, "authentication failed: %s", err)
	}

	n, err := storage.GetNetworkServer(config.C.PostgreSQL.DB, req.Id)
	if err != nil {
		return nil, errToRPCError(err)
	}

	var region string
	var version string

	nsClient, err := config.C.NetworkServer.Pool.Get(n.Server, []byte(n.CACert), []byte(n.TLSCert), []byte(n.TLSKey))
	if err == nil {
		resp, err := nsClient.GetVersion(context.Background(), &empty.Empty{})
		if err == nil {
			region = resp.Region.String()
			version = resp.Version
		}
	}

	resp := pb.GetNetworkServerResponse{
		NetworkServer: &pb.NetworkServer{
			Id:                          n.ID,
			Name:                        n.Name,
			Server:                      n.Server,
			CaCert:                      n.CACert,
			TlsCert:                     n.TLSCert,
			RoutingProfileCaCert:        n.RoutingProfileCACert,
			RoutingProfileTlsCert:       n.RoutingProfileTLSCert,
			GatewayDiscoveryEnabled:     n.GatewayDiscoveryEnabled,
			GatewayDiscoveryInterval:    uint32(n.GatewayDiscoveryInterval),
			GatewayDiscoveryTxFrequency: uint32(n.GatewayDiscoveryTXFrequency),
			GatewayDiscoveryDr:          uint32(n.GatewayDiscoveryDR),
		},
		Region:  region,
		Version: version,
	}

	resp.CreatedAt, err = ptypes.TimestampProto(n.CreatedAt)
	if err != nil {
		return nil, errToRPCError(err)
	}
	resp.UpdatedAt, err = ptypes.TimestampProto(n.UpdatedAt)
	if err != nil {
		return nil, errToRPCError(err)
	}

	return &resp, nil
}

// Update updates the given network-server.
func (a *NetworkServerAPI) Update(ctx context.Context, req *pb.UpdateNetworkServerRequest) (*empty.Empty, error) {
	if req.NetworkServer == nil {
		return nil, grpc.Errorf(codes.InvalidArgument, "network_server must not be nil")
	}

	if err := a.validator.Validate(ctx,
		auth.ValidateNetworkServerAccess(auth.Update, req.NetworkServer.Id),
	); err != nil {
		return nil, grpc.Errorf(codes.Unauthenticated, "authentication failed: %s", err)
	}

	ns, err := storage.GetNetworkServer(config.C.PostgreSQL.DB, req.NetworkServer.Id)
	if err != nil {
		return nil, errToRPCError(err)
	}

	ns.Name = req.NetworkServer.Name
	ns.Server = req.NetworkServer.Server
	ns.CACert = req.NetworkServer.CaCert
	ns.TLSCert = req.NetworkServer.TlsCert
	ns.RoutingProfileCACert = req.NetworkServer.RoutingProfileCaCert
	ns.RoutingProfileTLSCert = req.NetworkServer.RoutingProfileTlsCert
	ns.GatewayDiscoveryEnabled = req.NetworkServer.GatewayDiscoveryEnabled
	ns.GatewayDiscoveryInterval = int(req.NetworkServer.GatewayDiscoveryInterval)
	ns.GatewayDiscoveryTXFrequency = int(req.NetworkServer.GatewayDiscoveryTxFrequency)
	ns.GatewayDiscoveryDR = int(req.NetworkServer.GatewayDiscoveryDr)

	if req.NetworkServer.TlsKey != "" {
		ns.TLSKey = req.NetworkServer.TlsKey
	}
	if ns.TLSCert == "" {
		ns.TLSKey = ""
	}

	if req.NetworkServer.RoutingProfileTlsKey != "" {
		ns.RoutingProfileTLSKey = req.NetworkServer.RoutingProfileTlsKey
	}
	if ns.RoutingProfileTLSCert == "" {
		ns.RoutingProfileTLSKey = ""
	}

	err = storage.Transaction(config.C.PostgreSQL.DB, func(tx sqlx.Ext) error {
		return storage.UpdateNetworkServer(tx, &ns)
	})
	if err != nil {
		return nil, errToRPCError(err)
	}

	return &empty.Empty{}, nil
}

// Delete deletes the network-server matching the given id.
func (a *NetworkServerAPI) Delete(ctx context.Context, req *pb.DeleteNetworkServerRequest) (*empty.Empty, error) {
	if err := a.validator.Validate(ctx,
		auth.ValidateNetworkServerAccess(auth.Delete, req.Id),
	); err != nil {
		return nil, grpc.Errorf(codes.Unauthenticated, "authentication failed: %s", err)
	}

	err := storage.Transaction(config.C.PostgreSQL.DB, func(tx sqlx.Ext) error {
		return storage.DeleteNetworkServer(tx, req.Id)
	})
	if err != nil {
		return nil, errToRPCError(err)
	}

	return &empty.Empty{}, nil
}

// List lists the available network-servers.
func (a *NetworkServerAPI) List(ctx context.Context, req *pb.ListNetworkServerRequest) (*pb.ListNetworkServerResponse, error) {
	if err := a.validator.Validate(ctx,
		auth.ValidateNetworkServersAccess(auth.List, req.OrganizationId),
	); err != nil {
		return nil, grpc.Errorf(codes.Unauthenticated, "authentication failed: %s", err)
	}

	isAdmin, err := a.validator.GetIsAdmin(ctx)
	if err != nil {
		return nil, errToRPCError(err)
	}

	var count int
	var nss []storage.NetworkServer

	if req.OrganizationId == 0 {
		if isAdmin {
			count, err = storage.GetNetworkServerCount(config.C.PostgreSQL.DB)
			if err != nil {
				return nil, errToRPCError(err)
			}
			nss, err = storage.GetNetworkServers(config.C.PostgreSQL.DB, int(req.Limit), int(req.Offset))
			if err != nil {
				return nil, errToRPCError(err)
			}
		}
	} else {
		count, err = storage.GetNetworkServerCountForOrganizationID(config.C.PostgreSQL.DB, req.OrganizationId)
		if err != nil {
			return nil, errToRPCError(err)
		}
		nss, err = storage.GetNetworkServersForOrganizationID(config.C.PostgreSQL.DB, req.OrganizationId, int(req.Limit), int(req.Offset))
		if err != nil {
			return nil, errToRPCError(err)
		}
	}

	resp := pb.ListNetworkServerResponse{
		TotalCount: int64(count),
	}

	for _, ns := range nss {
		row := pb.NetworkServerListItem{
			Id:     ns.ID,
			Name:   ns.Name,
			Server: ns.Server,
		}

		row.CreatedAt, err = ptypes.TimestampProto(ns.CreatedAt)
		if err != nil {
			return nil, errToRPCError(err)
		}
		row.UpdatedAt, err = ptypes.TimestampProto(ns.UpdatedAt)
		if err != nil {
			return nil, errToRPCError(err)
		}

		resp.Result = append(resp.Result, &row)
	}

	return &resp, nil
}
