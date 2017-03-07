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

// ApplicationAPI exports the Application related functions.
type ApplicationAPI struct {
	ctx       common.Context
	validator auth.Validator
}

// NewApplicationAPI creates a new ApplicationAPI.
func NewApplicationAPI(ctx common.Context, validator auth.Validator) *ApplicationAPI {
	return &ApplicationAPI{
		ctx:       ctx,
		validator: validator,
	}
}

func (a *ApplicationAPI) Create(ctx context.Context, req *pb.CreateApplicationRequest) (*pb.CreateApplicationResponse, error) {
	if err := a.validator.Validate(ctx,
		auth.ValidateAPIMethod("Application.Create"),
	); err != nil {
		return nil, grpc.Errorf(codes.Unauthenticated, "authentication failed: %s", err)
	}

	app := storage.Application{
		Name:        req.Name,
		Description: req.Description,
	}

	if err := storage.CreateApplication(a.ctx.DB, &app); err != nil {
		return nil, grpc.Errorf(codes.Unknown, err.Error())
	}

	return &pb.CreateApplicationResponse{
		Id: app.ID,
	}, nil
}

func (a *ApplicationAPI) Get(ctx context.Context, req *pb.GetApplicationRequest) (*pb.GetApplicationResponse, error) {
	if err := a.validator.Validate(ctx,
		auth.ValidateAPIMethod("Application.Get"),
		auth.ValidateApplicationID(req.Id),
	); err != nil {
		return nil, grpc.Errorf(codes.Unauthenticated, "authentication failed: %s", err)
	}

	app, err := storage.GetApplication(a.ctx.DB, req.Id)
	if err != nil {
		return nil, grpc.Errorf(codes.Unknown, err.Error())
	}
	return &pb.GetApplicationResponse{
		Id:          app.ID,
		Name:        app.Name,
		Description: app.Description,
	}, nil
}

func (a *ApplicationAPI) Update(ctx context.Context, req *pb.UpdateApplicationRequest) (*pb.UpdateApplicationResponse, error) {
	if err := a.validator.Validate(ctx,
		auth.ValidateAPIMethod("Application.Update"),
		auth.ValidateApplicationID(req.Id),
	); err != nil {
		return nil, grpc.Errorf(codes.Unauthenticated, "authentication failed: %s", err)
	}

	app, err := storage.GetApplication(a.ctx.DB, req.Id)
	if err != nil {
		return nil, grpc.Errorf(codes.Unknown, err.Error())
	}

	// update the fields
	app.Name = req.Name
	app.Description = req.Description

	err = storage.UpdateApplication(a.ctx.DB, app)
	if err != nil {
		return nil, grpc.Errorf(codes.Unknown, err.Error())
	}

	return &pb.UpdateApplicationResponse{}, nil
}

func (a *ApplicationAPI) Delete(ctx context.Context, req *pb.DeleteApplicationRequest) (*pb.DeleteApplicationResponse, error) {
	if err := a.validator.Validate(ctx,
		auth.ValidateAPIMethod("Application.Delete"),
		auth.ValidateApplicationID(req.Id),
	); err != nil {
		return nil, grpc.Errorf(codes.Unauthenticated, "authentication failed: %s", err)
	}

	err := storage.DeleteApplication(a.ctx.DB, req.Id)
	if err != nil {
		return nil, grpc.Errorf(codes.Unknown, err.Error())
	}
	return &pb.DeleteApplicationResponse{}, nil
}

func (a *ApplicationAPI) List(ctx context.Context, req *pb.ListApplicationRequest) (*pb.ListApplicationResponse, error) {
	if err := a.validator.Validate(ctx,
		auth.ValidateAPIMethod("Application.List"),
	); err != nil {
		return nil, grpc.Errorf(codes.Unauthenticated, "authentication failed: %s", err)
	}

	apps, err := storage.GetApplications(a.ctx.DB, int(req.Limit), int(req.Offset))
	if err != nil {
		return nil, grpc.Errorf(codes.Internal, err.Error())
	}

	count, err := storage.GetApplicationCount(a.ctx.DB)
	if err != nil {
		return nil, grpc.Errorf(codes.Internal, err.Error())
	}

	resp := pb.ListApplicationResponse{
		TotalCount: int64(count),
	}
	for _, app := range apps {
		resp.Result = append(resp.Result, &pb.GetApplicationResponse{
			Id:          app.ID,
			Name:        app.Name,
			Description: app.Description,
		})
	}

	return &resp, nil
}

	// ListUsers lists the users for an application.
func (a *ApplicationAPI) ListUsers(ctx context.Context, in *pb.ListApplicationUsersRequest) (*pb.ListApplicationUsersResponse, error) {
	total, err := storage.GetApplicationUsersCount( a.ctx.DB, in.Id )
	if nil != err {
		return nil, grpc.Errorf(codes.Internal, err.Error())
	}
	
	userAccess, err := storage.GetApplicationUsers( a.ctx.DB, in.Id, int(in.Limit), int(in.Offset) )
	if nil != err {
		return nil, grpc.Errorf(codes.Internal, err.Error())
	}
	
	appUsers := make( []*pb.GetApplicationUserResponse, len( userAccess ) )
	for i, ua := range userAccess {
		// Get the user information
		appUsers[ i ] = &pb.GetApplicationUserResponse{
			Id: ua.UserId,
			Username: ua.Username,
			IsAdmin: ua.IsAdmin,
		}
	}
	return &pb.ListApplicationUsersResponse{ TotalCount: total, Result: appUsers }, nil
}

	// SetUsers sets the users for an application.  Any existing users are
	// dropped in favor of this list.
func (a *ApplicationAPI) AddUser(ctx context.Context, in *pb.AddApplicationUserRequest) (*pb.EmptyApplicationUserResponse, error) {
	err := storage.CreateUserForApplication( a.ctx.DB, in.Id, in.Userid, in.IsAdmin )
	if nil != err {
		return nil, grpc.Errorf( codes.InvalidArgument, err.Error() )
	}
	return &pb.EmptyApplicationUserResponse{}, nil
}

	// GetUser gets the user that is associated with the application.
func (a *ApplicationAPI) GetUser(ctx context.Context, in *pb.ApplicationUserRequest) (*pb.GetApplicationUserResponse, error) {
	ua, err := storage.GetUserForApplication( a.ctx.DB, in.Id, in.Userid )
	if nil != err {
		return nil, grpc.Errorf(codes.Internal, err.Error())
	}
	
	appUser := &pb.GetApplicationUserResponse{
			Id: ua.UserId,
			Username: ua.Username,
			IsAdmin: ua.IsAdmin,
		}

	return appUser, nil
}

	// PutUser sets the user's access to the associated application.
func (a *ApplicationAPI) UpdateUser(ctx context.Context, in *pb.UpdateApplicationUserRequest) (*pb.EmptyApplicationUserResponse, error) {
	err := storage.UpdateUserForApplication( a.ctx.DB, in.Id, in.Userid, in.IsAdmin )
	if nil != err {
		return nil, grpc.Errorf( codes.InvalidArgument, err.Error() )
	}
	return &pb.EmptyApplicationUserResponse{}, nil
}

	// DeleteUser deletes the user's access to the associated application.
func (a *ApplicationAPI) DeleteUser(ctx context.Context, in *pb.ApplicationUserRequest) (*pb.EmptyApplicationUserResponse, error) {
	err := storage.DeleteUserForApplication( a.ctx.DB, in.Id, in.Userid )
	if nil != err {
		return nil, grpc.Errorf( codes.InvalidArgument, err.Error() )
	}
	return &pb.EmptyApplicationUserResponse{}, nil
}
