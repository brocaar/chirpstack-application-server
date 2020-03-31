package external

import (
	"github.com/gofrs/uuid"
	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/empty"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

	pb "github.com/brocaar/chirpstack-api/go/v3/as/external/api"
	"github.com/brocaar/chirpstack-application-server/internal/api/external/auth"
	"github.com/brocaar/chirpstack-application-server/internal/api/helpers"
	"github.com/brocaar/chirpstack-application-server/internal/storage"
)

// InternalAPI exports the internal User related functions.
type InternalAPI struct {
	validator auth.Validator
}

// NewInternalAPI creates a new InternalAPI.
func NewInternalAPI(validator auth.Validator) *InternalAPI {
	return &InternalAPI{
		validator: validator,
	}
}

// Login validates the login request and returns a JWT token.
func (a *InternalAPI) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	jwt, err := storage.LoginUser(ctx, storage.DB(), req.Username, req.Password)
	if nil != err {
		return nil, helpers.ErrToRPCError(err)
	}

	return &pb.LoginResponse{Jwt: jwt}, nil
}

type claims struct {
	Username string `json:"username"`
}

// Profile returns the user profile.
func (a *InternalAPI) Profile(ctx context.Context, req *empty.Empty) (*pb.ProfileResponse, error) {
	if err := a.validator.Validate(ctx,
		auth.ValidateActiveUser()); err != nil {
		return nil, grpc.Errorf(codes.Unauthenticated, "authentication failed: %s", err)
	}

	username, err := a.validator.GetUsername(ctx)
	if nil != err {
		return nil, helpers.ErrToRPCError(err)
	}

	// Get the user id based on the username.
	user, err := storage.GetUserByUsername(ctx, storage.DB(), username)
	if nil != err {
		return nil, helpers.ErrToRPCError(err)
	}

	prof, err := storage.GetProfile(ctx, storage.DB(), user.ID)
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
			IsDeviceAdmin:    org.IsDeviceAdmin,
			IsGatewayAdmin:   org.IsGatewayAdmin,
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
func (a *InternalAPI) Branding(ctx context.Context, req *empty.Empty) (*pb.BrandingResponse, error) {
	resp := pb.BrandingResponse{
		Logo:         brandingHeader,
		Registration: brandingRegistration,
		Footer:       brandingFooter,
	}

	return &resp, nil
}

// GlobalSearch performs a global search.
func (a *InternalAPI) GlobalSearch(ctx context.Context, req *pb.GlobalSearchRequest) (*pb.GlobalSearchResponse, error) {
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

	results, err := storage.GlobalSearch(ctx, storage.DB(), username, isAdmin, req.Search, int(req.Limit), int(req.Offset))
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

// CreateAPIKey creates the given API key.
func (a *InternalAPI) CreateAPIKey(ctx context.Context, req *pb.CreateAPIKeyRequest) (*pb.CreateAPIKeyResponse, error) {
	apiKey := req.GetApiKey()

	if err := a.validator.Validate(ctx,
		auth.ValidateAPIKeysAccess(auth.Create, apiKey.GetOrganizationId(), apiKey.GetApplicationId())); err != nil {
		return nil, grpc.Errorf(codes.Unauthenticated, "authentication failed: %s", err)
	}

	if apiKey.GetIsAdmin() && (apiKey.GetOrganizationId() != 0 || apiKey.GetApplicationId() != 0) {
		return nil, grpc.Errorf(codes.InvalidArgument, "when is_admin is true, organization_id and application_id must be left blank")
	}

	var organizationID *int64
	var applicationID *int64

	if id := apiKey.GetOrganizationId(); id != 0 {
		organizationID = &id
	}

	if id := apiKey.GetApplicationId(); id != 0 {
		applicationID = &id
	}

	ak := storage.APIKey{
		Name:           apiKey.GetName(),
		IsAdmin:        apiKey.GetIsAdmin(),
		OrganizationID: organizationID,
		ApplicationID:  applicationID,
	}

	jwtToken, err := storage.CreateAPIKey(ctx, storage.DB(), &ak)
	if err != nil {
		return nil, helpers.ErrToRPCError(err)
	}

	return &pb.CreateAPIKeyResponse{
		Id:       ak.ID.String(),
		JwtToken: jwtToken,
	}, nil
}

// ListAPIKeys lists the API keys.
func (a *InternalAPI) ListAPIKeys(ctx context.Context, req *pb.ListAPIKeysRequest) (*pb.ListAPIKeysResponse, error) {
	if err := a.validator.Validate(ctx,
		auth.ValidateAPIKeysAccess(auth.List, req.GetOrganizationId(), req.GetApplicationId())); err != nil {
		return nil, grpc.Errorf(codes.Unauthenticated, "authentication failed: %s", err)
	}

	if req.GetIsAdmin() && (req.GetOrganizationId() != 0 || req.GetApplicationId() != 0) {
		return nil, grpc.Errorf(codes.InvalidArgument, "when is_admin is true, organization_id and application_id must be left blank")
	}

	filters := storage.APIKeyFilters{
		IsAdmin: req.GetIsAdmin(),
		Limit:   int(req.GetLimit()),
		Offset:  int(req.GetOffset()),
	}

	if id := req.GetOrganizationId(); id != 0 {
		filters.OrganizationID = &id
	}

	if id := req.GetApplicationId(); id != 0 {
		filters.ApplicationID = &id
	}

	count, err := storage.GetAPIKeyCount(ctx, storage.DB(), filters)
	if err != nil {
		return nil, helpers.ErrToRPCError(err)
	}

	apiKeys, err := storage.GetAPIKeys(ctx, storage.DB(), filters)
	if err != nil {
		return nil, helpers.ErrToRPCError(err)
	}

	resp := pb.ListAPIKeysResponse{
		TotalCount: int64(count),
	}

	for _, apiKey := range apiKeys {
		ak := pb.APIKey{
			Id:      apiKey.ID.String(),
			Name:    apiKey.Name,
			IsAdmin: apiKey.IsAdmin,
		}

		if apiKey.OrganizationID != nil {
			ak.OrganizationId = *apiKey.OrganizationID
		}

		if apiKey.ApplicationID != nil {
			ak.ApplicationId = *apiKey.ApplicationID
		}

		resp.Result = append(resp.Result, &ak)
	}

	return &resp, nil
}

// DeleteAPIKey deletes the given API key.
func (a *InternalAPI) DeleteAPIKey(ctx context.Context, req *pb.DeleteAPIKeyRequest) (*empty.Empty, error) {
	apiKeyID, err := uuid.FromString(req.Id)
	if err != nil {
		return nil, grpc.Errorf(codes.InvalidArgument, "api_key: %s", err)
	}

	if err := a.validator.Validate(ctx,
		auth.ValidateAPIKeyAccess(auth.Delete, apiKeyID)); err != nil {
		return nil, grpc.Errorf(codes.Unauthenticated, "authentication failed: %s", err)
	}

	if err := storage.DeleteAPIKey(ctx, storage.DB(), apiKeyID); err != nil {
		return nil, helpers.ErrToRPCError(err)
	}

	return &empty.Empty{}, nil
}
