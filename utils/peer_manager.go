package utils

type Peers struct {
	Peers   []uint
	Members []uint

	peerAddresses map[uint]string
}

func NewPeers(peers []uint, mappings map[uint]string) *Peers {

	return &Peers{
		Peers:         peers,
		Members:       make([]uint, 0),
		peerAddresses: mappings,
	}
}

func (p *Peers) GetAddr(peer uint) string {
	return p.peerAddresses[peer]
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
