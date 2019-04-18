module github.com/brocaar/lora-app-server

require (
	cloud.google.com/go v0.34.0
	github.com/Azure/azure-service-bus-go v0.2.0
	github.com/NickBall/go-aes-key-wrap v0.0.0-20170929221519-1c3aa3e4dfc5
	github.com/aws/aws-sdk-go v1.17.5
	github.com/brocaar/loraserver v0.0.0-20190411080028-d454003a0cc9
	github.com/brocaar/lorawan v0.0.0-20190308082318-5ed881e0a2d7
	github.com/dgrijalva/jwt-go v3.2.0+incompatible
	github.com/eclipse/paho.mqtt.golang v0.0.0-20190117150808-cb7eb9363b44
	github.com/elazarl/go-bindata-assetfs v1.0.0
	github.com/gofrs/uuid v3.2.0+incompatible
	github.com/golang/protobuf v1.3.1
	github.com/gomodule/redigo v2.0.0+incompatible
	github.com/goreleaser/goreleaser v0.101.0
	github.com/goreleaser/nfpm v0.9.7
	github.com/gorilla/context v1.1.1 // indirect
	github.com/gorilla/mux v1.6.2
	github.com/gorilla/websocket v1.4.0 // indirect
	github.com/grpc-ecosystem/go-grpc-middleware v1.0.0
	github.com/grpc-ecosystem/grpc-gateway v1.7.0
	github.com/jmoiron/sqlx v1.2.0
	github.com/jteeuwen/go-bindata v3.0.8-0.20180305030458-6025e8de665b+incompatible
	github.com/lib/pq v1.0.0
	github.com/mmcloughlin/geohash v0.0.0-20181009053802-f7f2bcae3294
	github.com/pkg/errors v0.8.1
	github.com/robertkrimen/otto v0.0.0-20180617131154-15f95af6e78d
	github.com/rubenv/sql-migrate v0.0.0-20181213081019-5a8808c14925
	github.com/sirupsen/logrus v1.3.0
	github.com/smartystreets/goconvey v0.0.0-20190306220146-200a235640ff
	github.com/spf13/cobra v0.0.3
	github.com/spf13/viper v1.3.1
	github.com/stretchr/testify v1.3.0
	github.com/tmc/grpc-websocket-proxy v0.0.0-20190109142713-0ad062ec5ee5
	golang.org/x/crypto v0.0.0-20190228161510-8dd112bcdc25
	golang.org/x/lint v0.0.0-20181026193005-c67002cb31c3
	golang.org/x/net v0.0.0-20190301231341-16b79f2e4e95
	golang.org/x/tools v0.0.0-20190118193359-16909d206f00
	google.golang.org/api v0.1.0
	google.golang.org/genproto v0.0.0-20190111180523-db91494dd46c
	google.golang.org/grpc v1.18.0
	gopkg.in/sourcemap.v1 v1.0.5 // indirect
)

replace github.com/grpc-ecosystem/grpc-gateway => github.com/brocaar/grpc-gateway v1.7.0-patched
