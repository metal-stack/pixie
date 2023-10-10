package pixiecore

import (
	"fmt"

	"github.com/metal-stack/pixie/dhcp6"
)

func (s *ServerV6) serveDHCP(conn *dhcp6.Conn) error {
	s.Log.Debug("dhcpv6", "Waiting for packets...\n")
	for {
		pkt, src, err := conn.RecvDHCP()
		if err != nil {
			return fmt.Errorf("error receiving DHCP packet: %w", err)
		}
		if err := pkt.ShouldDiscard(s.Duid); err != nil {
			s.Log.Debug("dhcpv6, discarding (%d) packet (%d): %s\n", pkt.Type, pkt.TransactionID, err))
			continue
		}

		s.Log.Debug("dhcpv6", fmt.Sprintf("Received (%d) packet (%d): %s\n", pkt.Type, pkt.TransactionID, pkt.Options.HumanReadable()))

		response, err := s.PacketBuilder.BuildResponse(pkt, s.Duid, s.BootConfig, s.AddressPool)
		if err != nil {
			s.Log.Info("dhcpv6", fmt.Sprintf("Error creating response for transaction: %d: %s", pkt.TransactionID, err))
			if response == nil {
				s.Log.Info("dhcpv6","dropping the packet")
				continue
			} else {
				s.Log.Info("dhcpv6", "will notify the client")
			}
		}
		if response == nil {
			s.Log.Info("dhcpv6", fmt.Sprintf("Don't know how to respond to packet type: %d (transaction id %d)", pkt.Type, pkt.TransactionID))
			continue
		}

		marshalledResponse, err := response.Marshal()
		if err != nil {
			s.Log.Info("dhcpv6", fmt.Sprintf("Error marshalling response (%d) (%d): %s", response.Type, response.TransactionID, err))
			continue
		}

		if err := conn.SendDHCP(src, marshalledResponse); err != nil {
			s.Log.Info("dhcpv6", fmt.Sprintf("Error sending reply (%d) (%d): %s", response.Type, response.TransactionID, err))
			continue
		}

		s.Log.Debug("dhcpv6", fmt.Sprintf("Sent (%d) packet (%d): %s\n", response.Type, response.TransactionID, response.Options.HumanReadable()))
	}
}
