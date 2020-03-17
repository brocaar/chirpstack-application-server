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

// UserAPI exports the User related functions.
type UserAPI struct {
	validator auth.Validator
}

// NewUserAPI creates a new UserAPI.
func NewUserAPI(validator auth.Validator) *UserAPI {
	return &UserAPI{
		validator: validator,
	}
}

// Create creates the given user.
func (a *UserAPI) Create(ctx context.Context, req *pb.CreateUserRequest) (*pb.CreateUserResponse, error) {
	if req.User == nil {
		return nil, grpc.Errorf(codes.InvalidArgument, "user must not be nil")
	}

	if err := a.validator.Validate(ctx,
		auth.ValidateUsersAccess(auth.Create)); err != nil {
		return nil, grpc.Errorf(codes.Unauthenticated, "authentication failed: %s", err)
	}

	// validate if the client has admin rights for the given organizations
	// to which the user must be linked
	for _, org := range req.Organizations {
		if err := a.validator.Validate(ctx,
			auth.ValidateIsOrganizationAdmin(org.OrganizationId)); err != nil {
			return nil, grpc.Errorf(codes.Unauthenticated, "authentication failed: %s", err)
		}
	}

	user := storage.User{
		Username:   req.User.Username,
		SessionTTL: req.User.SessionTtl,
		IsAdmin:    req.User.IsAdmin,
		IsActive:   req.User.IsActive,
		Email:      req.User.Email,
		Note:       req.User.Note,
	}

	isAdmin, err := a.validator.GetIsAdmin(ctx)
	if err != nil {
		return nil, helpers.ErrToRPCError(err)
	}

	if !isAdmin {
		// non-admin users are not able to modify the fields below
		user.IsAdmin = false
		user.IsActive = true
		user.SessionTTL = 0
	}

	var userID int64

	err = storage.Transaction(func(tx sqlx.Ext) error {
		userID, err = storage.CreateUser(ctx, tx, &user, req.Password)
		if err != nil {
			return err
		}

		for _, org := range req.Organizations {
			if err := storage.CreateOrganizationUser(ctx, tx, org.OrganizationId, userID, org.IsAdmin, org.IsDeviceAdmin, org.IsGatewayAdmin); err != nil {
				return err
			}
		}

		return nil
	})
	if err != nil {
		return nil, helpers.ErrToRPCError(err)
	}

	return &pb.CreateUserResponse{Id: userID}, nil
}

// Get returns the user matching the given ID.
func (a *UserAPI) Get(ctx context.Context, req *pb.GetUserRequest) (*pb.GetUserResponse, error) {
	if err := a.validator.Validate(ctx,
		auth.ValidateUserAccess(req.Id, auth.Read)); err != nil {
		return nil, grpc.Errorf(codes.Unauthenticated, "authentication failed: %s", err)
	}

	user, err := storage.GetUser(ctx, storage.DB(), req.Id)
	if err != nil {
		return nil, helpers.ErrToRPCError(err)
	}

	resp := pb.GetUserResponse{
		User: &pb.User{
			Id:         user.ID,
			Username:   user.Username,
			SessionTtl: user.SessionTTL,
			IsAdmin:    user.IsAdmin,
			IsActive:   user.IsActive,
			Email:      user.Email,
			Note:       user.Note,
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

// List lists the users.
func (a *UserAPI) List(ctx context.Context, req *pb.ListUserRequest) (*pb.ListUserResponse, error) {
	if err := a.validator.Validate(ctx,
		auth.ValidateUsersAccess(auth.List)); err != nil {
		return nil, grpc.Errorf(codes.Unauthenticated, "authentication failed: %s", err)
	}

	users, err := storage.GetUsers(ctx, storage.DB(), int(req.Limit), int(req.Offset), req.Search)
	if err != nil {
		return nil, helpers.ErrToRPCError(err)
	}

	totalUserCount, err := storage.GetUserCount(ctx, storage.DB(), req.Search)
	if err != nil {
		return nil, helpers.ErrToRPCError(err)
	}

	resp := pb.ListUserResponse{
		TotalCount: int64(totalUserCount),
	}

	for _, u := range users {
		row := pb.UserListItem{
			Id:         u.ID,
			Username:   u.Username,
			SessionTtl: u.SessionTTL,
			IsAdmin:    u.IsAdmin,
			IsActive:   u.IsActive,
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

// Update updates the given user.
func (a *UserAPI) Update(ctx context.Context, req *pb.UpdateUserRequest) (*empty.Empty, error) {
	if req.User == nil {
		return nil, grpc.Errorf(codes.InvalidArgument, "user must not be nil")
	}

	if err := a.validator.Validate(ctx,
		auth.ValidateUserAccess(req.User.Id, auth.Update)); err != nil {
		return nil, grpc.Errorf(codes.Unauthenticated, "authentication failed: %s", err)
	}

	userUpdate := storage.UserUpdate{
		ID:         req.User.Id,
		Username:   req.User.Username,
		IsAdmin:    req.User.IsAdmin,
		IsActive:   req.User.IsActive,
		SessionTTL: req.User.SessionTtl,
		Email:      req.User.Email,
		Note:       req.User.Note,
	}

	err := storage.UpdateUser(ctx, storage.DB(), userUpdate)
	if err != nil {
		return nil, helpers.ErrToRPCError(err)
	}

	return &empty.Empty{}, nil
}

// Delete deletes the user matching the given ID.
func (a *UserAPI) Delete(ctx context.Context, req *pb.DeleteUserRequest) (*empty.Empty, error) {
	if err := a.validator.Validate(ctx,
		auth.ValidateUserAccess(req.Id, auth.Delete)); err != nil {
		return nil, grpc.Errorf(codes.Unauthenticated, "authentication failed: %s", err)
	}

	err := storage.DeleteUser(ctx, storage.DB(), req.Id)
	if err != nil {
		return nil, helpers.ErrToRPCError(err)
	}

	return &empty.Empty{}, nil
}

// UpdatePassword updates the password for the user matching the given ID.
func (a *UserAPI) UpdatePassword(ctx context.Context, req *pb.UpdateUserPasswordRequest) (*empty.Empty, error) {
	if err := a.validator.Validate(ctx,
		auth.ValidateUserAccess(req.UserId, auth.UpdateProfile)); err != nil {
		return nil, grpc.Errorf(codes.Unauthenticated, "authentication failed: %s", err)
	}

	err := storage.UpdatePassword(ctx, storage.DB(), req.UserId, req.Password)
	if err != nil {
		return nil, helpers.ErrToRPCError(err)
	}

	return &empty.Empty{}, nil
}
