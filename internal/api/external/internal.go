package external

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gofrs/uuid"
	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

	pb "github.com/brocaar/chirpstack-api/go/v3/as/external/api"
	"github.com/brocaar/chirpstack-application-server/internal/api/external/auth"
	"github.com/brocaar/chirpstack-application-server/internal/api/external/oidc"
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
	jwt, err := storage.LoginUserByPassword(ctx, storage.DB(), req.Email, req.Password)
	if nil != err {
		return nil, helpers.ErrToRPCError(err)
	}

	return &pb.LoginResponse{Jwt: jwt}, nil
}

// Profile returns the user profile.
func (a *InternalAPI) Profile(ctx context.Context, req *empty.Empty) (*pb.ProfileResponse, error) {
	if err := a.validator.Validate(ctx,
		auth.ValidateActiveUser()); err != nil {
		return nil, grpc.Errorf(codes.Unauthenticated, "authentication failed: %s", err)
	}

	// Get the user
	user, err := a.validator.GetUser(ctx)
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
			Email:      prof.User.Email,
			SessionTtl: prof.User.SessionTTL,
			IsAdmin:    prof.User.IsAdmin,
			IsActive:   prof.User.IsActive,
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

// GlobalSearch performs a global search.
func (a *InternalAPI) GlobalSearch(ctx context.Context, req *pb.GlobalSearchRequest) (*pb.GlobalSearchResponse, error) {
	if err := a.validator.Validate(ctx,
		auth.ValidateActiveUser()); err != nil {
		return nil, grpc.Errorf(codes.Unauthenticated, "authentication failed: %s", err)
	}

	user, err := a.validator.GetUser(ctx)
	if err != nil {
		return nil, helpers.ErrToRPCError(err)
	}

	results, err := storage.GlobalSearch(ctx, storage.DB(), user.ID, user.IsAdmin, req.Search, int(req.Limit), int(req.Offset))
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

	if !ak.IsAdmin && ak.OrganizationID == nil && ak.ApplicationID == nil {
		return nil, grpc.Errorf(codes.InvalidArgument, "the api key must be either of type admin, organization or application")
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

// Settings returns the global settings.
func (a *InternalAPI) Settings(ctx context.Context, _ *empty.Empty) (*pb.SettingsResponse, error) {
	return &pb.SettingsResponse{
		Branding: &pb.Branding{
			Registration: brandingRegistration,
			Footer:       brandingFooter,
		},
		OpenidConnect: &pb.OpenIDConnect{
			Enabled:    openIDConnectEnabled,
			LoginLabel: openIDLoginLabel,
			LoginUrl:   "/auth/oidc/login",
			LogoutUrl:  logoutURL,
		},
	}, nil
}

// OpenIDConnectLogin performs an OpenID Connect login.
func (a *InternalAPI) OpenIDConnectLogin(ctx context.Context, req *pb.OpenIDConnectLoginRequest) (*pb.OpenIDConnectLoginResponse, error) {
	oidcUser, err := oidc.GetUser(ctx, req.Code, req.State)
	if err != nil {
		return nil, helpers.ErrToRPCError(err)
	}

	if !oidcUser.EmailVerified {
		return nil, grpc.Errorf(codes.FailedPrecondition, "email address must be verified before you can login")
	}

	var user storage.User

	// try to get the user by external ID.
	user, err = storage.GetUserByExternalID(ctx, storage.DB(), oidcUser.ExternalID)
	if err != nil {
		if err == storage.ErrDoesNotExist {
			// try to get the user by email and set the external id.
			user, err = storage.GetUserByEmail(ctx, storage.DB(), oidcUser.Email)
			if err != nil {
				// we did not find the user by external_id or email and registration is enabled.
				if err == storage.ErrDoesNotExist && registrationEnabled {
					user, err = a.createAndProvisionUser(ctx, oidcUser)
					if err != nil {
						return nil, helpers.ErrToRPCError(err)
					}
					// fetch user again because the provisioning callback url may have updated the user.
					user, err = storage.GetUser(ctx, storage.DB(), user.ID)
					if err != nil {
						return nil, helpers.ErrToRPCError(err)
					}
				} else {
					return nil, helpers.ErrToRPCError(err)
				}
			}
			user.ExternalID = &oidcUser.ExternalID
		} else {
			return nil, helpers.ErrToRPCError(err)
		}
	}

	// update the user
	user.Email = oidcUser.Email
	user.EmailVerified = oidcUser.EmailVerified
	if err := storage.UpdateUser(ctx, storage.DB(), &user); err != nil {
		return nil, helpers.ErrToRPCError(err)
	}

	// get the jwt token
	token, err := storage.GetUserToken(user)
	if err != nil {
		return nil, helpers.ErrToRPCError(err)
	}

	return &pb.OpenIDConnectLoginResponse{
		JwtToken: token,
	}, nil
}

// GetDevicesSummary returns an aggregated devices summary.
func (a *InternalAPI) GetDevicesSummary(ctx context.Context, req *pb.GetDevicesSummaryRequest) (*pb.GetDevicesSummaryResponse, error) {
	if err := a.validator.Validate(ctx,
		auth.ValidateOrganizationAccess(auth.Read, req.OrganizationId)); err != nil {
		return nil, grpc.Errorf(codes.Unauthenticated, "authentication failed: %s", err)
	}

	daic, err := storage.GetDevicesActiveInactive(ctx, storage.DB(), req.OrganizationId)
	if err != nil {
		return nil, helpers.ErrToRPCError(err)
	}

	ddr, err := storage.GetDevicesDataRates(ctx, storage.DB(), req.OrganizationId)
	if err != nil {
		return nil, helpers.ErrToRPCError(err)
	}

	out := pb.GetDevicesSummaryResponse{
		NeverSeenCount: daic.NeverSeenCount,
		ActiveCount:    daic.ActiveCount,
		InactiveCount:  daic.InactiveCount,
		DrCount:        ddr,
	}

	return &out, nil
}

// GetGatewaysSummary returns an aggregated gateways summary.
func (a *InternalAPI) GetGatewaysSummary(ctx context.Context, req *pb.GetGatewaysSummaryRequest) (*pb.GetGatewaysSummaryResponse, error) {
	if err := a.validator.Validate(ctx,
		auth.ValidateOrganizationAccess(auth.Read, req.OrganizationId)); err != nil {
		return nil, grpc.Errorf(codes.Unauthenticated, "authentication failed: %s", err)
	}

	gai, err := storage.GetGatewaysActiveInactive(ctx, storage.DB(), req.OrganizationId)
	if err != nil {
		return nil, helpers.ErrToRPCError(err)
	}

	out := pb.GetGatewaysSummaryResponse{
		NeverSeenCount: gai.NeverSeenCount,
		ActiveCount:    gai.ActiveCount,
		InactiveCount:  gai.InactiveCount,
	}

	return &out, nil
}

func (a *InternalAPI) createAndProvisionUser(ctx context.Context, user oidc.User) (storage.User, error) {
	u := storage.User{
		IsActive:      true,
		Email:         user.Email,
		EmailVerified: user.EmailVerified,
		ExternalID:    &user.ExternalID,
	}

	if err := storage.CreateUser(ctx, storage.DB(), &u); err != nil {
		return storage.User{}, errors.Wrap(err, "create user error")
	}

	if registrationCallbackURL == "" {
		return u, nil
	}

	if err := a.provisionUser(ctx, u, user); err != nil {
		if err := storage.DeleteUser(ctx, storage.DB(), u.ID); err != nil {
			return storage.User{}, errors.Wrap(err, "delete user error after failed user provisioning")
		}

		log.WithError(err).Error("api/external: provision user error")

		return storage.User{}, errors.New("error provisioning user")
	}

	return u, nil
}

func (a *InternalAPI) provisionUser(ctx context.Context, u storage.User, oidcUser oidc.User) error {
	req, err := http.NewRequestWithContext(ctx, "POST", registrationCallbackURL, nil)
	if err != nil {
		return errors.Wrap(err, "new request error")
	}
	q := req.URL.Query()
	q.Add("user_id", fmt.Sprintf("%d", u.ID))
	claims, err := json.Marshal(oidcUser.UserInfoClaims)
	if err != nil {
		return errors.Wrap(err, "marshal claims error")
	}
	q.Add("oidc_claims", string(claims))
	req.URL.RawQuery = q.Encode()

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return errors.Wrap(err, "make registration callback request error")
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("registration callback must return 200, received: %d (%s)", resp.StatusCode, resp.Status)
	}

	return nil
}
