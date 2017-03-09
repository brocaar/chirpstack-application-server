package api

import (
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

	pb "github.com/brocaar/lora-app-server/api"
	"github.com/brocaar/lora-app-server/internal/api/auth"
	"github.com/brocaar/lora-app-server/internal/common"
	"github.com/brocaar/lora-app-server/internal/storage"
	"github.com/jmoiron/sqlx"
)

// UserAPI exports the User related functions.
type UserAPI struct {
	ctx       common.Context
	validator auth.Validator
}

// InternalUserAPI exports the internal User related functions.
type InternalUserAPI struct {
	ctx       common.Context
	validator auth.Validator
}

// NewUserAPI creates a new UserAPI.
func NewUserAPI(ctx common.Context, validator auth.Validator) *UserAPI {
	return &UserAPI{
		ctx:       ctx,
		validator: validator,
	}
}

// Create creates the given user.
func (a *UserAPI) Create(ctx context.Context, req *pb.AddUserRequest) (*pb.AddUserResponse, error) {
	if err := a.validator.Validate(ctx,
		auth.ValidateUsersAccess(auth.Create)); err != nil {
		return nil, grpc.Errorf(codes.Unauthenticated, "authentication failed: %s", err)
	}

	user := storage.User{
		Username:   req.Username,
		SessionTTL: req.SessionTTL,
		IsAdmin:    req.IsAdmin,
	}

	id, err := storage.CreateUser(a.ctx.DB, &user, req.Password)
	if err != nil {
		return nil, errToRPCError(err)
	}

	return &pb.AddUserResponse{Id: id}, nil
}

// Get returns the user matching the given ID.
func (a *UserAPI) Get(ctx context.Context, req *pb.UserRequest) (*pb.UserResponse, error) {
	if err := a.validator.Validate(ctx,
		auth.ValidateUserAccess(req.Id, auth.Read)); err != nil {
		return nil, grpc.Errorf(codes.Unauthenticated, "authentication failed: %s", err)
	}

	user, err := storage.GetUser(a.ctx.DB, req.Id)
	if err != nil {
		return nil, errToRPCError(err)
	}

	profile, err := getProfile(a.ctx.DB, req.Id)
	if err != nil {
		return nil, errToRPCError(err)
	}

	return &pb.UserResponse{
		Info: &pb.UserInfo{
			UserSettings: &pb.UserSettings{
				Id:         user.ID,
				Username:   user.Username,
				SessionTTL: user.SessionTTL,
				IsAdmin:    user.IsAdmin,
			},
			UserProfile: profile.UserProfile,
		},
	}, nil
}

// List lists the users.
func (a *UserAPI) List(ctx context.Context, req *pb.ListUserRequest) (*pb.ListUserResponse, error) {
	if err := a.validator.Validate(ctx,
		auth.ValidateUsersAccess(auth.List)); err != nil {
		return nil, grpc.Errorf(codes.Unauthenticated, "authentication failed: %s", err)
	}

	users, err := storage.GetUsers(a.ctx.DB, req.Limit, req.Offset)
	if err != nil {
		return nil, errToRPCError(err)
	}

	totalUserCount, err := storage.GetUserCount(a.ctx.DB)
	if err != nil {
		return nil, errToRPCError(err)
	}
	usersResp := make([]*pb.UserResponse, len(users))
	for i, user := range users {
		profile, err := getProfile(a.ctx.DB, user.ID)
		if err != nil {
			return nil, errToRPCError(err)
		}

		usersResp[i] = &pb.UserResponse{
			Info: &pb.UserInfo{
				UserSettings: &pb.UserSettings{
					Id:         user.ID,
					Username:   user.Username,
					SessionTTL: user.SessionTTL,
					IsAdmin:    user.IsAdmin,
				},
				UserProfile: profile.UserProfile,
			},
		}
	}

	return &pb.ListUserResponse{
		TotalCount: totalUserCount,
		Users:      usersResp,
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
		SessionTTL: req.SessionTTL,
	}

	err := storage.UpdateUser(a.ctx.DB, userUpdate)
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

	err := storage.DeleteUser(a.ctx.DB, req.Id)
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

	err := storage.UpdatePassword(a.ctx.DB, req.Id, req.Password)
	if err != nil {
		return nil, errToRPCError(err)
	}
	return &pb.UserEmptyResponse{}, nil
}

// NewInternalUserAPI creates a new InternalUserAPI.
func NewInternalUserAPI(ctx common.Context, validator auth.Validator) *InternalUserAPI {
	return &InternalUserAPI{
		ctx:       ctx,
		validator: validator,
	}
}

// Login validates the login request and returns a JWT token.
func (a *InternalUserAPI) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	jwt, err := storage.LoginUser(a.ctx.DB, req.Username, req.Password)
	if nil != err {
		return nil, errToRPCError(err)
	}

	profile, _ := a.Profile(ctx, &pb.ProfileRequest{})
	// Ignore any error, take whatever profile we're given.

	return &pb.LoginResponse{Jwt: jwt, Profile: profile}, nil
}

type claims struct {
	Username string `json:"username"`
}

func (a *InternalUserAPI) Profile(ctx context.Context, req *pb.ProfileRequest) (*pb.ProfileResponse, error) {
	username, err := a.validator.GetUsername(ctx)
	if nil != err {
		return nil, errToRPCError(err)
	}

	// Get the user id based on the username.
	user, err := storage.GetUserByUsername(a.ctx.DB, username)
	if nil != err {
		return nil, errToRPCError(err)
	}

	return getProfile(a.ctx.DB, user.ID)
}

func getProfile(db *sqlx.DB, id int64) (*pb.ProfileResponse, error) {
	// Get the profile for the user.
	userProfile, err := storage.GetProfile(db, id)
	if nil != err {
		return nil, errToRPCError(err)
	}

	// Convert to the external form.
	profResp := make([]*pb.ApplicationLink, len(userProfile))
	for i, up := range userProfile {
		profResp[i] = &pb.ApplicationLink{
			ApplicationID:   up.ID,
			ApplicationName: up.Name,
			IsAdmin:         up.IsAdmin,
			CreatedAt:       up.CreatedAt.String(),
			UpdatedAt:       up.UpdatedAt.String(),
		}
	}

	userProf := &pb.UserProfile{
		Applications: profResp,
	}

	return &pb.ProfileResponse{UserProfile: userProf}, nil
}
