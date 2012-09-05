package main

import (
	"errors"
	"fmt"
)

type Mapper interface {
	WriteRamBank(start int, length int, offset int)
	WriteVramBank(start int, length int, offset int)
	Init(rom []byte) error
}

type Rom struct {
	PrgFlag   Word
	ChrFlag   Word
	Mirroring int
	Data      []byte
}

type Nrom Rom
type Mmc1 Rom

func (r *Nrom) WriteRamBank(start int, length int, offset int) {
	for i := 0; i < length; i++ {
		Ram.Write(i+start, Word(r.Data[i+offset]))
	}
}

func (r *Nrom) WriteVramBank(start int, length int, offset int) {
	for i := start; i < length; i++ {
		ppu.Vram[i] = Word(r.Data[i+offset])
	}
}

func (r *Nrom) Init(rom []byte) error {
	r.PrgFlag = Word(rom[4])
	r.ChrFlag = Word(rom[5])

    switch rom[6] & 0x1 {
    case 0x0:
        fmt.Println("Horizontal mirroring")
        r.Mirroring = MirroringHorizontal
    case 0x1:
        r.Mirroring = MirroringVertical
        fmt.Println("Vertical mirroring")
    }

    ppu.Mirroring = r.Mirroring

	// ROM data starts at byte 16
	r.Data = rom[16:]

	r.WriteRamBank(0x8000, 0x4000, 0x0)

	switch r.PrgFlag {
	case 0x01:
		r.WriteRamBank(0xC000, 0x4000, 0x0)
		r.WriteVramBank(0x0000, 0x2000, 0x4000)
	case 0x02:
		r.WriteRamBank(0xC000, 0x4000, 0x4000)
		r.WriteVramBank(0x0000, 0x2000, 0x8000)
	}

	return nil
}

func (r *Mmc1) WriteRamBank(start int, length int, offset int) {
	for i := 0; i < length; i++ {
		Ram.Write(i+start, Word(r.Data[i+offset]))
	}
}

func (r *Mmc1) WriteVramBank(start int, length int, offset int) {
	for i := start; i < length; i++ {
		ppu.Vram[i] = Word(r.Data[i+offset])
	}
}

func (r *Mmc1) Init(rom []byte) error {
	r.PrgFlag = Word(rom[4])
	r.ChrFlag = Word(rom[5])

	r.Data = rom[16:]

	// Write the first ROM bank
	r.WriteRamBank(0x8000, 0x4000, 0x0)
	// and the last ROM bank
	r.WriteRamBank(0xC000, 0x4000, len(r.Data)-0x4000)

	// r.WriteVramBank(0x0000, 0x2000, 0x0)

	return nil
}

func LoadRom(rom []byte) (r Mapper, e error) {
	if string(rom[0:3]) != "NES" {
		return r, errors.New("Invalid ROM file")

		if rom[3] != 0x1a {
			return r, errors.New("Invalid ROM file")
		}
	}

	mapper := (Word(rom[6])>>4 | (Word(rom[7]) & 0xF0))
	switch mapper {
	case 0x00:
		// NROM
		r = new(Nrom)
	case 0x01:
		// MMC1
		r = new(Mmc1)
	default:
		// Unsupported
		return r, errors.New("Unsupported memory mapper")
	}

	return
}
