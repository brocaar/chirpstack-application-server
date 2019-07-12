package storage

import (
	"testing"
	"time"

	"github.com/gofrs/uuid"
	. "github.com/smartystreets/goconvey/convey"

	"github.com/brocaar/lora-app-server/internal/backend/networkserver"
	"github.com/brocaar/lora-app-server/internal/backend/networkserver/mock"
	"github.com/brocaar/lora-app-server/internal/test"
	"github.com/brocaar/loraserver/api/ns"
)

func TestServiceProfile(t *testing.T) {
	conf := test.GetConfig()
	if err := Setup(conf); err != nil {
		t.Fatal(err)
	}

	nsClient := mock.NewClient()
	networkserver.SetPool(mock.NewPool(nsClient))

	Convey("Given a clean database with organization and network-server", t, func() {
		test.MustResetDB(DB().DB)

		org := Organization{
			Name: "test-org",
		}
		So(CreateOrganization(DB(), &org), ShouldBeNil)

		u := User{
			Username: "testuser",
			IsAdmin:  false,
			IsActive: true,
			Email:    "foo@bar.com",
		}
		uID, err := CreateUser(DB(), &u, "testpassword")
		So(err, ShouldBeNil)
		So(CreateOrganizationUser(DB(), org.ID, uID, false), ShouldBeNil)

		n := NetworkServer{
			Name:   "test-ns",
			Server: "test-ns:1234",
		}
		So(CreateNetworkServer(DB(), &n), ShouldBeNil)

		Convey("Then CreateServiceProfile creates the service-profile", func() {
			sp := ServiceProfile{
				OrganizationID:  org.ID,
				NetworkServerID: n.ID,
				Name:            "test-service-profile",
				ServiceProfile: ns.ServiceProfile{
					UlRate:                 100,
					UlBucketSize:           10,
					UlRatePolicy:           ns.RatePolicy_MARK,
					DlRate:                 200,
					DlBucketSize:           20,
					DlRatePolicy:           ns.RatePolicy_DROP,
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
			So(CreateServiceProfile(DB(), &sp), ShouldBeNil)
			So(nsClient.CreateServiceProfileChan, ShouldHaveLength, 1)
			So(<-nsClient.CreateServiceProfileChan, ShouldResemble, ns.CreateServiceProfileRequest{
				ServiceProfile: &sp.ServiceProfile,
			})
			sp.CreatedAt = sp.CreatedAt.UTC().Truncate(time.Millisecond)
			sp.UpdatedAt = sp.UpdatedAt.UTC().Truncate(time.Millisecond)
			spID, err := uuid.FromBytes(sp.ServiceProfile.Id)
			So(err, ShouldBeNil)

			Convey("Then GetServiceProfile returns the service-profile", func() {
				nsClient.GetServiceProfileResponse = ns.GetServiceProfileResponse{
					ServiceProfile: &sp.ServiceProfile,
				}

				spGet, err := GetServiceProfile(DB(), spID, false)
				So(err, ShouldBeNil)
				spGet.CreatedAt = spGet.CreatedAt.UTC().Truncate(time.Millisecond)
				spGet.UpdatedAt = spGet.UpdatedAt.UTC().Truncate(time.Millisecond)
				So(spGet, ShouldResemble, sp)
			})

			Convey("Then UpdateServiceProfile updates the service-profile", func() {
				sp.Name = "updated-service-profile"
				sp.ServiceProfile = ns.ServiceProfile{
					Id:                     sp.ServiceProfile.Id,
					UlRate:                 101,
					UlBucketSize:           11,
					UlRatePolicy:           ns.RatePolicy_DROP,
					DlRate:                 201,
					DlBucketSize:           21,
					DlRatePolicy:           ns.RatePolicy_MARK,
					AddGwMetadata:          true,
					DevStatusReqFreq:       5,
					ReportDevStatusBattery: true,
					ReportDevStatusMargin:  true,
					DrMin:                  4,
					DrMax:                  6,
					PrAllowed:              true,
					HrAllowed:              true,
					RaAllowed:              true,
					NwkGeoLoc:              true,
					TargetPer:              11,
					MinGwDiversity:         4,
				}
				So(UpdateServiceProfile(DB(), &sp), ShouldBeNil)
				sp.UpdatedAt = sp.UpdatedAt.UTC().Truncate(time.Millisecond)
				So(nsClient.UpdateServiceProfileChan, ShouldHaveLength, 1)
				So(<-nsClient.UpdateServiceProfileChan, ShouldResemble, ns.UpdateServiceProfileRequest{
					ServiceProfile: &sp.ServiceProfile,
				})

				spGet, err := GetServiceProfile(DB(), spID, false)
				So(err, ShouldBeNil)
				spGet.UpdatedAt = spGet.UpdatedAt.UTC().Truncate(time.Millisecond)
				So(spGet.Name, ShouldEqual, "updated-service-profile")
				So(spGet.UpdatedAt, ShouldResemble, sp.UpdatedAt)
			})

			Convey("Then DeleteServiceProfile deletes the service-profile", func() {
				So(DeleteServiceProfile(DB(), spID), ShouldBeNil)
				So(nsClient.DeleteServiceProfileChan, ShouldHaveLength, 1)
				So(<-nsClient.DeleteServiceProfileChan, ShouldResemble, ns.DeleteServiceProfileRequest{
					Id: sp.ServiceProfile.Id,
				})

				_, err := GetServiceProfile(DB(), spID, false)
				So(err, ShouldEqual, ErrDoesNotExist)
			})

			Convey("Then GetServiceProfileCount returns 1", func() {
				count, err := GetServiceProfileCount(DB())
				So(err, ShouldBeNil)
				So(count, ShouldEqual, 1)
			})

			Convey("Then GetServiceProfileCountForOrganizationID returns the number of service-profiles for the given organization", func() {
				count, err := GetServiceProfileCountForOrganizationID(DB(), org.ID)
				So(err, ShouldBeNil)
				So(count, ShouldEqual, 1)

				count, err = GetServiceProfileCountForOrganizationID(DB(), org.ID+1)
				So(err, ShouldBeNil)
				So(count, ShouldEqual, 0)
			})

			Convey("Then GetServiceProfileCountForUser returns the service-profile count accessible by the given user", func() {
				count, err := GetServiceProfileCountForUser(DB(), u.Username)
				So(err, ShouldBeNil)
				So(count, ShouldEqual, 1)

				count, err = GetServiceProfileCountForUser(DB(), "fakeuser")
				So(err, ShouldBeNil)
				So(count, ShouldEqual, 0)
			})

			Convey("Then GetServiceProfiles includes the created service-profile", func() {
				profiles, err := GetServiceProfiles(DB(), 10, 0)
				So(err, ShouldBeNil)
				So(profiles, ShouldHaveLength, 1)
				So(profiles[0].Name, ShouldEqual, sp.Name)
			})

			Convey("Then GetServiceProfilesForOrganizationID returns the service-profiles for the given organization", func() {
				sps, err := GetServiceProfilesForOrganizationID(DB(), org.ID, 10, 0)
				So(err, ShouldBeNil)
				So(sps, ShouldHaveLength, 1)

				sps, err = GetServiceProfilesForOrganizationID(DB(), org.ID+1, 10, 0)
				So(err, ShouldBeNil)
				So(sps, ShouldHaveLength, 0)
			})

			Convey("Then GetServiceProfilesForUser returns the service-profiles accessible by a given user", func() {
				sps, err := GetServiceProfilesForUser(DB(), u.Username, 10, 0)
				So(err, ShouldBeNil)
				So(sps, ShouldHaveLength, 1)

				sps, err = GetServiceProfilesForUser(DB(), "fakeuser", 10, 0)
				So(err, ShouldBeNil)
				So(sps, ShouldHaveLength, 0)
			})
		})
	})
}
