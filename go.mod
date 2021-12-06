module github.com/brocaar/chirpstack-application-server

go 1.16

require (
	github.com/NickBall/go-aes-key-wrap v0.0.0-20170929221519-1c3aa3e4dfc5
	github.com/aws/aws-sdk-go v1.40.34
	github.com/brocaar/chirpstack-api/go/v3 v3.12.4
	github.com/brocaar/lorawan v0.0.0-20210809075358-95fc1667572e
	github.com/coreos/go-oidc v2.2.1+incompatible
	github.com/eclipse/paho.mqtt.golang v1.3.1
	github.com/go-redis/redis/v8 v8.8.3
	github.com/gofrs/uuid v3.2.0+incompatible
	github.com/golang-jwt/jwt/v4 v4.0.0
	github.com/golang-migrate/migrate/v4 v4.14.1
	github.com/golang/protobuf v1.5.2
	github.com/gopherjs/gopherjs v0.0.0-20190430165422-3e4dfb77656c // indirect
	github.com/goreleaser/goreleaser v1.1.0
	github.com/goreleaser/nfpm v1.10.3
	github.com/gorilla/mux v1.7.4
	github.com/grpc-ecosystem/go-grpc-middleware v1.1.0
	github.com/grpc-ecosystem/go-grpc-prometheus v1.2.0
	github.com/grpc-ecosystem/grpc-gateway v1.16.0
	github.com/jmoiron/sqlx v1.2.1-0.20190826204134-d7d95172beb5
	github.com/lib/pq v1.10.2
	github.com/mmcloughlin/geohash v0.9.0
	github.com/pkg/errors v0.9.1
	github.com/pquerna/cachecontrol v0.0.0-20180517163645-1555304b9b35 // indirect
	github.com/prometheus/client_golang v1.2.1
	github.com/prometheus/client_model v0.0.0-20191202183732-d1d2010b5bee // indirect
	github.com/prometheus/procfs v0.0.8 // indirect
	github.com/robertkrimen/otto v0.0.0-20191217063420-37f8e9a2460c
	github.com/segmentio/kafka-go v0.4.17
	github.com/sirupsen/logrus v1.7.0
	github.com/smartystreets/goconvey v1.6.4
	github.com/spf13/cobra v1.2.1
	github.com/spf13/viper v1.8.1
	github.com/streadway/amqp v0.0.0-20190827072141-edfb9018d271
	github.com/stretchr/testify v1.7.0
	github.com/tmc/grpc-websocket-proxy v0.0.0-20201229170055-e5319fda7802
	golang.org/x/crypto v0.0.0-20210921155107-089bfa567519
	golang.org/x/lint v0.0.0-20210508222113-6edffad5e616
	golang.org/x/net v0.0.0-20211007125505-59d4e928ea9d
	golang.org/x/oauth2 v0.0.0-20210819190943-2bc19b11175f
	golang.org/x/tools v0.1.5
	google.golang.org/grpc v1.40.0
	gopkg.in/sourcemap.v1 v1.0.5 // indirect
	gopkg.in/square/go-jose.v2 v2.5.1 // indirect
)

// replace github.com/brocaar/chirpstack-api/go/v3 => ../chirpstack-api/go
