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
)

// UserAPI exports the User related functions.
type UserAPI struct {
	validator auth.Validator
}

// InternalUserAPI exports the internal User related functions.
type InternalUserAPI struct {
	validator auth.Validator
}

// NewUserAPI creates a new UserAPI.
func NewUserAPI(validator auth.Validator) *UserAPI {
	return &UserAPI{
		validator: validator,
	}
}

// Create creates the given user.
func (a *UserAPI) Create(ctx context.Context, req *pb.AddUserRequest) (*pb.AddUserResponse, error) {
	if err := a.validator.Validate(ctx,
		auth.ValidateUsersAccess(auth.Create)); err != nil {
		return nil, grpc.Errorf(codes.Unauthenticated, "authentication failed: %s", err)
	}

	// validate if the client has admin rights for the given organizations
	// to which the user must be linked
	for _, org := range req.Organizations {
		if err := a.validator.Validate(ctx,
			auth.ValidateIsOrganizationAdmin(org.OrganizationID)); err != nil {
			return nil, grpc.Errorf(codes.Unauthenticated, "authentication failed: %s", err)
		}
	}

	user := storage.User{
		Username:   req.Username,
		SessionTTL: req.SessionTTL,
		IsAdmin:    req.IsAdmin,
		IsActive:   req.IsActive,
		Email:      req.Email,
		Note:       req.Note,
	}

	isAdmin, err := a.validator.GetIsAdmin(ctx)
	if err != nil {
		return nil, errToRPCError(err)
	}

	if !isAdmin {
		// non-admin users are not able to modify the fields below
		user.IsAdmin = false
		user.IsActive = true
		user.SessionTTL = 0
	}

	var userID int64

	err = storage.Transaction(config.C.PostgreSQL.DB, func(tx sqlx.Ext) error {
		userID, err = storage.CreateUser(tx, &user, req.Password)
		if err != nil {
			return err
		}

		for _, org := range req.Organizations {
			if err := storage.CreateOrganizationUser(tx, org.OrganizationID, userID, org.IsAdmin); err != nil {
				return err
			}
		}

		return nil
	})
	if err != nil {
		return nil, errToRPCError(err)
	}

	return &pb.AddUserResponse{Id: userID}, nil
}

// Get returns the user matching the given ID.
func (a *UserAPI) Get(ctx context.Context, req *pb.UserRequest) (*pb.GetUserResponse, error) {
	if err := a.validator.Validate(ctx,
		auth.ValidateUserAccess(req.Id, auth.Read)); err != nil {
		return nil, grpc.Errorf(codes.Unauthenticated, "authentication failed: %s", err)
	}

	user, err := storage.GetUser(config.C.PostgreSQL.DB, req.Id)
	if err != nil {
		return nil, errToRPCError(err)
	}

	return &pb.GetUserResponse{
		Id:         user.ID,
		Username:   user.Username,
		SessionTTL: user.SessionTTL,
		IsAdmin:    user.IsAdmin,
		IsActive:   user.IsActive,
		CreatedAt:  user.CreatedAt.String(),
		UpdatedAt:  user.UpdatedAt.String(),
		Email:      user.Email,
		Note:       user.Note,
	}, nil
}

// List lists the users.
func (a *UserAPI) List(ctx context.Context, req *pb.ListUserRequest) (*pb.ListUserResponse, error) {
	if err := a.validator.Validate(ctx,
		auth.ValidateUsersAccess(auth.List)); err != nil {
		return nil, grpc.Errorf(codes.Unauthenticated, "authentication failed: %s", err)
	}

	users, err := storage.GetUsers(config.C.PostgreSQL.DB, req.Limit, req.Offset, req.Search)
	if err != nil {
		return nil, errToRPCError(err)
	}

	totalUserCount, err := storage.GetUserCount(config.C.PostgreSQL.DB, req.Search)
	if err != nil {
		return nil, errToRPCError(err)
	}

	result := make([]*pb.GetUserResponse, len(users))
	for i, user := range users {
		result[i] = &pb.GetUserResponse{
			Id:         user.ID,
			Username:   user.Username,
			SessionTTL: user.SessionTTL,
			IsAdmin:    user.IsAdmin,
			IsActive:   user.IsActive,
			CreatedAt:  user.CreatedAt.String(),
			UpdatedAt:  user.UpdatedAt.String(),
		}
	}

	return &pb.ListUserResponse{
		TotalCount: totalUserCount,
		Result:     result,
	}, nil
}

// Update updates the given user.
func (a *UserAPI) Update(ctx context.Context, req *pb.UpdateUserRequest) (*pb.UserEmptyResponse, error) {
	if err := a.validator.Validate(ctx,
		auth.ValidateUserAccess(req.Id, auth.Update)); err != nil {
		return nil, grpc.Errorf(codes.Unauthenticated, "authentication failed: %s", err)
	}

	userUpdate := storage.UserUpdate{
		ID:         req.Id,
		Username:   req.Username,
		IsAdmin:    req.IsAdmin,
		IsActive:   req.IsActive,
		SessionTTL: req.SessionTTL,
		Email:      req.Email,
		Note:       req.Note,
	}

	err := storage.UpdateUser(config.C.PostgreSQL.DB, userUpdate)
	if err != nil {
		return nil, errToRPCError(err)
	}

	return &pb.UserEmptyResponse{}, nil
}

// Delete deletes the user matching the given ID.
func (a *UserAPI) Delete(ctx context.Context, req *pb.UserRequest) (*pb.UserEmptyResponse, error) {
	if err := a.validator.Validate(ctx,
		auth.ValidateUserAccess(req.Id, auth.Delete)); err != nil {
		return nil, grpc.Errorf(codes.Unauthenticated, "authentication failed: %s", err)
	}

	err := storage.DeleteUser(config.C.PostgreSQL.DB, req.Id)
	if err != nil {
		return nil, errToRPCError(err)
	}
	return &pb.UserEmptyResponse{}, nil
}

// UpdatePassword updates the password for the user matching the given ID.
func (a *UserAPI) UpdatePassword(ctx context.Context, req *pb.UpdateUserPasswordRequest) (*pb.UserEmptyResponse, error) {
	if err := a.validator.Validate(ctx,
		auth.ValidateUserAccess(req.Id, auth.UpdateProfile)); err != nil {
		return nil, grpc.Errorf(codes.Unauthenticated, "authentication failed: %s", err)
	}

	err := storage.UpdatePassword(config.C.PostgreSQL.DB, req.Id, req.Password)
	if err != nil {
		return nil, errToRPCError(err)
	}
	return &pb.UserEmptyResponse{}, nil
}

// NewInternalUserAPI creates a new InternalUserAPI.
func NewInternalUserAPI(validator auth.Validator) *InternalUserAPI {
	return &InternalUserAPI{
		validator: validator,
	}
}

// Login validates the login request and returns a JWT token.
func (a *InternalUserAPI) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	jwt, err := storage.LoginUser(config.C.PostgreSQL.DB, req.Username, req.Password)
	if nil != err {
		return nil, errToRPCError(err)
	}

	return &pb.LoginResponse{Jwt: jwt}, nil
}

type claims struct {
	Username string `json:"username"`
}

// Profile returns the user profile.
func (a *InternalUserAPI) Profile(ctx context.Context, req *pb.ProfileRequest) (*pb.ProfileResponse, error) {
	if err := a.validator.Validate(ctx,
		auth.ValidateActiveUser()); err != nil {
		return nil, grpc.Errorf(codes.Unauthenticated, "authentication failed: %s", err)
	}

	username, err := a.validator.GetUsername(ctx)
	if nil != err {
		return nil, errToRPCError(err)
	}

	// Get the user id based on the username.
	user, err := storage.GetUserByUsername(config.C.PostgreSQL.DB, username)
	if nil != err {
		return nil, errToRPCError(err)
	}

	prof, err := storage.GetProfile(config.C.PostgreSQL.DB, user.ID)
	if err != nil {
		return nil, errToRPCError(err)
	}

	resp := pb.ProfileResponse{
		User: &pb.GetUserResponse{
			Id:         prof.User.ID,
			Username:   prof.User.Username,
			SessionTTL: prof.User.SessionTTL,
			IsAdmin:    prof.User.IsAdmin,
			IsActive:   prof.User.IsActive,
			CreatedAt:  prof.User.CreatedAt.Format(time.RFC3339Nano),
			UpdatedAt:  prof.User.UpdatedAt.Format(time.RFC3339Nano),
		},
		Organizations: make([]*pb.OrganizationLink, len(prof.Organizations)),
		Settings: &pb.ProfileSettings{
			DisableAssignExistingUsers: auth.DisableAssignExistingUsers,
		},
	}

	for i := range prof.Organizations {
		resp.Organizations[i] = &pb.OrganizationLink{
			OrganizationID:   prof.Organizations[i].ID,
			OrganizationName: prof.Organizations[i].Name,
			IsAdmin:          prof.Organizations[i].IsAdmin,
			UpdatedAt:        prof.Organizations[i].UpdatedAt.Format(time.RFC3339Nano),
			CreatedAt:        prof.Organizations[i].CreatedAt.Format(time.RFC3339Nano),
		}
	}

	return &resp, nil
}

// Branding returns UI branding.
func (a *InternalUserAPI) Branding(ctx context.Context, req *pb.BrandingRequest) (*pb.BrandingResponse, error) {
	resp := pb.BrandingResponse{
		Logo:         config.C.ApplicationServer.Branding.Header,
		Registration: config.C.ApplicationServer.Branding.Registration,
		Footer:       config.C.ApplicationServer.Branding.Footer,
	}

	return &resp, nil
}
