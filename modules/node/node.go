package node

import (
	"encoding/json"
	"fmt"
	"math/big"
	"net"
	"sync"
	"time"

	pec256 "github.com/polarysfoundation/pec-256"
	"github.com/polarysfoundation/polarys-chain/modules/common"
	"github.com/polarysfoundation/polarys-chain/modules/core/block"
	"github.com/polarysfoundation/polarys-chain/modules/crypto"
	"github.com/polarysfoundation/polarys-chain/modules/p2p"
	polarysdb "github.com/polarysfoundation/polarys_db"
	"github.com/sirupsen/logrus"
)

type Chain interface {
	AddRemoteBlock(block *block.Block) error
	HasBlock(hash common.Hash) bool
	GetBlockByHash(hash common.Hash) (*block.Block, error)
}

const (
	version = uint32(0x00000001)
)

type Node struct {
	self    *p2p.Peer
	peers   map[string]*p2p.Peer
	privKey pec256.PrivKey
	pubKey  pec256.PubKey

	trustedPeers map[string]bool

	bc Chain

	db  *polarysdb.Database
	log *logrus.Logger
	mu  sync.RWMutex
}

func NewNode(db *polarysdb.Database, log *logrus.Logger, bc Chain) *Node {
	priv, pub := crypto.GenerateKey()

	addr := &net.UDPAddr{
		IP:   net.ParseIP("0.0.0.0"),
		Port: 5865,
	}

	self := p2p.NewPeer(addr, version, pub, uint64(time.Now().Unix()))

	return &Node{
		self:         self,
		peers:        make(map[string]*p2p.Peer),
		trustedPeers: make(map[string]bool),
		privKey:      priv,
		pubKey:       pub,
		db:           db,
		log:          log,
		bc:           bc,
	}
}

func (nd *Node) Run() {
	conn, err := net.ListenUDP("udp", nd.self.Addr())
	if err != nil {
		nd.log.Fatal(err)
	}
	defer conn.Close()

	nd.log.WithField("client_id", common.EncodeToCXID(nd.self.ID())).Info("Node started")
	nd.log.WithField("client_id", common.EncodeToCXID(nd.self.ID())).Infof("Listening on: %s", nd.self.Addr().String())

	go nd.ping(conn)

	buf := make([]byte, 1024)

	for {
		n, addr, err := conn.ReadFromUDP(buf)
		if err != nil {
			nd.log.WithField("client_id", addr.String()).Error(err)
			continue
		}

		message := buf[:n]

		msg := &Message{}
		err = msg.Unmarshal(message)
		if err != nil {
			nd.log.WithField("client_id", addr.String()).Error(err)
			continue
		}

		nd.handleMessage(msg, addr, conn)
	}
}

func (n *Node) handleMessage(msg *Message, addr *net.UDPAddr, conn *net.UDPConn) {

	pubkey, err := msg.DecodePubKey()
	if err != nil {
		n.log.WithField("client_addr", addr.String()).Error(err)
		return
	}

	n.mu.Lock()
	id := crypto.Pm256(pubkey.Bytes())
	cxid := common.EncodeToCXID(id)

	if _, ok := n.peers[cxid]; !ok {
		peer := p2p.NewPeer(addr, version, pubkey, uint64(time.Now().Unix()))
		n.peers[cxid] = peer
	}
	n.mu.Unlock()

	n.log.WithField("client_id", cxid).Info("Message received")

	switch msg.Type {
	case BLOCK:
		ok, err := n.verifyMessage(msg)
		if err != nil {
			n.log.WithField("client_id", cxid).Error(err)
			return
		}

		if !ok {
			n.log.WithField("client_id", cxid).Error("Invalid signature")
			return
		}

		data, err := msg.DecodeData()
		if err != nil {
			n.log.WithField("client_id", cxid).Error(err)
			return
		}

		var blk block.Block
		err = json.Unmarshal(data, &blk)
		if err != nil {
			n.log.WithField("client_id", cxid).Error(err)
			return
		}

		err = n.bc.AddRemoteBlock(&blk)
		if err != nil {
			n.log.WithField("client_id", cxid).Error(err)
			return
		}

		newMessage := NewMessage(HASH, blk.Hash().Bytes(), n.pubKey)
		newMessage, err = n.signMessage(newMessage)
		if err != nil {
			n.log.WithField("client_id", cxid).Error(err)
			return
		}

		n.broadcast(newMessage, n.self, conn)
	case HASH:
		data, err := msg.DecodeData()
		if err != nil {
			n.log.WithField("client_id", cxid).Error(err)
			return
		}

		hashBlock := common.BytesToHash(data)
		if !n.bc.HasBlock(hashBlock) {
			newMessage := NewMessage(ASK, data, n.pubKey)
			newMessage, err = n.signMessage(newMessage)
			if err != nil {
				n.log.WithField("client_id", cxid).Error(err)
				return
			}

			n.response(newMessage, n.peers[cxid], conn)
		}
	case ASK:
		data, err := msg.DecodeData()
		if err != nil {
			n.log.WithField("client_id", cxid).Error(err)
			return
		}

		hashBlock := common.BytesToHash(data)
		if n.bc.HasBlock(hashBlock) {
			blk, err := n.bc.GetBlockByHash(hashBlock)
			if err != nil {
				n.log.WithField("client_id", cxid).Error(err)
				return
			}

			b, err := json.Marshal(blk)
			if err != nil {
				n.log.WithField("client_id", cxid).Error(err)
				return
			}

			newMessage := NewMessage(BLOCK, b, n.pubKey)
			newMessage, err = n.signMessage(newMessage)
			if err != nil {
				n.log.WithField("client_id", cxid).Error(err)
				return
			}

			n.response(newMessage, n.peers[cxid], conn)
		}
	}
}

func (n *Node) GetID() []byte {
	n.mu.RLock()
	defer n.mu.RUnlock()

	return n.self.ID()
}

func (n *Node) GetAddr() *net.UDPAddr {
	n.mu.RLock()
	defer n.mu.RUnlock()

	return n.self.Addr()
}

func (n *Node) GetPeers() map[string]*p2p.Peer {
	n.mu.RLock()
	defer n.mu.RUnlock()

	return n.peers
}

func (n *Node) GetPubKey() pec256.PubKey {
	n.mu.RLock()
	defer n.mu.RUnlock()

	return n.pubKey
}

func (n *Node) AddPeer(peer *p2p.Peer) error {
	n.mu.Lock()
	defer n.mu.Unlock()

	h := crypto.Pm256(peer.PubKey().Bytes())

	if common.Equal(peer.ID(), h) {
		return fmt.Errorf("peer id mismatch")
	}

	cxid := common.EncodeToCXID(peer.ID())

	if _, ok := n.peers[cxid]; ok {
		return fmt.Errorf("peer already exists")
	}

	n.peers[cxid] = peer

	return nil
}

func (n *Node) RemovePeer(peer *p2p.Peer) error {
	n.mu.Lock()
	defer n.mu.Unlock()

	h := crypto.Pm256(peer.PubKey().Bytes())

	if common.Equal(peer.ID(), h) {
		return fmt.Errorf("peer id mismatch")
	}

	cxid := common.EncodeToCXID(peer.ID())

	if _, ok := n.peers[cxid]; !ok {
		return fmt.Errorf("peer not found")
	}

	delete(n.peers, cxid)

	return nil
}

func (n *Node) GetPeerByID(id []byte) (*p2p.Peer, error) {
	n.mu.RLock()
	defer n.mu.RUnlock()

	h := crypto.Pm256(id)

	if common.Equal(n.self.ID(), h) {
		return nil, fmt.Errorf("peer id mismatch")
	}

	cxid := common.EncodeToCXID(id)

	if peer, ok := n.peers[cxid]; ok {
		return peer, nil
	}

	return nil, fmt.Errorf("peer not found")
}

func (n *Node) response(msg *Message, sender *p2p.Peer, conn *net.UDPConn) {
	_, err := conn.WriteToUDP(msg.Bytes(), sender.Addr())
	if err != nil {
		n.log.WithField("client_id", sender.ID()).Error(err)
		return
	}

	n.log.WithField("client_id", sender.ID()).Info("Message sent")
}

func (n *Node) broadcast(msg *Message, sender *p2p.Peer, conn *net.UDPConn) {
	for _, peer := range n.peers {
		if peer.Addr().String() != sender.Addr().String() {

			_, err := conn.WriteToUDP(msg.Bytes(), peer.Addr())
			if err != nil {
				n.log.WithField("client_id", peer.ID()).Error(err)
				continue
			}

			n.log.WithField("client_id", peer.ID()).Info("Message sent")
		}
	}
}

func (n *Node) signMessage(msg *Message) (*Message, error) {

	b, err := msg.Marshal()
	if err != nil {
		return nil, err
	}

	h := crypto.Pm256(b)

	r, s, err := crypto.Sign(common.BytesToHash(h), n.privKey)
	if err != nil {
		return nil, err
	}

	signature := make([]byte, 64)
	copy(signature[:32], r.Bytes())
	copy(signature[32:], s.Bytes())

	return msg.SignMessage(signature), nil
}

func (n *Node) verifyMessage(msg *Message) (bool, error) {
	b, err := msg.Marshal()
	if err != nil {
		return false, err
	}

	h := crypto.Pm256(b)
	signature := msg.Signature

	pubKey, err := msg.DecodePubKey()
	if err != nil {
		return false, err
	}

	r := new(big.Int).SetBytes(signature[:32])
	s := new(big.Int).SetBytes(signature[32:])

	return crypto.Verify(common.BytesToHash(h), r, s, pubKey)
}

func (n *Node) ping(conn *net.UDPConn) error {
	for {
		time.Sleep(5 * time.Second)

		n.mu.Lock()

		now := time.Now().Unix()
		for id, peer := range n.peers {
			if now-int64(peer.LastSeen()) > 10 {
				n.log.WithField("client_id", id).Info("Client disconnected")
				delete(n.peers, id)
				continue
			}

			ping := fmt.Sprintf("PING|%.2f", float64(now)/1e6)
			_, err := conn.WriteToUDP([]byte(ping), peer.Addr())
			if err != nil {
				n.log.WithField("client_id", id).Error(err)
				continue
			}

			n.log.WithField("client_id", id).Info("Ping sent")
		}

		n.mu.Unlock()
	}
}
