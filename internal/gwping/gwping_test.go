package gwping

import (
	"context"
	"testing"
	"time"

	"github.com/golang/protobuf/ptypes"
	. "github.com/smartystreets/goconvey/convey"

	"github.com/brocaar/chirpstack-api/go/v3/as"
	"github.com/brocaar/chirpstack-api/go/v3/common"
	"github.com/brocaar/chirpstack-api/go/v3/gw"
	"github.com/brocaar/chirpstack-application-server/internal/backend/networkserver"
	"github.com/brocaar/chirpstack-application-server/internal/backend/networkserver/mock"
	"github.com/brocaar/chirpstack-application-server/internal/storage"
	"github.com/brocaar/chirpstack-application-server/internal/test"
	"github.com/brocaar/lorawan"
)

func TestGatewayPing(t *testing.T) {
	conf := test.GetConfig()
	if err := storage.Setup(conf); err != nil {
		t.Fatal(err)
	}

	Convey("Given a clean database and a gateway", t, func() {
		nsClient := mock.NewClient()
		So(storage.MigrateDown(storage.DB().DB), ShouldBeNil)
		So(storage.MigrateUp(storage.DB().DB), ShouldBeNil)
		storage.RedisClient().FlushAll(context.Background())
		networkserver.SetPool(mock.NewPool(nsClient))

		org := storage.Organization{
			Name: "test-org",
		}
		So(storage.CreateOrganization(context.Background(), storage.DB(), &org), ShouldBeNil)

		n := storage.NetworkServer{
			Name:                        "test-ns",
			Server:                      "test-ns:1234",
			GatewayDiscoveryEnabled:     true,
			GatewayDiscoveryDR:          5,
			GatewayDiscoveryTXFrequency: 868100000,
			GatewayDiscoveryInterval:    1,
		}
		So(storage.CreateNetworkServer(context.Background(), storage.DB(), &n), ShouldBeNil)

		gw1 := storage.Gateway{
			MAC:             lorawan.EUI64{1, 2, 3, 4, 5, 6, 7, 8},
			Name:            "test-gw",
			Description:     "test gateway",
			OrganizationID:  org.ID,
			Ping:            true,
			NetworkServerID: n.ID,
		}
		So(storage.CreateGateway(context.Background(), storage.DB(), &gw1), ShouldBeNil)

		Convey("When gateway discovery is disabled on the network-server", func() {
			n.GatewayDiscoveryEnabled = false
			So(storage.UpdateNetworkServer(context.Background(), storage.DB(), &n), ShouldBeNil)

			Convey("When calling sendGatewayPing", func() {
				So(sendGatewayPing(context.Background()), ShouldBeNil)
			})

			Convey("Then no ping was sent", func() {
				gwGet, err := storage.GetGateway(context.Background(), storage.DB(), gw1.MAC, false)
				So(err, ShouldBeNil)
				So(gwGet.LastPingID, ShouldBeNil)
				So(gwGet.LastPingSentAt, ShouldBeNil)
			})
		})

		Convey("When calling sendGatewayPing", func() {
			So(sendGatewayPing(context.Background()), ShouldBeNil)

			Convey("Then the gateway ping fields have been set", func() {
				gwGet, err := storage.GetGateway(context.Background(), storage.DB(), gw1.MAC, false)
				So(err, ShouldBeNil)
				So(gwGet.LastPingID, ShouldNotBeNil)
				So(gwGet.LastPingSentAt, ShouldNotBeNil)

				Convey("Then a gateway ping records has been created", func() {
					gwPing, err := storage.GetGatewayPing(context.Background(), storage.DB(), *gwGet.LastPingID)
					So(err, ShouldBeNil)
					So(gwPing.GatewayMAC, ShouldEqual, gwGet.MAC)
					So(gwPing.DR, ShouldEqual, n.GatewayDiscoveryDR)
					So(gwPing.Frequency, ShouldEqual, n.GatewayDiscoveryTXFrequency)
				})

				Convey("Then the expected ping has been sent to the network-server", func() {
					So(nsClient.SendProprietaryPayloadChan, ShouldHaveLength, 1)
					req := <-nsClient.SendProprietaryPayloadChan
					So(req.Dr, ShouldEqual, uint32(n.GatewayDiscoveryDR))
					So(req.Frequency, ShouldEqual, uint32(n.GatewayDiscoveryTXFrequency))
					So(req.GatewayMacs, ShouldResemble, [][]byte{{1, 2, 3, 4, 5, 6, 7, 8}})
					So(req.PolarizationInversion, ShouldBeFalse)

					var mic lorawan.MIC
					copy(mic[:], req.Mic)
					So(mic, ShouldNotEqual, lorawan.MIC{})

					Convey("Then a ping lookup has been created", func() {
						id, err := getPingLookup(mic)
						So(err, ShouldBeNil)
						So(id, ShouldEqual, *gwGet.LastPingID)
					})

					Convey("When calling HandleReceivedPing", func() {
						gw2 := storage.Gateway{
							MAC:             lorawan.EUI64{8, 7, 6, 5, 4, 3, 2, 1},
							Name:            "test-gw-2",
							Description:     "test gateway 2",
							OrganizationID:  org.ID,
							NetworkServerID: n.ID,
						}
						So(storage.CreateGateway(context.Background(), storage.DB(), &gw2), ShouldBeNil)

						now := time.Now().UTC().Truncate(time.Millisecond)

						pong := as.HandleProprietaryUplinkRequest{
							Mic: mic[:],
							RxInfo: []*gw.UplinkRXInfo{
								{
									GatewayId: gw2.MAC[:],
									Rssi:      -10,
									LoraSnr:   5.5,
									Location: &common.Location{
										Latitude:  1.12345,
										Longitude: 1.23456,
										Altitude:  10,
									},
								},
							},
						}
						pong.RxInfo[0].Time, _ = ptypes.TimestampProto(now)
						So(HandleReceivedPing(context.Background(), &pong), ShouldBeNil)

						Convey("Then the ping lookup has been deleted", func() {
							_, err := getPingLookup(mic)
							So(err, ShouldNotBeNil)
						})

						Convey("Then the received ping has been stored to the database", func() {
							ping, rx, err := storage.GetLastGatewayPingAndRX(context.Background(), storage.DB(), gw1.MAC)
							So(err, ShouldBeNil)

							So(ping.ID, ShouldEqual, *gwGet.LastPingID)
							So(rx, ShouldHaveLength, 1)
							So(rx[0].GatewayMAC, ShouldEqual, gw2.MAC)
							So(rx[0].ReceivedAt.Equal(now), ShouldBeTrue)
							So(rx[0].RSSI, ShouldEqual, -10)
							So(rx[0].LoRaSNR, ShouldEqual, 5.5)
							So(rx[0].Location, ShouldResemble, storage.GPSPoint{
								Latitude:  1.12345,
								Longitude: 1.23456,
							})
							So(rx[0].Altitude, ShouldEqual, 10)
						})
					})
				})
			})
		})
	})
}
