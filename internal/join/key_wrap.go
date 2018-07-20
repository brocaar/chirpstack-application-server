package join

import (
	"crypto/aes"
	"encoding/hex"
	"fmt"

	keywrap "github.com/NickBall/go-aes-key-wrap"
	"github.com/pkg/errors"

	"github.com/brocaar/lora-app-server/internal/config"
	"github.com/brocaar/lorawan"
	"github.com/brocaar/lorawan/backend"
)

// getNSKeyEnvelope returns the KeyEnvelope for the given NS related key.
// When the join-server configuration has a KEK configured for the given
// NetID, it will wrap the key using this KEK.
func getNSKeyEnvelope(netID lorawan.NetID, key lorawan.AES128Key) (*backend.KeyEnvelope, error) {
	var err error
	var kek []byte

	for i := range config.C.JoinServer.KEK.Set {
		if netID.String() == config.C.JoinServer.KEK.Set[i].Label {
			kek, err = hex.DecodeString(config.C.JoinServer.KEK.Set[i].KEK)
			if err != nil {
				return nil, errors.Wrap(err, "decode kek error")
			}
		}
	}

	if len(kek) == 0 {
		return &backend.KeyEnvelope{
			AESKey: backend.HEXBytes(key[:]),
		}, nil
	}

	block, err := aes.NewCipher(kek)
	if err != nil {
		return nil, errors.Wrap(err, "new cipher error")
	}

	b, err := keywrap.Wrap(block, key[:])
	if err != nil {
		return nil, errors.Wrap(err, "key wrap error")
	}

	return &backend.KeyEnvelope{
		KEKLabel: netID.String(),
		AESKey:   backend.HEXBytes(b),
	}, nil
}

// getASKeyEnvelope returns the KeyEnvelope for the given AS related key.
// When the join-server configuration has a KEK configured for encrypting
// AS related keys, it will wrap the key using this KEK.
func getASKeyEnvelope(key lorawan.AES128Key) (*backend.KeyEnvelope, error) {
	if config.C.JoinServer.KEK.ASKEKLabel == "" {
		return &backend.KeyEnvelope{
			AESKey: backend.HEXBytes(key[:]),
		}, nil
	}

	var err error
	var kek []byte
	for i := range config.C.JoinServer.KEK.Set {
		if config.C.JoinServer.KEK.ASKEKLabel == config.C.JoinServer.KEK.Set[i].Label {
			kek, err = hex.DecodeString(config.C.JoinServer.KEK.Set[i].KEK)
			if err != nil {
				return nil, errors.Wrap(err, "decode kek error")
			}
		}
	}
	if len(kek) == 0 {
		return nil, fmt.Errorf("as kek label not found in set: %s", config.C.JoinServer.KEK.ASKEKLabel)
	}

	block, err := aes.NewCipher(kek)
	if err != nil {
		return nil, errors.Wrap(err, "new cipher error")
	}

	b, err := keywrap.Wrap(block, key[:])
	if err != nil {
		return nil, errors.Wrap(err, "key wrap error")
	}

	return &backend.KeyEnvelope{
		KEKLabel: config.C.JoinServer.KEK.ASKEKLabel,
		AESKey:   backend.HEXBytes(b),
	}, nil
}
