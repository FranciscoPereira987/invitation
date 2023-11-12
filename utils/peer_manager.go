package utils

type Peers struct {
	Peers   []uint
	Members []uint
	peerMap map[uint]bool

	peerAddresses map[uint]string
}

func NewPeers(peers []uint, mappings map[uint]string) *Peers {
	livePeers := make(map[uint]bool)
	for _, peer := range peers {
		livePeers[peer] = true
	}

	return &Peers{
		Peers:         peers,
		Members:       make([]uint, 0),
		peerMap:       livePeers,
		peerAddresses: mappings,
	}
}

func (p *Peers) KillPeer(peer uint) {
	if _, ok := p.peerMap[peer]; ok {
		p.peerMap[peer] = false
	}
}

func (p *Peers) GetAddr(peer uint) string {
	return p.peerAddresses[peer]
}

func (p *Peers) AlivePeer(peer uint) {
	if _, ok := p.peerMap[peer]; ok {
		p.peerMap[peer] = true
	}
}

func (p *Peers) AddMembers(peers ...uint) {
	p.Members = append(p.Members, peers...)
}

func (p *Peers) IsMember(peer uint) bool {
	for _, peerId := range p.Members {
		if peerId == peer {
			return true
		}
	}
	return false
}

func (p *Peers) GroupIsComplete() bool {
	return len(p.Members) == len(p.Peers)
}

func (p *Peers) isAlive(peer uint) bool {
	return p.peerMap[peer]
}
