package gwmigrate

import (
	"context"
	"time"

	log "github.com/Sirupsen/logrus"

	"github.com/brocaar/lora-app-server/internal/common"
	"github.com/brocaar/lora-app-server/internal/storage"
	"github.com/brocaar/loraserver/api/ns"
	"github.com/brocaar/lorawan"
	"github.com/pkg/errors"
)

// MigrateGateways is a temp. function to migrate gateway info from LoRa Server
// to LoRa App Server.
func MigrateGateways() error {
	count, err := storage.GetGatewayCount(common.DB)
	if err != nil {
		return errors.Wrap(err, "get gateway count error")
	}

	// don't migrate when there are already gateways present
	if count != 0 {
		return nil
	}

	orgs, err := storage.GetOrganizations(common.DB, 2, 0, "")
	if err != nil {
		return errors.Wrap(err, "get organizations error")
	}

	// don't migrate when the user has already created multiple organizations
	if len(orgs) != 1 {
		return nil
	}

	log.Info("gwmigrate: migrating gateway data from LoRa Server")
	for {
		var limit int32 = 100
		var offset int32 = 0

		gwRes, err := common.NetworkServer.ListGateways(context.Background(), &ns.ListGatewayRequest{
			Limit:  limit,
			Offset: offset,
		})
		if err != nil {
			log.Errorf("gwmigrate: list gateways error (will retry): %s", err)
			time.Sleep(2 * time.Second)
			continue
		}

		for _, gw := range gwRes.Result {
			var mac lorawan.EUI64
			copy(mac[:], gw.Mac)

			if gw.Name == "" {
				gw.Name = mac.String()
			}

			if gw.Description == "" {
				gw.Description = mac.String()
			}

			err = storage.CreateGateway(common.DB, &storage.Gateway{
				MAC:            mac,
				Name:           gw.Name,
				Description:    gw.Description,
				OrganizationID: orgs[0].ID,
			})
			if err != nil {
				return errors.Wrap(err, "create gateway error")
			}
		}

		offset += limit
		if offset > gwRes.TotalCount-1 {
			break
		}
	}

	return nil
}
