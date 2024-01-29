package mbr

import (
	"encoding/binary"
	"errors"
)

const (
	bootstrapLen    = 440
	uniqueDiskIDOff = bootstrapLen
	uniqueDiskIDLen = 4
	reservedLen     = 2
	pteOffset       = bootstrapLen + uniqueDiskIDLen + reservedLen
	pteLen          = 16 // partition table entry length
)

func ToBootSector(start []byte) (BootSector, error) {
	if len(start) < 512 {
		return BootSector{}, errors.New("boot sector too short")
	}
	bs := BootSector{
		data: start[:512:512],
	}
	return bs, nil
}

type BootSector struct {
	data []byte
}

type PartitionTableEntry struct {
	data [pteLen]byte
}

type DriveAttributes byte

// Bootstrap returns bytes 0..439 of the MBR containing the binary executable code.
func (mbr *BootSector) Bootstrap() []byte {
	return mbr.data[0:bootstrapLen]
}

func (mbr *BootSector) UniqueDiskID() uint32 {
	return binary.LittleEndian.Uint32(mbr.data[uniqueDiskIDOff : uniqueDiskIDOff+uniqueDiskIDLen])
}

func (mbr *BootSector) PartitionTable(i int) PartitionTableEntry {
	if i > 3 {
		panic("invalid partition table index")
	}
	return PartitionTableEntry{
		data: [pteLen]byte(mbr.data[pteOffset+i*pteLen : pteOffset+(i+1)*pteLen]),
	}
}

func (pte *PartitionTableEntry) Attributes() DriveAttributes {
	return DriveAttributes(pte.data[0])
}

func (pte *PartitionTableEntry) PartitionType() PartitionType {
	return PartitionType(pte.data[4])
}

// StartSector returns the starting sector of the partition in LBA format.
func (pte *PartitionTableEntry) StartSector() uint32 {
	return binary.LittleEndian.Uint32(pte.data[8:12])
}

// NumberOfSectors returns the number of sectors in the partition.
func (pte *PartitionTableEntry) NumberOfSectors() uint32 {
	return binary.LittleEndian.Uint32(pte.data[12:16])
}

// CHSStart returns the starting sector of the partition in CHS format.
func (pte *PartitionTableEntry) CHSStart() (cylinder, head, sector uint8) {
	return pte.data[1], pte.data[2], pte.data[3]
}

// CHSLast returns the last sector of the partition in CHS format.
func (pte *PartitionTableEntry) CHSLast() (cylinder, head, sector uint8) {
	return pte.data[5], pte.data[6], pte.data[7]
}

type PartitionType byte

const (
	PartitionTypeUnused   PartitionType = 0x00
	PartitionTypeFAT12    PartitionType = 0x01
	PartitionTypeFAT16    PartitionType = 0x04
	PartitionTypeExtended PartitionType = 0x05
	PartitionTypeFAT32    PartitionType = 0x0B
	PartitionTypeFAT32L   PartitionType = 0x0C
	PartitionTypeNTFS     PartitionType = 0x07 // Also includes exFAT.
	PartitionTypeLinux    PartitionType = 0x83
	PartitionTypeFreeBSD  PartitionType = 0xA5
	PartitionTypeAppleHFS PartitionType = 0xAF
)
