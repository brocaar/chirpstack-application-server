package storage

import (
	"testing"
	"time"

	"github.com/gusseleet/lora-app-server/internal/config"
	"github.com/gusseleet/lora-app-server/internal/test"
	"github.com/brocaar/lorawan"
	"github.com/pkg/errors"
	. "github.com/smartystreets/goconvey/convey"
)

func TestGateway(t *testing.T) {
	conf := test.GetConfig()

	Convey("Given a clean database woth a network-server and organization", t, func() {
		db, err := OpenDatabase(conf.PostgresDSN)
		So(err, ShouldBeNil)
		test.MustResetDB(db)

		nsClient := test.NewNetworkServerClient()
		config.C.NetworkServer.Pool = test.NewNetworkServerPool(nsClient)

		n := NetworkServer{
			Name:   "test-ns",
			Server: "test-ns:1234",
		}
		So(CreateNetworkServer(db, &n), ShouldBeNil)

		org := Organization{
			Name: "test-org",
		}
		So(CreateOrganization(db, &org), ShouldBeNil)

		Convey("When creating a gateway with an invalid name", func() {
			gw := Gateway{
				MAC:  lorawan.EUI64{1, 2, 3, 4, 5, 6, 7, 8},
				Name: "test gateway",
			}
			err := CreateGateway(db, &gw)

			Convey("Then an error is returned", func() {
				So(err, ShouldNotBeNil)
				So(errors.Cause(err), ShouldResemble, ErrGatewayInvalidName)
			})
		})

		Convey("When creating a gateway", func() {
			gw := Gateway{
				MAC:             lorawan.EUI64{1, 2, 3, 4, 5, 6, 7, 8},
				Name:            "test-gw",
				Description:     "test gateway",
				OrganizationID:  org.ID,
				Ping:            true,
				NetworkServerID: n.ID,
			}
			So(CreateGateway(db, &gw), ShouldBeNil)
			gw.CreatedAt = gw.CreatedAt.Truncate(time.Millisecond).UTC()
			gw.UpdatedAt = gw.UpdatedAt.Truncate(time.Millisecond).UTC()

			Convey("Then it can be get by its MAC", func() {
				gw2, err := GetGateway(db, gw.MAC, false)
				So(err, ShouldBeNil)
				gw2.CreatedAt = gw2.CreatedAt.Truncate(time.Millisecond).UTC()
				gw2.UpdatedAt = gw2.UpdatedAt.Truncate(time.Millisecond).UTC()
				So(gw2, ShouldResemble, gw)
			})

			Convey("Then it can be updated", func() {
				gw.Name = "test-gw2"
				gw.Description = "updated test gateway"
				gw.Ping = false
				So(UpdateGateway(db, &gw), ShouldBeNil)
				gw.CreatedAt = gw.CreatedAt.Truncate(time.Millisecond).UTC()
				gw.UpdatedAt = gw.UpdatedAt.Truncate(time.Millisecond).UTC()

				gw2, err := GetGateway(db, gw.MAC, false)
				So(err, ShouldBeNil)
				gw2.CreatedAt = gw2.CreatedAt.Truncate(time.Millisecond).UTC()
				gw2.UpdatedAt = gw2.UpdatedAt.Truncate(time.Millisecond).UTC()
				So(gw2, ShouldResemble, gw)
			})

			Convey("Then it can be deleted", func() {
				So(DeleteGateway(db, gw.MAC), ShouldBeNil)
				_, err := GetGateway(db, gw.MAC, false)
				So(errors.Cause(err), ShouldResemble, ErrDoesNotExist)
			})

			Convey("Then getting the total gateway count returns 1", func() {
				c, err := GetGatewayCount(db)
				So(err, ShouldBeNil)
				So(c, ShouldEqual, 1)
			})

			Convey("Then getting all gateways returns the expected gateway", func() {
				gws, err := GetGateways(db, 10, 0)
				So(err, ShouldBeNil)
				So(gws, ShouldHaveLength, 1)
				So(gws[0].MAC, ShouldEqual, gw.MAC)
			})

			Convey("Then getting the total gateway count for the organization returns 1", func() {
				c, err := GetGatewayCountForOrganizationID(db, org.ID)
				So(err, ShouldBeNil)
				So(c, ShouldEqual, 1)
			})

			Convey("Then getting all gateways for the organization returns the exepected gateway", func() {
				gws, err := GetGatewaysForOrganizationID(db, org.ID, 10, 0)
				So(err, ShouldBeNil)
				So(gws, ShouldHaveLength, 1)
				So(gws[0].MAC, ShouldEqual, gw.MAC)
			})

			Convey("When creating an user", func() {
				user := User{
					Username: "testuser",
					IsActive: true,
					Email:    "foo@bar.com",
				}
				_, err := CreateUser(db, &user, "password123")
				So(err, ShouldBeNil)

				Convey("Getting the gateway count for this user returns 0", func() {
					c, err := GetGatewayCountForUser(db, user.Username)
					So(err, ShouldBeNil)
					So(c, ShouldEqual, 0)
				})

				Convey("Then getting the gateways for this user returns 0 items", func() {
					gws, err := GetGatewaysForUser(db, user.Username, 10, 0)
					So(err, ShouldBeNil)
					So(gws, ShouldHaveLength, 0)
				})

				Convey("When assigning the user to the organization", func() {
					So(CreateOrganizationUser(db, org.ID, user.ID, false), ShouldBeNil)

					Convey("Getting the gateway count for this user returns 1", func() {
						c, err := GetGatewayCountForUser(db, user.Username)
						So(err, ShouldBeNil)
						So(c, ShouldEqual, 1)
					})

					Convey("Then getting the gateways for this user returns 1 item", func() {
						gws, err := GetGatewaysForUser(db, user.Username, 10, 0)
						So(err, ShouldBeNil)
						So(gws, ShouldHaveLength, 1)
						So(gws[0].MAC, ShouldEqual, gw.MAC)
					})
				})
			})

			Convey("When creating a gateway ping", func() {
				gwPing := GatewayPing{
					GatewayMAC: gw.MAC,
					Frequency:  868100000,
					DR:         5,
				}
				So(CreateGatewayPing(db, &gwPing), ShouldBeNil)
				gwPing.CreatedAt = gwPing.CreatedAt.UTC().Truncate(time.Millisecond)

				Convey("Then the ping can be retrieved by its ID", func() {
					gwPingGet, err := GetGatewayPing(db, gwPing.ID)
					So(err, ShouldBeNil)
					gwPingGet.CreatedAt = gwPingGet.CreatedAt.UTC().Truncate(time.Millisecond)

					So(gwPingGet, ShouldResemble, gwPing)
				})

				Convey("Then a gateway ping rx can be created", func() {
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
					So(CreateGatewayPingRX(db, &gwPingRX), ShouldBeNil)
					gwPingRX.CreatedAt = gwPingRX.CreatedAt.UTC().Truncate(time.Millisecond)

					Convey("Then the ping rx can be retrieved by its ping ID", func() {
						gw.LastPingID = &gwPing.ID
						gw.LastPingSentAt = &gwPing.CreatedAt
						So(UpdateGateway(db, &gw), ShouldBeNil)

						rx, err := GetGatewayPingRXForPingID(db, gwPing.ID)
						So(err, ShouldBeNil)
						So(rx, ShouldHaveLength, 1)
						So(rx[0].GatewayMAC, ShouldEqual, gw.MAC)
						So(rx[0].ReceivedAt.Equal(now), ShouldBeTrue)
						So(rx[0].RSSI, ShouldEqual, -10)
						So(rx[0].LoRaSNR, ShouldEqual, 5.5)
						So(rx[0].Location, ShouldResemble, GPSPoint{
							Latitude:  1.12345,
							Longitude: 1.23456,
						})
						So(rx[0].Altitude, ShouldEqual, 10)

						Convey("Then the same ping is returned by GetLastGatewayPingAndRX", func() {
							gwPing2, gwPingRX2, err := GetLastGatewayPingAndRX(db, gwPing.GatewayMAC)
							So(err, ShouldBeNil)
							So(gwPing2.ID, ShouldEqual, gwPing.ID)
							So(gwPingRX2, ShouldHaveLength, 1)
							So(gwPingRX2[0].ID, ShouldEqual, rx[0].ID)
						})
					})
				})
			})
		})
	})
}
