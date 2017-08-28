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
	validator auth.Validator
}

// NewApplicationAPI creates a new ApplicationAPI.
func NewApplicationAPI(validator auth.Validator) *ApplicationAPI {
	return &ApplicationAPI{
		validator: validator,
	}
}

func (a *ApplicationAPI) Create(ctx context.Context, req *pb.CreateApplicationRequest) (*pb.CreateApplicationResponse, error) {
	if err := a.validator.Validate(ctx,
		auth.ValidateApplicationsAccess(auth.Create, req.OrganizationID),
	); err != nil {
		return nil, grpc.Errorf(codes.Unauthenticated, "authentication failed: %s", err)
	}

	app := storage.Application{
		Name:               req.Name,
		Description:        req.Description,
		IsABP:              req.IsABP,
		IsClassC:           req.IsClassC,
		RelaxFCnt:          req.RelaxFCnt,
		RXDelay:            uint8(req.RxDelay),
		RX1DROffset:        uint8(req.Rx1DROffset),
		RXWindow:           storage.RXWindow(req.RxWindow),
		RX2DR:              uint8(req.Rx2DR),
		ADRInterval:        req.AdrInterval,
		InstallationMargin: req.InstallationMargin,
		OrganizationID:     req.OrganizationID,
	}

	if err := storage.CreateApplication(common.DB, &app); err != nil {
		return nil, errToRPCError(err)
	}

	return &pb.CreateApplicationResponse{
		Id: app.ID,
	}, nil
}

func (a *ApplicationAPI) Get(ctx context.Context, req *pb.GetApplicationRequest) (*pb.GetApplicationResponse, error) {
	if err := a.validator.Validate(ctx,
		auth.ValidateApplicationAccess(req.Id, auth.Read),
	); err != nil {
		return nil, grpc.Errorf(codes.Unauthenticated, "authentication failed: %s", err)
	}

	app, err := storage.GetApplication(common.DB, req.Id)
	if err != nil {
		return nil, errToRPCError(err)
	}
	resp := pb.GetApplicationResponse{
		Id:                 app.ID,
		Name:               app.Name,
		Description:        app.Description,
		IsABP:              app.IsABP,
		IsClassC:           app.IsClassC,
		RxDelay:            uint32(app.RXDelay),
		Rx1DROffset:        uint32(app.RX1DROffset),
		RxWindow:           pb.RXWindow(app.RXWindow),
		Rx2DR:              uint32(app.RX2DR),
		RelaxFCnt:          app.RelaxFCnt,
		AdrInterval:        app.ADRInterval,
		InstallationMargin: app.InstallationMargin,
		OrganizationID:     app.OrganizationID,
	}

	return &resp, nil
}

func (a *ApplicationAPI) Update(ctx context.Context, req *pb.UpdateApplicationRequest) (*pb.UpdateApplicationResponse, error) {
	if err := a.validator.Validate(ctx,
		auth.ValidateApplicationAccess(req.Id, auth.Update),
	); err != nil {
		return nil, grpc.Errorf(codes.Unauthenticated, "authentication failed: %s", err)
	}

	app, err := storage.GetApplication(common.DB, req.Id)
	if err != nil {
		return nil, errToRPCError(err)
	}

	// update the fields
	app.Name = req.Name
	app.Description = req.Description
	app.IsABP = req.IsABP
	app.IsClassC = req.IsClassC
	app.RXDelay = uint8(req.RxDelay)
	app.RX1DROffset = uint8(req.Rx1DROffset)
	app.RXWindow = storage.RXWindow(req.RxWindow)
	app.RX2DR = uint8(req.Rx2DR)
	app.RelaxFCnt = req.RelaxFCnt
	app.ADRInterval = req.AdrInterval
	app.InstallationMargin = req.InstallationMargin
	app.OrganizationID = req.OrganizationID

	err = storage.UpdateApplication(common.DB, app)
	if err != nil {
		return nil, errToRPCError(err)
	}

	return &pb.UpdateApplicationResponse{}, nil
}

func (a *ApplicationAPI) Delete(ctx context.Context, req *pb.DeleteApplicationRequest) (*pb.DeleteApplicationResponse, error) {
	if err := a.validator.Validate(ctx,
		auth.ValidateApplicationAccess(req.Id, auth.Delete),
	); err != nil {
		return nil, grpc.Errorf(codes.Unauthenticated, "authentication failed: %s", err)
	}

	err := storage.DeleteApplication(common.DB, req.Id)
	if err != nil {
		return nil, errToRPCError(err)
	}
	return &pb.DeleteApplicationResponse{}, nil
}

func (a *ApplicationAPI) List(ctx context.Context, req *pb.ListApplicationRequest) (*pb.ListApplicationResponse, error) {
	if err := a.validator.Validate(ctx,
		auth.ValidateApplicationsAccess(auth.List, req.OrganizationID),
	); err != nil {
		return nil, grpc.Errorf(codes.Unauthenticated, "authentication failed: %s", err)
	}

	isAdmin, err := a.validator.GetIsAdmin(ctx)
	if err != nil {
		return nil, errToRPCError(err)
	}

	username, err := a.validator.GetUsername(ctx)
	if err != nil {
		return nil, errToRPCError(err)
	}

	var count int
	var apps []storage.Application

	if req.OrganizationID == 0 {
		if isAdmin {
			apps, err = storage.GetApplications(common.DB, int(req.Limit), int(req.Offset))
			if err != nil {
				return nil, errToRPCError(err)
			}
			count, err = storage.GetApplicationCount(common.DB)
			if err != nil {
				return nil, errToRPCError(err)
			}
		} else {
			apps, err = storage.GetApplicationsForUser(common.DB, username, 0, int(req.Limit), int(req.Offset))
			if err != nil {
				return nil, errToRPCError(err)
			}
			count, err = storage.GetApplicationCountForUser(common.DB, username, 0)
			if err != nil {
				return nil, errToRPCError(err)
			}
		}
	} else {
		if isAdmin {
			apps, err = storage.GetApplicationsForOrganizationID(common.DB, req.OrganizationID, int(req.Limit), int(req.Offset))
			if err != nil {
				return nil, errToRPCError(err)
			}
			count, err = storage.GetApplicationCountForOrganizationID(common.DB, req.OrganizationID)
			if err != nil {
				return nil, errToRPCError(err)
			}
		} else {
			apps, err = storage.GetApplicationsForUser(common.DB, username, req.OrganizationID, int(req.Limit), int(req.Offset))
			if err != nil {
				return nil, errToRPCError(err)
			}
			count, err = storage.GetApplicationCountForUser(common.DB, username, req.OrganizationID)
			if err != nil {
				return nil, errToRPCError(err)
			}
		}
	}

	resp := pb.ListApplicationResponse{
		TotalCount: int64(count),
	}
	for _, app := range apps {
		item := pb.GetApplicationResponse{
			Id:                 app.ID,
			Name:               app.Name,
			Description:        app.Description,
			IsABP:              app.IsABP,
			IsClassC:           app.IsClassC,
			RxDelay:            uint32(app.RXDelay),
			Rx1DROffset:        uint32(app.RX1DROffset),
			RxWindow:           pb.RXWindow(app.RXWindow),
			Rx2DR:              uint32(app.RX2DR),
			RelaxFCnt:          app.RelaxFCnt,
			AdrInterval:        app.ADRInterval,
			InstallationMargin: app.InstallationMargin,
			OrganizationID:     app.OrganizationID,
		}

		resp.Result = append(resp.Result, &item)
	}

	return &resp, nil
}

// ListUsers lists the users for an application.
func (a *ApplicationAPI) ListUsers(ctx context.Context, in *pb.ListApplicationUsersRequest) (*pb.ListApplicationUsersResponse, error) {
	if err := a.validator.Validate(ctx,
		auth.ValidateApplicationUsersAccess(in.Id, auth.List),
	); err != nil {
		return nil, grpc.Errorf(codes.Unauthenticated, "authentication failed: %s", err)
	}

	total, err := storage.GetApplicationUsersCount(common.DB, in.Id)
	if nil != err {
		return nil, errToRPCError(err)
	}

	userAccess, err := storage.GetApplicationUsers(common.DB, in.Id, int(in.Limit), int(in.Offset))
	if nil != err {
		return nil, errToRPCError(err)
	}

	appUsers := make([]*pb.GetApplicationUserResponse, len(userAccess))
	for i, ua := range userAccess {
		// Get the user information
		appUsers[i] = &pb.GetApplicationUserResponse{
			Id:       ua.UserID,
			Username: ua.Username,
			IsAdmin:  ua.IsAdmin,
		}
	}
	return &pb.ListApplicationUsersResponse{TotalCount: total, Result: appUsers}, nil
}

// SetUsers sets the users for an application.  Any existing users are
// dropped in favor of this list.
func (a *ApplicationAPI) AddUser(ctx context.Context, in *pb.AddApplicationUserRequest) (*pb.EmptyApplicationUserResponse, error) {
	if err := a.validator.Validate(ctx,
		auth.ValidateApplicationUsersAccess(in.Id, auth.Create),
	); err != nil {
		return nil, grpc.Errorf(codes.Unauthenticated, "authentication failed: %s", err)
	}

	err := storage.CreateUserForApplication(common.DB, in.Id, in.UserID, in.IsAdmin)
	if nil != err {
		return nil, errToRPCError(err)
	}
	return &pb.EmptyApplicationUserResponse{}, nil
}

// GetUser gets the user that is associated with the application.
func (a *ApplicationAPI) GetUser(ctx context.Context, in *pb.ApplicationUserRequest) (*pb.GetApplicationUserResponse, error) {
	if err := a.validator.Validate(ctx,
		auth.ValidateApplicationUserAccess(in.Id, in.UserID, auth.Read),
	); err != nil {
		return nil, grpc.Errorf(codes.Unauthenticated, "authentication failed: %s", err)
	}

	ua, err := storage.GetUserForApplication(common.DB, in.Id, in.UserID)
	if nil != err {
		return nil, errToRPCError(err)
	}

	appUser := &pb.GetApplicationUserResponse{
		Id:       ua.UserID,
		Username: ua.Username,
		IsAdmin:  ua.IsAdmin,
	}

	return appUser, nil
}

// PutUser sets the user's access to the associated application.
func (a *ApplicationAPI) UpdateUser(ctx context.Context, in *pb.UpdateApplicationUserRequest) (*pb.EmptyApplicationUserResponse, error) {
	if err := a.validator.Validate(ctx,
		auth.ValidateApplicationUserAccess(in.Id, in.UserID, auth.Update),
	); err != nil {
		return nil, grpc.Errorf(codes.Unauthenticated, "authentication failed: %s", err)
	}

	err := storage.UpdateUserForApplication(common.DB, in.Id, in.UserID, in.IsAdmin)
	if nil != err {
		return nil, errToRPCError(err)
	}
	return &pb.EmptyApplicationUserResponse{}, nil
}

// DeleteUser deletes the user's access to the associated application.
func (a *ApplicationAPI) DeleteUser(ctx context.Context, in *pb.ApplicationUserRequest) (*pb.EmptyApplicationUserResponse, error) {
	if err := a.validator.Validate(ctx,
		auth.ValidateApplicationUserAccess(in.Id, in.UserID, auth.Delete),
	); err != nil {
		return nil, grpc.Errorf(codes.Unauthenticated, "authentication failed: %s", err)
	}

	err := storage.DeleteUserForApplication(common.DB, in.Id, in.UserID)
	if nil != err {
		return nil, errToRPCError(err)
	}
	return &pb.EmptyApplicationUserResponse{}, nil
}
