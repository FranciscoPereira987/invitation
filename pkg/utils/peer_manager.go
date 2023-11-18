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

func (p *Peers) Remove(peer uint) {
	for index, peerId := range p.Members {
		if peer == peerId {
			p.Members = append(p.Members[:index], p.Members[index+1:]...)
		}
	}
}

func (p *Peers) GetAddr(peer uint) string {
	return p.peerAddresses[peer]
}

func (p *Peers) IsInGroup(peer uint) bool {
	for _, peerId := range p.Members {
		if peerId == peer {
			return true
		}
	}
	return false
}

func (p *Peers) AddMembers(peers ...uint) {
	for _, peer := range peers {
		if !p.IsInGroup(peer) {
			p.Members = append(p.Members, peer)
		}
	}
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
