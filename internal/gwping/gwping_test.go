package gwping

import (
	"testing"
	"time"

	"github.com/brocaar/loraserver/api/as"

	"github.com/brocaar/lorawan"

	"github.com/gusseleet/lora-app-server/internal/config"
	"github.com/gusseleet/lora-app-server/internal/storage"
	"github.com/gusseleet/lora-app-server/internal/test"
	. "github.com/smartystreets/goconvey/convey"
)

func TestGatewayPing(t *testing.T) {
	conf := test.GetConfig()
	db, err := storage.OpenDatabase(conf.PostgresDSN)
	if err != nil {
		t.Fatal(err)
	}
	config.C.PostgreSQL.DB = db
	config.C.Redis.Pool = storage.NewRedisPool(conf.RedisURL)
	config.C.ApplicationServer.GatewayDiscovery.DR = 5
	config.C.ApplicationServer.GatewayDiscovery.Frequency = 868100000

	Convey("Given a clean database and a gateway", t, func() {
		nsClient := test.NewNetworkServerClient()
		test.MustResetDB(db)
		test.MustFlushRedis(config.C.Redis.Pool)
		config.C.NetworkServer.Pool = test.NewNetworkServerPool(nsClient)

		org := storage.Organization{
			Name: "test-org",
		}
		So(storage.CreateOrganization(config.C.PostgreSQL.DB, &org), ShouldBeNil)

		n := storage.NetworkServer{
			Name:   "test-ns",
			Server: "test-ns:1234",
		}
		So(storage.CreateNetworkServer(config.C.PostgreSQL.DB, &n), ShouldBeNil)

		gw := storage.Gateway{
			MAC:             lorawan.EUI64{1, 2, 3, 4, 5, 6, 7, 8},
			Name:            "test-gw",
			Description:     "test gateway",
			OrganizationID:  org.ID,
			Ping:            true,
			NetworkServerID: n.ID,
		}
		So(storage.CreateGateway(config.C.PostgreSQL.DB, &gw), ShouldBeNil)

		Convey("When calling sendGatewayPing", func() {
			So(sendGatewayPing(), ShouldBeNil)

			Convey("Then the gateway ping fields have been set", func() {
				gwGet, err := storage.GetGateway(config.C.PostgreSQL.DB, gw.MAC, false)
				So(err, ShouldBeNil)
				So(gwGet.LastPingID, ShouldNotBeNil)
				So(gwGet.LastPingSentAt, ShouldNotBeNil)

				Convey("Then a gateway ping records has been created", func() {
					gwPing, err := storage.GetGatewayPing(config.C.PostgreSQL.DB, *gwGet.LastPingID)
					So(err, ShouldBeNil)
					So(gwPing.GatewayMAC, ShouldEqual, gwGet.MAC)
					So(gwPing.DR, ShouldEqual, config.C.ApplicationServer.GatewayDiscovery.DR)
					So(gwPing.Frequency, ShouldEqual, config.C.ApplicationServer.GatewayDiscovery.Frequency)
				})

				Convey("Then the expected ping has been sent to the network-server", func() {
					So(nsClient.SendProprietaryPayloadChan, ShouldHaveLength, 1)
					req := <-nsClient.SendProprietaryPayloadChan
					So(req.Dr, ShouldEqual, uint32(config.C.ApplicationServer.GatewayDiscovery.DR))
					So(req.Frequency, ShouldEqual, uint32(config.C.ApplicationServer.GatewayDiscovery.Frequency))
					So(req.GatewayMACs, ShouldResemble, [][]byte{{1, 2, 3, 4, 5, 6, 7, 8}})
					So(req.IPol, ShouldBeFalse)

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
						So(storage.CreateGateway(config.C.PostgreSQL.DB, &gw2), ShouldBeNil)

						now := time.Now().UTC().Truncate(time.Millisecond)

						pong := as.HandleProprietaryUplinkRequest{
							Mic: mic[:],
							RxInfo: []*as.RXInfo{
								{
									Mac:       gw2.MAC[:],
									Time:      now.Format(time.RFC3339Nano),
									Rssi:      -10,
									LoRaSNR:   5.5,
									Name:      "test-gw-2",
									Latitude:  1.12345,
									Longitude: 1.23456,
									Altitude:  10,
								},
							},
						}
						So(HandleReceivedPing(&pong), ShouldBeNil)

						Convey("Then the ping lookup has been deleted", func() {
							_, err := getPingLookup(mic)
							So(err, ShouldNotBeNil)
						})

						Convey("Then the received ping has been stored to the database", func() {
							ping, rx, err := storage.GetLastGatewayPingAndRX(config.C.PostgreSQL.DB, gw.MAC)
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
