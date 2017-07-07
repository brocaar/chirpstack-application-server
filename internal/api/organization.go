package api

import (
	"time"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

	pb "github.com/brocaar/lora-app-server/api"
	"github.com/brocaar/lora-app-server/internal/api/auth"
	"github.com/brocaar/lora-app-server/internal/common"
	"github.com/brocaar/lora-app-server/internal/storage"
	"github.com/brocaar/loraserver/api/ns"
	"github.com/jmoiron/sqlx"
)

// OrganizationAPI exports the organization related functions.
type OrganizationAPI struct {
	ctx       common.Context
	validator auth.Validator
}

// NewOrganizationAPI creates a new OrganizationAPI.
func NewOrganizationAPI(ctx common.Context, validator auth.Validator) *OrganizationAPI {
	return &OrganizationAPI{
		ctx:       ctx,
		validator: validator,
	}
}

// Create creates the given organization.
func (a *OrganizationAPI) Create(ctx context.Context, req *pb.CreateOrganizationRequest) (*pb.CreateOrganizationResponse, error) {
	if err := a.validator.Validate(ctx,
		auth.ValidateOrganizationsAccess(auth.Create)); err != nil {
		return nil, grpc.Errorf(codes.Unauthenticated, "authentication failed: %s", err)
	}

	org := storage.Organization{
		Name:            req.Name,
		DisplayName:     req.DisplayName,
		CanHaveGateways: req.CanHaveGateways,
	}

	err := storage.CreateOrganization(a.ctx.DB, &org)
	if err != nil {
		return nil, errToRPCError(err)
	}

	return &pb.CreateOrganizationResponse{
		Id: org.ID,
	}, nil
}

// Get returns the organization matching the given ID.
func (a *OrganizationAPI) Get(ctx context.Context, req *pb.OrganizationRequest) (*pb.GetOrganizationResponse, error) {
	if err := a.validator.Validate(ctx,
		auth.ValidateOrganizationAccess(auth.Read, req.Id)); err != nil {
		return nil, grpc.Errorf(codes.Unauthenticated, "authentication failed: %s", err)
	}

	org, err := storage.GetOrganization(a.ctx.DB, req.Id)
	if err != nil {
		return nil, errToRPCError(err)
	}

	return &pb.GetOrganizationResponse{
		Id:              org.ID,
		Name:            org.Name,
		DisplayName:     org.DisplayName,
		CanHaveGateways: org.CanHaveGateways,
		CreatedAt:       org.CreatedAt.Format(time.RFC3339Nano),
		UpdatedAt:       org.UpdatedAt.Format(time.RFC3339Nano),
	}, nil
}

// List lists the organizations to which the user has access.
func (a *OrganizationAPI) List(ctx context.Context, req *pb.ListOrganizationRequest) (*pb.ListOrganizationResponse, error) {
	if err := a.validator.Validate(ctx,
		auth.ValidateOrganizationsAccess(auth.List)); err != nil {
		return nil, grpc.Errorf(codes.Unauthenticated, "authentication failed: %s", err)
	}

	isAdmin, err := a.validator.GetIsAdmin(ctx)
	if err != nil {
		return nil, errToRPCError(err)
	}

	var count int
	var orgs []storage.Organization

	if isAdmin {
		count, err = storage.GetOrganizationCount(a.ctx.DB, req.Search)
		if err != nil {
			return nil, errToRPCError(err)
		}

		orgs, err = storage.GetOrganizations(a.ctx.DB, int(req.Limit), int(req.Offset), req.Search)
		if err != nil {
			return nil, errToRPCError(err)
		}
	} else {
		username, err := a.validator.GetUsername(ctx)
		if err != nil {
			return nil, errToRPCError(err)
		}
		count, err = storage.GetOrganizationCountForUser(a.ctx.DB, username, req.Search)
		if err != nil {
			return nil, errToRPCError(err)
		}
		orgs, err = storage.GetOrganizationsForUser(a.ctx.DB, username, int(req.Limit), int(req.Offset), req.Search)
		if err != nil {
			return nil, errToRPCError(err)
		}
	}

	result := make([]*pb.GetOrganizationResponse, len(orgs))
	for i, org := range orgs {
		result[i] = &pb.GetOrganizationResponse{
			Id:              org.ID,
			Name:            org.Name,
			DisplayName:     org.DisplayName,
			CanHaveGateways: org.CanHaveGateways,
			CreatedAt:       org.CreatedAt.Format(time.RFC3339Nano),
			UpdatedAt:       org.UpdatedAt.Format(time.RFC3339Nano),
		}
	}

	return &pb.ListOrganizationResponse{
		TotalCount: int32(count),
		Result:     result,
	}, nil
}

// Update updates the given organization.
func (a *OrganizationAPI) Update(ctx context.Context, req *pb.UpdateOrganizationRequest) (*pb.OrganizationEmptyResponse, error) {
	if err := a.validator.Validate(ctx,
		auth.ValidateOrganizationAccess(auth.Update, req.Id)); err != nil {
		return nil, grpc.Errorf(codes.Unauthenticated, "authentication failed: %s", err)
	}

	isAdmin, err := a.validator.GetIsAdmin(ctx)
	if err != nil {
		return nil, errToRPCError(err)
	}

	org, err := storage.GetOrganization(a.ctx.DB, req.Id)
	if err != nil {
		return nil, errToRPCError(err)
	}

	org.Name = req.Name
	org.DisplayName = req.DisplayName
	if isAdmin {
		org.CanHaveGateways = req.CanHaveGateways
	}

	err = storage.UpdateOrganization(a.ctx.DB, &org)
	if err != nil {
		return nil, errToRPCError(err)
	}

	return &pb.OrganizationEmptyResponse{}, nil
}

// Delete deletes the organization matching the given ID.
func (a *OrganizationAPI) Delete(ctx context.Context, req *pb.OrganizationRequest) (*pb.OrganizationEmptyResponse, error) {
	if err := a.validator.Validate(ctx,
		auth.ValidateOrganizationAccess(auth.Delete, req.Id)); err != nil {
		return nil, grpc.Errorf(codes.Unauthenticated, "authentication failed: %s", err)
	}

	// deleting the organization will remove all gateways in the
	// LoRa App Server database, however we need to delete these gateways
	// also from the LoRa Server database.
	for {
		gws, err := storage.GetGatewaysForOrganizationID(a.ctx.DB, req.Id, 100, 0)
		if err != nil {
			return nil, errToRPCError(err)
		}

		if len(gws) == 0 {
			break
		}

		for _, gw := range gws {
			err = storage.Transaction(a.ctx.DB, func(tx *sqlx.Tx) error {
				err = storage.DeleteGateway(tx, gw.MAC)
				if err != nil {
					return errToRPCError(err)
				}

				_, err = a.ctx.NetworkServer.DeleteGateway(ctx, &ns.DeleteGatewayRequest{
					Mac: gw.MAC[:],
				})
				if err != nil && grpc.Code(err) != codes.NotFound {
					return err
				}
				return nil
			})
			if err != nil {
				return nil, err
			}
		}
	}

	err := storage.DeleteOrganization(a.ctx.DB, req.Id)
	if err != nil {
		return nil, errToRPCError(err)
	}
	return &pb.OrganizationEmptyResponse{}, nil
}

func (a *OrganizationAPI) ListUsers(ctx context.Context, req *pb.ListOrganizationUsersRequest) (*pb.ListOrganizationUsersResponse, error) {
	if err := a.validator.Validate(ctx,
		auth.ValidateOrganizationUsersAccess(auth.List, req.Id)); err != nil {
		return nil, grpc.Errorf(codes.Unauthenticated, "authentication failed: %s", err)
	}

	users, err := storage.GetOrganizationUsers(a.ctx.DB, req.Id, int(req.Limit), int(req.Offset))
	if err != nil {
		return nil, errToRPCError(err)
	}

	userCount, err := storage.GetOrganizationUserCount(a.ctx.DB, req.Id)
	if err != nil {
		return nil, errToRPCError(err)
	}

	result := make([]*pb.GetOrganizationUserResponse, len(users))
	for i, user := range users {
		result[i] = &pb.GetOrganizationUserResponse{
			Id:        user.UserID,
			Username:  user.Username,
			IsAdmin:   user.IsAdmin,
			CreatedAt: user.CreatedAt.Format(time.RFC3339Nano),
			UpdatedAt: user.UpdatedAt.Format(time.RFC3339Nano),
		}
	}

	return &pb.ListOrganizationUsersResponse{
		TotalCount: int32(userCount),
		Result:     result,
	}, nil
}

// Create creates the given organization-user link.
func (a *OrganizationAPI) AddUser(ctx context.Context, req *pb.OrganizationUserRequest) (*pb.OrganizationEmptyResponse, error) {
	if err := a.validator.Validate(ctx,
		auth.ValidateOrganizationUsersAccess(auth.Create, req.Id)); err != nil {
		return nil, grpc.Errorf(codes.Unauthenticated, "authentication failed: %s", err)
	}

	err := storage.CreateOrganizationUser(a.ctx.DB, req.Id, req.UserID, req.IsAdmin)
	if err != nil {
		return nil, errToRPCError(err)
	}

	return &pb.OrganizationEmptyResponse{}, nil
}

// Update updates the given user.
func (a *OrganizationAPI) UpdateUser(ctx context.Context, req *pb.OrganizationUserRequest) (*pb.OrganizationEmptyResponse, error) {
	if err := a.validator.Validate(ctx,
		auth.ValidateOrganizationUserAccess(auth.Update, req.Id, req.UserID)); err != nil {
		return nil, grpc.Errorf(codes.Unauthenticated, "authentication failed: %s", err)
	}

	err := storage.UpdateOrganizationUser(a.ctx.DB, req.Id, req.UserID, req.IsAdmin)
	if err != nil {
		return nil, errToRPCError(err)
	}

	return &pb.OrganizationEmptyResponse{}, nil
}

// Delete deletes the given user from the organization.
func (a *OrganizationAPI) DeleteUser(ctx context.Context, req *pb.DeleteOrganizationUserRequest) (*pb.OrganizationEmptyResponse, error) {
	if err := a.validator.Validate(ctx,
		auth.ValidateOrganizationUserAccess(auth.Delete, req.Id, req.UserID)); err != nil {
		return nil, grpc.Errorf(codes.Unauthenticated, "authentication failed: %s", err)
	}

	err := storage.DeleteOrganizationUser(a.ctx.DB, req.Id, req.UserID)
	if err != nil {
		return nil, errToRPCError(err)
	}

	return &pb.OrganizationEmptyResponse{}, nil
}

// GetUser returns the user details for the given user ID.
func (a *OrganizationAPI) GetUser(ctx context.Context, req *pb.GetOrganizationUserRequest) (*pb.GetOrganizationUserResponse, error) {
	if err := a.validator.Validate(ctx,
		auth.ValidateOrganizationUserAccess(auth.Read, req.Id, req.UserID)); err != nil {
		return nil, grpc.Errorf(codes.Unauthenticated, "authentication failed: %s", err)
	}

	user, err := storage.GetOrganizationUser(a.ctx.DB, req.Id, req.UserID)
	if err != nil {
		return nil, errToRPCError(err)
	}

	return &pb.GetOrganizationUserResponse{
		Id:        user.UserID,
		Username:  user.Username,
		IsAdmin:   user.IsAdmin,
		CreatedAt: user.CreatedAt.Format(time.RFC3339Nano),
		UpdatedAt: user.UpdatedAt.Format(time.RFC3339Nano),
	}, nil
}
