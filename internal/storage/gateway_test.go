package storage

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/gofrs/uuid"
	"github.com/lib/pq/hstore"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"

	"github.com/brocaar/chirpstack-application-server/internal/backend/networkserver"
	"github.com/brocaar/chirpstack-application-server/internal/backend/networkserver/mock"
	"github.com/brocaar/lorawan"
)

func (ts *StorageTestSuite) TestGateway() {
	assert := require.New(ts.T())

	nsClient := mock.NewClient()
	networkserver.SetPool(mock.NewPool(nsClient))

	n := NetworkServer{
		Name:   "test-ns",
		Server: "test-ns:1234",
	}
	assert.NoError(CreateNetworkServer(context.Background(), DB(), &n))

	org := Organization{
		Name: "test-org",
	}
	assert.NoError(CreateOrganization(context.Background(), DB(), &org))

	sp := ServiceProfile{
		Name:            "test-sp",
		NetworkServerID: n.ID,
		OrganizationID:  org.ID,
	}
	assert.NoError(CreateServiceProfile(context.Background(), DB(), &sp))
	var spID uuid.UUID
	copy(spID[:], sp.ServiceProfile.Id)

	ts.T().Run("Create with invalid name", func(t *testing.T) {
		assert := require.New(t)

		gw := Gateway{
			MAC:  lorawan.EUI64{1, 2, 3, 4, 5, 6, 7, 8},
			Name: "test gateway",
		}
		err := CreateGateway(context.Background(), ts.Tx(), &gw)
		assert.Equal(ErrGatewayInvalidName, errors.Cause(err))
	})

	ts.T().Run("Create", func(t *testing.T) {
		assert := require.New(t)
		now := time.Now().UTC().Round(time.Second)

		gw := Gateway{
			MAC:             lorawan.EUI64{1, 2, 3, 4, 5, 6, 7, 8},
			FirstSeenAt:     &now,
			LastSeenAt:      &now,
			Name:            "test-gw",
			Description:     "test gateway",
			OrganizationID:  org.ID,
			Ping:            true,
			NetworkServerID: n.ID,
			Latitude:        1,
			Longitude:       2,
			Altitude:        3,
			Tags: hstore.Hstore{
				Map: map[string]sql.NullString{
					"foo": sql.NullString{Valid: true, String: "bar"},
				},
			},
			Metadata: hstore.Hstore{
				Map: map[string]sql.NullString{
					"foo": sql.NullString{Valid: true, String: "bar"},
				},
			},
			ServiceProfileID: &spID,
		}
		assert.NoError(CreateGateway(context.Background(), ts.Tx(), &gw))
		gw.CreatedAt = gw.CreatedAt.Round(time.Millisecond).UTC()
		gw.UpdatedAt = gw.UpdatedAt.Round(time.Millisecond).UTC()

		t.Run("Get", func(t *testing.T) {
			assert := require.New(t)

			gwGet, err := GetGateway(context.Background(), ts.Tx(), gw.MAC, false)
			assert.NoError(err)

			gwGet.CreatedAt = gwGet.CreatedAt.Round(time.Millisecond).UTC()
			gwGet.UpdatedAt = gwGet.CreatedAt.Round(time.Millisecond).UTC()
			firstSeen := gwGet.FirstSeenAt.UTC()
			lastSeen := gwGet.LastSeenAt.UTC()
			gw.FirstSeenAt = &firstSeen
			gw.LastSeenAt = &lastSeen

			assert.Equal(gw, gwGet)
		})

		t.Run("Update", func(t *testing.T) {
			assert := require.New(t)

			gw.Name = "test-gw2"
			gw.Description = "updated test gateway"
			gw.Ping = false
			gw.ServiceProfileID = nil

			assert.NoError(UpdateGateway(context.Background(), ts.Tx(), &gw))
			gw.CreatedAt = gw.CreatedAt.Round(time.Millisecond).UTC()
			gw.UpdatedAt = gw.UpdatedAt.Round(time.Millisecond).UTC()

			gwGet, err := GetGateway(context.Background(), ts.Tx(), gw.MAC, false)
			assert.NoError(err)

			gwGet.UpdatedAt = gwGet.UpdatedAt.Round(time.Millisecond).UTC()
			gwGet.CreatedAt = gwGet.CreatedAt.Round(time.Millisecond).UTC()
			firstSeen := gwGet.FirstSeenAt.UTC()
			lastSeen := gwGet.LastSeenAt.UTC()
			gw.FirstSeenAt = &firstSeen
			gw.LastSeenAt = &lastSeen

			assert.Equal(gw, gwGet)
		})

		t.Run("GetGatewaysActiveInactive", func(t *testing.T) {
			assert := require.New(t)
			ls := time.Now()

			// gateway is never seen
			gw.LastSeenAt = nil
			assert.NoError(UpdateGateway(context.Background(), ts.Tx(), &gw))

			ga, err := GetGatewaysActiveInactive(context.Background(), ts.Tx(), gw.OrganizationID)
			assert.NoError(err)
			assert.Equal(GatewaysActiveInactive{
				NeverSeenCount: 1,
				ActiveCount:    0,
				InactiveCount:  0,
			}, ga)

			// gateway is active
			gw.LastSeenAt = &ls
			assert.NoError(UpdateGateway(context.Background(), ts.Tx(), &gw))

			ga, err = GetGatewaysActiveInactive(context.Background(), ts.Tx(), gw.OrganizationID)
			assert.NoError(err)
			assert.Equal(GatewaysActiveInactive{
				NeverSeenCount: 0,
				ActiveCount:    1,
				InactiveCount:  0,
			}, ga)

			// gateway is inactive
			ls = ls.Add(time.Second * -61)
			gw.LastSeenAt = &ls
			assert.NoError(UpdateGateway(context.Background(), ts.Tx(), &gw))

			ga, err = GetGatewaysActiveInactive(context.Background(), ts.Tx(), gw.OrganizationID)
			assert.NoError(err)
			assert.Equal(GatewaysActiveInactive{
				NeverSeenCount: 0,
				ActiveCount:    0,
				InactiveCount:  1,
			}, ga)
		})

		t.Run("Get gateways", func(t *testing.T) {
			assert := require.New(t)

			c, err := GetGatewayCount(context.Background(), ts.Tx(), GatewayFilters{})
			assert.NoError(err)
			assert.Equal(1, c)

			gws, err := GetGateways(context.Background(), ts.Tx(), GatewayFilters{
				Limit: 10,
			})
			assert.NoError(err)
			assert.Len(gws, 1)
			assert.Equal(gw.MAC, gws[0].MAC)
		})

		t.Run("Get get gateways for organization id", func(t *testing.T) {
			assert := require.New(t)

			c, err := GetGatewayCount(context.Background(), ts.Tx(), GatewayFilters{
				OrganizationID: org.ID,
			})
			assert.NoError(err)
			assert.Equal(1, c)

			gws, err := GetGateways(context.Background(), ts.Tx(), GatewayFilters{
				OrganizationID: org.ID,
				Limit:          10,
			})
			assert.NoError(err)
			assert.Len(gws, 1)
			assert.Equal(gw.MAC, gws[0].MAC)

			c, err = GetGatewayCount(context.Background(), ts.Tx(), GatewayFilters{
				OrganizationID: org.ID + 1,
			})
			assert.NoError(err)
			assert.Equal(0, c)

			gws, err = GetGateways(context.Background(), ts.Tx(), GatewayFilters{
				OrganizationID: org.ID + 1,
				Limit:          10,
			})
			assert.NoError(err)
			assert.Len(gws, 0)
		})

		t.Run("Get gateways for username", func(t *testing.T) {
			assert := require.New(t)
			user := User{
				IsActive: true,
				Email:    "foo@bar.com",
			}
			err := CreateUser(context.Background(), DB(), &user)
			assert.NoError(err)

			c, err := GetGatewayCount(context.Background(), ts.Tx(), GatewayFilters{
				UserID: user.ID,
			})
			assert.NoError(err)
			assert.Equal(0, c)

			gws, err := GetGateways(context.Background(), ts.Tx(), GatewayFilters{
				UserID: user.ID,
				Limit:  1,
			})
			assert.NoError(err)
			assert.Len(gws, 0)

			assert.NoError(CreateOrganizationUser(context.Background(), DB(), org.ID, user.ID, false, false, false))

			c, err = GetGatewayCount(context.Background(), ts.Tx(), GatewayFilters{
				UserID: user.ID,
			})
			assert.NoError(err)
			assert.Equal(1, c)

			gws, err = GetGateways(context.Background(), ts.Tx(), GatewayFilters{
				UserID: user.ID,
				Limit:  1,
			})
			assert.NoError(err)
			assert.Len(gws, 1)
			assert.Equal(gw.MAC, gws[0].MAC)
		})

		t.Run("Create gateway ping", func(t *testing.T) {
			assert := require.New(t)

			gwPing := GatewayPing{
				GatewayMAC: gw.MAC,
				Frequency:  868100000,
				DR:         5,
			}
			assert.NoError(CreateGatewayPing(context.Background(), ts.Tx(), &gwPing))
			gwPing.CreatedAt = gwPing.CreatedAt.UTC().Round(time.Millisecond)

			t.Run("Get", func(t *testing.T) {
				assert := require.New(t)

				gwPingGet, err := GetGatewayPing(context.Background(), ts.Tx(), gwPing.ID)
				assert.NoError(err)
				gwPingGet.CreatedAt = gwPingGet.CreatedAt.UTC().Round(time.Millisecond)

				assert.Equal(gwPing, gwPingGet)
			})

			t.Run("Create ping RX", func(t *testing.T) {
				assert := require.New(t)
				now := time.Now().Truncate(time.Millisecond)

				gwPingRX := GatewayPingRX{
					PingID:     gwPing.ID,
					GatewayMAC: gw.MAC,
					ReceivedAt: &now,
					RSSI:       -10,
					LoRaSNR:    5.5,
					Location: GPSPoint{
						Latitude:  1.12345,
						Longitude: 1.23456,
					},
					Altitude: 10,
				}
				assert.NoError(CreateGatewayPingRX(context.Background(), ts.Tx(), &gwPingRX))
				gwPingRX.CreatedAt = gwPingRX.CreatedAt.UTC().Round(time.Millisecond)

				t.Run("Get ping RX", func(t *testing.T) {
					assert := require.New(t)

					gw.LastPingID = &gwPing.ID
					gw.LastPingSentAt = &gwPing.CreatedAt
					assert.NoError(UpdateGateway(context.Background(), ts.Tx(), &gw))

					rx, err := GetGatewayPingRXForPingID(context.Background(), ts.Tx(), gwPing.ID)
					assert.NoError(err)
					assert.Len(rx, 1)
					assert.Equal(gw.MAC, rx[0].GatewayMAC)
					assert.True(rx[0].ReceivedAt.Equal(now))
					assert.Equal(-10, rx[0].RSSI)
					assert.Equal(5.5, rx[0].LoRaSNR)
					assert.Equal(GPSPoint{
						Latitude:  1.12345,
						Longitude: 1.23456,
					}, rx[0].Location)
					assert.Equal(10.0, rx[0].Altitude)

					t.Run("Get last ping and RX", func(t *testing.T) {
						assert := require.New(t)

						gwPing2, gwPingRX2, err := GetLastGatewayPingAndRX(context.Background(), ts.Tx(), gwPing.GatewayMAC)
						assert.NoError(err)
						assert.Equal(gwPing.ID, gwPing2.ID)
						assert.Len(gwPingRX2, 1)
						assert.Equal(rx[0].ID, gwPingRX2[0].ID)
					})
				})
			})
		})

		t.Run("Delete", func(t *testing.T) {
			assert := require.New(t)

			assert.NoError(DeleteGateway(context.Background(), ts.Tx(), gw.MAC))

			_, err := GetGateway(context.Background(), ts.Tx(), gw.MAC, false)
			assert.Equal(ErrDoesNotExist, errors.Cause(err))
		})
	})
}
