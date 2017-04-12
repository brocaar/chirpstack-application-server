package api

import (
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

	pb "github.com/brocaar/lora-app-server/api"
	"github.com/brocaar/lora-app-server/internal/api/auth"
	"github.com/brocaar/lora-app-server/internal/common"
	"github.com/brocaar/lora-app-server/internal/storage"
)

// OrganizationAPI exports the User related functions.
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
func (a *OrganizationAPI) Create(ctx context.Context, req *pb.AddOrganizationRequest) (*pb.OrganizationEmptyResponse, error) {
	if err := a.validator.Validate(ctx,
		auth.ValidateOrganizationsAccess(auth.Create)); err != nil {
		return nil, grpc.Errorf(codes.Unauthenticated, "authentication failed: %s", err)
	}

	org := storage.Organization{
		Name:   	 	 req.Name,
		DisplayName: 	 req.DisplayName,
		CanHaveGateways: req.CanHaveGateways,
	}

	err := storage.CreateOrganization(a.ctx.DB, &org)
	if err != nil {
		return nil, errToRPCError(err)
	}

	return &pb.OrganizationEmptyResponse{}, nil
}

// Get returns the user matching the given ID.
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
		Id:          	 org.ID,
		Name:        	 org.Name,
		DisplayName: 	 org.DisplayName,
		CanHaveGateways: org.CanHaveGateways,
		CreatedAt:  	 org.CreatedAt.String(),
		UpdatedAt:  	 org.UpdatedAt.String(),
	}, nil
}

// List lists the users.
func (a *OrganizationAPI) List(ctx context.Context, req *pb.ListOrganizationRequest) (*pb.ListOrganizationResponse, error) {
	if err := a.validator.Validate(ctx,
		auth.ValidateOrganizationsAccess(auth.List)); err != nil {
		return nil, grpc.Errorf(codes.Unauthenticated, "authentication failed: %s", err)
	}

	orgs, err := storage.GetOrganizations(a.ctx.DB, int(req.Limit), int(req.Offset))
	if err != nil {
		return nil, errToRPCError(err)
	}

	totalOrgCount, err := storage.GetOrganizationCount(a.ctx.DB)
	if err != nil {
		return nil, errToRPCError(err)
	}

	result := make([]*pb.GetOrganizationResponse, len(orgs))
	for i, org := range orgs {
		result[i] = &pb.GetOrganizationResponse{
			Id:          	 org.ID,
			Name:        	 org.Name,
			DisplayName: 	 org.DisplayName,
			CanHaveGateways: org.CanHaveGateways,
			CreatedAt:  	 org.CreatedAt.String(),
			UpdatedAt:  	 org.UpdatedAt.String(),
		}
	}

	return &pb.ListOrganizationResponse{
		TotalCount: int32(totalOrgCount),
		Result:     result,
	}, nil
}

// Update updates the given user.
func (a *OrganizationAPI) Update(ctx context.Context, req *pb.UpdateOrganizationRequest) (*pb.OrganizationEmptyResponse, error) {
	if err := a.validator.Validate(ctx,
		auth.ValidateOrganizationAccess(auth.Update, req.Id)); err != nil {
		return nil, grpc.Errorf(codes.Unauthenticated, "authentication failed: %s", err)
	}

	orgUpdate := storage.Organization{
		ID:				 req.Id,
		Name:   	 	 req.Name,
		DisplayName: 	 req.DisplayName,
		CanHaveGateways: req.CanHaveGateways,
	}

	err := storage.UpdateOrganization(a.ctx.DB, &orgUpdate)
	if err != nil {
		return nil, errToRPCError(err)
	}

	return &pb.OrganizationEmptyResponse{}, nil
}

// Delete deletes the user matching the given ID.
func (a *OrganizationAPI) Delete(ctx context.Context, req *pb.OrganizationRequest) (*pb.OrganizationEmptyResponse, error) {
	if err := a.validator.Validate(ctx,
		auth.ValidateOrganizationAccess(auth.Delete, req.Id)); err != nil {
		return nil, grpc.Errorf(codes.Unauthenticated, "authentication failed: %s", err)
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

	orgCount, err := storage.GetOrganizationUserCount(a.ctx.DB, req.Id)
	if err != nil {
		return nil, errToRPCError(err)
	}

	result := make([]*pb.GetOrganizationUserResponse, len(users))
	for i, user := range users {
		result[i] = &pb.GetOrganizationUserResponse{
			UserId:     user.UserID,
			Username:   user.Username,
			IsAdmin:    user.IsAdmin,
			CreatedAt:  user.CreatedAt.String(),
			UpdatedAt:  user.UpdatedAt.String(),
		}
	}

	return &pb.ListOrganizationUsersResponse{
		TotalCount: int32(orgCount),
		Result:     result,
	}, nil
}

// Create creates the given organization-user link.
func (a *OrganizationAPI) AddUser(ctx context.Context, req *pb.OrganizationUserRequest) (*pb.OrganizationEmptyResponse, error) {
	if err := a.validator.Validate(ctx,
		auth.ValidateOrganizationUsersAccess(auth.Create, req.Id)); err != nil {
		return nil, grpc.Errorf(codes.Unauthenticated, "authentication failed: %s", err)
	}

	err := storage.CreateOrganizationUser(a.ctx.DB, req.Id, req.UserId, req.IsAdmin)
	if err != nil {
		return nil, errToRPCError(err)
	}

	return &pb.OrganizationEmptyResponse{}, nil
}

// Update updates the given user.
func (a *OrganizationAPI) UpdateUser(ctx context.Context, req *pb.OrganizationUserRequest) (*pb.OrganizationEmptyResponse, error) {
	if err := a.validator.Validate(ctx,
		auth.ValidateOrganizationUserAccess(auth.Update, req.Id, req.UserId)); err != nil {
		return nil, grpc.Errorf(codes.Unauthenticated, "authentication failed: %s", err)
	}

	err := storage.UpdateOrganizationUser(a.ctx.DB, req.Id, req.UserId, req.IsAdmin)
	if err != nil {
		return nil, errToRPCError(err)
	}

	return &pb.OrganizationEmptyResponse{}, nil
}

// Delete deletes the given user from the organization.
func (a *OrganizationAPI) DeleteUser(ctx context.Context, req *pb.DeleteOrganizationUserRequest) (*pb.OrganizationEmptyResponse, error) {
	if err := a.validator.Validate(ctx,
		auth.ValidateOrganizationUserAccess(auth.Delete, req.Id, req.UserId)); err != nil {
		return nil, grpc.Errorf(codes.Unauthenticated, "authentication failed: %s", err)
	}

	err := storage.DeleteOrganizationUser(a.ctx.DB, req.Id, req.UserId)
	if err != nil {
		return nil, errToRPCError(err)
	}

	return &pb.OrganizationEmptyResponse{}, nil
}

