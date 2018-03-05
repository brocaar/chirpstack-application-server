package profilesmigrate

import (
	"context"
	"database/sql/driver"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"

	"github.com/gusseleet/lora-app-server/internal/config"
	"github.com/gusseleet/lora-app-server/internal/storage"
	"github.com/brocaar/loraserver/api/ns"
	"github.com/brocaar/lorawan"
	"github.com/brocaar/lorawan/backend"
)

// DevNonceList represents a list of dev nonces
type DevNonceList [][2]byte

// Scan implements the sql.Scanner interface.
func (l *DevNonceList) Scan(src interface{}) error {
	if src == nil {
		*l = make([][2]byte, 0)
		return nil
	}

	b, ok := src.([]byte)
	if !ok {
		return fmt.Errorf("src must be of type []byte, got: %T", src)
	}
	if len(b)%2 != 0 {
		return errors.New("the length of src must be a multiple of 2")
	}
	for i := 0; i < len(b); i += 2 {
		*l = append(*l, [2]byte{b[i], b[i+1]})
	}
	return nil
}

// Value implements the driver.Valuer interface.
func (l DevNonceList) Value() (driver.Value, error) {
	b := make([]byte, 0, len(l)/2)
	for _, n := range l {
		b = append(b, n[:]...)
	}
	return b, nil
}

// RXWindow defines the RX window option.
type RXWindow int8

// Available RX window options.
const (
	RX1 = iota
	RX2
)

// Scan implements the sql.Scanner interface.
func (r *RXWindow) Scan(src interface{}) error {
	i, ok := src.(int64)
	if !ok {
		return fmt.Errorf("expected int64, got: %T", src)
	}
	*r = RXWindow(i)
	return nil
}

// Value implements the driver.Valuer interface.
func (r RXWindow) Value() (driver.Value, error) {
	return int64(r), nil
}

// Node contains the information of a node.
type Node struct {
	ApplicationID          int64             `db:"application_id"`
	UseApplicationSettings bool              `db:"use_application_settings"`
	Name                   string            `db:"name"`
	Description            string            `db:"description"`
	DevEUI                 lorawan.EUI64     `db:"dev_eui"`
	AppEUI                 lorawan.EUI64     `db:"app_eui"`
	AppKey                 lorawan.AES128Key `db:"app_key"`
	IsABP                  bool              `db:"is_abp"`
	IsClassC               bool              `db:"is_class_c"`
	DevAddr                lorawan.DevAddr   `db:"dev_addr"`
	NwkSKey                lorawan.AES128Key `db:"nwk_s_key"`
	AppSKey                lorawan.AES128Key `db:"app_s_key"`
	UsedDevNonces          DevNonceList      `db:"used_dev_nonces"`
	RelaxFCnt              bool              `db:"relax_fcnt"`

	RXWindow    RXWindow `db:"rx_window"`
	RXDelay     uint8    `db:"rx_delay"`
	RX1DROffset uint8    `db:"rx1_dr_offset"`
	RX2DR       uint8    `db:"rx2_dr"`

	ADRInterval        uint32  `db:"adr_interval"`
	InstallationMargin float64 `db:"installation_margin"`
}

// StartProfilesMigration starts the profiles migration.
func StartProfilesMigration(nsServer string) error {
	nsCount, err := storage.GetNetworkServerCount(config.C.PostgreSQL.DB)
	if err != nil {
		return errors.Wrap(err, "get network-server count error")
	}

	appCount, err := storage.GetApplicationCount(config.C.PostgreSQL.DB)
	if err != nil {
		return errors.Wrap(err, "get applications count error")
	}

	// skip migration in case there are already network-servers in the database
	// or when no applications exist (clean installation)
	if nsCount != 0 || appCount == 0 {
		return nil
	}

	n := storage.NetworkServer{
		Name:   "LoRa Server",
		Server: nsServer,
	}

	return storage.Transaction(config.C.PostgreSQL.DB, func(tx sqlx.Ext) error {
		err = storage.CreateNetworkServer(tx, &n)
		if err != nil {
			return errors.Wrap(err, "create network-server error")
		}

		err = assignNSToGateways(tx, n)
		if err != nil {
			return errors.Wrap(err, "assign network-server to gateways error")
		}

		err = createServiceProfiles(tx, n)
		if err != nil {
			return errors.Wrap(err, "create service-profiles error")
		}

		return nil
	})
}

func assignNSToGateways(tx sqlx.Execer, n storage.NetworkServer) error {
	_, err := tx.Exec("update gateway set network_server_id = $1", n.ID)
	if err != nil {
		return errors.Wrap(err, "update gateways error")
	}
	return nil
}

func createServiceProfiles(tx sqlx.Ext, n storage.NetworkServer) error {
	var orgs []storage.Organization
	err := sqlx.Select(tx, &orgs, "select * from organization")
	if err != nil {
		return errors.Wrap(err, "get organizations error")
	}

	for _, org := range orgs {
		sp := storage.ServiceProfile{
			NetworkServerID: n.ID,
			OrganizationID:  org.ID,
			Name:            org.Name + " service-profile",
			ServiceProfile: backend.ServiceProfile{
				AddGWMetadata: true,
			},
		}

		if err := storage.CreateServiceProfile(tx, &sp); err != nil {
			return errors.Wrap(err, "create service-profile error")
		}

		err = updateApplicationsForOrganization(tx, n, org, sp)
		if err != nil {
			return errors.Wrap(err, "update applications for organization error")
		}
	}

	return nil
}

func updateApplicationsForOrganization(tx sqlx.Ext, n storage.NetworkServer, org storage.Organization, sp storage.ServiceProfile) error {
	var apps []storage.Application
	err := sqlx.Select(tx, &apps, `
		select
			id,
			name,
			description,
			organization_id,
			coalesce(service_profile_id, '6d5db27e-4ce2-4b2b-b5d7-91f069397978') as service_profile_id
		from application
		where
			organization_id = $1`,
		org.ID,
	)
	if err != nil {
		return errors.Wrap(err, "select applications error")
	}

	for _, app := range apps {
		app.ServiceProfileID = sp.ServiceProfile.ServiceProfileID
		if err = storage.UpdateApplication(tx, app); err != nil {
			return errors.Wrap(err, "update application error")
		}

		if err = migrateDevicesForApplication(tx, n, app); err != nil {
			return errors.Wrap(err, "migrate devices for application error")
		}
	}

	return nil
}

func migrateDevicesForApplication(tx sqlx.Ext, n storage.NetworkServer, app storage.Application) error {
	var nodes []Node
	var appDeviceProfile *storage.DeviceProfile

	err := sqlx.Select(tx, &nodes, "select * from node where application_id = $1", app.ID)
	if err != nil {
		return errors.Wrap(err, "select nodes for application error")
	}

	for _, node := range nodes {
		d := storage.Device{
			DevEUI:        node.DevEUI,
			ApplicationID: node.ApplicationID,
			Name:          node.Name,
			Description:   node.Description,
		}

		// create or assign device-profile
		if node.UseApplicationSettings {
			if appDeviceProfile == nil {
				dp := getDeviceProfileForNode(node, n, app)
				if err := storage.CreateDeviceProfile(tx, &dp); err != nil {
					return errors.Wrap(err, "create device-profile error")
				}
				appDeviceProfile = &dp
			}

			d.DeviceProfileID = appDeviceProfile.DeviceProfile.DeviceProfileID
		} else {
			dp := getDeviceProfileForNode(node, n, app)
			if err := storage.CreateDeviceProfile(tx, &dp); err != nil {
				return errors.Wrap(err, "create device-profile error")
			}
			d.DeviceProfileID = dp.DeviceProfile.DeviceProfileID
		}

		// create device
		if err := storage.CreateDevice(tx, &d); err != nil {
			return errors.Wrap(err, "create device error")
		}

		// create device keys
		dk := storage.DeviceKeys{
			DevEUI: d.DevEUI,
			AppKey: node.AppKey,
		}
		if err = storage.CreateDeviceKeys(tx, &dk); err != nil {
			return errors.Wrap(err, "create device-keys error")
		}

		// create device-activation
		emptyKey := lorawan.AES128Key{}
		if node.NwkSKey == emptyKey && node.AppSKey == emptyKey {
			continue
		}

		da := storage.DeviceActivation{
			DevEUI:  d.DevEUI,
			DevAddr: node.DevAddr,
			AppSKey: node.AppSKey,
			NwkSKey: node.NwkSKey,
		}
		if err = storage.CreateDeviceActivation(tx, &da); err != nil {
			return errors.Wrap(err, "create device-activation error")
		}

		// migrate node-session to device-session
		n, err := storage.GetNetworkServerForDevEUI(tx, d.DevEUI)
		if err != nil {
			return errors.Wrap(err, "get network-server error")
		}

		var nonces [][]byte
		for i := range node.UsedDevNonces {
			nonces = append(nonces, node.UsedDevNonces[i][:])
		}

		nsClient, err := config.C.NetworkServer.Pool.Get(n.Server, []byte(n.CACert), []byte(n.TLSCert), []byte(n.TLSKey))
		if err != nil {
			return errors.Wrap(err, "get network-server client error")
		}

		_, err = nsClient.MigrateNodeToDeviceSession(context.Background(), &ns.MigrateNodeToDeviceSessionRequest{
			DevEUI:    d.DevEUI[:],
			JoinEUI:   node.AppEUI[:],
			DevNonces: nonces,
		})
		if err != nil {
			return errors.Wrap(err, "migrate node-session to device-session error")
		}
	}

	return nil
}

func getDeviceProfileForNode(node Node, n storage.NetworkServer, app storage.Application) storage.DeviceProfile {
	var name string
	if node.UseApplicationSettings {
		name = app.Name + " device-profile"
	} else {
		name = node.DevEUI.String() + " device-profile"
	}

	return storage.DeviceProfile{
		NetworkServerID: n.ID,
		OrganizationID:  app.OrganizationID,
		Name:            name,
		DeviceProfile: backend.DeviceProfile{
			SupportsClassC:    node.IsClassC,
			MACVersion:        "1.0.2",
			RegParamsRevision: "B",
			RXDelay1:          int(node.RXDelay),
			RXDROffset1:       int(node.RX1DROffset),
			RXDataRate2:       int(node.RX2DR),
			SupportsJoin:      !node.IsABP,
			Supports32bitFCnt: true,
		},
	}
}
