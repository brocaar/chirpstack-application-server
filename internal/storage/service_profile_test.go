package storage

import (
	"testing"
	"time"

	"github.com/gusseleet/lora-app-server/internal/config"
	"github.com/gusseleet/lora-app-server/internal/test"
	"github.com/brocaar/loraserver/api/ns"
	"github.com/brocaar/lorawan/backend"
	. "github.com/smartystreets/goconvey/convey"
)

func TestServiceProfile(t *testing.T) {
	conf := test.GetConfig()
	db, err := OpenDatabase(conf.PostgresDSN)
	if err != nil {
		t.Fatal(err)
	}
	config.C.PostgreSQL.DB = db
	nsClient := test.NewNetworkServerClient()
	config.C.NetworkServer.Pool = test.NewNetworkServerPool(nsClient)

	Convey("Given a clean database with organization and network-server", t, func() {
		test.MustResetDB(config.C.PostgreSQL.DB)

		org := Organization{
			Name: "test-org",
		}
		So(CreateOrganization(config.C.PostgreSQL.DB, &org), ShouldBeNil)

		u := User{
			Username: "testuser",
			IsAdmin:  false,
			IsActive: true,
			Email:    "foo@bar.com",
		}
		uID, err := CreateUser(config.C.PostgreSQL.DB, &u, "testpassword")
		So(err, ShouldBeNil)
		So(CreateOrganizationUser(config.C.PostgreSQL.DB, org.ID, uID, false), ShouldBeNil)

		n := NetworkServer{
			Name:   "test-ns",
			Server: "test-ns:1234",
		}
		So(CreateNetworkServer(config.C.PostgreSQL.DB, &n), ShouldBeNil)

		Convey("Then CreateServiceProfile creates the service-profile", func() {
			sp := ServiceProfile{
				OrganizationID:  org.ID,
				NetworkServerID: n.ID,
				Name:            "test-service-profile",
				ServiceProfile: backend.ServiceProfile{
					ULRate:                 100,
					ULBucketSize:           10,
					ULRatePolicy:           backend.Mark,
					DLRate:                 200,
					DLBucketSize:           20,
					DLRatePolicy:           backend.Drop,
					AddGWMetadata:          true,
					DevStatusReqFreq:       4,
					ReportDevStatusBattery: true,
					ReportDevStatusMargin:  true,
					DRMin:          3,
					DRMax:          5,
					PRAllowed:      true,
					HRAllowed:      true,
					RAAllowed:      true,
					NwkGeoLoc:      true,
					TargetPER:      10,
					MinGWDiversity: 3,
				},
			}
			So(CreateServiceProfile(config.C.PostgreSQL.DB, &sp), ShouldBeNil)
			So(nsClient.CreateServiceProfileChan, ShouldHaveLength, 1)
			So(<-nsClient.CreateServiceProfileChan, ShouldResemble, ns.CreateServiceProfileRequest{
				ServiceProfile: &ns.ServiceProfile{
					ServiceProfileID:       sp.ServiceProfile.ServiceProfileID,
					UlRate:                 100,
					UlBucketSize:           10,
					UlRatePolicy:           ns.RatePolicy_MARK,
					DlRate:                 200,
					DlBucketSize:           20,
					DlRatePolicy:           ns.RatePolicy_DROP,
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
			})
			sp.CreatedAt = sp.CreatedAt.UTC().Truncate(time.Millisecond)
			sp.UpdatedAt = sp.UpdatedAt.UTC().Truncate(time.Millisecond)

			Convey("Then GetServiceProfile returns the service-profile", func() {
				nsClient.GetServiceProfileResponse = ns.GetServiceProfileResponse{
					ServiceProfile: &ns.ServiceProfile{
						ServiceProfileID:       sp.ServiceProfile.ServiceProfileID,
						UlRate:                 100,
						UlBucketSize:           10,
						UlRatePolicy:           ns.RatePolicy_MARK,
						DlRate:                 200,
						DlBucketSize:           20,
						DlRatePolicy:           ns.RatePolicy_DROP,
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

				spGet, err := GetServiceProfile(config.C.PostgreSQL.DB, sp.ServiceProfile.ServiceProfileID)
				So(err, ShouldBeNil)
				spGet.CreatedAt = spGet.CreatedAt.UTC().Truncate(time.Millisecond)
				spGet.UpdatedAt = spGet.UpdatedAt.UTC().Truncate(time.Millisecond)

				So(spGet, ShouldResemble, sp)
			})

			Convey("Then UpdateServiceProfile updates the service-profile", func() {
				sp.Name = "updated-service-profile"
				sp.ServiceProfile = backend.ServiceProfile{
					ServiceProfileID:       sp.ServiceProfile.ServiceProfileID,
					ULRate:                 101,
					ULBucketSize:           11,
					ULRatePolicy:           backend.Drop,
					DLRate:                 201,
					DLBucketSize:           21,
					DLRatePolicy:           backend.Mark,
					AddGWMetadata:          true,
					DevStatusReqFreq:       5,
					ReportDevStatusBattery: true,
					ReportDevStatusMargin:  true,
					DRMin:          4,
					DRMax:          6,
					PRAllowed:      true,
					HRAllowed:      true,
					RAAllowed:      true,
					NwkGeoLoc:      true,
					TargetPER:      11,
					MinGWDiversity: 4,
				}
				So(UpdateServiceProfile(config.C.PostgreSQL.DB, &sp), ShouldBeNil)
				sp.UpdatedAt = sp.UpdatedAt.UTC().Truncate(time.Millisecond)
				So(nsClient.UpdateServiceProfileChan, ShouldHaveLength, 1)
				So(<-nsClient.UpdateServiceProfileChan, ShouldResemble, ns.UpdateServiceProfileRequest{
					ServiceProfile: &ns.ServiceProfile{
						ServiceProfileID:       sp.ServiceProfile.ServiceProfileID,
						UlRate:                 101,
						UlBucketSize:           11,
						UlRatePolicy:           ns.RatePolicy_DROP,
						DlRate:                 201,
						DlBucketSize:           21,
						DlRatePolicy:           ns.RatePolicy_MARK,
						AddGWMetadata:          true,
						DevStatusReqFreq:       5,
						ReportDevStatusBattery: true,
						ReportDevStatusMargin:  true,
						DrMin:          4,
						DrMax:          6,
						PrAllowed:      true,
						HrAllowed:      true,
						RaAllowed:      true,
						NwkGeoLoc:      true,
						TargetPER:      11,
						MinGWDiversity: 4,
					},
				})

				spGet, err := GetServiceProfile(config.C.PostgreSQL.DB, sp.ServiceProfile.ServiceProfileID)
				So(err, ShouldBeNil)
				spGet.UpdatedAt = spGet.UpdatedAt.UTC().Truncate(time.Millisecond)
				So(spGet.Name, ShouldEqual, "updated-service-profile")
				So(spGet.UpdatedAt, ShouldResemble, sp.UpdatedAt)
			})

			Convey("Then DeleteServiceProfile deletes the service-profile", func() {
				So(DeleteServiceProfile(config.C.PostgreSQL.DB, sp.ServiceProfile.ServiceProfileID), ShouldBeNil)
				So(nsClient.DeleteServiceProfileChan, ShouldHaveLength, 1)
				So(<-nsClient.DeleteServiceProfileChan, ShouldResemble, ns.DeleteServiceProfileRequest{
					ServiceProfileID: sp.ServiceProfile.ServiceProfileID,
				})

				_, err := GetServiceProfile(config.C.PostgreSQL.DB, sp.ServiceProfile.ServiceProfileID)
				So(err, ShouldEqual, ErrDoesNotExist)
			})

			Convey("Then GetServiceProfileCount returns 1", func() {
				count, err := GetServiceProfileCount(config.C.PostgreSQL.DB)
				So(err, ShouldBeNil)
				So(count, ShouldEqual, 1)
			})

			Convey("Then GetServiceProfileCountForOrganizationID returns the number of service-profiles for the given organization", func() {
				count, err := GetServiceProfileCountForOrganizationID(config.C.PostgreSQL.DB, org.ID)
				So(err, ShouldBeNil)
				So(count, ShouldEqual, 1)

				count, err = GetServiceProfileCountForOrganizationID(config.C.PostgreSQL.DB, org.ID+1)
				So(err, ShouldBeNil)
				So(count, ShouldEqual, 0)
			})

			Convey("Then GetServiceProfileCountForUser returns the service-profile count accessible by the given user", func() {
				count, err := GetServiceProfileCountForUser(config.C.PostgreSQL.DB, u.Username)
				So(err, ShouldBeNil)
				So(count, ShouldEqual, 1)

				count, err = GetServiceProfileCountForUser(config.C.PostgreSQL.DB, "fakeuser")
				So(err, ShouldBeNil)
				So(count, ShouldEqual, 0)
			})

			Convey("Then GetServiceProfiles includes the created service-profile", func() {
				profiles, err := GetServiceProfiles(config.C.PostgreSQL.DB, 10, 0)
				So(err, ShouldBeNil)
				So(profiles, ShouldHaveLength, 1)
				So(profiles[0].Name, ShouldEqual, sp.Name)
			})

			Convey("Then GetServiceProfilesForOrganizationID returns the service-profiles for the given organization", func() {
				sps, err := GetServiceProfilesForOrganizationID(config.C.PostgreSQL.DB, org.ID, 10, 0)
				So(err, ShouldBeNil)
				So(sps, ShouldHaveLength, 1)

				sps, err = GetServiceProfilesForOrganizationID(config.C.PostgreSQL.DB, org.ID+1, 10, 0)
				So(err, ShouldBeNil)
				So(sps, ShouldHaveLength, 0)
			})

			Convey("Then GetServiceProfilesForUser returns the service-profiles accessible by a given user", func() {
				sps, err := GetServiceProfilesForUser(config.C.PostgreSQL.DB, u.Username, 10, 0)
				So(err, ShouldBeNil)
				So(sps, ShouldHaveLength, 1)

				sps, err = GetServiceProfilesForUser(config.C.PostgreSQL.DB, "fakeuser", 10, 0)
				So(err, ShouldBeNil)
				So(sps, ShouldHaveLength, 0)
			})
		})
	})
}
