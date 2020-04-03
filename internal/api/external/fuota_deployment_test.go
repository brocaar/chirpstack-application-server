package external

import (
	"testing"
	"time"

	"github.com/gofrs/uuid"
	"github.com/golang/protobuf/ptypes"
	"github.com/stretchr/testify/require"
	"golang.org/x/net/context"

	pb "github.com/brocaar/chirpstack-api/go/v3/as/external/api"
	"github.com/brocaar/chirpstack-application-server/internal/backend/networkserver"
	"github.com/brocaar/chirpstack-application-server/internal/backend/networkserver/mock"
	"github.com/brocaar/chirpstack-application-server/internal/storage"
	"github.com/brocaar/chirpstack-api/go/v3/common"
	"github.com/brocaar/chirpstack-api/go/v3/ns"
	"github.com/brocaar/lorawan"
)

func (ts *APITestSuite) TestFUOTADeployment() {
	assert := require.New(ts.T())

	nsClient := mock.NewClient()
	networkserver.SetPool(mock.NewPool(nsClient))

	validator := &TestValidator{}
	api := NewFUOTADeploymentAPI(validator)

	n := storage.NetworkServer{
		Name:   "test",
		Server: "test:1234",
	}
	assert.NoError(storage.CreateNetworkServer(context.Background(), storage.DB(), &n))

	org := storage.Organization{
		Name: "test-org",
	}
	assert.NoError(storage.CreateOrganization(context.Background(), storage.DB(), &org))

	sp := storage.ServiceProfile{
		Name:            "test-sp",
		OrganizationID:  org.ID,
		NetworkServerID: n.ID,
	}
	assert.NoError(storage.CreateServiceProfile(context.Background(), storage.DB(), &sp))
	var spID uuid.UUID
	copy(spID[:], sp.ServiceProfile.Id)

	app := storage.Application{
		Name:           "test-app",
		OrganizationID: org.ID,
	}
	copy(app.ServiceProfileID[:], sp.ServiceProfile.Id)
	assert.NoError(storage.CreateApplication(context.Background(), storage.DB(), &app))

	dp := storage.DeviceProfile{
		Name:            "test-dp",
		OrganizationID:  org.ID,
		NetworkServerID: n.ID,
	}
	assert.NoError(storage.CreateDeviceProfile(context.Background(), storage.DB(), &dp))
	var dpID uuid.UUID
	copy(dpID[:], dp.DeviceProfile.Id)

	d := storage.Device{
		DevEUI:          lorawan.EUI64{1, 2, 3, 4, 5, 6, 7, 8},
		ApplicationID:   app.ID,
		DeviceProfileID: dpID,
		Name:            "test-device",
	}
	assert.NoError(storage.CreateDevice(context.Background(), storage.DB(), &d))

	ts.T().Run("CreateForDevice", func(t *testing.T) {
		assert := require.New(t)

		nsClient.GetVersionResponse = ns.GetVersionResponse{
			Region: common.Region_EU868,
		}

		req := pb.CreateFUOTADeploymentForDeviceRequest{
			DevEui: d.DevEUI.String(),
			FuotaDeployment: &pb.FUOTADeployment{
				Name:             "test-deployment",
				GroupType:        pb.MulticastGroupType_CLASS_C,
				Dr:               5,
				Frequency:        868100000,
				Payload:          []byte{1, 2, 3, 4},
				Redundancy:       2,
				MulticastTimeout: 3,
				UnicastTimeout:   ptypes.DurationProto(5 * time.Second),
			},
		}

		resp, err := api.CreateForDevice(context.Background(), &req)
		assert.NoError(err)
		assert.NotEqual("", resp.Id)

		req.FuotaDeployment.Id = resp.Id

		t.Run("Get", func(t *testing.T) {
			assert := require.New(t)

			resp, err := api.Get(context.Background(), &pb.GetFUOTADeploymentRequest{
				Id: resp.Id,
			})
			assert.NoError(err)

			assert.NotEqual("", resp.CreatedAt)
			assert.NotEqual("", resp.UpdatedAt)
			assert.NotNil(resp.FuotaDeployment)
			assert.NotNil(resp.FuotaDeployment.NextStepAfter)

			req.FuotaDeployment.State = "MC_CREATE"
			resp.FuotaDeployment.NextStepAfter = nil
			assert.Equal(req.FuotaDeployment, resp.FuotaDeployment)
		})

		t.Run("List", func(t *testing.T) {
			t.Run("By ApplicationID", func(t *testing.T) {
				assert := require.New(t)

				resp, err := api.List(context.Background(), &pb.ListFUOTADeploymentRequest{
					Limit:         10,
					ApplicationId: app.ID,
				})
				assert.NoError(err)

				assert.EqualValues(1, resp.TotalCount)
				assert.Len(resp.Result, 1)
			})

			t.Run("By DevEUI", func(t *testing.T) {
				assert := require.New(t)

				resp, err := api.List(context.Background(), &pb.ListFUOTADeploymentRequest{
					Limit:  10,
					DevEui: d.DevEUI.String(),
				})
				assert.NoError(err)

				assert.EqualValues(1, resp.TotalCount)
				assert.Len(resp.Result, 1)
			})

			t.Run("No filters - no admin", func(t *testing.T) {
				assert := require.New(t)
				validator.returnIsAdmin = false

				_, err := api.List(context.Background(), &pb.ListFUOTADeploymentRequest{
					Limit: 10,
				})
				assert.NotNil(err)
			})

			t.Run("No filters - admin", func(t *testing.T) {
				assert := require.New(t)
				validator.returnIsAdmin = true

				resp, err := api.List(context.Background(), &pb.ListFUOTADeploymentRequest{
					Limit: 10,
				})
				assert.NoError(err)

				assert.EqualValues(1, resp.TotalCount)
				assert.Len(resp.Result, 1)
			})
		})

		t.Run("GetDeploymentDevice", func(t *testing.T) {
			assert := require.New(t)

			resp, err := api.GetDeploymentDevice(context.Background(), &pb.GetFUOTADeploymentDeviceRequest{
				FuotaDeploymentId: resp.Id,
				DevEui:            d.DevEUI.String(),
			})
			assert.NoError(err)

			assert.NotNil(resp.DeploymentDevice)
			assert.Equal(d.DevEUI.String(), resp.DeploymentDevice.DevEui)
			assert.Equal(d.Name, resp.DeploymentDevice.DeviceName)
			assert.Equal(pb.FUOTADeploymentDeviceState_PENDING, resp.DeploymentDevice.State)
			assert.Equal("", resp.DeploymentDevice.ErrorMessage)
			assert.NotNil(resp.DeploymentDevice.CreatedAt)
			assert.NotNil(resp.DeploymentDevice.UpdatedAt)
		})

		t.Run("ListDeploymentDevices", func(t *testing.T) {
			assert := require.New(t)

			resp, err := api.ListDeploymentDevices(context.Background(), &pb.ListFUOTADeploymentDevicesRequest{
				FuotaDeploymentId: resp.Id,
				Limit:             10,
			})
			assert.NoError(err)

			assert.EqualValues(1, resp.TotalCount)
			assert.Len(resp.Result, 1)
		})
	})
}
