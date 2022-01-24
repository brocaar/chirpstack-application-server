package helpers

import (
	"testing"

	"github.com/golang/protobuf/ptypes"
	"github.com/stretchr/testify/require"

	"github.com/brocaar/chirpstack-api/go/v3/gw"
)

func TestHEXBytes(t *testing.T) {
	assert := require.New(t)

	hb1 := HEXBytes{1, 2, 3}
	b, err := hb1.MarshalText()
	assert.NoError(err)
	assert.Equal("010203", string(b))

	var hb2 HEXBytes
	assert.NoError(hb2.UnmarshalText(b))
	assert.Equal(hb1, hb2)
}

func TestGetTimestamp(t *testing.T) {
	assert := require.New(t)
	nowPB := ptypes.TimestampNow()

	rxInfo := []*gw.UplinkRXInfo{
		{
			Time: nil,
		},
		{
			Time: nowPB,
		},
	}

	now, err := ptypes.Timestamp(nowPB)
	assert.NoError(err)
	assert.True(GetTimestamp(rxInfo).Equal(now))
}

func TestEUI64(t *testing.T) {
	assert := require.New(t)

	eui1 := EUI64{1, 2, 3, 4, 5, 6, 7, 8}
	b, err := eui1.MarshalText()
	assert.NoError(err)
	assert.Equal("01-02-03-04-05-06-07-08", string(b))

	var eui2 EUI64
	assert.NoError(eui2.UnmarshalText(b))
	assert.Equal(eui1, eui2)
}
