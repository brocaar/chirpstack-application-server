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
	}
	if req.ChannelListID > 0 {
		app.ChannelListID = &req.ChannelListID
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
	}

	if app.ChannelListID != nil {
		resp.ChannelListID = *app.ChannelListID
	}

	return &resp, nil
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
	app.IsABP = req.IsABP
	app.IsClassC = req.IsClassC
	app.RXDelay = uint8(req.RxDelay)
	app.RX1DROffset = uint8(req.Rx1DROffset)
	app.RXWindow = storage.RXWindow(req.RxWindow)
	app.RX2DR = uint8(req.Rx2DR)
	app.RelaxFCnt = req.RelaxFCnt
	app.ADRInterval = req.AdrInterval
	app.InstallationMargin = req.InstallationMargin
	if req.ChannelListID > 0 {
		app.ChannelListID = &req.ChannelListID
	} else {
		app.ChannelListID = nil
	}

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
		}

		if app.ChannelListID != nil {
			item.ChannelListID = *app.ChannelListID
		}

		resp.Result = append(resp.Result, &item)
	}

	return &resp, nil
}
