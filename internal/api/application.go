package api

import (
	"encoding/json"
	"strings"

	"github.com/gofrs/uuid"
	"github.com/golang/protobuf/ptypes/empty"

	"github.com/jmoiron/sqlx"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

	pb "github.com/brocaar/lora-app-server/api"
	"github.com/brocaar/lora-app-server/internal/api/auth"
	"github.com/brocaar/lora-app-server/internal/codec"
	"github.com/brocaar/lora-app-server/internal/config"
	"github.com/brocaar/lora-app-server/internal/integration"
	"github.com/brocaar/lora-app-server/internal/integration/http"
	"github.com/brocaar/lora-app-server/internal/integration/influxdb"
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

// Create creates the given application.
func (a *ApplicationAPI) Create(ctx context.Context, req *pb.CreateApplicationRequest) (*pb.CreateApplicationResponse, error) {
	if req.Application == nil {
		return nil, grpc.Errorf(codes.InvalidArgument, "application must not be nil")
	}

	if err := a.validator.Validate(ctx,
		auth.ValidateApplicationsAccess(auth.Create, req.Application.OrganizationId),
	); err != nil {
		return nil, grpc.Errorf(codes.Unauthenticated, "authentication failed: %s", err)
	}

	spID, err := uuid.FromString(req.Application.ServiceProfileId)
	if err != nil {
		return nil, errToRPCError(err)
	}

	app := storage.Application{
		Name:                 req.Application.Name,
		Description:          req.Application.Description,
		OrganizationID:       req.Application.OrganizationId,
		ServiceProfileID:     spID,
		PayloadCodec:         codec.Type(req.Application.PayloadCodec),
		PayloadEncoderScript: req.Application.PayloadEncoderScript,
		PayloadDecoderScript: req.Application.PayloadDecoderScript,
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
		Application: &pb.Application{
			Id:                   app.ID,
			Name:                 app.Name,
			Description:          app.Description,
			OrganizationId:       app.OrganizationID,
			ServiceProfileId:     app.ServiceProfileID.String(),
			PayloadCodec:         string(app.PayloadCodec),
			PayloadEncoderScript: app.PayloadEncoderScript,
			PayloadDecoderScript: app.PayloadDecoderScript,
		},
	}

	return &resp, nil
}

// Update updates the given application.
func (a *ApplicationAPI) Update(ctx context.Context, req *pb.UpdateApplicationRequest) (*empty.Empty, error) {
	if req.Application == nil {
		return nil, grpc.Errorf(codes.InvalidArgument, "application must not be nil")
	}

	if err := a.validator.Validate(ctx,
		auth.ValidateApplicationAccess(req.Application.Id, auth.Update),
	); err != nil {
		return nil, grpc.Errorf(codes.Unauthenticated, "authentication failed: %s", err)
	}

	app, err := storage.GetApplication(config.C.PostgreSQL.DB, req.Application.Id)
	if err != nil {
		return nil, errToRPCError(err)
	}

	spID, err := uuid.FromString(req.Application.ServiceProfileId)
	if err != nil {
		return nil, errToRPCError(err)
	}

	// update the fields
	app.Name = req.Application.Name
	app.Description = req.Application.Description
	app.ServiceProfileID = spID
	app.PayloadCodec = codec.Type(req.Application.PayloadCodec)
	app.PayloadEncoderScript = req.Application.PayloadEncoderScript
	app.PayloadDecoderScript = req.Application.PayloadDecoderScript

	err = storage.UpdateApplication(config.C.PostgreSQL.DB, app)
	if err != nil {
		return nil, errToRPCError(err)
	}

	return &empty.Empty{}, nil
}

// Delete deletes the given application.
func (a *ApplicationAPI) Delete(ctx context.Context, req *pb.DeleteApplicationRequest) (*empty.Empty, error) {
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

	return &empty.Empty{}, nil
}

// List lists the available applications.
func (a *ApplicationAPI) List(ctx context.Context, req *pb.ListApplicationRequest) (*pb.ListApplicationResponse, error) {
	if err := a.validator.Validate(ctx,
		auth.ValidateApplicationsAccess(auth.List, req.OrganizationId),
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

	if req.OrganizationId == 0 {
		if isAdmin {
			apps, err = storage.GetApplications(config.C.PostgreSQL.DB, int(req.Limit), int(req.Offset), req.Search)
			if err != nil {
				return nil, errToRPCError(err)
			}
			count, err = storage.GetApplicationCount(config.C.PostgreSQL.DB, req.Search)
			if err != nil {
				return nil, errToRPCError(err)
			}
		} else {
			apps, err = storage.GetApplicationsForUser(config.C.PostgreSQL.DB, username, 0, int(req.Limit), int(req.Offset), req.Search)
			if err != nil {
				return nil, errToRPCError(err)
			}
			count, err = storage.GetApplicationCountForUser(config.C.PostgreSQL.DB, username, 0, req.Search)
			if err != nil {
				return nil, errToRPCError(err)
			}
		}
	} else {
		if isAdmin {
			apps, err = storage.GetApplicationsForOrganizationID(config.C.PostgreSQL.DB, req.OrganizationId, int(req.Limit), int(req.Offset), req.Search)
			if err != nil {
				return nil, errToRPCError(err)
			}
			count, err = storage.GetApplicationCountForOrganizationID(config.C.PostgreSQL.DB, req.OrganizationId, req.Search)
			if err != nil {
				return nil, errToRPCError(err)
			}
		} else {
			apps, err = storage.GetApplicationsForUser(config.C.PostgreSQL.DB, username, req.OrganizationId, int(req.Limit), int(req.Offset), req.Search)
			if err != nil {
				return nil, errToRPCError(err)
			}
			count, err = storage.GetApplicationCountForUser(config.C.PostgreSQL.DB, username, req.OrganizationId, req.Search)
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
			OrganizationId:     app.OrganizationID,
			ServiceProfileId:   app.ServiceProfileID.String(),
			ServiceProfileName: app.ServiceProfileName,
		}

		resp.Result = append(resp.Result, &item)
	}

	return &resp, nil
}

// CreateHTTPIntegration creates an HTTP application-integration.
func (a *ApplicationAPI) CreateHTTPIntegration(ctx context.Context, in *pb.CreateHTTPIntegrationRequest) (*empty.Empty, error) {
	if in.Integration == nil {
		return nil, grpc.Errorf(codes.InvalidArgument, "integration must not be nil")
	}

	if err := a.validator.Validate(ctx,
		auth.ValidateApplicationAccess(in.Integration.ApplicationId, auth.Update),
	); err != nil {
		return nil, grpc.Errorf(codes.Unauthenticated, "authentication failed: %s", err)
	}

	headers := make(map[string]string)
	for _, h := range in.Integration.Headers {
		headers[h.Key] = h.Value
	}

	conf := http.Config{
		Headers:                 headers,
		DataUpURL:               in.Integration.UplinkDataUrl,
		JoinNotificationURL:     in.Integration.JoinNotificationUrl,
		ACKNotificationURL:      in.Integration.AckNotificationUrl,
		ErrorNotificationURL:    in.Integration.ErrorNotificationUrl,
		StatusNotificationURL:   in.Integration.StatusNotificationUrl,
		LocationNotificationURL: in.Integration.LocationNotificationUrl,
	}
	if err := conf.Validate(); err != nil {
		return nil, errToRPCError(err)
	}

	confJSON, err := json.Marshal(conf)
	if err != nil {
		return nil, errToRPCError(err)
	}

	integration := storage.Integration{
		ApplicationID: in.Integration.ApplicationId,
		Kind:          integration.HTTP,
		Settings:      confJSON,
	}
	if err = storage.CreateIntegration(config.C.PostgreSQL.DB, &integration); err != nil {
		return nil, errToRPCError(err)
	}

	return &empty.Empty{}, nil
}

// GetHTTPIntegration returns the HTTP application-itegration.
func (a *ApplicationAPI) GetHTTPIntegration(ctx context.Context, in *pb.GetHTTPIntegrationRequest) (*pb.GetHTTPIntegrationResponse, error) {
	if err := a.validator.Validate(ctx,
		auth.ValidateApplicationAccess(in.ApplicationId, auth.Update),
	); err != nil {
		return nil, grpc.Errorf(codes.Unauthenticated, "authentication failed: %s", err)
	}

	integration, err := storage.GetIntegrationByApplicationID(config.C.PostgreSQL.DB, in.ApplicationId, integration.HTTP)
	if err != nil {
		return nil, errToRPCError(err)
	}

	var conf http.Config
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

	return &pb.GetHTTPIntegrationResponse{
		Integration: &pb.HTTPIntegration{
			ApplicationId:           integration.ApplicationID,
			Headers:                 headers,
			UplinkDataUrl:           conf.DataUpURL,
			JoinNotificationUrl:     conf.JoinNotificationURL,
			AckNotificationUrl:      conf.ACKNotificationURL,
			ErrorNotificationUrl:    conf.ErrorNotificationURL,
			StatusNotificationUrl:   conf.StatusNotificationURL,
			LocationNotificationUrl: conf.LocationNotificationURL,
		},
	}, nil
}

// UpdateHTTPIntegration updates the HTTP application-integration.
func (a *ApplicationAPI) UpdateHTTPIntegration(ctx context.Context, in *pb.UpdateHTTPIntegrationRequest) (*empty.Empty, error) {
	if in.Integration == nil {
		return nil, grpc.Errorf(codes.InvalidArgument, "integration must not be nil")
	}

	if err := a.validator.Validate(ctx,
		auth.ValidateApplicationAccess(in.Integration.ApplicationId, auth.Update),
	); err != nil {
		return nil, grpc.Errorf(codes.Unauthenticated, "authentication failed: %s", err)
	}

	integration, err := storage.GetIntegrationByApplicationID(config.C.PostgreSQL.DB, in.Integration.ApplicationId, integration.HTTP)
	if err != nil {
		return nil, errToRPCError(err)
	}

	headers := make(map[string]string)
	for _, h := range in.Integration.Headers {
		headers[h.Key] = h.Value
	}

	conf := http.Config{
		Headers:                 headers,
		DataUpURL:               in.Integration.UplinkDataUrl,
		JoinNotificationURL:     in.Integration.JoinNotificationUrl,
		ACKNotificationURL:      in.Integration.AckNotificationUrl,
		ErrorNotificationURL:    in.Integration.ErrorNotificationUrl,
		StatusNotificationURL:   in.Integration.StatusNotificationUrl,
		LocationNotificationURL: in.Integration.LocationNotificationUrl,
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

	return &empty.Empty{}, nil
}

// DeleteHTTPIntegration deletes the application-integration of the given type.
func (a *ApplicationAPI) DeleteHTTPIntegration(ctx context.Context, in *pb.DeleteHTTPIntegrationRequest) (*empty.Empty, error) {
	if err := a.validator.Validate(ctx,
		auth.ValidateApplicationAccess(in.ApplicationId, auth.Update),
	); err != nil {
		return nil, grpc.Errorf(codes.Unauthenticated, "authentication failed: %s", err)
	}

	integration, err := storage.GetIntegrationByApplicationID(config.C.PostgreSQL.DB, in.ApplicationId, integration.HTTP)
	if err != nil {
		return nil, errToRPCError(err)
	}

	if err = storage.DeleteIntegration(config.C.PostgreSQL.DB, integration.ID); err != nil {
		return nil, errToRPCError(err)
	}

	return &empty.Empty{}, nil
}

// CreateInfluxDBIntegration create an InfluxDB application-integration.
func (a *ApplicationAPI) CreateInfluxDBIntegration(ctx context.Context, in *pb.CreateInfluxDBIntegrationRequest) (*empty.Empty, error) {
	if in.Integration == nil {
		return nil, grpc.Errorf(codes.InvalidArgument, "integration must not be nil")
	}

	if err := a.validator.Validate(ctx,
		auth.ValidateApplicationAccess(in.Integration.ApplicationId, auth.Update),
	); err != nil {
		return nil, grpc.Errorf(codes.Unauthenticated, "authentication failed: %s", err)
	}

	conf := influxdb.Config{
		Endpoint:            in.Integration.Endpoint,
		DB:                  in.Integration.Db,
		Username:            in.Integration.Username,
		Password:            in.Integration.Password,
		RetentionPolicyName: in.Integration.RetentionPolicyName,
		Precision:           strings.ToLower(in.Integration.Precision.String()),
	}
	if err := conf.Validate(); err != nil {
		return nil, errToRPCError(err)
	}

	confJSON, err := json.Marshal(conf)
	if err != nil {
		return nil, errToRPCError(err)
	}

	integration := storage.Integration{
		ApplicationID: in.Integration.ApplicationId,
		Kind:          integration.InfluxDB,
		Settings:      confJSON,
	}
	if err := storage.CreateIntegration(config.C.PostgreSQL.DB, &integration); err != nil {
		return nil, errToRPCError(err)
	}

	return &empty.Empty{}, nil
}

// GetInfluxDBIntegration returns the InfluxDB application-integration.
func (a *ApplicationAPI) GetInfluxDBIntegration(ctx context.Context, in *pb.GetInfluxDBIntegrationRequest) (*pb.GetInfluxDBIntegrationResponse, error) {
	if err := a.validator.Validate(ctx,
		auth.ValidateApplicationAccess(in.ApplicationId, auth.Update),
	); err != nil {
		return nil, grpc.Errorf(codes.Unauthenticated, "authentication failed: %s", err)
	}

	integration, err := storage.GetIntegrationByApplicationID(config.C.PostgreSQL.DB, in.ApplicationId, integration.InfluxDB)
	if err != nil {
		return nil, errToRPCError(err)
	}

	var conf influxdb.Config
	if err = json.Unmarshal(integration.Settings, &conf); err != nil {
		return nil, errToRPCError(err)
	}

	prec, _ := pb.InfluxDBPrecision_value[strings.ToUpper(conf.Precision)]

	return &pb.GetInfluxDBIntegrationResponse{
		Integration: &pb.InfluxDBIntegration{
			ApplicationId:       in.ApplicationId,
			Endpoint:            conf.Endpoint,
			Db:                  conf.DB,
			Username:            conf.Username,
			Password:            conf.Password,
			RetentionPolicyName: conf.RetentionPolicyName,
			Precision:           pb.InfluxDBPrecision(prec),
		},
	}, nil
}

// UpdateInfluxDBIntegration updates the InfluxDB application-integration.
func (a *ApplicationAPI) UpdateInfluxDBIntegration(ctx context.Context, in *pb.UpdateInfluxDBIntegrationRequest) (*empty.Empty, error) {
	if in.Integration == nil {
		return nil, grpc.Errorf(codes.InvalidArgument, "integration must not be nil")
	}

	if err := a.validator.Validate(ctx,
		auth.ValidateApplicationAccess(in.Integration.ApplicationId, auth.Update),
	); err != nil {
		return nil, grpc.Errorf(codes.Unauthenticated, "authentication failed: %s", err)
	}

	integration, err := storage.GetIntegrationByApplicationID(config.C.PostgreSQL.DB, in.Integration.ApplicationId, integration.InfluxDB)
	if err != nil {
		return nil, errToRPCError(err)
	}

	conf := influxdb.Config{
		Endpoint:            in.Integration.Endpoint,
		DB:                  in.Integration.Db,
		Username:            in.Integration.Username,
		Password:            in.Integration.Password,
		RetentionPolicyName: in.Integration.RetentionPolicyName,
		Precision:           strings.ToLower(in.Integration.Precision.String()),
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

	return &empty.Empty{}, nil
}

// DeleteInfluxDBIntegration deletes the InfluxDB application-integration.
func (a *ApplicationAPI) DeleteInfluxDBIntegration(ctx context.Context, in *pb.DeleteInfluxDBIntegrationRequest) (*empty.Empty, error) {
	if err := a.validator.Validate(ctx,
		auth.ValidateApplicationAccess(in.ApplicationId, auth.Update),
	); err != nil {
		return nil, grpc.Errorf(codes.Unauthenticated, "authentication failed: %s", err)
	}

	integration, err := storage.GetIntegrationByApplicationID(config.C.PostgreSQL.DB, in.ApplicationId, integration.InfluxDB)
	if err != nil {
		return nil, errToRPCError(err)
	}

	if err = storage.DeleteIntegration(config.C.PostgreSQL.DB, integration.ID); err != nil {
		return nil, errToRPCError(err)
	}

	return &empty.Empty{}, nil
}

// ListIntegrations lists all configured integrations.
func (a *ApplicationAPI) ListIntegrations(ctx context.Context, in *pb.ListIntegrationRequest) (*pb.ListIntegrationResponse, error) {
	if err := a.validator.Validate(ctx,
		auth.ValidateApplicationAccess(in.ApplicationId, auth.Update),
	); err != nil {
		return nil, grpc.Errorf(codes.Unauthenticated, "authentication failed: %s", err)
	}

	integrations, err := storage.GetIntegrationsForApplicationID(config.C.PostgreSQL.DB, in.ApplicationId)
	if err != nil {
		return nil, errToRPCError(err)
	}

	out := pb.ListIntegrationResponse{
		TotalCount: int64(len(integrations)),
	}

	for _, intgr := range integrations {
		switch intgr.Kind {
		case integration.HTTP:
			out.Result = append(out.Result, &pb.IntegrationListItem{Kind: pb.IntegrationKind_HTTP})
		case integration.InfluxDB:
			out.Result = append(out.Result, &pb.IntegrationListItem{Kind: pb.IntegrationKind_INFLUXDB})
		default:
			return nil, grpc.Errorf(codes.Internal, "unknown integration kind: %s", intgr.Kind)
		}
	}

	return &out, nil
}
