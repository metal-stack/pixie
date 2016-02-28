package pcap

import (
	"encoding/binary"
	"io"
)

// Writer serializes Packets to an io.Writer.
type Writer struct {
	Writer    io.Writer
	LinkType  LinkType
	SnapLen   uint32
	ByteOrder binary.ByteOrder // defaults to binary.LittleEndian

	headerWritten bool
}

func (w *Writer) order() binary.ByteOrder {
	if w.ByteOrder != nil {
		return w.ByteOrder
	}
	return binary.LittleEndian
}

func (w *Writer) header() error {
	hdr := struct {
		Magic   uint32
		Major   uint16
		Minor   uint16
		Ignored uint64
		Snaplen uint32
		Type    uint32
	}{
		Magic:   0xa1b23c4d,
		Major:   2,
		Minor:   4,
		Snaplen: w.SnapLen,
		Type:    uint32(w.LinkType),
	}

	if err := binary.Write(w.Writer, w.order(), hdr); err != nil {
		return err
	}
	w.headerWritten = true
	return nil
}

// Put serializes pkt to w.Writer.
func (w *Writer) Put(pkt *Packet) error {
	if !w.headerWritten {
		if err := w.header(); err != nil {
			return err
		}
	}
	hdr := struct {
		Sec     uint32
		NSec    uint32
		Len     uint32
		OrigLen uint32
	}{
		Sec:     uint32(pkt.Timestamp.Unix()),
		NSec:    uint32(pkt.Timestamp.Nanosecond()),
		Len:     uint32(len(pkt.Bytes)),
		OrigLen: uint32(pkt.Length),
	}

	if err := binary.Write(w.Writer, w.order(), hdr); err != nil {
		return err
	}
	if _, err := w.Writer.Write(pkt.Bytes); err != nil {
		return err
	}
	return nil
}
