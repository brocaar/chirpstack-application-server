package api

import (
	"context"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"

	pb "github.com/brocaar/lora-app-server/api"
	"github.com/brocaar/lora-app-server/internal/common"
	"github.com/brocaar/lora-app-server/internal/test"
	"github.com/brocaar/loraserver/api/ns"
)

func TestGatewayAPI(t *testing.T) {

	Convey("Given a mocked network-server api", t, func() {
		nsClient := test.NewNetworkServerClient()
		ctx := context.Background()
		lsCtx := common.Context{NetworkServer: nsClient}
		validator := &TestValidator{}
		api := NewGatewayAPI(lsCtx, validator)

		now := time.Now().UTC()

		getGatewayResponseNS := ns.GetGatewayResponse{
			Mac:         []byte{1, 2, 3, 4, 5, 6, 7, 8},
			Name:        "test-gateway",
			Description: "test gateway",
			Latitude:    1.1234,
			Longitude:   1.1235,
			Altitude:    5.5,
			CreatedAt:   now.UTC().Format(time.RFC3339Nano),
			UpdatedAt:   now.UTC().Add(1 * time.Second).Format(time.RFC3339Nano),
			FirstSeenAt: now.UTC().Add(2 * time.Second).Format(time.RFC3339Nano),
			LastSeenAt:  now.UTC().Add(3 * time.Second).Format(time.RFC3339Nano),
		}

		getGatewayResponseAS := pb.GetGatewayResponse{
			Mac:         "0102030405060708",
			Name:        "test-gateway",
			Description: "test gateway",
			Latitude:    1.1234,
			Longitude:   1.1235,
			Altitude:    5.5,
			CreatedAt:   now.UTC().Format(time.RFC3339Nano),
			UpdatedAt:   now.UTC().Add(1 * time.Second).Format(time.RFC3339Nano),
			FirstSeenAt: now.UTC().Add(2 * time.Second).Format(time.RFC3339Nano),
			LastSeenAt:  now.UTC().Add(3 * time.Second).Format(time.RFC3339Nano),
		}

		Convey("When calling Create", func() {
			_, err := api.Create(ctx, &pb.CreateGatewayRequest{
				Mac:         "0102030405060708",
				Name:        "test-gateway",
				Description: "test gateway",
				Latitude:    1.1234,
				Longitude:   1.1235,
				Altitude:    5.5,
			})
			So(err, ShouldBeNil)
			So(validator.ctx, ShouldResemble, ctx)
			So(validator.validatorFuncs, ShouldHaveLength, 1)

			Convey("Then the correct request was forwarded to the network-server api", func() {
				So(nsClient.CreateGatewayChan, ShouldHaveLength, 1)
				So(<-nsClient.CreateGatewayChan, ShouldResemble, ns.CreateGatewayRequest{
					Mac:         []byte{1, 2, 3, 4, 5, 6, 7, 8},
					Name:        "test-gateway",
					Description: "test gateway",
					Latitude:    1.1234,
					Longitude:   1.1235,
					Altitude:    5.5,
				})
			})
		})

		Convey("When calling Get", func() {
			nsClient.GetGatewayResponse = getGatewayResponseNS
			resp, err := api.Get(ctx, &pb.GetGatewayRequest{
				Mac: "0102030405060708",
			})
			So(err, ShouldBeNil)
			So(validator.ctx, ShouldResemble, ctx)
			So(validator.validatorFuncs, ShouldHaveLength, 1)

			Convey("Then the correct request / response was forwarded to / from the network-server api", func() {
				So(nsClient.GetGatewayChan, ShouldHaveLength, 1)
				So(<-nsClient.GetGatewayChan, ShouldResemble, ns.GetGatewayRequest{
					Mac: []byte{1, 2, 3, 4, 5, 6, 7, 8},
				})

				So(resp, ShouldResemble, &getGatewayResponseAS)
			})
		})

		Convey("When calling List", func() {
			nsClient.ListGatewayResponse = ns.ListGatewayResponse{
				TotalCount: 1,
				Result:     []*ns.GetGatewayResponse{&getGatewayResponseNS},
			}
			resp, err := api.List(ctx, &pb.ListGatewayRequest{
				Offset: 10,
				Limit:  20,
			})
			So(err, ShouldBeNil)
			So(validator.ctx, ShouldResemble, ctx)
			So(validator.validatorFuncs, ShouldHaveLength, 1)

			Convey("Then the correct / response was forwarded to / from the network-server api", func() {
				So(nsClient.ListGatewayChan, ShouldHaveLength, 1)
				So(<-nsClient.ListGatewayChan, ShouldResemble, ns.ListGatewayRequest{
					Offset: 10,
					Limit:  20,
				})

				So(resp, ShouldResemble, &pb.ListGatewayResponse{
					TotalCount: 1,
					Result:     []*pb.GetGatewayResponse{&getGatewayResponseAS},
				})
			})
		})

		Convey("When calling Update", func() {
			_, err := api.Update(ctx, &pb.UpdateGatewayRequest{
				Mac:         "0102030405060708",
				Name:        "test-gateway",
				Description: "test gateway",
				Latitude:    1.1234,
				Longitude:   1.1235,
				Altitude:    5.5,
			})
			So(err, ShouldBeNil)
			So(validator.ctx, ShouldResemble, ctx)
			So(validator.validatorFuncs, ShouldHaveLength, 1)

			Convey("Then the correct request was forwarded to the network-server api", func() {
				So(nsClient.UpdateGatewayChan, ShouldHaveLength, 1)
				So(<-nsClient.UpdateGatewayChan, ShouldResemble, ns.UpdateGatewayRequest{
					Mac:         []byte{1, 2, 3, 4, 5, 6, 7, 8},
					Name:        "test-gateway",
					Description: "test gateway",
					Latitude:    1.1234,
					Longitude:   1.1235,
					Altitude:    5.5,
				})
			})
		})

		Convey("When calling Delete", func() {
			_, err := api.Delete(ctx, &pb.DeleteGatewayRequest{
				Mac: "0102030405060708",
			})
			So(err, ShouldBeNil)
			So(validator.ctx, ShouldResemble, ctx)
			So(validator.validatorFuncs, ShouldHaveLength, 1)

			Convey("Then the correct request was forwarded to the network-server api", func() {
				So(nsClient.DeleteGatewayChan, ShouldHaveLength, 1)
				So(<-nsClient.DeleteGatewayChan, ShouldResemble, ns.DeleteGatewayRequest{
					Mac: []byte{1, 2, 3, 4, 5, 6, 7, 8},
				})
			})
		})

		Convey("When calling GetStats", func() {
			now := time.Now().UTC()

			nsClient.GetGatewayStatsResponse = ns.GetGatewayStatsResponse{
				Result: []*ns.GatewayStats{
					{
						Timestamp:           now.Format(time.RFC3339Nano),
						RxPacketsReceived:   10,
						RxPacketsReceivedOK: 9,
						TxPacketsReceived:   8,
						TxPacketsEmitted:    7,
					},
				},
			}

			resp, err := api.GetStats(ctx, &pb.GetGatewayStatsRequest{
				Mac:            "0102030405060708",
				Interval:       "DAY",
				StartTimestamp: now.Format(time.RFC3339Nano),
				EndTimestamp:   now.Add(24 * time.Hour).Format(time.RFC3339Nano),
			})
			So(err, ShouldBeNil)
			So(validator.ctx, ShouldResemble, ctx)
			So(validator.validatorFuncs, ShouldHaveLength, 1)

			Convey("Then the correct request / response was forwarded to / from the network-server api", func() {
				So(nsClient.GetGatewayStatsChan, ShouldHaveLength, 1)
				So(<-nsClient.GetGatewayStatsChan, ShouldResemble, ns.GetGatewayStatsRequest{
					Mac:            []byte{1, 2, 3, 4, 5, 6, 7, 8},
					Interval:       ns.AggregationInterval_DAY,
					StartTimestamp: now.Format(time.RFC3339Nano),
					EndTimestamp:   now.Add(24 * time.Hour).Format(time.RFC3339Nano),
				})

				So(resp, ShouldResemble, &pb.GetGatewayStatsResponse{
					Result: []*pb.GatewayStats{
						{
							Timestamp:           now.Format(time.RFC3339Nano),
							RxPacketsReceived:   10,
							RxPacketsReceivedOK: 9,
							TxPacketsReceived:   8,
							TxPacketsEmitted:    7,
						},
					},
				})
			})

		})

	})
}
