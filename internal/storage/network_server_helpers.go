package storage

import (
	"github.com/gofrs/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"

	"github.com/brocaar/lora-app-server/internal/backend/networkserver"
	"github.com/brocaar/loraserver/api/ns"
)

func getNSClientForServiceProfile(db sqlx.Queryer, id uuid.UUID) (ns.NetworkServerServiceClient, error) {
	n, err := GetNetworkServerForServiceProfileID(db, id)
	if err != nil {
		return nil, errors.Wrap(err, "get network-server error")
	}

	return getNSClient(n)
}

func getNSClientForMulticastGroup(db sqlx.Queryer, id uuid.UUID) (ns.NetworkServerServiceClient, error) {
	n, err := GetNetworkServerForMulticastGroupID(db, id)
	if err != nil {
		return nil, errors.Wrap(err, "get network-server error")
	}
	return getNSClient(n)
}

func getNSClient(n NetworkServer) (ns.NetworkServerServiceClient, error) {
	return networkserver.GetPool().Get(n.Server, []byte(n.CACert), []byte(n.TLSCert), []byte(n.TLSKey))
}
