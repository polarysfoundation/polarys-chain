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
	GetBlockByHeight(height uint64) (*block.Block, error)
	GetLatestBlock() (*block.Block, error)
	ChainID() uint64
	ProtocolHash() common.Hash
}

const (
	version        = uint32(0x00000001)
	readDeadline   = 30 * time.Second
	writeDeadline  = 30 * time.Second
	connectTimeout = 10 * time.Second
)

type Node struct {
	self             *p2p.Peer
	peers            map[string]*p2p.Peer
	peerConnections  map[string]net.Conn // Track TCP connections
	privKey          pec256.PrivKey
	pubKey           pec256.PubKey
	blocksTransmited map[common.Hash]bool
	blocksReceived   map[common.Hash]bool

	trustedPeers map[string]bool

	bc Chain

	db  *polarysdb.Database
	log *logrus.Logger
	mu  sync.RWMutex
}

func NewNode(db *polarysdb.Database, log *logrus.Logger, bc Chain) *Node {
	priv, pub := crypto.GenerateKey()

	addr := &net.TCPAddr{
		IP:   net.ParseIP("0.0.0.0"),
		Port: 5865,
	}

	self := p2p.NewPeer(addr, version, pub, uint64(time.Now().Unix()))

	return &Node{
		self:             self,
		peers:            make(map[string]*p2p.Peer),
		peerConnections:  make(map[string]net.Conn),
		trustedPeers:     make(map[string]bool),
		blocksTransmited: make(map[common.Hash]bool),
		blocksReceived:   make(map[common.Hash]bool),
		privKey:          priv,
		pubKey:           pub,
		db:               db,
		log:              log,
		bc:               bc,
	}
}

func (n *Node) SetPort(port int) {
	n.self.Addr().Port = port
}

func (n *Node) Run() {
	listener, err := net.ListenTCP("tcp", n.self.Addr())
	if err != nil {
		n.log.Fatal(err)
	}
	defer listener.Close()

	n.log.WithField("client_id", common.EncodeToCXID(n.self.ID())).Info("Node started")
	n.log.WithField("client_id", common.EncodeToCXID(n.self.ID())).Infof("Listening on: %s", n.self.Addr().String())

	// Start goroutine to handle incoming connections
	go n.acceptConnections(listener)

	// Start ping and block propagation in separate goroutines
	go n.ping()
	go n.propagateBlock()

	// Block forever
	select {}
}

func (n *Node) acceptConnections(listener *net.TCPListener) {
	for {
		conn, err := listener.AcceptTCP()
		if err != nil {
			n.log.Error("Error accepting connection: ", err)
			continue
		}

		go n.handleConnection(conn)
	}
}

func (nd *Node) handleConnection(conn *net.TCPConn) {
	defer conn.Close()

	// Set read deadline to detect dead connections
	conn.SetReadDeadline(time.Now().Add(readDeadline))

	buf := make([]byte, 1024)
	for {
		n, err := conn.Read(buf)
		if err != nil {
			nd.log.WithField("remote_addr", conn.RemoteAddr().String()).Error("Error reading from connection: ", err)
			return
		}

		message := buf[:n]

		msg := &Message{}
		err = msg.Unmarshal(message)
		if err != nil {
			nd.log.WithField("remote_addr", conn.RemoteAddr().String()).Error(err)
			continue
		}

		// Reset read deadline
		conn.SetReadDeadline(time.Now().Add(readDeadline))

		// Handle the message
		nd.handleMessage(msg, conn)
	}
}

func (n *Node) handleMessage(msg *Message, conn net.Conn) {
	pubkey, err := msg.DecodePubKey()
	if err != nil {
		n.log.WithField("remote_addr", conn.RemoteAddr().String()).Error(err)
		return
	}

	n.mu.Lock()
	id := crypto.Pm256(pubkey.Bytes())
	cxid := common.EncodeToCXID(id)

	// Update or create peer information
	if peer, ok := n.peers[cxid]; !ok {
		addr := conn.RemoteAddr().(*net.TCPAddr)
		peer = p2p.NewPeer(addr, version, pubkey, uint64(time.Now().Unix()))
		n.peers[cxid] = peer
		n.peerConnections[cxid] = conn
	} else {
		peer.SetLastSeen(uint64(time.Now().Unix()))
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

		n.blocksReceived[blk.Hash()] = true

		newMessage := NewMessage(HASH, blk.Hash().Bytes(), n.pubKey)
		newMessage, err = n.signMessage(newMessage)
		if err != nil {
			n.log.WithField("client_id", cxid).Error(err)
			return
		}

		n.broadcast(newMessage, cxid)
	case HASH:
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

		hashBlock := common.BytesToHash(data)
		if !n.bc.HasBlock(hashBlock) {
			newMessage := NewMessage(ASK, data, n.pubKey)
			newMessage, err = n.signMessage(newMessage)
			if err != nil {
				n.log.WithField("client_id", cxid).Error(err)
				return
			}

			n.response(newMessage, cxid)
		}
	case ASK:
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

			n.blocksTransmited[hashBlock] = true

			n.response(newMessage, cxid)
		}
	case PEER_INFO:
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

		peerInfo := struct {
			ChainID      uint64      `json:"chain_id"`
			ProtocolHash common.Hash `json:"protocol_hash"`
			LatestBlock  common.Hash `json:"latest_block"`
		}{}

		err = json.Unmarshal(data, &peerInfo)
		if err != nil {
			n.log.WithField("client_id", cxid).Error(err)
			return
		}

		if peerInfo.ChainID != n.bc.ChainID() {
			n.log.WithField("client_id", cxid).Error("Invalid chain ID")
			return
		}

		if peerInfo.ProtocolHash != n.bc.ProtocolHash() {
			n.log.WithField("client_id", cxid).Error("Invalid protocol hash")
			return
		}

		if peerInfo.LatestBlock.IsValid() {
			latestBlock, err := n.bc.GetLatestBlock()
			if err != nil {
				n.log.WithField("client_id", cxid).Error(err)
				return
			}

			blk, err := n.bc.GetBlockByHash(peerInfo.LatestBlock)
			if err != nil {
				n.log.WithField("client_id", cxid).Error(err)
				return
			}

			if blk.Hash() != peerInfo.LatestBlock {
				n.log.WithField("client_id", cxid).Error("Invalid block")
				return
			}

			diffBlocks := blk.Height() - latestBlock.Height()
			if diffBlocks == 0 {
				n.log.WithField("client_id", cxid).Error("Node synced")
				return
			}

			for i := 1; i < int(diffBlocks); i++ {
				rBlk, err := n.bc.GetBlockByHeight(blk.Height() + uint64(i))
				if err != nil {
					n.log.WithField("client_id", cxid).Error(err)
					return
				}

				b, err := json.Marshal(rBlk)
				if err != nil {
					n.log.WithField("client_id", cxid).Error(err)
					return
				}

				nMsg := NewMessage(BLOCK, b, n.pubKey)
				signedMsg, err := n.signMessage(nMsg)
				if err != nil {
					n.log.WithField("client_id", cxid).Error(err)
					return
				}

				n.response(signedMsg, cxid)

			}

		}

	}
}

// ConnectToPeer establishes a TCP connection to another peer
func (n *Node) ConnectToPeer(addr *net.TCPAddr) error {
	conn, err := net.DialTimeout("tcp", addr.String(), connectTimeout)
	if err != nil {
		return err
	}

	latestBlock, err := n.bc.GetLatestBlock()
	if err != nil {
		conn.Close()
		return err
	}

	// Send our peer information immediately after connecting
	peerInfo := struct {
		ChainID      uint64      `json:"chain_id"`
		ProtocolHash common.Hash `json:"protocol_hash"`
		LatestBlock  common.Hash `json:"latest_block"`
	}{
		ChainID:      n.bc.ChainID(),
		ProtocolHash: n.bc.ProtocolHash(),
		LatestBlock:  latestBlock.Hash(),
	}

	data, err := json.Marshal(peerInfo)
	if err != nil {
		conn.Close()
		return err
	}

	msg := NewMessage(PEER_INFO, data, n.pubKey)
	signedMsg, err := n.signMessage(msg)
	if err != nil {
		conn.Close()
		return err
	}

	_, err = conn.Write(signedMsg.Bytes())
	if err != nil {
		conn.Close()
		return err
	}

	// Start a goroutine to handle this connection
	go n.handleConnection(conn.(*net.TCPConn))

	return nil
}

func (n *Node) GetID() []byte {
	n.mu.RLock()
	defer n.mu.RUnlock()

	return n.self.ID()
}

func (n *Node) GetAddr() *net.TCPAddr {
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

	// Close the connection if it exists
	if conn, ok := n.peerConnections[cxid]; ok {
		conn.Close()
		delete(n.peerConnections, cxid)
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

func (n *Node) propagateBlock() error {
	for {
		time.Sleep(5 * time.Second)

		latestBlock, err := n.bc.GetLatestBlock()
		if err != nil {
			n.log.Error("Error getting latest block: ", err)
			continue
		}

		n.mu.Lock()

		for cxid, peer := range n.peers {
			if peer.CXID() != n.self.CXID() {
				if _, ok := n.blocksTransmited[latestBlock.Hash()]; ok {
					continue
				}
				newMessage := NewMessage(HASH, latestBlock.Hash().Bytes(), n.pubKey)
				newMessage, err = n.signMessage(newMessage)
				if err != nil {
					n.log.WithField("client_id", peer.CXID()).Error(err)
					continue
				}

				if err := n.sendMessage(cxid, newMessage); err != nil {
					n.log.WithField("client_id", peer.CXID()).Error("Error sending block hash: ", err)
					continue
				}

				n.log.WithField("client_id", peer.CXID()).Info("Block proposed")
			}
		}
		n.mu.Unlock()
	}
}

func (n *Node) response(msg *Message, cxid string) {
	if err := n.sendMessage(cxid, msg); err != nil {
		n.log.WithField("client_id", cxid).Error("Error sending response: ", err)
		return
	}
	n.log.WithField("client_id", cxid).Info("Message sent")
}

func (n *Node) broadcast(msg *Message, senderCXID string) {
	n.mu.RLock()
	defer n.mu.RUnlock()

	for cxid := range n.peers {
		if cxid != senderCXID {
			if err := n.sendMessage(cxid, msg); err != nil {
				n.log.WithField("client_id", cxid).Error("Error broadcasting message: ", err)
				continue
			}
			n.log.WithField("client_id", cxid).Info("Message sent")
		}
	}
}

func (n *Node) sendMessage(cxid string, msg *Message) error {
	n.mu.RLock()
	conn, ok := n.peerConnections[cxid]
	n.mu.RUnlock()

	if !ok {
		// Try to establish a new connection if we don't have one
		peer, err := n.GetPeerByID(common.DecodeCXID(cxid))
		if err != nil {
			return err
		}

		tcpAddr := peer.Addr()
		newConn, err := net.DialTCP("tcp", nil, tcpAddr)
		if err != nil {
			return err
		}

		n.mu.Lock()
		n.peerConnections[cxid] = newConn
		conn = newConn
		n.mu.Unlock()

		// Start handling the new connection
		go n.handleConnection(newConn)
	}

	conn.SetWriteDeadline(time.Now().Add(writeDeadline))
	_, err := conn.Write(msg.Bytes())
	return err
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

func (n *Node) ping() {
	for {
		time.Sleep(5 * time.Second)

		n.mu.Lock()

		now := time.Now().Unix()
		for cxid, peer := range n.peers {
			if now-int64(peer.LastSeen()) > 10 {
				n.log.WithField("client_id", cxid).Info("Client disconnected")
				if conn, ok := n.peerConnections[cxid]; ok {
					conn.Close()
					delete(n.peerConnections, cxid)
				}
				delete(n.peers, cxid)
				continue
			}

			pingMsg := NewMessage(PING, []byte(fmt.Sprintf("%d", now)), n.pubKey)
			signedPing, err := n.signMessage(pingMsg)
			if err != nil {
				n.log.WithField("client_id", cxid).Error(err)
				continue
			}

			if err := n.sendMessage(cxid, signedPing); err != nil {
				n.log.WithField("client_id", cxid).Error("Error sending ping: ", err)
				continue
			}

			n.log.WithField("client_id", cxid).Info("Ping sent")
		}

		n.mu.Unlock()
	}
}
