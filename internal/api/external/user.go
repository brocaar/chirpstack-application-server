package external

import (
	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/jmoiron/sqlx"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

	pb "github.com/brocaar/lora-app-server/api"
	"github.com/brocaar/lora-app-server/internal/api/external/auth"
	"github.com/brocaar/lora-app-server/internal/api/helpers"
	"github.com/brocaar/lora-app-server/internal/storage"
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
		userID, err = storage.CreateUser(tx, &user, req.Password)
		if err != nil {
			return err
		}

		for _, org := range req.Organizations {
			if err := storage.CreateOrganizationUser(tx, org.OrganizationId, userID, org.IsAdmin); err != nil {
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

	user, err := storage.GetUser(storage.DB(), req.Id)
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

	users, err := storage.GetUsers(storage.DB(), int(req.Limit), int(req.Offset), req.Search)
	if err != nil {
		return nil, helpers.ErrToRPCError(err)
	}

	totalUserCount, err := storage.GetUserCount(storage.DB(), req.Search)
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

	err := storage.UpdateUser(storage.DB(), userUpdate)
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

	err := storage.DeleteUser(storage.DB(), req.Id)
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

	err := storage.UpdatePassword(storage.DB(), req.UserId, req.Password)
	if err != nil {
		return nil, helpers.ErrToRPCError(err)
	}

	return &empty.Empty{}, nil
}

// NewInternalUserAPI creates a new InternalUserAPI.
func NewInternalUserAPI(validator auth.Validator) *InternalUserAPI {
	return &InternalUserAPI{
		validator: validator,
	}
}

// Login validates the login request and returns a JWT token.
func (a *InternalUserAPI) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	jwt, err := storage.LoginUser(storage.DB(), req.Username, req.Password)
	if nil != err {
		return nil, helpers.ErrToRPCError(err)
	}

	return &pb.LoginResponse{Jwt: jwt}, nil
}

type claims struct {
	Username string `json:"username"`
}

// Profile returns the user profile.
func (a *InternalUserAPI) Profile(ctx context.Context, req *empty.Empty) (*pb.ProfileResponse, error) {
	if err := a.validator.Validate(ctx,
		auth.ValidateActiveUser()); err != nil {
		return nil, grpc.Errorf(codes.Unauthenticated, "authentication failed: %s", err)
	}

	username, err := a.validator.GetUsername(ctx)
	if nil != err {
		return nil, helpers.ErrToRPCError(err)
	}

	// Get the user id based on the username.
	user, err := storage.GetUserByUsername(storage.DB(), username)
	if nil != err {
		return nil, helpers.ErrToRPCError(err)
	}

	prof, err := storage.GetProfile(storage.DB(), user.ID)
	if err != nil {
		return nil, helpers.ErrToRPCError(err)
	}

	resp := pb.ProfileResponse{
		User: &pb.User{
			Id:         prof.User.ID,
			Username:   prof.User.Username,
			SessionTtl: prof.User.SessionTTL,
			IsAdmin:    prof.User.IsAdmin,
			IsActive:   prof.User.IsActive,
		},
		Settings: &pb.ProfileSettings{
			DisableAssignExistingUsers: auth.DisableAssignExistingUsers,
		},
	}

	for _, org := range prof.Organizations {
		row := pb.OrganizationLink{
			OrganizationId:   org.ID,
			OrganizationName: org.Name,
			IsAdmin:          org.IsAdmin,
		}

		row.CreatedAt, err = ptypes.TimestampProto(org.CreatedAt)
		if err != nil {
			return nil, helpers.ErrToRPCError(err)
		}
		row.UpdatedAt, err = ptypes.TimestampProto(org.UpdatedAt)
		if err != nil {
			return nil, helpers.ErrToRPCError(err)
		}

		resp.Organizations = append(resp.Organizations, &row)
	}

	return &resp, nil
}

// Branding returns UI branding.
func (a *InternalUserAPI) Branding(ctx context.Context, req *empty.Empty) (*pb.BrandingResponse, error) {
	resp := pb.BrandingResponse{
		Logo:         brandingHeader,
		Registration: brandingRegistration,
		Footer:       brandingFooter,
	}

	return &resp, nil
}

// GlobalSearch performs a global search.
func (a *InternalUserAPI) GlobalSearch(ctx context.Context, req *pb.GlobalSearchRequest) (*pb.GlobalSearchResponse, error) {
	if err := a.validator.Validate(ctx,
		auth.ValidateActiveUser()); err != nil {
		return nil, grpc.Errorf(codes.Unauthenticated, "authentication failed: %s", err)
	}

	isAdmin, err := a.validator.GetIsAdmin(ctx)
	if err != nil {
		return nil, helpers.ErrToRPCError(err)
	}

	username, err := a.validator.GetUsername(ctx)
	if err != nil {
		return nil, helpers.ErrToRPCError(err)
	}

	results, err := storage.GlobalSearch(storage.DB(), username, isAdmin, req.Search, int(req.Limit), int(req.Offset))
	if err != nil {
		return nil, helpers.ErrToRPCError(err)
	}

	var out pb.GlobalSearchResponse

	for _, r := range results {
		res := pb.GlobalSearchResult{
			Kind:  r.Kind,
			Score: float32(r.Score),
		}

		if r.OrganizationID != nil {
			res.OrganizationId = *r.OrganizationID
		}
		if r.OrganizationName != nil {
			res.OrganizationName = *r.OrganizationName
		}

		if r.ApplicationID != nil {
			res.ApplicationId = *r.ApplicationID
		}
		if r.ApplicationName != nil {
			res.ApplicationName = *r.ApplicationName
		}

		if r.DeviceDevEUI != nil {
			res.DeviceDevEui = r.DeviceDevEUI.String()
		}
		if r.DeviceName != nil {
			res.DeviceName = *r.DeviceName
		}

		if r.GatewayMAC != nil {
			res.GatewayMac = r.GatewayMAC.String()
		}
		if r.GatewayName != nil {
			res.GatewayName = *r.GatewayName
		}

		out.Result = append(out.Result, &res)
	}

	return &out, nil
}
