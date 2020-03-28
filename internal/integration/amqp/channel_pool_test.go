package amqp

import (
	"testing"

	"github.com/streadway/amqp"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/brocaar/chirpstack-application-server/internal/test"
)

type ChannelPoolTestSuite struct {
	suite.Suite

	conn *amqp.Connection
}

func (ts *ChannelPoolTestSuite) SetupSuite() {
	assert := require.New(ts.T())
	conf := test.GetConfig()

	var err error
	ts.conn, err = amqp.Dial(conf.ApplicationServer.Integration.AMQP.URL)
	assert.NoError(err)
}

func (ts *ChannelPoolTestSuite) TestNew() {
	assert := require.New(ts.T())

	p, err := newPool(10, ts.conn)
	assert.NoError(err)
	defer p.close()
	assert.Len(p.chans, 10)
}

func (ts *ChannelPoolTestSuite) TestGet() {
	assert := require.New(ts.T())

	p, err := newPool(10, ts.conn)
	assert.NoError(err)
	defer p.close()
	assert.Len(p.chans, 10)

	_, err = p.get()
	assert.NoError(err)
	assert.Len(p.chans, 9)

	for i := 0; i < 9; i++ {
		_, err = p.get()
		assert.NoError(err)
	}

	assert.Len(p.chans, 0)

	_, err = p.get()
	assert.NoError(err)
}

func (ts *ChannelPoolTestSuite) TestPut() {
	assert := require.New(ts.T())

	p, err := newPool(10, ts.conn)
	assert.NoError(err)

	chans := make([]*poolChannel, 10)
	for i := 0; i < 10; i++ {
		pc, err := p.get()
		assert.NoError(err)
		chans[i] = pc
	}

	assert.Len(p.chans, 0)

	for _, pc := range chans {
		assert.NoError(pc.close())
	}

	assert.Len(p.chans, 10)

	pc, err := p.get()
	assert.NoError(err)
	p.close()

	assert.Len(p.chans, 0)
	assert.NoError(pc.close())
	assert.Len(p.chans, 0)
}

func (ts *ChannelPoolTestSuite) TestPutUnusable() {
	assert := require.New(ts.T())

	p, err := newPool(10, ts.conn)
	assert.NoError(err)
	defer p.close()

	assert.Len(p.chans, 10)

	pc, err := p.get()
	assert.NoError(err)

	pc.markUnusable()

	assert.NoError(pc.close())

	assert.Len(p.chans, 9)
}

func TestChannelPool(t *testing.T) {
	suite.Run(t, new(ChannelPoolTestSuite))
}
