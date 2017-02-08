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
		auth.ValidateApplicationName(req.Name),
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

	return &pb.CreateApplicationResponse{}, nil
}

func (a *ApplicationAPI) Get(ctx context.Context, req *pb.GetApplicationRequest) (*pb.GetApplicationResponse, error) {
	if err := a.validator.Validate(ctx,
		auth.ValidateAPIMethod("Application.Get"),
		auth.ValidateApplicationName(req.ApplicationName),
	); err != nil {
		return nil, grpc.Errorf(codes.Unauthenticated, "authentication failed: %s", err)
	}

	app, err := storage.GetApplicationByName(a.ctx.DB, req.ApplicationName)
	if err != nil {
		return nil, grpc.Errorf(codes.Unknown, err.Error())
	}
	return &pb.GetApplicationResponse{
		Name:        app.Name,
		Description: app.Description,
	}, nil
}

func (a *ApplicationAPI) Update(ctx context.Context, req *pb.UpdateApplicationRequest) (*pb.UpdateApplicationResponse, error) {
	if err := a.validator.Validate(ctx,
		auth.ValidateAPIMethod("Application.Update"),
		auth.ValidateApplicationName(req.ApplicationName),
	); err != nil {
		return nil, grpc.Errorf(codes.Unauthenticated, "authentication failed: %s", err)
	}

	app, err := storage.GetApplicationByName(a.ctx.DB, req.ApplicationName)
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
		auth.ValidateApplicationName(req.ApplicationName),
	); err != nil {
		return nil, grpc.Errorf(codes.Unauthenticated, "authentication failed: %s", err)
	}

	err := storage.DeleteApplicationByname(a.ctx.DB, req.ApplicationName)
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
			Name:        app.Name,
			Description: app.Description,
		})
	}

	return &resp, nil
}
