package external

import (
	"testing"

	uuid "github.com/gofrs/uuid"
	. "github.com/smartystreets/goconvey/convey"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

	pb "github.com/brocaar/lora-app-server/api"
	"github.com/brocaar/lora-app-server/internal/backend/networkserver"
	"github.com/brocaar/lora-app-server/internal/backend/networkserver/mock"
	"github.com/brocaar/lora-app-server/internal/storage"
	"github.com/brocaar/lora-app-server/internal/test"
	"github.com/brocaar/loraserver/api/ns"
)

func TestServiceProfileServiceAPI(t *testing.T) {
	conf := test.GetConfig()
	if err := storage.Setup(conf); err != nil {
		t.Fatal(err)
	}

	Convey("Given a clean database and api instance", t, func() {
		test.MustResetDB(storage.DB().DB)

		nsClient := mock.NewClient()
		networkserver.SetPool(mock.NewPool(nsClient))

		ctx := context.Background()
		validator := &TestValidator{}
		api := NewServiceProfileServiceAPI(validator)

		n := storage.NetworkServer{
			Name:   "test-ns",
			Server: "test-ns:1234",
		}
		So(storage.CreateNetworkServer(storage.DB(), &n), ShouldBeNil)

		org := storage.Organization{
			Name: "test-org",
		}
		So(storage.CreateOrganization(storage.DB(), &org), ShouldBeNil)

		Convey("Then Create creates a service-profile", func() {
			createReq := pb.CreateServiceProfileRequest{
				ServiceProfile: &pb.ServiceProfile{
					Name:                   "test-sp",
					OrganizationId:         org.ID,
					NetworkServerId:        n.ID,
					UlRate:                 100,
					UlBucketSize:           10,
					UlRatePolicy:           pb.RatePolicy_MARK,
					DlRate:                 200,
					DlBucketSize:           20,
					DlRatePolicy:           pb.RatePolicy_DROP,
					AddGwMetadata:          true,
					DevStatusReqFreq:       4,
					ReportDevStatusBattery: true,
					ReportDevStatusMargin:  true,
					DrMin:                  3,
					DrMax:                  5,
					PrAllowed:              true,
					HrAllowed:              true,
					RaAllowed:              true,
					NwkGeoLoc:              true,
					TargetPer:              10,
					MinGwDiversity:         3,
				},
			}

			createResp, err := api.Create(ctx, &createReq)
			So(err, ShouldBeNil)
			So(createResp.Id, ShouldNotEqual, "")
			So(createResp.Id, ShouldNotEqual, uuid.Nil.String())
			So(nsClient.CreateServiceProfileChan, ShouldHaveLength, 1)

			// set network-server mock
			nsCreate := <-nsClient.CreateServiceProfileChan
			nsClient.GetServiceProfileResponse = ns.GetServiceProfileResponse{
				ServiceProfile: nsCreate.ServiceProfile,
			}

			Convey("Then Get returns the service-profile", func() {
				getResp, err := api.Get(ctx, &pb.GetServiceProfileRequest{
					Id: createResp.Id,
				})
				So(err, ShouldBeNil)

				createReq.ServiceProfile.Id = createResp.Id
				So(getResp.ServiceProfile, ShouldResemble, createReq.ServiceProfile)
			})

			Convey("Then Update updates the service-profile", func() {
				updateReq := pb.UpdateServiceProfileRequest{
					ServiceProfile: &pb.ServiceProfile{
						Id:                     createResp.Id,
						Name:                   "updated-sp",
						OrganizationId:         org.ID,
						NetworkServerId:        n.ID,
						UlRate:                 200,
						UlBucketSize:           20,
						UlRatePolicy:           pb.RatePolicy_DROP,
						DlRate:                 300,
						DlBucketSize:           30,
						DlRatePolicy:           pb.RatePolicy_MARK,
						AddGwMetadata:          true,
						DevStatusReqFreq:       5,
						ReportDevStatusBattery: true,
						ReportDevStatusMargin:  true,
						DrMin:                  2,
						DrMax:                  4,
						PrAllowed:              true,
						HrAllowed:              true,
						RaAllowed:              true,
						NwkGeoLoc:              true,
						TargetPer:              20,
						MinGwDiversity:         4,
					},
				}

				_, err := api.Update(ctx, &updateReq)
				So(err, ShouldBeNil)
				So(nsClient.UpdateServiceProfileChan, ShouldHaveLength, 1)

				nsUpdate := <-nsClient.UpdateServiceProfileChan
				nsClient.GetServiceProfileResponse = ns.GetServiceProfileResponse{
					ServiceProfile: nsUpdate.ServiceProfile,
				}

				getResp, err := api.Get(ctx, &pb.GetServiceProfileRequest{
					Id: createResp.Id,
				})
				So(err, ShouldBeNil)
				So(getResp.ServiceProfile, ShouldResemble, updateReq.ServiceProfile)
			})

			Convey("Then Delete deletes the service-profile", func() {
				_, err := api.Delete(ctx, &pb.DeleteServiceProfileRequest{
					Id: createResp.Id,
				})
				So(err, ShouldBeNil)

				_, err = api.Get(ctx, &pb.GetServiceProfileRequest{
					Id: createResp.Id,
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
				userID, err := storage.CreateUser(storage.DB(), &storage.User{
					Username: "testuser",
					IsActive: true,
					Email:    "foo@bar.com",
				}, "testpassword")
				So(err, ShouldBeNil)
				So(storage.CreateOrganizationUser(storage.DB(), org.ID, userID, false), ShouldBeNil)

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
					OrganizationId: org.ID,
				})
				So(err, ShouldBeNil)
				So(listResp.TotalCount, ShouldEqual, 1)
				So(listResp.Result, ShouldHaveLength, 1)

				listResp, err = api.List(ctx, &pb.ListServiceProfileRequest{
					Limit:          10,
					OrganizationId: org.ID + 1,
				})
				So(err, ShouldBeNil)
				So(listResp.TotalCount, ShouldEqual, 0)
				So(listResp.Result, ShouldHaveLength, 0)
			})
		})
	})
}
