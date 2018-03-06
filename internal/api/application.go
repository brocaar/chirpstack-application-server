package api

import (
	"encoding/json"

	"github.com/jmoiron/sqlx"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

	pb "github.com/gusseleet/lora-app-server/api"
	"github.com/gusseleet/lora-app-server/internal/api/auth"
	"github.com/gusseleet/lora-app-server/internal/codec"
	"github.com/gusseleet/lora-app-server/internal/config"
	"github.com/gusseleet/lora-app-server/internal/handler"
	"github.com/gusseleet/lora-app-server/internal/handler/httphandler"
	"github.com/gusseleet/lora-app-server/internal/storage"
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

// Create creates the given application.
func (a *ApplicationAPI) Create(ctx context.Context, req *pb.CreateApplicationRequest) (*pb.CreateApplicationResponse, error) {
	if err := a.validator.Validate(ctx,
		auth.ValidateApplicationsAccess(auth.Create, req.OrganizationID),
	); err != nil {
		return nil, grpc.Errorf(codes.Unauthenticated, "authentication failed: %s", err)
	}

	app := storage.Application{
		Name:                 req.Name,
		Description:          req.Description,
		OrganizationID:       req.OrganizationID,
		ServiceProfileID:     req.ServiceProfileID,
		PayloadCodec:         codec.Type(req.PayloadCodec),
		PayloadEncoderScript: req.PayloadEncoderScript,
		PayloadDecoderScript: req.PayloadDecoderScript,
	}

	if err := storage.CreateApplication(config.C.PostgreSQL.DB, &app); err != nil {
		return nil, errToRPCError(err)
	}

	return &pb.CreateApplicationResponse{
		Id: app.ID,
	}, nil
}

// Get returns the requested application.
func (a *ApplicationAPI) Get(ctx context.Context, req *pb.GetApplicationRequest) (*pb.GetApplicationResponse, error) {
	if err := a.validator.Validate(ctx,
		auth.ValidateApplicationAccess(req.Id, auth.Read),
	); err != nil {
		return nil, grpc.Errorf(codes.Unauthenticated, "authentication failed: %s", err)
	}

	app, err := storage.GetApplication(config.C.PostgreSQL.DB, req.Id)
	if err != nil {
		return nil, errToRPCError(err)
	}
	resp := pb.GetApplicationResponse{
		Id:                   app.ID,
		Name:                 app.Name,
		Description:          app.Description,
		OrganizationID:       app.OrganizationID,
		ServiceProfileID:     app.ServiceProfileID,
		PayloadCodec:         string(app.PayloadCodec),
		PayloadEncoderScript: app.PayloadEncoderScript,
		PayloadDecoderScript: app.PayloadDecoderScript,
	}

	return &resp, nil
}

// Update updates the given application.
func (a *ApplicationAPI) Update(ctx context.Context, req *pb.UpdateApplicationRequest) (*pb.UpdateApplicationResponse, error) {
	if err := a.validator.Validate(ctx,
		auth.ValidateApplicationAccess(req.Id, auth.Update),
	); err != nil {
		return nil, grpc.Errorf(codes.Unauthenticated, "authentication failed: %s", err)
	}

	app, err := storage.GetApplication(config.C.PostgreSQL.DB, req.Id)
	if err != nil {
		return nil, errToRPCError(err)
	}

	// update the fields
	app.Name = req.Name
	app.Description = req.Description
	app.ServiceProfileID = req.ServiceProfileID
	app.PayloadCodec = codec.Type(req.PayloadCodec)
	app.PayloadEncoderScript = req.PayloadEncoderScript
	app.PayloadDecoderScript = req.PayloadDecoderScript

	err = storage.UpdateApplication(config.C.PostgreSQL.DB, app)
	if err != nil {
		return nil, errToRPCError(err)
	}

	return &pb.UpdateApplicationResponse{}, nil
}

// Delete deletes the given application.
func (a *ApplicationAPI) Delete(ctx context.Context, req *pb.DeleteApplicationRequest) (*pb.DeleteApplicationResponse, error) {
	if err := a.validator.Validate(ctx,
		auth.ValidateApplicationAccess(req.Id, auth.Delete),
	); err != nil {
		return nil, grpc.Errorf(codes.Unauthenticated, "authentication failed: %s", err)
	}

	err := storage.Transaction(config.C.PostgreSQL.DB, func(tx sqlx.Ext) error {
		err := storage.DeleteApplication(tx, req.Id)
		if err != nil {
			return errToRPCError(err)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return &pb.DeleteApplicationResponse{}, nil
}

// List lists the available applications.
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
	var apps []storage.ApplicationListItem

	if req.OrganizationID == 0 {
		if isAdmin {
			apps, err = storage.GetApplications(config.C.PostgreSQL.DB, int(req.Limit), int(req.Offset))
			if err != nil {
				return nil, errToRPCError(err)
			}
			count, err = storage.GetApplicationCount(config.C.PostgreSQL.DB)
			if err != nil {
				return nil, errToRPCError(err)
			}
		} else {
			apps, err = storage.GetApplicationsForUser(config.C.PostgreSQL.DB, username, 0, int(req.Limit), int(req.Offset))
			if err != nil {
				return nil, errToRPCError(err)
			}
			count, err = storage.GetApplicationCountForUser(config.C.PostgreSQL.DB, username, 0)
			if err != nil {
				return nil, errToRPCError(err)
			}
		}
	} else {
		if isAdmin {
			apps, err = storage.GetApplicationsForOrganizationID(config.C.PostgreSQL.DB, req.OrganizationID, int(req.Limit), int(req.Offset))
			if err != nil {
				return nil, errToRPCError(err)
			}
			count, err = storage.GetApplicationCountForOrganizationID(config.C.PostgreSQL.DB, req.OrganizationID)
			if err != nil {
				return nil, errToRPCError(err)
			}
		} else {
			apps, err = storage.GetApplicationsForUser(config.C.PostgreSQL.DB, username, req.OrganizationID, int(req.Limit), int(req.Offset))
			if err != nil {
				return nil, errToRPCError(err)
			}
			count, err = storage.GetApplicationCountForUser(config.C.PostgreSQL.DB, username, req.OrganizationID)
			if err != nil {
				return nil, errToRPCError(err)
			}
		}
	}

	resp := pb.ListApplicationResponse{
		TotalCount: int64(count),
	}
	for _, app := range apps {
		item := pb.ApplicationListItem{
			Id:                 app.ID,
			Name:               app.Name,
			Description:        app.Description,
			OrganizationID:     app.OrganizationID,
			ServiceProfileID:   app.ServiceProfileID,
			ServiceProfileName: app.ServiceProfileName,
		}

		resp.Result = append(resp.Result, &item)
	}

	return &resp, nil
}

// CreateHTTPIntegration creates an HTTP application-integration.
func (a *ApplicationAPI) CreateHTTPIntegration(ctx context.Context, in *pb.HTTPIntegration) (*pb.EmptyResponse, error) {
	if err := a.validator.Validate(ctx,
		auth.ValidateApplicationAccess(in.Id, auth.Update),
	); err != nil {
		return nil, grpc.Errorf(codes.Unauthenticated, "authentication failed: %s", err)
	}

	headers := make(map[string]string)
	for _, h := range in.Headers {
		headers[h.Key] = h.Value
	}

	conf := httphandler.HandlerConfig{
		Headers:              headers,
		DataUpURL:            in.DataUpURL,
		JoinNotificationURL:  in.JoinNotificationURL,
		ACKNotificationURL:   in.AckNotificationURL,
		ErrorNotificationURL: in.ErrorNotificationURL,
	}
	if err := conf.Validate(); err != nil {
		return nil, errToRPCError(err)
	}

	confJSON, err := json.Marshal(conf)
	if err != nil {
		return nil, errToRPCError(err)
	}

	integration := storage.Integration{
		ApplicationID: in.Id,
		Kind:          handler.HTTPHandlerKind,
		Settings:      confJSON,
	}
	if err = storage.CreateIntegration(config.C.PostgreSQL.DB, &integration); err != nil {
		return nil, errToRPCError(err)
	}

	return &pb.EmptyResponse{}, nil
}

// GetHTTPIntegration returns the HTTP application-itegration.
func (a *ApplicationAPI) GetHTTPIntegration(ctx context.Context, in *pb.GetHTTPIntegrationRequest) (*pb.HTTPIntegration, error) {
	if err := a.validator.Validate(ctx,
		auth.ValidateApplicationAccess(in.Id, auth.Update),
	); err != nil {
		return nil, grpc.Errorf(codes.Unauthenticated, "authentication failed: %s", err)
	}

	integration, err := storage.GetIntegrationByApplicationID(config.C.PostgreSQL.DB, in.Id, handler.HTTPHandlerKind)
	if err != nil {
		return nil, errToRPCError(err)
	}

	var conf httphandler.HandlerConfig
	if err = json.Unmarshal(integration.Settings, &conf); err != nil {
		return nil, errToRPCError(err)
	}

	var headers []*pb.HTTPIntegrationHeader
	for k, v := range conf.Headers {
		headers = append(headers, &pb.HTTPIntegrationHeader{
			Key:   k,
			Value: v,
		})

	}

	return &pb.HTTPIntegration{
		Id:                   integration.ApplicationID,
		Headers:              headers,
		DataUpURL:            conf.DataUpURL,
		JoinNotificationURL:  conf.JoinNotificationURL,
		AckNotificationURL:   conf.ACKNotificationURL,
		ErrorNotificationURL: conf.ErrorNotificationURL,
	}, nil
}

// UpdateHTTPIntegration updates the HTTP application-integration.
func (a *ApplicationAPI) UpdateHTTPIntegration(ctx context.Context, in *pb.HTTPIntegration) (*pb.EmptyResponse, error) {
	if err := a.validator.Validate(ctx,
		auth.ValidateApplicationAccess(in.Id, auth.Update),
	); err != nil {
		return nil, grpc.Errorf(codes.Unauthenticated, "authentication failed: %s", err)
	}

	integration, err := storage.GetIntegrationByApplicationID(config.C.PostgreSQL.DB, in.Id, handler.HTTPHandlerKind)
	if err != nil {
		return nil, errToRPCError(err)
	}

	headers := make(map[string]string)
	for _, h := range in.Headers {
		headers[h.Key] = h.Value
	}

	conf := httphandler.HandlerConfig{
		Headers:              headers,
		DataUpURL:            in.DataUpURL,
		JoinNotificationURL:  in.JoinNotificationURL,
		ACKNotificationURL:   in.AckNotificationURL,
		ErrorNotificationURL: in.ErrorNotificationURL,
	}
	if err := conf.Validate(); err != nil {
		return nil, errToRPCError(err)
	}

	confJSON, err := json.Marshal(conf)
	if err != nil {
		return nil, errToRPCError(err)
	}
	integration.Settings = confJSON

	if err = storage.UpdateIntegration(config.C.PostgreSQL.DB, &integration); err != nil {
		return nil, errToRPCError(err)
	}

	return &pb.EmptyResponse{}, nil
}

// DeleteHTTPIntegration deletes the application-integration of the given type.
func (a *ApplicationAPI) DeleteHTTPIntegration(ctx context.Context, in *pb.DeleteIntegrationRequest) (*pb.EmptyResponse, error) {
	if err := a.validator.Validate(ctx,
		auth.ValidateApplicationAccess(in.Id, auth.Update),
	); err != nil {
		return nil, grpc.Errorf(codes.Unauthenticated, "authentication failed: %s", err)
	}

	integration, err := storage.GetIntegrationByApplicationID(config.C.PostgreSQL.DB, in.Id, handler.HTTPHandlerKind)
	if err != nil {
		return nil, errToRPCError(err)
	}

	if err = storage.DeleteIntegration(config.C.PostgreSQL.DB, integration.ID); err != nil {
		return nil, errToRPCError(err)
	}

	return &pb.EmptyResponse{}, nil
}

// ListIntegrations lists all configured integrations.
func (a *ApplicationAPI) ListIntegrations(ctx context.Context, in *pb.ListIntegrationRequest) (*pb.ListIntegrationResponse, error) {
	if err := a.validator.Validate(ctx,
		auth.ValidateApplicationAccess(in.Id, auth.Update),
	); err != nil {
		return nil, grpc.Errorf(codes.Unauthenticated, "authentication failed: %s", err)
	}

	integrations, err := storage.GetIntegrationsForApplicationID(config.C.PostgreSQL.DB, in.Id)
	if err != nil {
		return nil, errToRPCError(err)
	}

	var out pb.ListIntegrationResponse
	for _, integration := range integrations {
		switch integration.Kind {
		case handler.HTTPHandlerKind:
			out.Kinds = append(out.Kinds, pb.IntegrationKind_HTTP)
		default:
			return nil, grpc.Errorf(codes.Internal, "unknown integration kind: %s", integration.Kind)
		}
	}

	return &out, nil
}
