package external

import (
	"encoding/json"
	"strings"

	"github.com/gofrs/uuid"
	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/empty"
	log "github.com/sirupsen/logrus"

	"github.com/jmoiron/sqlx"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

	pb "github.com/brocaar/chirpstack-api/go/v3/as/external/api"
	"github.com/brocaar/chirpstack-application-server/internal/api/external/auth"
	"github.com/brocaar/chirpstack-application-server/internal/api/helpers"
	"github.com/brocaar/chirpstack-application-server/internal/codec"
	"github.com/brocaar/chirpstack-application-server/internal/config"
	"github.com/brocaar/chirpstack-application-server/internal/integration"
	"github.com/brocaar/chirpstack-application-server/internal/integration/http"
	"github.com/brocaar/chirpstack-application-server/internal/integration/influxdb"
	"github.com/brocaar/chirpstack-application-server/internal/integration/loracloud"
	"github.com/brocaar/chirpstack-application-server/internal/integration/mqtt"
	"github.com/brocaar/chirpstack-application-server/internal/integration/mydevices"
	"github.com/brocaar/chirpstack-application-server/internal/integration/pilotthings"
	"github.com/brocaar/chirpstack-application-server/internal/integration/thingsboard"
	"github.com/brocaar/chirpstack-application-server/internal/storage"
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
		return nil, helpers.ErrToRPCError(err)
	}

	sp, err := storage.GetServiceProfile(ctx, storage.DB(), spID, true)
	if err != nil {
		return nil, helpers.ErrToRPCError(err)
	}

	if sp.OrganizationID != req.Application.OrganizationId {
		return nil, grpc.Errorf(codes.InvalidArgument, "application and service-profile must be under the same organization")
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

	if err := storage.CreateApplication(ctx, storage.DB(), &app); err != nil {
		return nil, helpers.ErrToRPCError(err)
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

	app, err := storage.GetApplication(ctx, storage.DB(), req.Id)
	if err != nil {
		return nil, helpers.ErrToRPCError(err)
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

	app, err := storage.GetApplication(ctx, storage.DB(), req.Application.Id)
	if err != nil {
		return nil, helpers.ErrToRPCError(err)
	}

	spID, err := uuid.FromString(req.Application.ServiceProfileId)
	if err != nil {
		return nil, helpers.ErrToRPCError(err)
	}

	sp, err := storage.GetServiceProfile(ctx, storage.DB(), spID, true)
	if err != nil {
		return nil, helpers.ErrToRPCError(err)
	}

	if sp.OrganizationID != app.OrganizationID {
		return nil, grpc.Errorf(codes.InvalidArgument, "application and service-profile must be under the same organization")
	}

	// update the fields
	app.Name = req.Application.Name
	app.Description = req.Application.Description
	app.ServiceProfileID = spID
	app.PayloadCodec = codec.Type(req.Application.PayloadCodec)
	app.PayloadEncoderScript = req.Application.PayloadEncoderScript
	app.PayloadDecoderScript = req.Application.PayloadDecoderScript

	err = storage.UpdateApplication(ctx, storage.DB(), app)
	if err != nil {
		return nil, helpers.ErrToRPCError(err)
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

	err := storage.Transaction(func(tx sqlx.Ext) error {
		err := storage.DeleteApplication(ctx, tx, req.Id)
		if err != nil {
			return helpers.ErrToRPCError(err)
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

	filters := storage.ApplicationFilters{
		Search:         req.Search,
		Limit:          int(req.Limit),
		Offset:         int(req.Offset),
		OrganizationID: req.OrganizationId,
	}

	sub, err := a.validator.GetSubject(ctx)
	if err != nil {
		return nil, helpers.ErrToRPCError(err)
	}

	switch sub {
	case auth.SubjectUser:
		user, err := a.validator.GetUser(ctx)
		if err != nil {
			return nil, helpers.ErrToRPCError(err)
		}

		// Filter on user ID when OrganizationID is not set and the user is
		// not a global admin.
		if !user.IsAdmin && filters.OrganizationID == 0 {
			filters.UserID = user.ID
		}

	case auth.SubjectAPIKey:
		// Nothing to do as the validator function already validated that the
		// API Key has access to the given OrganizationID.
	default:
		return nil, grpc.Errorf(codes.Unauthenticated, "invalid token subject: %s", sub)
	}

	count, err := storage.GetApplicationCount(ctx, storage.DB(), filters)
	if err != nil {
		return nil, helpers.ErrToRPCError(err)
	}

	apps, err := storage.GetApplications(ctx, storage.DB(), filters)
	if err != nil {
		return nil, helpers.ErrToRPCError(err)
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
		Headers:          headers,
		EventEndpointURL: in.Integration.EventEndpointUrl,
		Marshaler:        in.Integration.Marshaler.String(),

		// Backwards compatibility.
		DataUpURL:                  in.Integration.UplinkDataUrl,
		JoinNotificationURL:        in.Integration.JoinNotificationUrl,
		ACKNotificationURL:         in.Integration.AckNotificationUrl,
		ErrorNotificationURL:       in.Integration.ErrorNotificationUrl,
		StatusNotificationURL:      in.Integration.StatusNotificationUrl,
		LocationNotificationURL:    in.Integration.LocationNotificationUrl,
		TxAckNotificationURL:       in.Integration.TxAckNotificationUrl,
		IntegrationNotificationURL: in.Integration.IntegrationNotificationUrl,
	}
	if err := conf.Validate(); err != nil {
		return nil, helpers.ErrToRPCError(err)
	}

	confJSON, err := json.Marshal(conf)
	if err != nil {
		return nil, helpers.ErrToRPCError(err)
	}

	integration := storage.Integration{
		ApplicationID: in.Integration.ApplicationId,
		Kind:          integration.HTTP,
		Settings:      confJSON,
	}
	if err = storage.CreateIntegration(ctx, storage.DB(), &integration); err != nil {
		return nil, helpers.ErrToRPCError(err)
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

	integration, err := storage.GetIntegrationByApplicationID(ctx, storage.DB(), in.ApplicationId, integration.HTTP)
	if err != nil {
		return nil, helpers.ErrToRPCError(err)
	}

	var conf http.Config
	if err = json.Unmarshal(integration.Settings, &conf); err != nil {
		return nil, helpers.ErrToRPCError(err)
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
			ApplicationId:    integration.ApplicationID,
			Headers:          headers,
			EventEndpointUrl: conf.EventEndpointURL,
			Marshaler:        pb.Marshaler(pb.Marshaler_value[conf.Marshaler]),

			// Backwards compatibility.
			UplinkDataUrl:              conf.DataUpURL,
			JoinNotificationUrl:        conf.JoinNotificationURL,
			AckNotificationUrl:         conf.ACKNotificationURL,
			ErrorNotificationUrl:       conf.ErrorNotificationURL,
			StatusNotificationUrl:      conf.StatusNotificationURL,
			LocationNotificationUrl:    conf.LocationNotificationURL,
			TxAckNotificationUrl:       conf.TxAckNotificationURL,
			IntegrationNotificationUrl: conf.IntegrationNotificationURL,
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

	integration, err := storage.GetIntegrationByApplicationID(ctx, storage.DB(), in.Integration.ApplicationId, integration.HTTP)
	if err != nil {
		return nil, helpers.ErrToRPCError(err)
	}

	headers := make(map[string]string)
	for _, h := range in.Integration.Headers {
		headers[h.Key] = h.Value
	}

	conf := http.Config{
		Headers:          headers,
		EventEndpointURL: in.Integration.EventEndpointUrl,
		Marshaler:        in.Integration.Marshaler.String(),

		// Backwards compatibility.
		DataUpURL:                  in.Integration.UplinkDataUrl,
		JoinNotificationURL:        in.Integration.JoinNotificationUrl,
		ACKNotificationURL:         in.Integration.AckNotificationUrl,
		ErrorNotificationURL:       in.Integration.ErrorNotificationUrl,
		StatusNotificationURL:      in.Integration.StatusNotificationUrl,
		LocationNotificationURL:    in.Integration.LocationNotificationUrl,
		TxAckNotificationURL:       in.Integration.TxAckNotificationUrl,
		IntegrationNotificationURL: in.Integration.IntegrationNotificationUrl,
	}
	if err := conf.Validate(); err != nil {
		return nil, helpers.ErrToRPCError(err)
	}

	confJSON, err := json.Marshal(conf)
	if err != nil {
		return nil, helpers.ErrToRPCError(err)
	}
	integration.Settings = confJSON

	if err = storage.UpdateIntegration(ctx, storage.DB(), &integration); err != nil {
		return nil, helpers.ErrToRPCError(err)
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

	integration, err := storage.GetIntegrationByApplicationID(ctx, storage.DB(), in.ApplicationId, integration.HTTP)
	if err != nil {
		return nil, helpers.ErrToRPCError(err)
	}

	if err = storage.DeleteIntegration(ctx, storage.DB(), integration.ID); err != nil {
		return nil, helpers.ErrToRPCError(err)
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

	version := 1
	if in.Integration.Version == pb.InfluxDBVersion_INFLUXDB_2 {
		version = 2
	}

	conf := influxdb.Config{
		Endpoint:            in.Integration.Endpoint,
		Version:             version,
		DB:                  in.Integration.Db,
		Username:            in.Integration.Username,
		Password:            in.Integration.Password,
		RetentionPolicyName: in.Integration.RetentionPolicyName,
		Precision:           strings.ToLower(in.Integration.Precision.String()),
		Token:               in.Integration.Token,
		Organization:        in.Integration.Organization,
		Bucket:              in.Integration.Bucket,
	}
	if err := conf.Validate(); err != nil {
		return nil, helpers.ErrToRPCError(err)
	}

	confJSON, err := json.Marshal(conf)
	if err != nil {
		return nil, helpers.ErrToRPCError(err)
	}

	integration := storage.Integration{
		ApplicationID: in.Integration.ApplicationId,
		Kind:          integration.InfluxDB,
		Settings:      confJSON,
	}
	if err := storage.CreateIntegration(ctx, storage.DB(), &integration); err != nil {
		return nil, helpers.ErrToRPCError(err)
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

	integration, err := storage.GetIntegrationByApplicationID(ctx, storage.DB(), in.ApplicationId, integration.InfluxDB)
	if err != nil {
		return nil, helpers.ErrToRPCError(err)
	}

	var conf influxdb.Config
	if err = json.Unmarshal(integration.Settings, &conf); err != nil {
		return nil, helpers.ErrToRPCError(err)
	}

	prec, _ := pb.InfluxDBPrecision_value[strings.ToUpper(conf.Precision)]
	version := pb.InfluxDBVersion_INFLUXDB_1

	if conf.Version == 2 {
		version = pb.InfluxDBVersion_INFLUXDB_2
	}

	return &pb.GetInfluxDBIntegrationResponse{
		Integration: &pb.InfluxDBIntegration{
			ApplicationId:       in.ApplicationId,
			Endpoint:            conf.Endpoint,
			Version:             version,
			Db:                  conf.DB,
			Username:            conf.Username,
			Password:            conf.Password,
			RetentionPolicyName: conf.RetentionPolicyName,
			Precision:           pb.InfluxDBPrecision(prec),
			Token:               conf.Token,
			Organization:        conf.Organization,
			Bucket:              conf.Bucket,
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

	integration, err := storage.GetIntegrationByApplicationID(ctx, storage.DB(), in.Integration.ApplicationId, integration.InfluxDB)
	if err != nil {
		return nil, helpers.ErrToRPCError(err)
	}

	version := 1
	if in.Integration.Version == pb.InfluxDBVersion_INFLUXDB_2 {
		version = 2
	}

	conf := influxdb.Config{
		Endpoint:            in.Integration.Endpoint,
		Version:             version,
		DB:                  in.Integration.Db,
		Username:            in.Integration.Username,
		Password:            in.Integration.Password,
		RetentionPolicyName: in.Integration.RetentionPolicyName,
		Precision:           strings.ToLower(in.Integration.Precision.String()),
		Token:               in.Integration.Token,
		Organization:        in.Integration.Organization,
		Bucket:              in.Integration.Bucket,
	}
	if err := conf.Validate(); err != nil {
		return nil, helpers.ErrToRPCError(err)
	}

	confJSON, err := json.Marshal(conf)
	if err != nil {
		return nil, helpers.ErrToRPCError(err)
	}

	integration.Settings = confJSON
	if err = storage.UpdateIntegration(ctx, storage.DB(), &integration); err != nil {
		return nil, helpers.ErrToRPCError(err)
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

	integration, err := storage.GetIntegrationByApplicationID(ctx, storage.DB(), in.ApplicationId, integration.InfluxDB)
	if err != nil {
		return nil, helpers.ErrToRPCError(err)
	}

	if err = storage.DeleteIntegration(ctx, storage.DB(), integration.ID); err != nil {
		return nil, helpers.ErrToRPCError(err)
	}

	return &empty.Empty{}, nil
}

// CreateThingsBoardIntegration creates a ThingsBoard application-integration.
func (a *ApplicationAPI) CreateThingsBoardIntegration(ctx context.Context, in *pb.CreateThingsBoardIntegrationRequest) (*empty.Empty, error) {
	if in.Integration == nil {
		return nil, grpc.Errorf(codes.InvalidArgument, "integration must not be nil")
	}

	if err := a.validator.Validate(ctx,
		auth.ValidateApplicationAccess(in.Integration.ApplicationId, auth.Update),
	); err != nil {
		return nil, grpc.Errorf(codes.Unauthenticated, "authentication failed: %s", err)
	}

	conf := thingsboard.Config{
		Server: in.Integration.Server,
	}
	if err := conf.Validate(); err != nil {
		return nil, helpers.ErrToRPCError(err)
	}

	confJSON, err := json.Marshal(conf)
	if err != nil {
		return nil, helpers.ErrToRPCError(err)
	}

	integration := storage.Integration{
		ApplicationID: in.Integration.ApplicationId,
		Kind:          integration.ThingsBoard,
		Settings:      confJSON,
	}
	if err := storage.CreateIntegration(ctx, storage.DB(), &integration); err != nil {
		return nil, helpers.ErrToRPCError(err)
	}

	return &empty.Empty{}, nil
}

// GetThingsBoardIntegration returns the ThingsBoard application-integration.
func (a *ApplicationAPI) GetThingsBoardIntegration(ctx context.Context, in *pb.GetThingsBoardIntegrationRequest) (*pb.GetThingsBoardIntegrationResponse, error) {
	if err := a.validator.Validate(ctx,
		auth.ValidateApplicationAccess(in.ApplicationId, auth.Update),
	); err != nil {
		return nil, grpc.Errorf(codes.Unauthenticated, "authentication failed: %s", err)
	}

	integration, err := storage.GetIntegrationByApplicationID(ctx, storage.DB(), in.ApplicationId, integration.ThingsBoard)
	if err != nil {
		return nil, helpers.ErrToRPCError(err)
	}

	var conf thingsboard.Config
	if err = json.Unmarshal(integration.Settings, &conf); err != nil {
		return nil, helpers.ErrToRPCError(err)
	}

	return &pb.GetThingsBoardIntegrationResponse{
		Integration: &pb.ThingsBoardIntegration{
			ApplicationId: in.ApplicationId,
			Server:        conf.Server,
		},
	}, nil
}

// UpdateThingsBoardIntegration updates the ThingsBoard application-integration.
func (a *ApplicationAPI) UpdateThingsBoardIntegration(ctx context.Context, in *pb.UpdateThingsBoardIntegrationRequest) (*empty.Empty, error) {
	if in.Integration == nil {
		return nil, grpc.Errorf(codes.InvalidArgument, "integration must not be nil")
	}

	if err := a.validator.Validate(ctx,
		auth.ValidateApplicationAccess(in.Integration.ApplicationId, auth.Update),
	); err != nil {
		return nil, grpc.Errorf(codes.Unauthenticated, "authentication failed: %s", err)
	}

	integration, err := storage.GetIntegrationByApplicationID(ctx, storage.DB(), in.Integration.ApplicationId, integration.ThingsBoard)
	if err != nil {
		return nil, helpers.ErrToRPCError(err)
	}

	conf := thingsboard.Config{
		Server: in.Integration.Server,
	}
	if err := conf.Validate(); err != nil {
		return nil, helpers.ErrToRPCError(err)
	}

	confJSON, err := json.Marshal(conf)
	if err != nil {
		return nil, helpers.ErrToRPCError(err)
	}

	integration.Settings = confJSON
	if err = storage.UpdateIntegration(ctx, storage.DB(), &integration); err != nil {
		return nil, helpers.ErrToRPCError(err)
	}

	return &empty.Empty{}, nil
}

// DeleteThingsBoardIntegration deletes the ThingsBoard application-integration.
func (a *ApplicationAPI) DeleteThingsBoardIntegration(ctx context.Context, in *pb.DeleteThingsBoardIntegrationRequest) (*empty.Empty, error) {
	if err := a.validator.Validate(ctx,
		auth.ValidateApplicationAccess(in.ApplicationId, auth.Update),
	); err != nil {
		return nil, grpc.Errorf(codes.Unauthenticated, "authentication failed: %s", err)
	}

	integration, err := storage.GetIntegrationByApplicationID(ctx, storage.DB(), in.ApplicationId, integration.ThingsBoard)
	if err != nil {
		return nil, helpers.ErrToRPCError(err)
	}

	if err = storage.DeleteIntegration(ctx, storage.DB(), integration.ID); err != nil {
		return nil, helpers.ErrToRPCError(err)
	}

	return &empty.Empty{}, nil
}

// CreateMyDevicesIntegration creates a MyDevices application-integration.
func (a *ApplicationAPI) CreateMyDevicesIntegration(ctx context.Context, in *pb.CreateMyDevicesIntegrationRequest) (*empty.Empty, error) {
	if in.Integration == nil {
		return nil, grpc.Errorf(codes.InvalidArgument, "integration must not be nil")
	}

	if err := a.validator.Validate(ctx,
		auth.ValidateApplicationAccess(in.Integration.ApplicationId, auth.Update),
	); err != nil {
		return nil, grpc.Errorf(codes.Unauthenticated, "authentication failed: %s", err)
	}

	config := mydevices.Config{
		Endpoint: in.Integration.Endpoint,
	}
	confJSON, err := json.Marshal(config)
	if err != nil {
		return nil, helpers.ErrToRPCError(err)
	}

	integration := storage.Integration{
		ApplicationID: in.Integration.ApplicationId,
		Kind:          integration.MyDevices,
		Settings:      confJSON,
	}
	if err := storage.CreateIntegration(ctx, storage.DB(), &integration); err != nil {
		return nil, helpers.ErrToRPCError(err)
	}

	return &empty.Empty{}, nil
}

// GetMyDevicesIntegration returns the MyDevices application-integration.
func (a *ApplicationAPI) GetMyDevicesIntegration(ctx context.Context, in *pb.GetMyDevicesIntegrationRequest) (*pb.GetMyDevicesIntegrationResponse, error) {
	if err := a.validator.Validate(ctx,
		auth.ValidateApplicationAccess(in.ApplicationId, auth.Update),
	); err != nil {
		return nil, grpc.Errorf(codes.Unauthenticated, "authentication failed: %s", err)
	}

	integration, err := storage.GetIntegrationByApplicationID(ctx, storage.DB(), in.ApplicationId, integration.MyDevices)
	if err != nil {
		return nil, helpers.ErrToRPCError(err)
	}

	var conf mydevices.Config
	if err := json.Unmarshal(integration.Settings, &conf); err != nil {
		return nil, helpers.ErrToRPCError(err)
	}

	return &pb.GetMyDevicesIntegrationResponse{
		Integration: &pb.MyDevicesIntegration{
			ApplicationId: in.ApplicationId,
			Endpoint:      conf.Endpoint,
		},
	}, nil
}

// UpdateMyDevicesIntegration updates the MyDevices application-integration.
func (a *ApplicationAPI) UpdateMyDevicesIntegration(ctx context.Context, in *pb.UpdateMyDevicesIntegrationRequest) (*empty.Empty, error) {
	if in.Integration == nil {
		return nil, grpc.Errorf(codes.InvalidArgument, "integration must not be nil")
	}

	if err := a.validator.Validate(ctx,
		auth.ValidateApplicationAccess(in.Integration.ApplicationId, auth.Update),
	); err != nil {
		return nil, grpc.Errorf(codes.Unauthenticated, "authentication failed: %s", err)
	}

	integration, err := storage.GetIntegrationByApplicationID(ctx, storage.DB(), in.Integration.ApplicationId, integration.MyDevices)
	if err != nil {
		return nil, helpers.ErrToRPCError(err)
	}

	conf := mydevices.Config{
		Endpoint: in.Integration.Endpoint,
	}

	confJSON, err := json.Marshal(conf)
	if err != nil {
		return nil, helpers.ErrToRPCError(err)
	}

	integration.Settings = confJSON
	if err = storage.UpdateIntegration(ctx, storage.DB(), &integration); err != nil {
		return nil, helpers.ErrToRPCError(err)
	}

	return &empty.Empty{}, nil
}

// DeleteMyDevicesIntegration deletes the MyDevices application-integration.
func (a *ApplicationAPI) DeleteMyDevicesIntegration(ctx context.Context, in *pb.DeleteMyDevicesIntegrationRequest) (*empty.Empty, error) {
	if err := a.validator.Validate(ctx,
		auth.ValidateApplicationAccess(in.ApplicationId, auth.Update),
	); err != nil {
		return nil, grpc.Errorf(codes.Unauthenticated, "authentication failed: %s", err)
	}

	integration, err := storage.GetIntegrationByApplicationID(ctx, storage.DB(), in.ApplicationId, integration.MyDevices)
	if err != nil {
		return nil, helpers.ErrToRPCError(err)
	}

	if err = storage.DeleteIntegration(ctx, storage.DB(), integration.ID); err != nil {
		return nil, helpers.ErrToRPCError(err)
	}

	return &empty.Empty{}, nil
}

// CreateLoRaCloudIntegration creates a LoRaCloud application-integration.
func (a *ApplicationAPI) CreateLoRaCloudIntegration(ctx context.Context, in *pb.CreateLoRaCloudIntegrationRequest) (*empty.Empty, error) {
	if in.Integration == nil {
		return nil, grpc.Errorf(codes.InvalidArgument, "integration must not be nil")
	}

	if err := a.validator.Validate(ctx,
		auth.ValidateApplicationAccess(in.GetIntegration().ApplicationId, auth.Update),
	); err != nil {
		return nil, grpc.Errorf(codes.Unauthenticated, "authentication failed: %s", err)
	}

	config := loracloud.Config{
		Geolocation:                  in.GetIntegration().Geolocation,
		GeolocationToken:             in.GetIntegration().GeolocationToken,
		GeolocationBufferTTL:         int(in.GetIntegration().GeolocationBufferTtl),
		GeolocationMinBufferSize:     int(in.GetIntegration().GeolocationMinBufferSize),
		GeolocationTDOA:              in.GetIntegration().GeolocationTdoa,
		GeolocationRSSI:              in.GetIntegration().GeolocationRssi,
		GeolocationGNSS:              in.GetIntegration().GeolocationGnss,
		GeolocationGNSSPayloadField:  in.GetIntegration().GeolocationGnssPayloadField,
		GeolocationGNSSUseRxTime:     in.GetIntegration().GeolocationGnssUseRxTime,
		GeolocationWifi:              in.GetIntegration().GeolocationWifi,
		GeolocationWifiPayloadField:  in.GetIntegration().GeolocationWifiPayloadField,
		DAS:                          in.GetIntegration().Das,
		DASToken:                     in.GetIntegration().DasToken,
		DASModemPort:                 uint8(in.GetIntegration().DasModemPort),
		DASGNSSPort:                  uint8(in.GetIntegration().DasGnssPort),
		DASGNSSUseRxTime:             in.GetIntegration().DasGnssUseRxTime,
		DASStreamingGeolocWorkaround: in.GetIntegration().DasStreamingGeolocWorkaround,
	}
	confJSON, err := json.Marshal(config)
	if err != nil {
		return nil, helpers.ErrToRPCError(err)
	}

	integration := storage.Integration{
		ApplicationID: in.GetIntegration().ApplicationId,
		Kind:          integration.LoRaCloud,
		Settings:      confJSON,
	}
	if err := storage.CreateIntegration(ctx, storage.DB(), &integration); err != nil {
		return nil, helpers.ErrToRPCError(err)
	}

	return &empty.Empty{}, nil
}

// GetLoRaCloudIntegration returns the LoRaCloud application-integration.
func (a *ApplicationAPI) GetLoRaCloudIntegration(ctx context.Context, in *pb.GetLoRaCloudIntegrationRequest) (*pb.GetLoRaCloudIntegrationResponse, error) {
	if err := a.validator.Validate(ctx,
		auth.ValidateApplicationAccess(in.ApplicationId, auth.Update),
	); err != nil {
		return nil, grpc.Errorf(codes.Unauthenticated, "authentication failed: %s", err)
	}

	integration, err := storage.GetIntegrationByApplicationID(ctx, storage.DB(), in.ApplicationId, integration.LoRaCloud)
	if err != nil {
		return nil, helpers.ErrToRPCError(err)
	}

	var conf loracloud.Config
	if err := json.Unmarshal(integration.Settings, &conf); err != nil {
		return nil, helpers.ErrToRPCError(err)
	}

	return &pb.GetLoRaCloudIntegrationResponse{
		Integration: &pb.LoRaCloudIntegration{
			ApplicationId:                in.ApplicationId,
			Geolocation:                  conf.Geolocation,
			GeolocationToken:             conf.GeolocationToken,
			GeolocationBufferTtl:         uint32(conf.GeolocationBufferTTL),
			GeolocationMinBufferSize:     uint32(conf.GeolocationMinBufferSize),
			GeolocationTdoa:              conf.GeolocationTDOA,
			GeolocationRssi:              conf.GeolocationRSSI,
			GeolocationGnss:              conf.GeolocationGNSS,
			GeolocationGnssPayloadField:  conf.GeolocationGNSSPayloadField,
			GeolocationGnssUseRxTime:     conf.GeolocationGNSSUseRxTime,
			GeolocationWifi:              conf.GeolocationWifi,
			GeolocationWifiPayloadField:  conf.GeolocationWifiPayloadField,
			Das:                          conf.DAS,
			DasToken:                     conf.DASToken,
			DasModemPort:                 uint32(conf.DASModemPort),
			DasGnssPort:                  uint32(conf.DASGNSSPort),
			DasGnssUseRxTime:             conf.DASGNSSUseRxTime,
			DasStreamingGeolocWorkaround: conf.DASStreamingGeolocWorkaround,
		},
	}, nil
}

// UpdateLoRaCloudIntegration updates the LoRaCloud application-integration.
func (a *ApplicationAPI) UpdateLoRaCloudIntegration(ctx context.Context, in *pb.UpdateLoRaCloudIntegrationRequest) (*empty.Empty, error) {
	if in.Integration == nil {
		return nil, grpc.Errorf(codes.InvalidArgument, "integration must not be nil")
	}

	if err := a.validator.Validate(ctx,
		auth.ValidateApplicationAccess(in.Integration.ApplicationId, auth.Update),
	); err != nil {
		return nil, grpc.Errorf(codes.Unauthenticated, "authentication failed: %s", err)
	}

	integration, err := storage.GetIntegrationByApplicationID(ctx, storage.DB(), in.GetIntegration().ApplicationId, integration.LoRaCloud)
	if err != nil {
		return nil, helpers.ErrToRPCError(err)
	}

	var conf loracloud.Config
	if err := json.Unmarshal(integration.Settings, &conf); err != nil {
		return nil, helpers.ErrToRPCError(err)
	}

	conf = loracloud.Config{
		Geolocation:                  in.GetIntegration().Geolocation,
		GeolocationToken:             in.GetIntegration().GeolocationToken,
		GeolocationBufferTTL:         int(in.GetIntegration().GeolocationBufferTtl),
		GeolocationMinBufferSize:     int(in.GetIntegration().GeolocationMinBufferSize),
		GeolocationTDOA:              in.GetIntegration().GeolocationTdoa,
		GeolocationRSSI:              in.GetIntegration().GeolocationRssi,
		GeolocationGNSS:              in.GetIntegration().GeolocationGnss,
		GeolocationGNSSPayloadField:  in.GetIntegration().GeolocationGnssPayloadField,
		GeolocationGNSSUseRxTime:     in.GetIntegration().GeolocationGnssUseRxTime,
		GeolocationWifi:              in.GetIntegration().GeolocationWifi,
		GeolocationWifiPayloadField:  in.GetIntegration().GeolocationWifiPayloadField,
		DAS:                          in.GetIntegration().Das,
		DASToken:                     in.GetIntegration().DasToken,
		DASModemPort:                 uint8(in.GetIntegration().DasModemPort),
		DASGNSSPort:                  uint8(in.GetIntegration().DasGnssPort),
		DASGNSSUseRxTime:             in.GetIntegration().DasGnssUseRxTime,
		DASStreamingGeolocWorkaround: in.GetIntegration().DasStreamingGeolocWorkaround,
	}
	confJSON, err := json.Marshal(conf)
	if err != nil {
		return nil, helpers.ErrToRPCError(err)
	}

	integration.Settings = confJSON
	if err = storage.UpdateIntegration(ctx, storage.DB(), &integration); err != nil {
		return nil, helpers.ErrToRPCError(err)
	}

	return &empty.Empty{}, nil
}

// DeleteLoRaCloudIntegration deletes the LoRaCloud application-integration.
func (a *ApplicationAPI) DeleteLoRaCloudIntegration(ctx context.Context, in *pb.DeleteLoRaCloudIntegrationRequest) (*empty.Empty, error) {
	if err := a.validator.Validate(ctx,
		auth.ValidateApplicationAccess(in.ApplicationId, auth.Update),
	); err != nil {
		return nil, grpc.Errorf(codes.Unauthenticated, "authentication failed: %s", err)
	}

	integration, err := storage.GetIntegrationByApplicationID(ctx, storage.DB(), in.ApplicationId, integration.LoRaCloud)
	if err != nil {
		return nil, helpers.ErrToRPCError(err)
	}

	if err = storage.DeleteIntegration(ctx, storage.DB(), integration.ID); err != nil {
		return nil, helpers.ErrToRPCError(err)
	}

	return &empty.Empty{}, nil
}

// CreateGCPPubSubIntegration creates a GCP PubSub application-integration.
func (a *ApplicationAPI) CreateGCPPubSubIntegration(ctx context.Context, in *pb.CreateGCPPubSubIntegrationRequest) (*empty.Empty, error) {
	if in.Integration == nil {
		return nil, grpc.Errorf(codes.InvalidArgument, "integration must not be nil")
	}

	if err := a.validator.Validate(ctx,
		auth.ValidateApplicationAccess(in.GetIntegration().ApplicationId, auth.Update),
	); err != nil {
		return nil, grpc.Errorf(codes.Unauthenticated, "authentication failed: %s", err)
	}

	conf := config.IntegrationGCPConfig{
		Marshaler:            in.GetIntegration().Marshaler.String(),
		CredentialsFileBytes: []byte(in.GetIntegration().CredentialsFile),
		TopicName:            in.GetIntegration().TopicName,
		ProjectID:            in.GetIntegration().ProjectId,
	}
	confJSON, err := json.Marshal(conf)
	if err != nil {
		return nil, helpers.ErrToRPCError(err)
	}

	integration := storage.Integration{
		ApplicationID: in.GetIntegration().ApplicationId,
		Kind:          integration.GCPPubSub,
		Settings:      confJSON,
	}
	if err := storage.CreateIntegration(ctx, storage.DB(), &integration); err != nil {
		return nil, helpers.ErrToRPCError(err)
	}

	return &empty.Empty{}, nil
}

// GetGCPPubSubIntegration returns the GCP PubSub application-integration.
func (a *ApplicationAPI) GetGCPPubSubIntegration(ctx context.Context, in *pb.GetGCPPubSubIntegrationRequest) (*pb.GetGCPPubSubIntegrationResponse, error) {
	if err := a.validator.Validate(ctx,
		auth.ValidateApplicationAccess(in.ApplicationId, auth.Update),
	); err != nil {
		return nil, grpc.Errorf(codes.Unauthenticated, "authentication failed: %s", err)
	}

	integration, err := storage.GetIntegrationByApplicationID(ctx, storage.DB(), in.ApplicationId, integration.GCPPubSub)
	if err != nil {
		return nil, helpers.ErrToRPCError(err)
	}

	var conf config.IntegrationGCPConfig
	if err = json.Unmarshal(integration.Settings, &conf); err != nil {
		return nil, helpers.ErrToRPCError(err)
	}

	return &pb.GetGCPPubSubIntegrationResponse{
		Integration: &pb.GCPPubSubIntegration{
			ApplicationId:   integration.ApplicationID,
			Marshaler:       pb.Marshaler(pb.Marshaler_value[conf.Marshaler]),
			CredentialsFile: string(conf.CredentialsFileBytes),
			TopicName:       conf.TopicName,
			ProjectId:       conf.ProjectID,
		},
	}, nil
}

// UpdateGCPPubSubIntegration updates the GCP PubSub application-integration.
func (a *ApplicationAPI) UpdateGCPPubSubIntegration(ctx context.Context, in *pb.UpdateGCPPubSubIntegrationRequest) (*empty.Empty, error) {
	if in.Integration == nil {
		return nil, grpc.Errorf(codes.InvalidArgument, "integration must not be nil")
	}

	if err := a.validator.Validate(ctx,
		auth.ValidateApplicationAccess(in.Integration.ApplicationId, auth.Update),
	); err != nil {
		return nil, grpc.Errorf(codes.Unauthenticated, "authentication failed: %s", err)
	}

	integration, err := storage.GetIntegrationByApplicationID(ctx, storage.DB(), in.GetIntegration().ApplicationId, integration.GCPPubSub)
	if err != nil {
		return nil, helpers.ErrToRPCError(err)
	}

	conf := config.IntegrationGCPConfig{
		Marshaler:            in.GetIntegration().Marshaler.String(),
		CredentialsFileBytes: []byte(in.GetIntegration().CredentialsFile),
		TopicName:            in.GetIntegration().TopicName,
		ProjectID:            in.GetIntegration().ProjectId,
	}
	confJSON, err := json.Marshal(conf)
	if err != nil {
		return nil, helpers.ErrToRPCError(err)
	}

	integration.Settings = confJSON

	if err = storage.UpdateIntegration(ctx, storage.DB(), &integration); err != nil {
		return nil, helpers.ErrToRPCError(err)
	}

	return &empty.Empty{}, nil
}

// DeleteGCPPubSubIntegration deletes the GCP PubSub application-integration.
func (a *ApplicationAPI) DeleteGCPPubSubIntegration(ctx context.Context, in *pb.DeleteGCPPubSubIntegrationRequest) (*empty.Empty, error) {
	if err := a.validator.Validate(ctx,
		auth.ValidateApplicationAccess(in.ApplicationId, auth.Update),
	); err != nil {
		return nil, grpc.Errorf(codes.Unauthenticated, "authentication failed: %s", err)
	}

	integration, err := storage.GetIntegrationByApplicationID(ctx, storage.DB(), in.ApplicationId, integration.GCPPubSub)
	if err != nil {
		return nil, helpers.ErrToRPCError(err)
	}

	if err = storage.DeleteIntegration(ctx, storage.DB(), integration.ID); err != nil {
		return nil, helpers.ErrToRPCError(err)
	}

	return &empty.Empty{}, nil
}

// CreateAWSSNSIntegration creates a AWS SNS application-integration.
func (a *ApplicationAPI) CreateAWSSNSIntegration(ctx context.Context, in *pb.CreateAWSSNSIntegrationRequest) (*empty.Empty, error) {
	if in.Integration == nil {
		return nil, grpc.Errorf(codes.InvalidArgument, "integration must not be nil")
	}

	if err := a.validator.Validate(ctx,
		auth.ValidateApplicationAccess(in.GetIntegration().ApplicationId, auth.Update),
	); err != nil {
		return nil, grpc.Errorf(codes.Unauthenticated, "authentication failed: %s", err)
	}

	conf := config.IntegrationAWSSNSConfig{
		Marshaler:          in.GetIntegration().Marshaler.String(),
		AWSRegion:          in.GetIntegration().Region,
		AWSAccessKeyID:     in.GetIntegration().AccessKeyId,
		AWSSecretAccessKey: in.GetIntegration().SecretAccessKey,
		TopicARN:           in.GetIntegration().TopicArn,
	}
	confJSON, err := json.Marshal(conf)
	if err != nil {
		return nil, helpers.ErrToRPCError(err)
	}

	integration := storage.Integration{
		ApplicationID: in.GetIntegration().ApplicationId,
		Kind:          integration.AWSSNS,
		Settings:      confJSON,
	}
	if err = storage.CreateIntegration(ctx, storage.DB(), &integration); err != nil {
		return nil, helpers.ErrToRPCError(err)
	}

	return &empty.Empty{}, nil
}

// GetAWSSNSIntegration returns the AWS SNS application-integration.
func (a *ApplicationAPI) GetAWSSNSIntegration(ctx context.Context, in *pb.GetAWSSNSIntegrationRequest) (*pb.GetAWSSNSIntegrationResponse, error) {
	if err := a.validator.Validate(ctx,
		auth.ValidateApplicationAccess(in.ApplicationId, auth.Update),
	); err != nil {
		return nil, grpc.Errorf(codes.Unauthenticated, "authentication failed: %s", err)
	}

	integration, err := storage.GetIntegrationByApplicationID(ctx, storage.DB(), in.ApplicationId, integration.AWSSNS)
	if err != nil {
		return nil, helpers.ErrToRPCError(err)
	}

	var conf config.IntegrationAWSSNSConfig
	if err = json.Unmarshal(integration.Settings, &conf); err != nil {
		return nil, helpers.ErrToRPCError(err)
	}

	return &pb.GetAWSSNSIntegrationResponse{
		Integration: &pb.AWSSNSIntegration{
			ApplicationId:   integration.ApplicationID,
			Marshaler:       pb.Marshaler(pb.Marshaler_value[conf.Marshaler]),
			Region:          conf.AWSRegion,
			AccessKeyId:     conf.AWSAccessKeyID,
			SecretAccessKey: conf.AWSSecretAccessKey,
			TopicArn:        conf.TopicARN,
		},
	}, nil
}

// UpdateAWSSNSIntegration updates the AWS SNS application-integration.
func (a *ApplicationAPI) UpdateAWSSNSIntegration(ctx context.Context, in *pb.UpdateAWSSNSIntegrationRequest) (*empty.Empty, error) {
	if in.Integration == nil {
		return nil, grpc.Errorf(codes.InvalidArgument, "integration must not be nil")
	}

	if err := a.validator.Validate(ctx,
		auth.ValidateApplicationAccess(in.GetIntegration().ApplicationId, auth.Update),
	); err != nil {
		return nil, grpc.Errorf(codes.Unauthenticated, "authentication failed: %s", err)
	}

	integration, err := storage.GetIntegrationByApplicationID(ctx, storage.DB(), in.GetIntegration().ApplicationId, integration.AWSSNS)
	if err != nil {
		return nil, helpers.ErrToRPCError(err)
	}

	conf := config.IntegrationAWSSNSConfig{
		Marshaler:          in.GetIntegration().Marshaler.String(),
		AWSRegion:          in.GetIntegration().Region,
		AWSAccessKeyID:     in.GetIntegration().AccessKeyId,
		AWSSecretAccessKey: in.GetIntegration().SecretAccessKey,
		TopicARN:           in.GetIntegration().TopicArn,
	}
	confJSON, err := json.Marshal(conf)
	if err != nil {
		return nil, helpers.ErrToRPCError(err)
	}
	integration.Settings = confJSON

	if err = storage.UpdateIntegration(ctx, storage.DB(), &integration); err != nil {
		return nil, helpers.ErrToRPCError(err)
	}

	return &empty.Empty{}, nil
}

// DeleteAWSSNSIntegration deletes the AWS SNS application-integration.
func (a *ApplicationAPI) DeleteAWSSNSIntegration(ctx context.Context, in *pb.DeleteAWSSNSIntegrationRequest) (*empty.Empty, error) {
	if err := a.validator.Validate(ctx,
		auth.ValidateApplicationAccess(in.ApplicationId, auth.Update),
	); err != nil {
		return nil, grpc.Errorf(codes.Unauthenticated, "authentication failed: %s", err)
	}

	integration, err := storage.GetIntegrationByApplicationID(ctx, storage.DB(), in.ApplicationId, integration.AWSSNS)
	if err != nil {
		return nil, helpers.ErrToRPCError(err)
	}

	if err = storage.DeleteIntegration(ctx, storage.DB(), integration.ID); err != nil {
		return nil, helpers.ErrToRPCError(err)
	}

	return &empty.Empty{}, nil
}

// CreateAzureServiceBusIntegration creates an Azure Service-Bus application-integration.
func (a *ApplicationAPI) CreateAzureServiceBusIntegration(ctx context.Context, in *pb.CreateAzureServiceBusIntegrationRequest) (*empty.Empty, error) {
	if in.Integration == nil {
		return nil, grpc.Errorf(codes.InvalidArgument, "integration must not be nil")
	}

	if err := a.validator.Validate(ctx,
		auth.ValidateApplicationAccess(in.GetIntegration().ApplicationId, auth.Update),
	); err != nil {
		return nil, grpc.Errorf(codes.Unauthenticated, "authentication failed: %s", err)
	}

	conf := config.IntegrationAzureConfig{
		Marshaler:        in.GetIntegration().Marshaler.String(),
		ConnectionString: in.GetIntegration().ConnectionString,
		PublishName:      in.GetIntegration().PublishName,
	}
	confJSON, err := json.Marshal(conf)
	if err != nil {
		return nil, helpers.ErrToRPCError(err)
	}

	integration := storage.Integration{
		ApplicationID: in.GetIntegration().ApplicationId,
		Kind:          integration.AzureServiceBus,
		Settings:      confJSON,
	}
	if err = storage.CreateIntegration(ctx, storage.DB(), &integration); err != nil {
		return nil, helpers.ErrToRPCError(err)
	}

	return &empty.Empty{}, nil
}

// GetAzureServiceBusIntegration returns the Azure Service-Bus application-integration.
func (a *ApplicationAPI) GetAzureServiceBusIntegration(ctx context.Context, in *pb.GetAzureServiceBusIntegrationRequest) (*pb.GetAzureServiceBusIntegrationResponse, error) {
	if err := a.validator.Validate(ctx,
		auth.ValidateApplicationAccess(in.ApplicationId, auth.Update),
	); err != nil {
		return nil, grpc.Errorf(codes.Unauthenticated, "authentication failed: %s", err)
	}

	integration, err := storage.GetIntegrationByApplicationID(ctx, storage.DB(), in.ApplicationId, integration.AzureServiceBus)
	if err != nil {
		return nil, helpers.ErrToRPCError(err)
	}

	var conf config.IntegrationAzureConfig
	if err = json.Unmarshal(integration.Settings, &conf); err != nil {
		return nil, helpers.ErrToRPCError(err)
	}

	return &pb.GetAzureServiceBusIntegrationResponse{
		Integration: &pb.AzureServiceBusIntegration{
			ApplicationId:    integration.ApplicationID,
			Marshaler:        pb.Marshaler(pb.Marshaler_value[conf.Marshaler]),
			ConnectionString: conf.ConnectionString,
			PublishName:      conf.PublishName,
		},
	}, nil
}

// UpdateAzureServiceBusIntegration updates the Azure Service-Bus application-integration.
func (a *ApplicationAPI) UpdateAzureServiceBusIntegration(ctx context.Context, in *pb.UpdateAzureServiceBusIntegrationRequest) (*empty.Empty, error) {
	if in.Integration == nil {
		return nil, grpc.Errorf(codes.InvalidArgument, "integration must not be nil")
	}

	if err := a.validator.Validate(ctx,
		auth.ValidateApplicationAccess(in.GetIntegration().ApplicationId, auth.Update),
	); err != nil {
		return nil, grpc.Errorf(codes.Unauthenticated, "authentication failed: %s", err)
	}

	integration, err := storage.GetIntegrationByApplicationID(ctx, storage.DB(), in.GetIntegration().ApplicationId, integration.AzureServiceBus)
	if err != nil {
		return nil, helpers.ErrToRPCError(err)
	}

	conf := config.IntegrationAzureConfig{
		Marshaler:        in.GetIntegration().Marshaler.String(),
		ConnectionString: in.GetIntegration().ConnectionString,
		PublishName:      in.GetIntegration().PublishName,
	}
	confJSON, err := json.Marshal(conf)
	if err != nil {
		return nil, helpers.ErrToRPCError(err)
	}
	integration.Settings = confJSON

	if err = storage.UpdateIntegration(ctx, storage.DB(), &integration); err != nil {
		return nil, helpers.ErrToRPCError(err)
	}

	return &empty.Empty{}, nil
}

// DeleteAzureServiceBusIntegration deletes the Azure Service-Bus application-integration.
func (a *ApplicationAPI) DeleteAzureServiceBusIntegration(ctx context.Context, in *pb.DeleteAzureServiceBusIntegrationRequest) (*empty.Empty, error) {
	if err := a.validator.Validate(ctx,
		auth.ValidateApplicationAccess(in.ApplicationId, auth.Update),
	); err != nil {
		return nil, grpc.Errorf(codes.Unauthenticated, "authentication failed: %s", err)
	}

	integration, err := storage.GetIntegrationByApplicationID(ctx, storage.DB(), in.ApplicationId, integration.AzureServiceBus)
	if err != nil {
		return nil, helpers.ErrToRPCError(err)
	}

	if err = storage.DeleteIntegration(ctx, storage.DB(), integration.ID); err != nil {
		return nil, helpers.ErrToRPCError(err)
	}

	return &empty.Empty{}, nil
}

// CreatePilotThingsIntegration creates an Pilot Things application-integration.
func (a *ApplicationAPI) CreatePilotThingsIntegration(ctx context.Context, in *pb.CreatePilotThingsIntegrationRequest) (*empty.Empty, error) {
	if in.Integration == nil {
		return nil, grpc.Errorf(codes.InvalidArgument, "integration must not be nil")
	}

	if err := a.validator.Validate(ctx,
		auth.ValidateApplicationAccess(in.GetIntegration().ApplicationId, auth.Update),
	); err != nil {
		return nil, grpc.Errorf(codes.Unauthenticated, "authentication failed: %s", err)
	}

	conf := pilotthings.Config{
		Server: in.GetIntegration().Server,
		Token:  in.GetIntegration().Token,
	}
	if err := conf.Validate(); err != nil {
		return nil, helpers.ErrToRPCError(err)
	}

	confJSON, err := json.Marshal(conf)
	if err != nil {
		return nil, helpers.ErrToRPCError(err)
	}

	integration := storage.Integration{
		ApplicationID: in.GetIntegration().ApplicationId,
		Kind:          integration.PilotThings,
		Settings:      confJSON,
	}
	if err = storage.CreateIntegration(ctx, storage.DB(), &integration); err != nil {
		return nil, helpers.ErrToRPCError(err)
	}

	return &empty.Empty{}, nil
}

// GetPilotThingsIntegration returns the Pilot Things application-integration.
func (a *ApplicationAPI) GetPilotThingsIntegration(ctx context.Context, in *pb.GetPilotThingsIntegrationRequest) (*pb.GetPilotThingsIntegrationResponse, error) {
	if err := a.validator.Validate(ctx,
		auth.ValidateApplicationAccess(in.ApplicationId, auth.Update),
	); err != nil {
		return nil, grpc.Errorf(codes.Unauthenticated, "authentication failed: %s", err)
	}

	integration, err := storage.GetIntegrationByApplicationID(ctx, storage.DB(), in.ApplicationId, integration.PilotThings)
	if err != nil {
		return nil, helpers.ErrToRPCError(err)
	}

	var conf pilotthings.Config
	if err = json.Unmarshal(integration.Settings, &conf); err != nil {
		return nil, helpers.ErrToRPCError(err)
	}

	return &pb.GetPilotThingsIntegrationResponse{
		Integration: &pb.PilotThingsIntegration{
			ApplicationId: integration.ApplicationID,
			Server:        conf.Server,
			Token:         conf.Token,
		},
	}, nil
}

// UpdatePilotThingsIntegration updates the Pilot Things application-integration.
func (a *ApplicationAPI) UpdatePilotThingsIntegration(ctx context.Context, in *pb.UpdatePilotThingsIntegrationRequest) (*empty.Empty, error) {
	if in.Integration == nil {
		return nil, grpc.Errorf(codes.InvalidArgument, "integration must not be nil")
	}

	if err := a.validator.Validate(ctx,
		auth.ValidateApplicationAccess(in.GetIntegration().ApplicationId, auth.Update),
	); err != nil {
		return nil, grpc.Errorf(codes.Unauthenticated, "authentication failed: %s", err)
	}

	integration, err := storage.GetIntegrationByApplicationID(ctx, storage.DB(), in.GetIntegration().ApplicationId, integration.PilotThings)
	if err != nil {
		return nil, helpers.ErrToRPCError(err)
	}

	conf := pilotthings.Config{
		Server: in.GetIntegration().Server,
		Token:  in.GetIntegration().Token,
	}
	if err := conf.Validate(); err != nil {
		return nil, helpers.ErrToRPCError(err)
	}

	confJSON, err := json.Marshal(conf)
	if err != nil {
		return nil, helpers.ErrToRPCError(err)
	}
	integration.Settings = confJSON

	if err = storage.UpdateIntegration(ctx, storage.DB(), &integration); err != nil {
		return nil, helpers.ErrToRPCError(err)
	}

	return &empty.Empty{}, nil
}

// DeletePilotThingsIntegration deletes an Pilot Things application-integration.
func (a *ApplicationAPI) DeletePilotThingsIntegration(ctx context.Context, in *pb.DeletePilotThingsIntegrationRequest) (*empty.Empty, error) {
	if err := a.validator.Validate(ctx,
		auth.ValidateApplicationAccess(in.ApplicationId, auth.Update),
	); err != nil {
		return nil, grpc.Errorf(codes.Unauthenticated, "authentication failed: %s", err)
	}

	integration, err := storage.GetIntegrationByApplicationID(ctx, storage.DB(), in.ApplicationId, integration.PilotThings)
	if err != nil {
		return nil, helpers.ErrToRPCError(err)
	}

	if err = storage.DeleteIntegration(ctx, storage.DB(), integration.ID); err != nil {
		return nil, helpers.ErrToRPCError(err)
	}

	return &empty.Empty{}, nil
}

// GenerateMQTTIntegrationClientCertificate generates an application ID specific TLS certificate
// to connect to the MQTT broker.
func (a *ApplicationAPI) GenerateMQTTIntegrationClientCertificate(ctx context.Context, in *pb.GenerateMQTTIntegrationClientCertificateRequest) (*pb.GenerateMQTTIntegrationClientCertificateResponse, error) {
	if err := a.validator.Validate(ctx,
		auth.ValidateApplicationAccess(in.ApplicationId, auth.Update),
	); err != nil {
		return nil, grpc.Errorf(codes.Unauthenticated, "authentication failed: %s", err)
	}

	expiresAt, ca, cert, key, err := mqtt.GenerateClientCertificate(in.ApplicationId)
	if err != nil {
		return nil, helpers.ErrToRPCError(err)
	}

	app, err := storage.GetApplication(ctx, storage.DB(), in.ApplicationId)
	if err != nil {
		return nil, helpers.ErrToRPCError(err)
	}

	app.MQTTTLSCert = cert

	if err := storage.UpdateApplication(ctx, storage.DB(), app); err != nil {
		return nil, helpers.ErrToRPCError(err)
	}

	expiresAtPB, err := ptypes.TimestampProto(expiresAt)
	if err != nil {
		return nil, helpers.ErrToRPCError(err)
	}

	return &pb.GenerateMQTTIntegrationClientCertificateResponse{
		TlsCert:   string(cert),
		TlsKey:    string(key),
		CaCert:    string(ca),
		ExpiresAt: expiresAtPB,
	}, nil
}

// ListIntegrations lists all configured integrations.
func (a *ApplicationAPI) ListIntegrations(ctx context.Context, in *pb.ListIntegrationRequest) (*pb.ListIntegrationResponse, error) {
	if err := a.validator.Validate(ctx,
		auth.ValidateApplicationAccess(in.ApplicationId, auth.Update),
	); err != nil {
		return nil, grpc.Errorf(codes.Unauthenticated, "authentication failed: %s", err)
	}

	// Application integrations
	integrations, err := storage.GetIntegrationsForApplicationID(ctx, storage.DB(), in.ApplicationId)
	if err != nil {
		return nil, helpers.ErrToRPCError(err)
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
		case integration.ThingsBoard:
			out.Result = append(out.Result, &pb.IntegrationListItem{Kind: pb.IntegrationKind_THINGSBOARD})
		case integration.MyDevices:
			out.Result = append(out.Result, &pb.IntegrationListItem{Kind: pb.IntegrationKind_MYDEVICES})
		case integration.LoRaCloud:
			out.Result = append(out.Result, &pb.IntegrationListItem{Kind: pb.IntegrationKind_LORACLOUD})
		case integration.GCPPubSub:
			out.Result = append(out.Result, &pb.IntegrationListItem{Kind: pb.IntegrationKind_GCP_PUBSUB})
		case integration.AWSSNS:
			out.Result = append(out.Result, &pb.IntegrationListItem{Kind: pb.IntegrationKind_AWS_SNS})
		case integration.AzureServiceBus:
			out.Result = append(out.Result, &pb.IntegrationListItem{Kind: pb.IntegrationKind_AZURE_SERVICE_BUS})
		case integration.PilotThings:
			out.Result = append(out.Result, &pb.IntegrationListItem{Kind: pb.IntegrationKind_PILOT_THINGS})
		default:
			log.WithFields(log.Fields{
				"kind": intgr.Kind,
			}).Warning("api/external: unknown integration kind")
		}
	}

	// Global integrations
	conf := config.Get()
	for _, name := range conf.ApplicationServer.Integration.Enabled {
		switch name {
		case "mqtt":
			out.Result = append(out.Result, &pb.IntegrationListItem{Kind: pb.IntegrationKind_MQTT_GLOBAL})
		}

	}

	return &out, nil
}
