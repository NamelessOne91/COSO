package network

// Veth represents a virtual Ethernet device pair. Veth devices are always created in interconnected pairs.
//
// Veth devices can act as a tunnel etween network namespaces to create a bridge to a
// physical network device in another namespace. That's because packets transmitted on one device in the pair
// are immediately received on the other device
type Veth struct{}

func NewVeth() *Veth {
	return &Veth{}
}
