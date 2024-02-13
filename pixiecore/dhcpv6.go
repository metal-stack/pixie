package pixiecore

import (
	"fmt"

	"github.com/metal-stack/pixie/dhcp6"
)

func (s *ServerV6) serveDHCP(conn *dhcp6.Conn) error {
	s.Log.Debug("waiting for packets...")
	for {
		pkt, src, err := conn.RecvDHCP()
		if err != nil {
			return fmt.Errorf("error receiving DHCP packet: %w", err)
		}
		if err := pkt.ShouldDiscard(s.Duid); err != nil {
			s.Log.Debug("discarding packet", "type", pkt.Type, "packet", pkt.TransactionID, "error", err)
			continue
		}

		s.Log.Debug("received packet", "type", pkt.Type, "packet", pkt.TransactionID, "options", pkt.Options.HumanReadable())

		response, err := s.PacketBuilder.BuildResponse(pkt, s.Duid, s.BootConfig, s.AddressPool)
		if err != nil {
			s.Log.Info("error creating response for transaction", "transaction", pkt.TransactionID, "error", err)
			if response == nil {
				s.Log.Info("dropping the packet")
				continue
			} else {
				s.Log.Info("will notify the client")
			}
		}
		if response == nil {
			s.Log.Info("don't know how to respond to packet", "type", pkt.Type, "packet", pkt.TransactionID)
			continue
		}

		marshalledResponse, err := response.Marshal()
		if err != nil {
			s.Log.Error("error marshalling response", "type", response.Type, "packet", response.TransactionID, "error", err)
			continue
		}

		if err := conn.SendDHCP(src, marshalledResponse); err != nil {
			s.Log.Error("error sending reply", "type", response.Type, "packet", response.TransactionID, "error", err)
			continue
		}

		s.Log.Debug("sent packet", "type", response.Type, "packet", response.TransactionID, "options", response.Options.HumanReadable())
	}
}
