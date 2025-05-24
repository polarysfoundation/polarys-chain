package p2p

import (
	"net"

	pec256 "github.com/polarysfoundation/pec-256"
	"github.com/polarysfoundation/polarys-chain/modules/crypto"
)

type Peer struct {
	id       []byte        // ID node
	addr     *net.UDPAddr  // ip:port
	version  uint32        // version
	pubKey   pec256.PubKey // public key
	lastSeen uint64        // last seen time
}

func NewPeer(addr *net.UDPAddr, version uint32, pubKey pec256.PubKey, lastSeen uint64) *Peer {

	id := crypto.Pm256(pubKey.Bytes())

	return &Peer{
		id:       id,
		addr:     addr,
		version:  version,
		pubKey:   pubKey,
		lastSeen: lastSeen,
	}
}

func (p *Peer) ID() []byte {
	return p.id
}

func (p *Peer) Addr() *net.UDPAddr {
	return p.addr
}

func (p *Peer) Version() uint32 {
	return p.version
}

func (p *Peer) PubKey() pec256.PubKey {
	return p.pubKey
}

func (p *Peer) LastSeen() uint64 {
	return p.lastSeen
}

func (p *Peer) SetLastSeen(lastSeen uint64) {
	p.lastSeen = lastSeen
}

