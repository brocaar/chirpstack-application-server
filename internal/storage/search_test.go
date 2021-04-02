package storage

import (
	"context"
	"testing"

	"github.com/gofrs/uuid"
	. "github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/require"

	"github.com/brocaar/chirpstack-application-server/internal/backend/networkserver"
	"github.com/brocaar/chirpstack-application-server/internal/backend/networkserver/mock"
	"github.com/brocaar/chirpstack-application-server/internal/test"
	"github.com/brocaar/lorawan"
)

func TestSearch(t *testing.T) {
	conf := test.GetConfig()
	if err := Setup(conf); err != nil {
		t.Fatal(err)
	}

	nsClient := mock.NewClient()
	networkserver.SetPool(mock.NewPool(nsClient))

	Convey("Given a clean database with test-data", t, func() {
		So(MigrateDown(DB().DB), ShouldBeNil)
		So(MigrateUp(DB().DB), ShouldBeNil)

		u := User{
			Email: "test@example.com",
		}
		err := CreateUser(context.Background(), DB(), &u)
		So(err, ShouldBeNil)

		n := NetworkServer{
			Name:   "test-ns",
			Server: "test-ns:1234",
		}
		So(CreateNetworkServer(context.Background(), DB(), &n), ShouldBeNil)

		org := Organization{
			Name: "test-org",
		}
		So(CreateOrganization(context.Background(), DB(), &org), ShouldBeNil)

		sp := ServiceProfile{
			OrganizationID:  org.ID,
			NetworkServerID: n.ID,
			Name:            "test-sp",
		}
		So(CreateServiceProfile(context.Background(), DB(), &sp), ShouldBeNil)
		spID, err := uuid.FromBytes(sp.ServiceProfile.Id)
		So(err, ShouldBeNil)

		dp := DeviceProfile{
			OrganizationID:  org.ID,
			NetworkServerID: n.ID,
			Name:            "test-dp",
		}
		So(CreateDeviceProfile(context.Background(), DB(), &dp), ShouldBeNil)
		dpID, err := uuid.FromBytes(dp.DeviceProfile.Id)
		So(err, ShouldBeNil)

		a := Application{
			Name:             "test-app",
			OrganizationID:   org.ID,
			ServiceProfileID: spID,
		}
		So(CreateApplication(context.Background(), DB(), &a), ShouldBeNil)

		gw := Gateway{
			MAC:             lorawan.EUI64{1, 2, 3, 4, 5, 6, 7, 8},
			Name:            "test-gateway",
			OrganizationID:  org.ID,
			NetworkServerID: n.ID,
		}
		So(CreateGateway(context.Background(), DB(), &gw), ShouldBeNil)

		d := Device{
			DevEUI:          lorawan.EUI64{2, 3, 4, 5, 6, 7, 8, 9},
			Name:            "test-device",
			ApplicationID:   a.ID,
			DeviceProfileID: dpID,
		}
		So(CreateDevice(context.Background(), DB(), &d), ShouldBeNil)

		Convey("When the user is not global admin, this does not return any results", func() {
			queries := []string{
				"test",
				"org",
				"app",
				"010203",
				"020304",
				"device",
			}

			for _, q := range queries {
				res, err := GlobalSearch(context.Background(), DB(), u.ID, false, q, 10, 0)
				So(err, ShouldBeNil)
				So(res, ShouldHaveLength, 0)
			}
		})

		Convey("When the user is global admin, this returns results", func() {
			queries := map[string]int{
				"test":   4,
				"org":    1,
				"app":    1,
				"010203": 1,
				"020304": 2,
				"device": 1,
				"dev":    1,
				"gatew":  1,
			}

			for q, c := range queries {
				res, err := GlobalSearch(context.Background(), DB(), u.ID, true, q, 10, 0)
				So(err, ShouldBeNil)
				So(res, ShouldHaveLength, c)
			}
		})

		Convey("When the user is part of the organization, this returns results", func() {
			So(CreateOrganizationUser(context.Background(), DB(), org.ID, u.ID, false, false, false), ShouldBeNil)

			queries := map[string]int{
				"test":   4,
				"org":    1,
				"app":    1,
				"010203": 1,
				"020304": 2,
				"device": 1,
				"dev":    1,
				"gatew":  1,
			}

			for q, c := range queries {
				res, err := GlobalSearch(context.Background(), DB(), u.ID, false, q, 10, 0)
				So(err, ShouldBeNil)
				So(res, ShouldHaveLength, c)
			}
		})
	})
}

func TestParseSearchQuery(t *testing.T) {
	tests := []struct {
		Input string
		Query string
		Tags  map[string]string
	}{
		{
			Input: "foo bar",
			Query: "foo bar",
		},
		{
			Input: "foo: bar",
			Query: "foo: bar",
		},
		{
			Input: "foo:bar",
			Tags: map[string]string{
				"foo": "bar",
			},
		},
		{
			Input: "foo:bar test",
			Query: "test",
			Tags: map[string]string{
				"foo": "bar",
			},
		},
		{
			Input: "test foo:bar",
			Query: "test",
			Tags: map[string]string{
				"foo": "bar",
			},
		},
		{
			Input: "test foo:bar alice:bob",
			Query: "test",
			Tags: map[string]string{
				"foo":   "bar",
				"alice": "bob",
			},
		},
	}

	assert := require.New(t)

	for _, tst := range tests {
		query, tags := parseSearchQuery(tst.Input)
		assert.Equal(tst.Query, query)
		assert.Equal(tst.Tags, tags)
	}
}
