package api

import (
	"testing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

	"github.com/brocaar/loraserver/api/ns"

	. "github.com/smartystreets/goconvey/convey"
	"golang.org/x/net/context"

	pb "github.com/gusseleet/lora-app-server/api"
	"github.com/gusseleet/lora-app-server/internal/config"
	"github.com/gusseleet/lora-app-server/internal/storage"
	"github.com/gusseleet/lora-app-server/internal/test"
)

func TestServiceProfileServiceAPI(t *testing.T) {
	conf := test.GetConfig()
	db, err := storage.OpenDatabase(conf.PostgresDSN)
	if err != nil {
		t.Fatal(err)
	}

	config.C.PostgreSQL.DB = db

	Convey("Given a clean database and api instance", t, func() {
		test.MustResetDB(config.C.PostgreSQL.DB)

		nsClient := test.NewNetworkServerClient()
		config.C.NetworkServer.Pool = test.NewNetworkServerPool(nsClient)

		ctx := context.Background()
		validator := &TestValidator{}
		api := NewServiceProfileServiceAPI(validator)

		n := storage.NetworkServer{
			Name:   "test-ns",
			Server: "test-ns:1234",
		}
		So(storage.CreateNetworkServer(config.C.PostgreSQL.DB, &n), ShouldBeNil)

		org := storage.Organization{
			Name: "test-org",
		}
		So(storage.CreateOrganization(config.C.PostgreSQL.DB, &org), ShouldBeNil)

		Convey("Then Create creates a service-profile", func() {
			createReq := pb.CreateServiceProfileRequest{
				Name:            "test-sp",
				OrganizationID:  org.ID,
				NetworkServerID: n.ID,
				ServiceProfile: &pb.ServiceProfile{
					UlRate:                 100,
					UlBucketSize:           10,
					UlRatePolicy:           pb.RatePolicy_MARK,
					DlRate:                 200,
					DlBucketSize:           20,
					DlRatePolicy:           pb.RatePolicy_DROP,
					AddGWMetadata:          true,
					DevStatusReqFreq:       4,
					ReportDevStatusBattery: true,
					ReportDevStatusMargin:  true,
					DrMin:          3,
					DrMax:          5,
					PrAllowed:      true,
					HrAllowed:      true,
					RaAllowed:      true,
					NwkGeoLoc:      true,
					TargetPER:      10,
					MinGWDiversity: 3,
				},
			}

			createResp, err := api.Create(ctx, &createReq)
			So(err, ShouldBeNil)
			So(createResp.ServiceProfileID, ShouldNotEqual, "")
			So(nsClient.CreateServiceProfileChan, ShouldHaveLength, 1)

			// set network-server mock
			nsCreate := <-nsClient.CreateServiceProfileChan
			nsClient.GetServiceProfileResponse = ns.GetServiceProfileResponse{
				ServiceProfile: nsCreate.ServiceProfile,
			}

			Convey("Then Get returns the service-profile", func() {
				getResp, err := api.Get(ctx, &pb.GetServiceProfileRequest{
					ServiceProfileID: createResp.ServiceProfileID,
				})
				So(err, ShouldBeNil)
				So(getResp.Name, ShouldEqual, createReq.Name)
				So(getResp.OrganizationID, ShouldEqual, createReq.OrganizationID)
				So(getResp.NetworkServerID, ShouldEqual, createReq.NetworkServerID)
				So(getResp.ServiceProfile, ShouldResemble, &pb.ServiceProfile{
					ServiceProfileID:       createResp.ServiceProfileID,
					UlRate:                 100,
					UlBucketSize:           10,
					UlRatePolicy:           pb.RatePolicy_MARK,
					DlRate:                 200,
					DlBucketSize:           20,
					DlRatePolicy:           pb.RatePolicy_DROP,
					AddGWMetadata:          true,
					DevStatusReqFreq:       4,
					ReportDevStatusBattery: true,
					ReportDevStatusMargin:  true,
					DrMin:          3,
					DrMax:          5,
					PrAllowed:      true,
					HrAllowed:      true,
					RaAllowed:      true,
					NwkGeoLoc:      true,
					TargetPER:      10,
					MinGWDiversity: 3,
				})
			})

			Convey("Then Update updates the service-profile", func() {
				_, err := api.Update(ctx, &pb.UpdateServiceProfileRequest{
					Name: "updated-sp",
					ServiceProfile: &pb.ServiceProfile{
						ServiceProfileID:       createResp.ServiceProfileID,
						UlRate:                 200,
						UlBucketSize:           20,
						UlRatePolicy:           pb.RatePolicy_DROP,
						DlRate:                 300,
						DlBucketSize:           30,
						DlRatePolicy:           pb.RatePolicy_MARK,
						AddGWMetadata:          true,
						DevStatusReqFreq:       5,
						ReportDevStatusBattery: true,
						ReportDevStatusMargin:  true,
						DrMin:          2,
						DrMax:          4,
						PrAllowed:      true,
						HrAllowed:      true,
						RaAllowed:      true,
						NwkGeoLoc:      true,
						TargetPER:      20,
						MinGWDiversity: 4,
					},
				})
				So(err, ShouldBeNil)
				So(nsClient.UpdateServiceProfileChan, ShouldHaveLength, 1)

				nsUpdate := <-nsClient.UpdateServiceProfileChan
				nsClient.GetServiceProfileResponse = ns.GetServiceProfileResponse{
					ServiceProfile: nsUpdate.ServiceProfile,
				}

				getResp, err := api.Get(ctx, &pb.GetServiceProfileRequest{
					ServiceProfileID: createResp.ServiceProfileID,
				})
				So(err, ShouldBeNil)
				So(getResp.Name, ShouldEqual, "updated-sp")
				So(getResp.OrganizationID, ShouldEqual, org.ID)
				So(getResp.NetworkServerID, ShouldEqual, n.ID)
				So(getResp.ServiceProfile, ShouldResemble, &pb.ServiceProfile{
					ServiceProfileID:       createResp.ServiceProfileID,
					UlRate:                 200,
					UlBucketSize:           20,
					UlRatePolicy:           pb.RatePolicy_DROP,
					DlRate:                 300,
					DlBucketSize:           30,
					DlRatePolicy:           pb.RatePolicy_MARK,
					AddGWMetadata:          true,
					DevStatusReqFreq:       5,
					ReportDevStatusBattery: true,
					ReportDevStatusMargin:  true,
					DrMin:          2,
					DrMax:          4,
					PrAllowed:      true,
					HrAllowed:      true,
					RaAllowed:      true,
					NwkGeoLoc:      true,
					TargetPER:      20,
					MinGWDiversity: 4,
				})
			})

			Convey("Then Delete deletes the service-profile", func() {
				_, err := api.Delete(ctx, &pb.DeleteServiceProfileRequest{
					ServiceProfileID: createResp.ServiceProfileID,
				})
				So(err, ShouldBeNil)

				_, err = api.Get(ctx, &pb.GetServiceProfileRequest{
					ServiceProfileID: createResp.ServiceProfileID,
				})
				So(err, ShouldNotBeNil)
				So(grpc.Code(err), ShouldEqual, codes.NotFound)
			})

			Convey("Given a global admin user", func() {
				validator.returnIsAdmin = true

				Convey("Then List without organization id returns all service-profiles", func() {
					listResp, err := api.List(ctx, &pb.ListServiceProfileRequest{
						Limit: 10,
					})
					So(err, ShouldBeNil)
					So(listResp.TotalCount, ShouldEqual, 1)
					So(listResp.Result, ShouldHaveLength, 1)
				})
			})

			Convey("GIven an organization user", func() {
				userID, err := storage.CreateUser(config.C.PostgreSQL.DB, &storage.User{
					Username: "testuser",
					IsActive: true,
					Email:    "foo@bar.com",
				}, "testpassword")
				So(err, ShouldBeNil)
				So(storage.CreateOrganizationUser(config.C.PostgreSQL.DB, org.ID, userID, false), ShouldBeNil)

				Convey("Then List without organization id returns all service-profiles related to the user", func() {
					validator.returnUsername = "testuser"
					listResp, err := api.List(ctx, &pb.ListServiceProfileRequest{
						Limit: 10,
					})
					So(err, ShouldBeNil)
					So(listResp.TotalCount, ShouldEqual, 1)
					So(listResp.Result, ShouldHaveLength, 1)
				})

				Convey("Then calling List using a different username returns no items", func() {
					validator.returnUsername = "differentuser"
					listResp, err := api.List(ctx, &pb.ListServiceProfileRequest{
						Limit: 10,
					})
					So(err, ShouldBeNil)
					So(listResp.TotalCount, ShouldEqual, 0)
					So(listResp.Result, ShouldHaveLength, 0)
				})
			})

			Convey("Then List returns the service-profiles for the given organization id", func() {
				listResp, err := api.List(ctx, &pb.ListServiceProfileRequest{
					Limit:          10,
					OrganizationID: org.ID,
				})
				So(err, ShouldBeNil)
				So(listResp.TotalCount, ShouldEqual, 1)
				So(listResp.Result, ShouldHaveLength, 1)

				listResp, err = api.List(ctx, &pb.ListServiceProfileRequest{
					Limit:          10,
					OrganizationID: org.ID + 1,
				})
				So(err, ShouldBeNil)
				So(listResp.TotalCount, ShouldEqual, 0)
				So(listResp.Result, ShouldHaveLength, 0)
			})
		})
	})
}
