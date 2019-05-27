// Package blockchain implements an blockchain integration.
package blockchain

import (
	"bufio"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	mrand "math/rand"
	"strings"
	"sync"
	"time"

	"github.com/brocaar/lora-app-server/internal/integration"

	"github.com/davecgh/go-spew/spew"
	libp2p "github.com/libp2p/go-libp2p"
	crypto "github.com/libp2p/go-libp2p-crypto"
	host "github.com/libp2p/go-libp2p-host"
	net "github.com/libp2p/go-libp2p-net"
	peer "github.com/libp2p/go-libp2p-peer"
	pstore "github.com/libp2p/go-libp2p-peerstore"
	ma "github.com/multiformats/go-multiaddr"
)

// Config holds the AWS SNS integration configuration.
type Config struct {
	ListenPort     int    `mapstructure:"ListenPort"`
	DialConnection string `mapstructure:"DialConnectiom"`
	Seed           int64  `mapstructure:"Seed"`
	Difficulty     string `mapstructure:"Difficulty"`
}

type Block struct {
	Index      int
	Timestamp  string
	Data       []byte
	Nonce      int
	Hash       string
	PrevHash   string
	Difficulty string
}

// Integration implements the blockchain integration.
type Integration struct {
	Blockchain []Block
	mutex      sync.Mutex
	Difficulty string
	Channel    chan []byte
}

// New creates a new blockchain integration.
func New(conf Config) (*Integration, error) {
	i := Integration{
		Difficulty: conf.Difficulty,
	}
	go i.Running(conf)
	return &i, nil
}

func (i *Integration) Running(conf Config) {
	go func() {
		listenF := conf.ListenPort
		target := conf.DialConnection
		seed := conf.Seed
		if listenF == 0 {
			log.Fatal("Please provide a port to bind with on -l")
		}

		//Make a host that listens on the multiaddress
		ha, err := makeBasicHost(listenF, seed)
		if err != nil {
			log.Fatal(err)
		}
		if conf.DialConnection == "" {
			log.Println("listening for connections")
			//Set the stream on host A, where we are using a user defined protocol name
			ha.SetStreamHandler("/p2p/1.0.0", i.handleStream)

			select {}
		} else {
			ha.SetStreamHandler("/p2p/1.0.0", i.handleStream)
			ipfsaddr, err := ma.NewMultiaddr(target)
			if err != nil {
				log.Fatalln(err)
			}

			pid, err := ipfsaddr.ValueForProtocol(ma.P_IPFS)
			if err != nil {
				log.Fatalln(err)
			}

			peerid, err := peer.IDB58Decode(pid)
			if err != nil {
				log.Fatalln(err)
			}
			// Decapsulate the /ipfs/<peerID> part from the target
			// /ip4/<a.b.c.d>/ipfs/<peer> becomes /ip4/<a.b.c.d>
			targetPeerAddr, _ := ma.NewMultiaddr(
				fmt.Sprintf("/ipfs/%s", peer.IDB58Encode(peerid)))
			targetAddr := ipfsaddr.Decapsulate(targetPeerAddr)

			// We have a peer ID and a targetAddr so we add it to the peerstore
			// so LibP2P knows how to contact it
			ha.Peerstore().AddAddr(peerid, targetAddr, pstore.PermanentAddrTTL)

			log.Println("opening stream")
			// make a new stream from host B to host A
			// it should be handled on host A by the handler we set above because
			// we use the same /p2p/1.0.0 protocol
			s, err := ha.NewStream(context.Background(), peerid, "/p2p/1.0.0")
			if err != nil {
				log.Fatalln(err)
			}
			// Create a buffered stream so that read and writes are non blocking.
			rw := bufio.NewReadWriter(bufio.NewReader(s), bufio.NewWriter(s))

			// Create a thread to read and write data.
			go i.writeData(rw)
			go i.readData(rw)

			select {} // hang forever
		}
	}()
}

// SendDataUp sends an uplink data payload.
func (i *Integration) SendDataUp(pl integration.DataUpPayload) error {
	data, err := json.Marshal(pl)
	if err != nil {
		log.Fatal(err)
	}
	return i.publish(data)
}

// SendJoinNotification sends a join notification.
func (i *Integration) SendJoinNotification(pl integration.JoinNotification) error {
	data, err := json.Marshal(pl)
	if err != nil {
		log.Fatal(err)
	}
	return i.publish(data)
}

// SendACKNotification sends an ack notification.
func (i *Integration) SendACKNotification(pl integration.ACKNotification) error {
	data, err := json.Marshal(pl)
	if err != nil {
		log.Fatal(err)
	}
	return i.publish(data)
}

// SendErrorNotification sends an error notification.
func (i *Integration) SendErrorNotification(pl integration.ErrorNotification) error {
	data, err := json.Marshal(pl)
	if err != nil {
		log.Fatal(err)
	}
	return i.publish(data)
}

// SendStatusNotification sends a status notification.
func (i *Integration) SendStatusNotification(pl integration.StatusNotification) error {
	data, err := json.Marshal(pl)
	if err != nil {
		log.Fatal(err)
	}
	return i.publish(data)
}

// SendLocationNotification sends a location notification.
func (i *Integration) SendLocationNotification(pl integration.LocationNotification) error {
	data, err := json.Marshal(pl)
	if err != nil {
		log.Fatal(err)
	}
	return i.publish(data)
}

// DataDownChan return nil.
func (i *Integration) DataDownChan() chan integration.DataDownPayload {
	return nil
}

// Close closes the integration.
func (i *Integration) Close() error {
	return nil
}

func calcHash(block Block) string {
	record, err := json.Marshal(block)
	if err != nil {
		log.Fatal(err)
	}
	h := sha256.New()
	h.Write([]byte(record))
	hash := h.Sum(nil)
	return hex.EncodeToString(hash)
}

func (i *Integration) publish(data []byte) error {
	i.Channel <- data
	return nil
}

func generateBlock(i *Integration, prevBlock Block, data []byte) (Block, error) {
	var newBlock Block

	t := time.Now()

	newBlock.Index = prevBlock.Index + 1
	newBlock.Timestamp = t.String()
	newBlock.Data = data
	newBlock.Nonce = 0
	newBlock.PrevHash = prevBlock.Hash
	newBlock.Difficulty = i.Difficulty
	newBlock.Hash = calcHash(newBlock)
	fmt.Println("Generating Proof")
	newBlock = proofOfWork(i, newBlock)
	fmt.Println("Returning Block")
	return newBlock, nil
}

func isValid(i *Integration, prevBlock Block, newBlock Block) bool {

	if prevBlock.Index+1 != newBlock.Index {
		return false
	}

	if prevBlock.Hash != newBlock.PrevHash {
		return false
	}

	if calcHash(newBlock) != newBlock.Hash {
		return false
	}
	if !strings.HasPrefix(newBlock.Hash, i.Difficulty) {
		return false
	}

	return true
}

func replaceChain(i *Integration, newBlocks []Block) {
	if len(newBlocks) > len(i.Blockchain) {
		i.Blockchain = newBlocks
	}
}

func proofOfWork(i *Integration, newBlock Block) Block {
	newBlock.Nonce = 0
	fmt.Println("Working")
	for !(strings.HasPrefix(newBlock.Hash, i.Difficulty)) {
		newBlock.Nonce = newBlock.Nonce + 1
		fmt.Println(fmt.Sprintf("Nonce didn't work, trying nonce value of : %d", newBlock.Nonce))
		newBlock.Hash = calcHash(newBlock)
		fmt.Println(newBlock.Hash)
	}
	return newBlock
}

func makeBasicHost(listenPort int, randseed int64) (host.Host, error) {
	var r io.Reader
	if randseed == 0 {
		r = rand.Reader
	} else {
		r = mrand.New(mrand.NewSource(randseed))
	}

	priv, _, err := crypto.GenerateKeyPairWithReader(crypto.RSA, 2048, r)
	if err != nil {
		return nil, err
	}

	opts := []libp2p.Option{
		libp2p.ListenAddrStrings(fmt.Sprintf("/ip4/127.0.0.1/tcp/%d", listenPort)),
		libp2p.Identity(priv),
	}

	basicHost, err := libp2p.New(context.Background(), opts...)
	if err != nil {
		return nil, err
	}

	hostAddr, _ := ma.NewMultiaddr(fmt.Sprintf("/ipfs/%s", basicHost.ID().Pretty()))

	addr := basicHost.Addrs()[0]
	fulldAddr := addr.Encapsulate(hostAddr)
	log.Printf("Connect %s on port %d to on any other terminal,\n", fulldAddr, listenPort+1)

	return basicHost, nil
}

func run() error {

	return nil
}

func (i *Integration) handleStream(s net.Stream) {
	log.Println("New Stream!")

	rw := bufio.NewReadWriter(bufio.NewReader(s), bufio.NewWriter(s))
	go i.readData(rw)
	go i.writeData(rw)
}

func (i *Integration) readData(rw *bufio.ReadWriter) {

	for {
		str, err := rw.ReadString('\n')
		if err != nil {
			log.Fatal(err)
		}

		if str == "" {
			return
		}
		if str != "\n" {
			chain := make([]Block, 0)
			if err := json.Unmarshal([]byte(str), &chain); err != nil {
				log.Fatal(err)
			}
			i.mutex.Lock()
			// This is where consesnus is to be added
			if len(chain) > len(i.Blockchain) {
				i.Blockchain = chain
				i.Difficulty = chain[len(chain)-1].Difficulty
				bytes, err := json.MarshalIndent(i.Blockchain, "", " ")
				if err != nil {
					log.Fatal(err)
				}
				fmt.Printf("\x1b[32m%s\x1b[0m> ", string(bytes))
			}

			i.mutex.Unlock()
		}
	}
}

func (i *Integration) writeData(rw *bufio.ReadWriter) {

	go func() {
		for {
			time.Sleep(1 * time.Second)
			i.mutex.Lock()
			bytes, err := json.Marshal(i.Blockchain)
			if err != nil {
				log.Println(err)
			}
			i.mutex.Unlock()

			i.mutex.Lock()
			rw.WriteString(fmt.Sprintf("%s\n", string(bytes)))
			rw.Flush()
			i.mutex.Unlock()
		}
	}()

	for {
		data := <-i.Channel
		newBlock, err := generateBlock(i, i.Blockchain[len(i.Blockchain)-1], data)
		if err != nil {
			log.Fatal(err)
		}

		if isValid(i, i.Blockchain[len(i.Blockchain)-1], newBlock) {
			i.mutex.Lock()
			i.Blockchain = append(i.Blockchain, newBlock)
			i.mutex.Unlock()
		}

		bytes, err := json.Marshal(i.Blockchain)
		if err != nil {
			log.Println(err)
		}

		spew.Dump(i.Blockchain)

		i.mutex.Lock()
		rw.WriteString(fmt.Sprintf("%s\n", string(bytes)))
		rw.Flush()
		i.mutex.Unlock()
	}
}
