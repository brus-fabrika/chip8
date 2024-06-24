package chip8

import (
	"bufio"
	"fmt"
	"math/rand"
	"os"
)

const (
	DISPLAY_WIDTH          = 64
	DISPLAY_HEIGHT         = 32
	MEMORY_SIZE     uint16 = 0x1000
	MEMORY_FONT     uint16 = 0x01B0
	MEMORY_USER     uint16 = 0x0200
	MEMORY_STACK    uint16 = MEMORY_SIZE - 0x0200 + 0x00a0
	MEMORY_INT_AREA uint16 = MEMORY_STACK + 0x0030
	MEMORY_REG_AREA uint16 = MEMORY_INT_AREA + 0x0020
	MEMORY_DISPLAY  uint16 = MEMORY_REG_AREA + 0x0010
)

var font_data = []uint8{
	0xF0, 0x90, 0x90, 0x90, 0xF0, // 0
	0x20, 0x60, 0x20, 0x20, 0x70, // 1
	0xF0, 0x10, 0xF0, 0x80, 0xF0, // 2
	0xF0, 0x10, 0xF0, 0x10, 0xF0, // 3
	0x90, 0x90, 0xF0, 0x10, 0x10, // 4
	0xF0, 0x80, 0xF0, 0x10, 0xF0, // 5
	0xF0, 0x80, 0xF0, 0x90, 0xF0, // 6
	0xF0, 0x10, 0x20, 0x40, 0x40, // 7
	0xF0, 0x90, 0xF0, 0x90, 0xF0, // 8
	0xF0, 0x90, 0xF0, 0x10, 0xF0, // 9
	0xF0, 0x90, 0xF0, 0x90, 0x90, // A
	0xE0, 0x90, 0xE0, 0x90, 0xE0, // B
	0xF0, 0x80, 0x80, 0x80, 0xF0, // C
	0xE0, 0x90, 0x90, 0x90, 0xE0, // D
	0xF0, 0x80, 0xF0, 0x80, 0xF0, // E
	0xF0, 0x80, 0xF0, 0x80, 0x80, // F
}

type Register int

const (
	RegV0 Register = iota
	RegV1
	RegV2
	RegV3
	RegV4
	RegV5
	RegV6
	RegV7
	RegV8
	RegV9
	RegVA
	RegVB
	RegVC
	RegVD
	RegVE
	RegVF
	RegI
	RegSP
	RegT0
	RegT1
	RegPC
)

type ChipVersion int

const (
	Chip_8 ChipVersion = iota
	Super_Chip_Modern
	XO_Chip
	Super_Chip_Legacy
)

type Chip8 struct {
	Ver           ChipVersion
	Memory        [MEMORY_SIZE]uint8
	DisplayBuffer [DISPLAY_WIDTH * DISPLAY_HEIGHT]bool
	Keyboard      [0x10]bool
	Reg           RegisterSet

	RomSize uint16 // just for control and debug

	State struct {
		Running bool
		Paused  bool
	}
}

type RegisterSet struct {
	PC uint16    // programm counter
	SP uint16    // stack pointer
	I  uint16    // index
	T0 uint8     // timer register (decrement counter)
	T1 uint8     // tone register (decrement counter)
	V  [16]uint8 // general purpose registers
}

func (chip *Chip8) getRegister(r Register) uint16 {
	switch r {
	case RegI:
		return chip.Reg.I
	case RegSP:
		return chip.Reg.SP
	case RegT0:
		return uint16(chip.Reg.T0)
	case RegT1:
		return uint16(chip.Reg.T1)
	case RegV0, RegV1, RegV2, RegV3,
		RegV4, RegV5, RegV6, RegV7,
		RegV8, RegV9, RegVA, RegVB,
		RegVC, RegVD, RegVE, RegVF:
		return uint16(chip.Reg.V[r])
	}

	return 0x0000
}

func (chip *Chip8) setRegister(r Register, val uint16) {
	switch r {
	case RegI:
		chip.Reg.I = val
	case RegSP:
		chip.Reg.SP = val
	case RegT0:
		chip.Reg.T0 = uint8(val)
	case RegT1:
		chip.Reg.T1 = uint8(val)
	case RegV0, RegV1, RegV2, RegV3,
		RegV4, RegV5, RegV6, RegV7,
		RegV8, RegV9, RegVA, RegVB,
		RegVC, RegVD, RegVE, RegVF:
		chip.Reg.V[r] = uint8(val)
	}
}

func (chip *Chip8) ClearScreen() {
	for i := 0; i < DISPLAY_WIDTH*DISPLAY_HEIGHT; i++ {
		chip.DisplayBuffer[i] = false
	}
	chip.Reg.PC += 2
}

func (chip *Chip8) DisplayAt(xr, yr Register, h int) {
	x := int(chip.getRegister(xr)) & (DISPLAY_WIDTH - 1)
	y := int(chip.getRegister(yr)) & (DISPLAY_HEIGHT - 1)

	dataOffset := int(chip.Reg.I)
	data := chip.Memory[dataOffset : dataOffset+h]

	chip.Reg.V[0x0F] = 0x00

	for yOffset, v := range data {
		if y+yOffset >= DISPLAY_HEIGHT {
			break
		}
		for xOffset := 0; xOffset < 8; xOffset++ {
			if x+xOffset >= DISPLAY_WIDTH {
				break
			}

			spriteBit := v & (1 << (7 - xOffset))

			if spriteBit != 0 {
				index := x + xOffset + (y+yOffset)*DISPLAY_WIDTH
				cv := chip.DisplayBuffer[index]
				if cv {
					chip.DisplayBuffer[index] = false
					chip.Reg.V[0x0F] = 0x01
				} else {
					chip.DisplayBuffer[index] = true
				}
			}
		}
	}

	chip.Reg.PC += 2
}

func (chip *Chip8) MovRegVal(r Register, val uint16) {
	chip.setRegister(r, val)
	chip.Reg.PC += 2
}

func (chip *Chip8) MovRegRnd(r Register, mask uint8) {
	rnd := uint8(rand.Intn(0x100)) & mask
	chip.setRegister(r, uint16(rnd))
	chip.Reg.PC += 2
}

func (chip *Chip8) MovRegReg(rdst, rsrc Register) {
	chip.setRegister(rdst, chip.getRegister(rsrc))
	chip.Reg.PC += 2
}

func (chip *Chip8) AddRegVal(r Register, val uint8) {
	// no carry flag modification!

	chip.setRegister(r, chip.getRegister(r)+uint16(val))
	chip.Reg.PC += 2
}

func (chip *Chip8) AddRegReg(r1, r2 Register) {
	result := chip.getRegister(r1) + chip.getRegister(r2)

	chip.setRegister(r1, result)

	// carry flag modified if needed
	if result > 0x00ff {
		chip.setRegister(RegVF, 0x01)
	} else {
		chip.setRegister(RegVF, 0x00)
	}

	chip.Reg.PC += 2
}

func (chip *Chip8) SubRegReg(r1, r2 Register) {
	// carry flag modified if needed
	result := chip.getRegister(r1) - chip.getRegister(r2)
	cf := chip.getRegister(r1) >= chip.getRegister(r2)

	chip.setRegister(r1, result)

	if cf {
		// carry flag is set if there was NO underflow
		chip.setRegister(RegVF, 0x01)
	} else {
		chip.setRegister(RegVF, 0x00)
	}

	chip.Reg.PC += 2
}

func (chip *Chip8) SubNegRegReg(r1, r2 Register) {
	// carry flag modified if needed
	result := chip.getRegister(r2) - chip.getRegister(r1)
	cf := chip.getRegister(r2) >= chip.getRegister(r1)

	chip.setRegister(r1, result)

	if cf {
		// carry flag is set if there was NO underflow
		chip.setRegister(RegVF, 0x01)
	} else {
		chip.setRegister(RegVF, 0x00)
	}

	chip.Reg.PC += 2
}

func (chip *Chip8) Or(r1, r2 Register) {
	result := chip.getRegister(r1) | chip.getRegister(r2)
	chip.setRegister(r1, result)
	chip.setRegister(RegVF, 0x00)
	chip.Reg.PC += 2
}

func (chip *Chip8) Xor(r1, r2 Register) {
	result := chip.getRegister(r1) ^ chip.getRegister(r2)
	chip.setRegister(r1, result)
	chip.setRegister(RegVF, 0x00)
	chip.Reg.PC += 2
}

func (chip *Chip8) And(r1, r2 Register) {
	result := chip.getRegister(r1) & chip.getRegister(r2)
	chip.setRegister(r1, result)
	chip.setRegister(RegVF, 0x00)
	chip.Reg.PC += 2
}

func (chip *Chip8) ShiftR(r1, r2 Register) {
	if chip.Ver == Chip_8 {
		chip.setRegister(r1, chip.getRegister(r2))
	}

	cf := uint8(chip.getRegister(r1) & 0x01)

	result := chip.getRegister(r1) >> 1
	chip.setRegister(r1, result)

	chip.Reg.V[0x0F] = cf

	chip.Reg.PC += 2
}

func (chip *Chip8) ShiftL(r1, r2 Register) {
	if chip.Ver == Chip_8 {
		chip.setRegister(r1, chip.getRegister(r2))
	}

	cf := uint8(chip.getRegister(r1) & 0x80 >> 7)

	result := chip.getRegister(r1) << 1
	chip.setRegister(r1, result)

	chip.Reg.V[0x0F] = cf

	chip.Reg.PC += 2
}

func (chip *Chip8) Jump(adr uint16) {
	chip.Reg.PC = adr
}

func (chip *Chip8) JumpV(adr uint16) {
	chip.Reg.PC = adr + uint16(chip.Reg.V[0])
}

func (chip *Chip8) Call(adr uint16) {
	chip.Memory[chip.Reg.SP] = uint8(chip.Reg.PC)
	chip.Memory[chip.Reg.SP-1] = uint8(chip.Reg.PC >> 8)
	chip.Reg.SP -= 2

	chip.Reg.PC = adr
}

func (chip *Chip8) Ret() {
	chip.Reg.PC = uint16(chip.Memory[chip.Reg.SP+1])<<8 + uint16(chip.Memory[chip.Reg.SP+2])
	chip.Reg.SP += 2

	chip.Reg.PC += 2
}

func (chip *Chip8) SkipEqualVal(reg Register, val uint8) {
	if chip.getRegister(reg) != uint16(val) {
		chip.Reg.PC += 2
	} else {
		chip.Reg.PC += 4
	}
}

func (chip *Chip8) SkipNotEqualVal(reg Register, val uint8) {
	if chip.getRegister(reg) == uint16(val) {
		chip.Reg.PC += 2
	} else {
		chip.Reg.PC += 4
	}
}

func (chip *Chip8) SkipEqualReg(r1 Register, r2 Register) {
	if chip.getRegister(r1) != chip.getRegister(r2) {
		chip.Reg.PC += 2
	} else {
		chip.Reg.PC += 4
	}
}

func (chip *Chip8) SkipNotEqualReg(r1 Register, r2 Register) {
	if chip.getRegister(r1) == chip.getRegister(r2) {
		chip.Reg.PC += 2
	} else {
		chip.Reg.PC += 4
	}
}

func (chip *Chip8) SkipKeyPressedAtReg(r Register) {
	if chip.Keyboard[chip.getRegister(r)] {
		chip.Reg.PC += 4
	} else {
		chip.Reg.PC += 2
	}
}

func (chip *Chip8) SkipKeyNotPressedAtReg(r Register) {
	if !chip.Keyboard[chip.getRegister(r)] {
		chip.Reg.PC += 4
	} else {
		chip.Reg.PC += 2
	}
}

func (chip *Chip8) BcdReg(r Register) {

	origVal := uint8(chip.getRegister(r))

	chip.Memory[chip.Reg.I+0] = origVal / 100
	chip.Memory[chip.Reg.I+1] = (origVal % 100) / 10
	chip.Memory[chip.Reg.I+2] = (origVal % 10)
	chip.Reg.PC += 2
}

func (chip *Chip8) CopyRegToMem(r Register) {
	for x := 0; x <= int(r); x++ {
		chip.Memory[chip.Reg.I] = chip.Reg.V[Register(x)]
		chip.Reg.I++
	}
	chip.Reg.PC += 2
}

func (chip *Chip8) CopyMemToReg(r Register) {
	for x := 0; x <= int(r); x++ {
		chip.Reg.V[Register(x)] = chip.Memory[chip.Reg.I]
		chip.Reg.I++
	}
	chip.Reg.PC += 2
}

func (chip *Chip8) SetCharReg(r Register) {
	val := chip.getRegister(r) & 0x000F
	chip.Reg.I = MEMORY_FONT + val

	chip.Reg.PC += 2
}

func (chip *Chip8) GetKeyReg(r Register) {
	anyKeyPressed := false

	for key, pressed := range chip.Keyboard {
		if pressed {
			chip.setRegister(r, uint16(key))
			anyKeyPressed = true
			break
		}
	}

	if anyKeyPressed {
		chip.Reg.PC += 2
	}
}

func (chip *Chip8) ClearKeyboard() {
	// clear keaboard state
	for i := 0; i < 16; i++ {
		chip.Keyboard[i] = false
	}
}

func (chip *Chip8) Init(ver ChipVersion) {
	chip.Ver = ver

	chip.ClearScreen()

	// clear memory
	for i := 0; i < int(MEMORY_SIZE); i++ {
		chip.Memory[i] = 0
	}

	chip.ClearKeyboard()

	// clear registers state
	for i := 0; i < 16; i++ {
		chip.Reg.V[i] = 0
	}

	chip.LoadFontFromData(font_data)

	chip.Reg.PC = MEMORY_USER           // set programm counter at the beginning of user prog area
	chip.Reg.SP = MEMORY_STACK + 0x002f // set stack pointer at the last byte of stack area
	chip.Reg.I = 0
	chip.Reg.T0 = 0
	chip.Reg.T1 = 0

	chip.State.Running = true
	chip.State.Paused = false

}

func (chip *Chip8) Execute() {
	var cmd uint16
	for i := 0; i < 40; /*int(chip.RomSize)*/ i += 2 {
		cmd = uint16(chip.Memory[int(chip.Reg.PC)])<<8 + uint16(chip.Memory[int(chip.Reg.PC+1)])
		chip.ProcessCmd(cmd)
	}
}

func (chip *Chip8) ProcessCmd(cmd uint16) {
	var cmdStr string

	// for debug print purpose - save the current PC
	curPC := chip.Reg.PC
	// just debug print for now
	switch cmd & 0xf000 {
	case 0x0000:
		if cmd == 0x00e0 {
			cmdStr = "CLS"
			chip.ClearScreen()
		} else if cmd == 0x00ee {
			cmdStr = "RET"
			chip.Ret()
		} else {
			cmdStr = fmt.Sprintf("MCALL 0x%04x", cmd&0x0fff)
		}
	case 0x1000:
		cmdStr = fmt.Sprintf("JMP 0x%04X", cmd&0x0fff)
		chip.Jump(cmd & 0x0fff)
	case 0x2000:
		cmdStr = fmt.Sprintf("CALL 0x%04x", cmd&0x0fff)
		chip.Call(cmd & 0x0fff)
	case 0x3000:
		cmdStr = fmt.Sprintf("SE  V%X, %02x", cmd&0x0f00>>8, cmd&0x00ff)
		chip.SkipEqualVal(Register(cmd&0x0f00>>8), uint8(cmd&0x00ff))
	case 0x4000:
		cmdStr = fmt.Sprintf("SNE V%X, %02x", cmd&0x0f00>>8, cmd&0x00ff)
		chip.SkipNotEqualVal(Register(cmd&0x0f00>>8), uint8(cmd&0x00ff))
	case 0x5000:
		cmdStr = fmt.Sprintf("SE  V%X, V%x", cmd&0x0f00>>8, cmd&0x00f0>>4)
		chip.SkipEqualReg(Register(cmd&0x0f00>>8), Register(cmd&0x00f0>>4))
	case 0x6000:
		cmdStr = fmt.Sprintf("MOV V%X, %02x", cmd&0x0f00>>8, cmd&0x00ff)
		chip.MovRegVal(Register(cmd&0x0f00>>8), cmd&0x00ff)
	case 0x7000:
		cmdStr = fmt.Sprintf("ADD V%x, %02x", cmd&0x0f00>>8, cmd&0x00ff)
		chip.AddRegVal(Register(cmd&0x0f00>>8), uint8(cmd&0x00ff))
	case 0x8000:
		// ALU command
		opCode := cmd & 0x000f
		xr := Register(cmd & 0x0f00 >> 8)
		yr := Register(cmd & 0x00f0 >> 4)

		switch opCode {
		case 0x00:
			cmdStr = fmt.Sprintf("MOV V%x, V%x", xr, yr)
			chip.MovRegReg(xr, yr)
		case 0x01:
			cmdStr = fmt.Sprintf("OR V%x, V%x", xr, yr)
			chip.Or(xr, yr)
		case 0x02:
			cmdStr = fmt.Sprintf("AND V%x, V%x", xr, yr)
			chip.And(xr, yr)
		case 0x03:
			cmdStr = fmt.Sprintf("XOR V%x, V%x", xr, yr)
			chip.Xor(xr, yr)
		case 0x04:
			cmdStr = fmt.Sprintf("ADD V%x, V%x", xr, yr)
			chip.AddRegReg(xr, yr)
		case 0x05:
			cmdStr = fmt.Sprintf("SUB V%x, V%x", xr, yr)
			chip.SubRegReg(xr, yr)
		case 0x06:
			cmdStr = fmt.Sprintf("SHR V%x, V%x", xr, yr)
			chip.ShiftR(xr, yr)
		case 0x07:
			cmdStr = fmt.Sprintf("SUBN V%x, V%x", xr, yr)
			chip.SubNegRegReg(xr, yr)
		case 0x0E:
			cmdStr = fmt.Sprintf("SHL V%x, V%x", xr, yr)
			chip.ShiftL(xr, yr)
		default:
			cmdStr = "NVO"
		}
	case 0x9000:
		cmdStr = fmt.Sprintf("SNE V%x, V%x", cmd&0x0f00, cmd&0x00f0)
		chip.SkipNotEqualReg(Register(cmd&0x0f00>>8), Register(cmd&0x00f0>>4))
	case 0xa000:
		cmdStr = fmt.Sprintf("MOV I, 0x%04x", cmd&0x0fff)
		chip.MovRegVal(RegI, cmd&0x0fff)
	case 0xb000:
		cmdStr = fmt.Sprintf("JMPV 0x%04x", cmd&0x0fff)
		chip.JumpV(cmd & 0x0fff)
	case 0xc000:
		cmdStr = fmt.Sprintf("RND V%X, %2X", cmd&0x0f00>>8, cmd&0x00ff)
		chip.MovRegRnd(Register(cmd&0x0f00>>8), uint8(cmd&0x00ff))
	case 0xd000:
		cmdStr = fmt.Sprintf("DRAW %x, V%X, V%X ; (%d, %d)", cmd&0x000f, cmd&0x0f00>>8, cmd&0x00f0>>4, chip.getRegister(Register(cmd&0x0f00>>8)), chip.getRegister(Register(cmd&0x00f0>>4)))
		xr := Register(cmd & 0x0f00 >> 8)
		yr := Register(cmd & 0x00f0 >> 4)
		chip.DisplayAt(xr, yr, int(cmd&0x000f))
	case 0xe000:
		xr := Register(cmd & 0x0f00 >> 8)

		if cmd&0x00ff == 0x009E {
			cmdStr = fmt.Sprintf("SK V%x", xr)
			chip.SkipKeyPressedAtReg(xr)
		} else if cmd&0x00ff == 0x00A1 {
			cmdStr = fmt.Sprintf("SNK V%x", xr)
			chip.SkipKeyNotPressedAtReg(xr)
		} else {
			cmdStr = "NVO"
		}
	case 0xf000:
		opCode := cmd & 0x00ff
		switch opCode {
		case 0x07:
			cmdStr = fmt.Sprintf("MOV V%X, T0", cmd&0x0f00>>8)
			chip.MovRegReg(Register(cmd&0x0f00>>8), RegT0)
		case 0x0A:
			// wait for key pressed
			cmdStr = fmt.Sprintf("KEY V%x", cmd&0x0f00>>8)
			chip.GetKeyReg(Register(cmd & 0x0f00 >> 8))
		case 0x15:
			cmdStr = fmt.Sprintf("MOV T0, V%X", cmd&0x0f00>>8)
			chip.MovRegReg(RegT0, Register(cmd&0x0f00>>8))
		case 0x18:
			cmdStr = fmt.Sprintf("MOV T1, V%X", cmd&0x0f00>>8)
			chip.MovRegReg(RegT1, Register(cmd&0x0f00>>8))
		case 0x1E:
			cmdStr = fmt.Sprintf("ADD I, V%X", cmd&0x0f00>>8)
			chip.AddRegVal(RegI, uint8(chip.getRegister(Register(cmd&0x0f00>>8))))
		case 0x29:
			cmdStr = fmt.Sprintf("STC V%X", cmd&0x0f00>>8)
			chip.SetCharReg(Register(cmd & 0x0f00 >> 8))
		case 0x33:
			cmdStr = fmt.Sprintf("BCD V%X", cmd&0x0f00>>8)
			chip.BcdReg(Register(cmd & 0x0f00 >> 8))
		case 0x55:
			cmdStr = fmt.Sprintf("CAM V%X", cmd&0x0f00>>8)
			chip.CopyRegToMem(Register(cmd & 0x0f00 >> 8))
		case 0x65:
			cmdStr = fmt.Sprintf("CAR V%X", cmd&0x0f00>>8)
			chip.CopyMemToReg(Register(cmd & 0x0f00 >> 8))
		default:
			cmdStr = "NVO"
		}
	default:
		cmdStr = "NVO"
	}

	fmt.Printf("\t%04x:\t%04x\t;%s\n", curPC, cmd, cmdStr)
}

func (chip *Chip8) LoadRomFromFile(fileName string) (uint16, error) {
	file, err := os.Open(fileName)
	if err != nil {
		return 0, err
	}
	defer file.Close()

	stats, statsErr := file.Stat()
	if statsErr != nil {
		return 0, statsErr
	}

	var size int64 = stats.Size()

	bufr := bufio.NewReader(file)
	_, err = bufr.Read(chip.Memory[MEMORY_USER:])

	chip.RomSize = uint16(size)

	return chip.RomSize, err
}

func (chip *Chip8) LoadRomFromData(data []uint8) (uint16, error) {
	for i, v := range data {
		chip.Memory[int(MEMORY_USER)+i] = v
	}
	chip.RomSize = uint16(len(data))

	return chip.RomSize, nil
}

func (chip *Chip8) LoadFontFromData(data []uint8) (uint16, error) {
	for i, v := range data {
		chip.Memory[MEMORY_FONT+uint16(i)] = v
	}

	return uint16(len(data)), nil
}

/*
*

	Dumps memory into output using startPos and endPos as memory dump boundaries.
	Both startPos and endPos are rounded to the 16 bytes, startPos to nearest below, endPos to nearest up
*/
func (chip *Chip8) MemoryDump(startPos uint16, endPos uint16) {
	startPos = startPos & 0xfff0
	endPos = (endPos + 16) & 0xfff0

	fmt.Printf("Memory dump %04x - %04x:\n", startPos, endPos)
	fmt.Printf("\t00 01 02 03 04 05 06 07 08 09 0a 0b 0c 0d 0e 0f\n")
	for startPos < endPos {
		fmt.Printf("%04x", startPos)
		fmt.Printf("\t%02x %02x %02x %02x", chip.Memory[startPos], chip.Memory[startPos+1], chip.Memory[startPos+2], chip.Memory[startPos+3])
		fmt.Printf(" %02x %02x %02x %02x", chip.Memory[startPos+4], chip.Memory[startPos+5], chip.Memory[startPos+6], chip.Memory[startPos+7])
		fmt.Printf(" %02x %02x %02x %02x", chip.Memory[startPos+8], chip.Memory[startPos+9], chip.Memory[startPos+10], chip.Memory[startPos+11])
		fmt.Printf(" %02x %02x %02x %02x", chip.Memory[startPos+12], chip.Memory[startPos+13], chip.Memory[startPos+14], chip.Memory[startPos+15])
		fmt.Println()
		startPos += 16
	}
}

func (chip *Chip8) DisplayDump() {
	drawHeader := func() {
		print("   |")
		for i := 0; i < DISPLAY_WIDTH; i++ {
			print("-")
		}
		println("|")
	}

	print("    ")
	for i := 0; i < DISPLAY_WIDTH; i++ {
		if i&0xf == 0 {
			fmt.Printf("%X", i&0xf0>>4)
		} else {
			print(" ")
		}
	}
	println()

	print("    ")
	for i := 0; i < DISPLAY_WIDTH; i++ {
		fmt.Printf("%X", i&0x0f)
	}
	println()

	drawHeader()
	for y := 0; y < DISPLAY_HEIGHT; y++ {
		fmt.Printf("%2X |", y)
		for x := 0; x < DISPLAY_WIDTH; x++ {
			if chip.DisplayBuffer[x+y*DISPLAY_WIDTH] {
				print("*")
			} else {
				print(" ")
			}
		}
		println("|")
	}
	drawHeader()
}

func (chip *Chip8) RegistryDump() {
	fmt.Println("Registry Dump:")
	fmt.Printf("PC:\t%04x\n", chip.Reg.PC)
	fmt.Printf("SP:\t%04x\n", chip.Reg.SP)
	fmt.Printf("T0:\t%x02\n", chip.Reg.T0)
	fmt.Printf("T1:\t%02x\n", chip.Reg.T1)
	fmt.Printf(" I:\t%04x\n", chip.Reg.I)
	fmt.Print(" V:\t")
	for i := 0; i < 16; i++ {
		fmt.Printf("%02x ", chip.Reg.V[Register(i)])
	}
	fmt.Println()
}
