package external

import (
	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/jmoiron/sqlx"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

	pb "github.com/brocaar/chirpstack-api/go/v3/as/external/api"
	"github.com/brocaar/chirpstack-application-server/internal/api/external/auth"
	"github.com/brocaar/chirpstack-application-server/internal/api/helpers"
	"github.com/brocaar/chirpstack-application-server/internal/storage"
)

// OrganizationAPI exports the organization related functions.
type OrganizationAPI struct {
	validator auth.Validator
}

// NewOrganizationAPI creates a new OrganizationAPI.
func NewOrganizationAPI(validator auth.Validator) *OrganizationAPI {
	return &OrganizationAPI{
		validator: validator,
	}
}

// Create creates the given organization.
func (a *OrganizationAPI) Create(ctx context.Context, req *pb.CreateOrganizationRequest) (*pb.CreateOrganizationResponse, error) {
	if req.Organization == nil {
		return nil, grpc.Errorf(codes.InvalidArgument, "organization must not be nil")
	}

	if err := a.validator.Validate(ctx,
		auth.ValidateOrganizationsAccess(auth.Create)); err != nil {
		return nil, grpc.Errorf(codes.Unauthenticated, "authentication failed: %s", err)
	}

	org := storage.Organization{
		Name:            req.Organization.Name,
		DisplayName:     req.Organization.DisplayName,
		CanHaveGateways: req.Organization.CanHaveGateways,
		MaxDeviceCount:  int(req.Organization.MaxDeviceCount),
		MaxGatewayCount: int(req.Organization.MaxGatewayCount),
	}

	err := storage.CreateOrganization(ctx, storage.DB(), &org)
	if err != nil {
		return nil, helpers.ErrToRPCError(err)
	}

	return &pb.CreateOrganizationResponse{
		Id: org.ID,
	}, nil
}

// Get returns the organization matching the given ID.
func (a *OrganizationAPI) Get(ctx context.Context, req *pb.GetOrganizationRequest) (*pb.GetOrganizationResponse, error) {
	if err := a.validator.Validate(ctx,
		auth.ValidateOrganizationAccess(auth.Read, req.Id)); err != nil {
		return nil, grpc.Errorf(codes.Unauthenticated, "authentication failed: %s", err)
	}

	org, err := storage.GetOrganization(ctx, storage.DB(), req.Id, false)
	if err != nil {
		return nil, helpers.ErrToRPCError(err)
	}

	resp := pb.GetOrganizationResponse{
		Organization: &pb.Organization{
			Id:              org.ID,
			Name:            org.Name,
			DisplayName:     org.DisplayName,
			CanHaveGateways: org.CanHaveGateways,
			MaxDeviceCount:  uint32(org.MaxDeviceCount),
			MaxGatewayCount: uint32(org.MaxGatewayCount),
		},
	}

	resp.CreatedAt, err = ptypes.TimestampProto(org.CreatedAt)
	if err != nil {
		return nil, helpers.ErrToRPCError(err)
	}
	resp.UpdatedAt, err = ptypes.TimestampProto(org.UpdatedAt)
	if err != nil {
		return nil, helpers.ErrToRPCError(err)
	}

	return &resp, nil
}

// List lists the organizations to which the user has access.
func (a *OrganizationAPI) List(ctx context.Context, req *pb.ListOrganizationRequest) (*pb.ListOrganizationResponse, error) {
	if err := a.validator.Validate(ctx,
		auth.ValidateOrganizationsAccess(auth.List)); err != nil {
		return nil, grpc.Errorf(codes.Unauthenticated, "authentication failed: %s", err)
	}

	filters := storage.OrganizationFilters{
		Search: req.Search,
		Limit:  int(req.Limit),
		Offset: int(req.Offset),
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

		if !user.IsAdmin {
			filters.UserID = user.ID
		}
	case auth.SubjectAPIKey:
		// Nothing to do as the validator function already validated that the
		// API key must be a global admin key.
	default:
		return nil, grpc.Errorf(codes.Unauthenticated, "invalid token subject: %s", err)
	}

	count, err := storage.GetOrganizationCount(ctx, storage.DB(), filters)
	if err != nil {
		return nil, helpers.ErrToRPCError(err)
	}

	orgs, err := storage.GetOrganizations(ctx, storage.DB(), filters)
	if err != nil {
		return nil, helpers.ErrToRPCError(err)
	}

	resp := pb.ListOrganizationResponse{
		TotalCount: int64(count),
	}

	for _, org := range orgs {
		row := pb.OrganizationListItem{
			Id:              org.ID,
			Name:            org.Name,
			DisplayName:     org.DisplayName,
			CanHaveGateways: org.CanHaveGateways,
		}

		row.CreatedAt, err = ptypes.TimestampProto(org.CreatedAt)
		if err != nil {
			return nil, helpers.ErrToRPCError(err)
		}
		row.UpdatedAt, err = ptypes.TimestampProto(org.UpdatedAt)
		if err != nil {
			return nil, helpers.ErrToRPCError(err)
		}

		resp.Result = append(resp.Result, &row)
	}

	return &resp, nil
}

// Update updates the given organization.
func (a *OrganizationAPI) Update(ctx context.Context, req *pb.UpdateOrganizationRequest) (*empty.Empty, error) {
	if req.Organization == nil {
		return nil, grpc.Errorf(codes.InvalidArgument, "organization must not be nil")
	}

	if err := a.validator.Validate(ctx,
		auth.ValidateOrganizationAccess(auth.Update, req.Organization.Id)); err != nil {
		return nil, grpc.Errorf(codes.Unauthenticated, "authentication failed: %s", err)
	}

	sub, err := a.validator.GetSubject(ctx)
	if err != nil {
		return nil, helpers.ErrToRPCError(err)
	}

	org, err := storage.GetOrganization(ctx, storage.DB(), req.Organization.Id, false)
	if err != nil {
		return nil, helpers.ErrToRPCError(err)
	}

	org.Name = req.Organization.Name
	org.DisplayName = req.Organization.DisplayName

	switch sub {
	case auth.SubjectUser:
		user, err := a.validator.GetUser(ctx)
		if err != nil {
			return nil, helpers.ErrToRPCError(err)
		}

		if user.IsAdmin {
			org.CanHaveGateways = req.Organization.CanHaveGateways
			org.MaxGatewayCount = int(req.Organization.MaxGatewayCount)
			org.MaxDeviceCount = int(req.Organization.MaxDeviceCount)
		}
	case auth.SubjectAPIKey:
		// The validator function already validated that the
		// API key must be a global admin key.
		org.CanHaveGateways = req.Organization.CanHaveGateways
		org.MaxGatewayCount = int(req.Organization.MaxGatewayCount)
		org.MaxDeviceCount = int(req.Organization.MaxDeviceCount)
	}

	err = storage.UpdateOrganization(ctx, storage.DB(), &org)
	if err != nil {
		return nil, helpers.ErrToRPCError(err)
	}

	return &empty.Empty{}, nil
}

// Delete deletes the organization matching the given ID.
func (a *OrganizationAPI) Delete(ctx context.Context, req *pb.DeleteOrganizationRequest) (*empty.Empty, error) {
	if err := a.validator.Validate(ctx,
		auth.ValidateOrganizationAccess(auth.Delete, req.Id)); err != nil {
		return nil, grpc.Errorf(codes.Unauthenticated, "authentication failed: %s", err)
	}

	err := storage.Transaction(func(tx sqlx.Ext) error {
		if err := storage.DeleteAllGatewaysForOrganizationID(ctx, tx, req.Id); err != nil {
			return helpers.ErrToRPCError(err)
		}

		if err := storage.DeleteOrganization(ctx, tx, req.Id); err != nil {
			return helpers.ErrToRPCError(err)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return &empty.Empty{}, nil
}

// ListUsers lists the users assigned to the given organization.
func (a *OrganizationAPI) ListUsers(ctx context.Context, req *pb.ListOrganizationUsersRequest) (*pb.ListOrganizationUsersResponse, error) {
	if err := a.validator.Validate(ctx,
		auth.ValidateOrganizationUsersAccess(auth.List, req.OrganizationId)); err != nil {
		return nil, grpc.Errorf(codes.Unauthenticated, "authentication failed: %s", err)
	}

	users, err := storage.GetOrganizationUsers(ctx, storage.DB(), req.OrganizationId, int(req.Limit), int(req.Offset))
	if err != nil {
		return nil, helpers.ErrToRPCError(err)
	}

	userCount, err := storage.GetOrganizationUserCount(ctx, storage.DB(), req.OrganizationId)
	if err != nil {
		return nil, helpers.ErrToRPCError(err)
	}

	resp := pb.ListOrganizationUsersResponse{
		TotalCount: int64(userCount),
	}

	for _, u := range users {
		row := pb.OrganizationUserListItem{
			UserId:         u.UserID,
			Email:          u.Email,
			IsAdmin:        u.IsAdmin,
			IsDeviceAdmin:  u.IsDeviceAdmin,
			IsGatewayAdmin: u.IsGatewayAdmin,
		}

		row.CreatedAt, err = ptypes.TimestampProto(u.CreatedAt)
		if err != nil {
			return nil, helpers.ErrToRPCError(err)
		}
		row.UpdatedAt, err = ptypes.TimestampProto(u.UpdatedAt)
		if err != nil {
			return nil, helpers.ErrToRPCError(err)
		}

		resp.Result = append(resp.Result, &row)
	}

	return &resp, nil
}

// AddUser creates the given organization-user link. The user is matched by email, not user id.
func (a *OrganizationAPI) AddUser(ctx context.Context, req *pb.AddOrganizationUserRequest) (*empty.Empty, error) {
	if req.OrganizationUser == nil {
		return nil, grpc.Errorf(codes.InvalidArgument, "organization_user must not be nil")
	}

	if err := a.validator.Validate(ctx,
		auth.ValidateOrganizationUsersAccess(auth.Create, req.OrganizationUser.OrganizationId)); err != nil {
		return nil, grpc.Errorf(codes.Unauthenticated, "authentication failed: %s", err)
	}

	user, err := storage.GetUserByEmail(ctx, storage.DB(), req.OrganizationUser.Email)
	if err != nil {
		return nil, helpers.ErrToRPCError(err)
	}

	err = storage.CreateOrganizationUser(ctx,
		storage.DB(),
		req.OrganizationUser.OrganizationId,
		user.ID,
		req.OrganizationUser.IsAdmin,
		req.OrganizationUser.IsDeviceAdmin,
		req.OrganizationUser.IsGatewayAdmin,
	)
	if err != nil {
		return nil, helpers.ErrToRPCError(err)
	}

	return &empty.Empty{}, nil
}

// UpdateUser updates the given user.
func (a *OrganizationAPI) UpdateUser(ctx context.Context, req *pb.UpdateOrganizationUserRequest) (*empty.Empty, error) {
	if req.OrganizationUser == nil {
		return nil, grpc.Errorf(codes.InvalidArgument, "organization_user must not be nil")
	}

	if err := a.validator.Validate(ctx,
		auth.ValidateOrganizationUserAccess(auth.Update, req.OrganizationUser.OrganizationId, req.OrganizationUser.UserId)); err != nil {
		return nil, grpc.Errorf(codes.Unauthenticated, "authentication failed: %s", err)
	}

	err := storage.UpdateOrganizationUser(ctx,
		storage.DB(),
		req.OrganizationUser.OrganizationId,
		req.OrganizationUser.UserId,
		req.OrganizationUser.IsAdmin,
		req.OrganizationUser.IsDeviceAdmin,
		req.OrganizationUser.IsGatewayAdmin,
	)
	if err != nil {
		return nil, helpers.ErrToRPCError(err)
	}

	return &empty.Empty{}, nil
}

// DeleteUser deletes the given user from the organization.
func (a *OrganizationAPI) DeleteUser(ctx context.Context, req *pb.DeleteOrganizationUserRequest) (*empty.Empty, error) {
	if err := a.validator.Validate(ctx,
		auth.ValidateOrganizationUserAccess(auth.Delete, req.OrganizationId, req.UserId)); err != nil {
		return nil, grpc.Errorf(codes.Unauthenticated, "authentication failed: %s", err)
	}

	sub, err := a.validator.GetSubject(ctx)
	if err != nil {
		return nil, helpers.ErrToRPCError(err)
	}

	if sub == auth.SubjectUser {
		user, err := a.validator.GetUser(ctx)
		if err != nil {
			return nil, helpers.ErrToRPCError(err)
		}

		if !user.IsAdmin && user.ID == req.UserId {
			return nil, grpc.Errorf(codes.InvalidArgument, "you can not delete yourself from an organization")
		}
	}

	err = storage.DeleteOrganizationUser(ctx, storage.DB(), req.OrganizationId, req.UserId)
	if err != nil {
		return nil, helpers.ErrToRPCError(err)
	}

	return &empty.Empty{}, nil
}

// GetUser returns the user details for the given user ID.
func (a *OrganizationAPI) GetUser(ctx context.Context, req *pb.GetOrganizationUserRequest) (*pb.GetOrganizationUserResponse, error) {
	if err := a.validator.Validate(ctx,
		auth.ValidateOrganizationUserAccess(auth.Read, req.OrganizationId, req.UserId)); err != nil {
		return nil, grpc.Errorf(codes.Unauthenticated, "authentication failed: %s", err)
	}

	user, err := storage.GetOrganizationUser(ctx, storage.DB(), req.OrganizationId, req.UserId)
	if err != nil {
		return nil, helpers.ErrToRPCError(err)
	}

	resp := pb.GetOrganizationUserResponse{
		OrganizationUser: &pb.OrganizationUser{
			OrganizationId: req.OrganizationId,
			UserId:         req.UserId,
			IsAdmin:        user.IsAdmin,
			IsDeviceAdmin:  user.IsDeviceAdmin,
			IsGatewayAdmin: user.IsGatewayAdmin,
			Email:          user.Email,
		},
	}

	resp.CreatedAt, err = ptypes.TimestampProto(user.CreatedAt)
	if err != nil {
		return nil, helpers.ErrToRPCError(err)
	}
	resp.UpdatedAt, err = ptypes.TimestampProto(user.UpdatedAt)
	if err != nil {
		return nil, helpers.ErrToRPCError(err)
	}

	return &resp, nil
}
