//go:generate protoc -I=/protobuf/src -I=/tmp/chirpstack-api/protobuf -I=. --go_out=. frame_rx_info.proto

package loracloud
