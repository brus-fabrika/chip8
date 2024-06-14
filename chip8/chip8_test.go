package chip8_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/brus-fabrika/chip8/chip8"
)

func TestLoadRomFromData(t *testing.T) {
	displayTest := []uint8{0xa2, 0x0a, 0x61, 0x00, 0x62, 0x0a, 0xd1, 0x25, 0x12, 0x08, 0xf0, 0x90, 0xf0, 0x90, 0xf0, 0x00}

	ch := chip8.Chip8{}
	ch.Init(chip8.Chip_8)

	bytesLoaded, err := ch.LoadRomFromData(displayTest)

	if assert.NoError(t, err) {
		assert.Equal(t, uint16(len(displayTest)), bytesLoaded)
		assert.Equal(t, displayTest, ch.Memory[0x0200:0x200+len(displayTest)])
	}
}

func TestInit(t *testing.T) {
	ch := chip8.Chip8{}

	// set some rnd values for registers
	ch.Reg.V[0] = 0x55
	ch.Reg.V[0xF] = 0x01
	ch.Reg.SP = 0x0dfa
	ch.Reg.I = 0x0123
	ch.Reg.PC = 0x0123
	ch.Reg.T0 = 0x12
	ch.Reg.T1 = 0x34

	// set some rnd values for keyboard state
	ch.Keyboard[0] = true
	ch.Keyboard[10] = true
	ch.Keyboard[15] = true

	// set some rnd values for video buffer state
	ch.DisplayBuffer[0] = true
	ch.DisplayBuffer[123] = true
	ch.DisplayBuffer[255] = true

	// set some random state for memory
	bytes, err := ch.LoadRomFromData([]uint8{1, 2, 3, 4, 5, 6, 7, 8, 9, 10})

	if assert.NoError(t, err) {
		assert.Equal(t, uint16(bytes), uint16(10))
	}

	ch.Init(chip8.Chip_8)

	// make sure init sets register to correct initail values
	assert.Equal(t, chip8.MEMORY_USER, ch.Reg.PC)
	assert.Equal(t, chip8.MEMORY_STACK+47, ch.Reg.SP)
	assert.Equal(t, uint16(0), ch.Reg.I)
	assert.Equal(t, uint8(0), ch.Reg.T0)
	assert.Equal(t, uint8(0), ch.Reg.T1)
	assert.Equal(t, [16]uint8{}, ch.Reg.V)
	// make sure keyboard state reset
	assert.Equal(t, [16]bool{}, ch.Keyboard)
	// make sure video buffer state reset
	assert.Equal(t, [chip8.DISPLAY_WIDTH * chip8.DISPLAY_HEIGHT]bool{}, ch.DisplayBuffer)
	// make sure memory state reset, but skiping first 0x0200 bytes for now
	assert.Equal(t, make([]uint8, chip8.MEMORY_SIZE)[chip8.MEMORY_USER:], ch.Memory[chip8.MEMORY_USER:])

}

// TODO: add table test with all the registers and getRegValue
func TestMoveRegVal(t *testing.T) {
	ch := chip8.Chip8{}
	ch.Init(chip8.Chip_8)

	expected16 := uint16(0x1234)
	expected08 := uint8(0x34)

	ch.MovRegVal(chip8.RegI, expected16)
	ch.MovRegVal(chip8.RegT0, expected16)

	cmd := 0x6832
	ch.MovRegVal(chip8.Register(cmd&0x0f00>>8), uint16(cmd&0x00ff))

	assert.Equal(t, uint8(0x32), ch.Reg.V[8])

	assert.Equal(t, expected16, ch.Reg.I)
	assert.Equal(t, expected08, ch.Reg.T0)
	assert.Equal(t, uint16(0x0206), ch.Reg.PC)
}

func TestMoveRegReg(t *testing.T) {
	ch := chip8.Chip8{}
	ch.Init(chip8.Chip_8)

	expected16 := uint16(0x1234)
	expected08 := uint8(0x34)

	ch.MovRegVal(chip8.RegSP, expected16)

	assert.Equal(t, expected16, ch.Reg.SP)
	assert.Equal(t, uint16(0x0202), ch.Reg.PC)

	ch.MovRegReg(chip8.RegI, chip8.RegSP)
	ch.MovRegReg(chip8.RegVA, chip8.RegSP) // should this be possible?
	ch.MovRegReg(chip8.RegT1, chip8.RegVA)

	assert.Equal(t, expected16, ch.Reg.I)
	assert.Equal(t, expected08, ch.Reg.V[chip8.RegVA])
	assert.Equal(t, expected08, ch.Reg.T1)

	//ch.MovRegReg(chip8.RegSP, chip8.RegV0) // should this be possible? if yes, what is the result expected

}

func TestJump(t *testing.T) {
	ch := chip8.Chip8{}
	ch.Init(chip8.Chip_8)

	expected16 := uint16(0x0123)

	ch.Jump(expected16)
	assert.Equal(t, expected16, ch.Reg.PC)
	assert.Equal(t, chip8.MEMORY_STACK+47, ch.Reg.SP) // SP not affected by JMP
}

func TestJumpV(t *testing.T) {
	ch := chip8.Chip8{}
	ch.Init(chip8.Chip_8)

	expected16 := uint16(0x0123)

	ch.JumpV(expected16)

	assert.Equal(t, expected16, ch.Reg.PC)
	assert.Equal(t, chip8.MEMORY_STACK+47, ch.Reg.SP) // SP not affected by JMPV
}

func TestCall(t *testing.T) {
	ch := chip8.Chip8{}
	ch.Init(chip8.Chip_8)

	expected16 := uint16(0x0123)

	ch.Call(expected16)

	assert.Equal(t, chip8.MEMORY_STACK+45, ch.Reg.SP)
	assert.Equal(t, expected16, ch.Reg.PC)
}

func TestCallRet(t *testing.T) {
	ch := chip8.Chip8{}
	ch.Init(chip8.Chip_8)

	// emulate we are inside a function call
	// address of call instruction
	expected16 := uint16(0x0123)

	// set return address
	ch.Memory[chip8.MEMORY_STACK+23] = uint8(expected16)
	ch.Memory[chip8.MEMORY_STACK+22] = uint8(expected16 >> 8)

	// set stack to next empty position
	ch.Reg.SP = uint16(chip8.MEMORY_STACK + 21)
	ch.Reg.PC = uint16(0x0322) // no matter what the PC before RET

	ch.Ret()

	assert.Equal(t, chip8.MEMORY_STACK+23, ch.Reg.SP) // SP pointer is shifted back
	assert.Equal(t, expected16+2, ch.Reg.PC)          // instruction counter points to the next address after call insruction

}

func TestSkipEqualVal(t *testing.T) {
	ch := chip8.Chip8{}
	ch.Init(chip8.Chip_8)
	ch.Reg.PC = 0x0222 // emulate non start address
	ch.Reg.V[1] = 0x42

	expectedPcOnEqual := uint16(0x0226)

	t.Run("Equal", func(t *testing.T) {
		ch.SkipEqualVal(chip8.RegV1, 0x42)
		assert.Equal(t, expectedPcOnEqual, ch.Reg.PC)
	})

	expectedPcOnNonEqual := ch.Reg.PC + 2

	t.Run("NotEqual", func(t *testing.T) {
		ch.SkipEqualVal(chip8.RegV1, 0x43)
		assert.Equal(t, expectedPcOnNonEqual, ch.Reg.PC)
	})
}

func TestSkipNotEqualVal(t *testing.T) {
	ch := chip8.Chip8{}
	ch.Init(chip8.Chip_8)
	ch.Reg.PC = 0x0222 // emulate non start address
	ch.Reg.V[1] = 0x42

	expectedPcOnEqual := uint16(0x0224)

	t.Run("Equal", func(t *testing.T) {
		ch.SkipNotEqualVal(chip8.RegV1, 0x42)
		assert.Equal(t, expectedPcOnEqual, ch.Reg.PC)
	})

	expectedPcOnNonEqual := ch.Reg.PC + 4

	t.Run("NotEqual", func(t *testing.T) {
		ch.SkipNotEqualVal(chip8.RegV1, 0x43)
		assert.Equal(t, expectedPcOnNonEqual, ch.Reg.PC)
	})
}

func TestSkipEqualReg(t *testing.T) {
	ch := chip8.Chip8{}
	ch.Init(chip8.Chip_8)
	ch.Reg.PC = 0x0222 // emulate non start address
	ch.Reg.V[1] = 0x42
	ch.Reg.V[10] = 0x42

	expectedPcOnEqual := uint16(0x0226)

	t.Run("Equal", func(t *testing.T) {
		ch.SkipEqualReg(chip8.Register(0x0001), chip8.Register(0x000a))
		assert.Equal(t, expectedPcOnEqual, ch.Reg.PC)
	})

	expectedPcOnNonEqual := ch.Reg.PC + 2

	t.Run("NotEqual", func(t *testing.T) {
		ch.SkipEqualReg(0x0001, 0x0009)
		assert.Equal(t, expectedPcOnNonEqual, ch.Reg.PC)
	})
}

func TestSkipNotEqualReg(t *testing.T) {
	ch := chip8.Chip8{}
	ch.Init(chip8.Chip_8)
	ch.Reg.PC = 0x0222 // emulate non start address
	ch.Reg.V[1] = 0x42
	ch.Reg.V[10] = 0x42

	expectedPcOnEqual := uint16(0x0224)

	t.Run("Equal", func(t *testing.T) {
		ch.SkipNotEqualReg(chip8.Register(0x0001), chip8.Register(0x000a))
		assert.Equal(t, expectedPcOnEqual, ch.Reg.PC)
	})

	expectedPcOnNonEqual := ch.Reg.PC + 4

	t.Run("NotEqual", func(t *testing.T) {
		ch.SkipNotEqualReg(0x0001, 0x0009)
		assert.Equal(t, expectedPcOnNonEqual, ch.Reg.PC)
	})
}

func TestSkipKeyPressedAtReg(t *testing.T) {
	ch := chip8.Chip8{}
	ch.Init(chip8.Chip_8)

	// init setup
	ch.Reg.PC = 0x0234
	// key pressed
	ch.Keyboard[0x0A] = true
	// register set up
	ch.Reg.V[0x01] = 0x0A

	expectedPcOnPressed := uint16(0x0238)

	t.Run("Equal", func(t *testing.T) {
		ch.SkipKeyPressedAtReg(0x01)
		assert.Equal(t, expectedPcOnPressed, ch.Reg.PC)
	})

	expectedPcOnNotPressed := ch.Reg.PC + 2

	t.Run("Equal", func(t *testing.T) {
		ch.SkipKeyPressedAtReg(0x02)
		assert.Equal(t, expectedPcOnNotPressed, ch.Reg.PC)
	})
}

func TestSkipKeyNotPressedAtReg(t *testing.T) {
	ch := chip8.Chip8{}
	ch.Init(chip8.Chip_8)

	// init setup
	ch.Reg.PC = 0x0234
	// key pressed
	ch.Keyboard[0x0A] = true
	// register set up
	ch.Reg.V[0x01] = 0x0A

	expectedPcOnPressed := uint16(0x0236)

	t.Run("Equal", func(t *testing.T) {
		ch.SkipKeyNotPressedAtReg(0x01)
		assert.Equal(t, expectedPcOnPressed, ch.Reg.PC)
	})

	expectedPcOnNotPressed := ch.Reg.PC + 4

	t.Run("Equal", func(t *testing.T) {
		ch.SkipKeyNotPressedAtReg(0x02)
		assert.Equal(t, expectedPcOnNotPressed, ch.Reg.PC)
	})
}

func TestAddRegVal(t *testing.T) {

	testTable := []struct {
		Name       string
		Reg        chip8.Register
		RegVal     uint16
		AddVal     uint8
		Expected16 uint8
	}{
		{Name: "V5", Reg: chip8.RegV5, RegVal: uint16(0x0c), AddVal: uint8(0x09), Expected16: uint8(0x15)},
		{Name: "VA", Reg: chip8.RegVA, RegVal: uint16(0x0d), AddVal: uint8(0x0a), Expected16: uint8(0x17)},
		{Name: "VF", Reg: chip8.RegVF, RegVal: uint16(0x0e), AddVal: uint8(0x0b), Expected16: uint8(0x19)},
		{Name: "V0_overflow", Reg: chip8.RegV0, RegVal: uint16(0x0c), AddVal: uint8(0xf9), Expected16: uint8(0x05)},
		{Name: "V3_unchanged", Reg: chip8.RegV0, RegVal: uint16(0xbd), AddVal: uint8(0x00), Expected16: uint8(0xbd)},
		{Name: "V1_overToZero", Reg: chip8.RegV1, RegVal: uint16(0xff), AddVal: uint8(0x01), Expected16: uint8(0x00)},
	}

	ch := chip8.Chip8{}

	for _, tc := range testTable {
		t.Run(tc.Name, func(t *testing.T) {
			ch.Init(chip8.Chip_8)

			expectedPC := uint16(0x0204)

			ch.MovRegVal(tc.Reg, tc.RegVal)
			ch.AddRegVal(tc.Reg, tc.AddVal)

			assert.Equal(t, tc.Expected16, ch.Reg.V[tc.Reg])
			assert.Equal(t, expectedPC, ch.Reg.PC)

		})
	}
}

func TestAddRegReg(t *testing.T) {

	testTable := []struct {
		Name        string
		Reg1        chip8.Register
		RegVal1     uint16
		Reg2        chip8.Register
		RegVal2     uint16
		RegExpected uint8
		ExpectedOF  uint8
	}{
		{Name: "V1V2", Reg1: chip8.RegV1, RegVal1: uint16(0x0c), Reg2: chip8.RegV2, RegVal2: uint16(0x09), RegExpected: uint8(0x15), ExpectedOF: uint8(0)},
		{Name: "V1V2OF", Reg1: chip8.RegV1, RegVal1: uint16(0x0c), Reg2: chip8.RegV2, RegVal2: uint16(0xF9), RegExpected: uint8(0x05), ExpectedOF: uint8(1)},
		{Name: "V1V2TwoZero", Reg1: chip8.RegV1, RegVal1: uint16(0x00), Reg2: chip8.RegV2, RegVal2: uint16(0x00), RegExpected: uint8(0x0), ExpectedOF: uint8(0)},
		{Name: "V1V2OFtoZero", Reg1: chip8.RegV1, RegVal1: uint16(0x85), Reg2: chip8.RegV2, RegVal2: uint16(0x7b), RegExpected: uint8(0), ExpectedOF: uint8(1)},
		{Name: "VFInputNoCarry", Reg1: chip8.RegVF, RegVal1: uint16(0x85), Reg2: chip8.RegV2, RegVal2: uint16(0x5), RegExpected: uint8(0x00), ExpectedOF: uint8(0)},
		{Name: "VFInputCarry", Reg1: chip8.RegVF, RegVal1: uint16(0x85), Reg2: chip8.RegV2, RegVal2: uint16(0x7d), RegExpected: uint8(0x01), ExpectedOF: uint8(1)},
	}

	ch := chip8.Chip8{}

	for _, tc := range testTable {
		t.Run(tc.Name, func(t *testing.T) {
			ch.Init(chip8.Chip_8)

			expectedPC := uint16(0x0206)

			ch.MovRegVal(tc.Reg1, tc.RegVal1)
			ch.MovRegVal(tc.Reg2, tc.RegVal2)
			ch.AddRegReg(tc.Reg1, tc.Reg2)

			assert.Equal(t, tc.RegExpected, ch.Reg.V[tc.Reg1])
			assert.Equal(t, tc.ExpectedOF, ch.Reg.V[0x0F])
			assert.Equal(t, expectedPC, ch.Reg.PC)

		})
	}
}

func TestSubRegReg(t *testing.T) {

	testTable := []struct {
		Name        string
		Reg1        chip8.Register
		RegVal1     uint16
		Reg2        chip8.Register
		RegVal2     uint16
		RegExpected uint8
		ExpectedOF  uint8
	}{
		{Name: "V1V2", Reg1: chip8.RegV1, RegVal1: uint16(0x3c), Reg2: chip8.RegV2, RegVal2: uint16(0x19), RegExpected: uint8(0x23), ExpectedOF: uint8(1)},
		{Name: "V1V2OF", Reg1: chip8.RegV1, RegVal1: uint16(0x0c), Reg2: chip8.RegV2, RegVal2: uint16(0xF9), RegExpected: uint8(0x13), ExpectedOF: uint8(0)},
		{Name: "V1V2TwoZero", Reg1: chip8.RegV1, RegVal1: uint16(0x00), Reg2: chip8.RegV2, RegVal2: uint16(0x00), RegExpected: uint8(0x0), ExpectedOF: uint8(1)},
		{Name: "V1V2SubtoZero", Reg1: chip8.RegV1, RegVal1: uint16(0x85), Reg2: chip8.RegV2, RegVal2: uint16(0x85), RegExpected: uint8(0), ExpectedOF: uint8(1)},
		{Name: "VFCF", Reg1: chip8.RegVF, RegVal1: uint16(0x3c), Reg2: chip8.RegV2, RegVal2: uint16(0x19), RegExpected: uint8(0x01), ExpectedOF: uint8(1)},
		{Name: "VFNoCF", Reg1: chip8.RegVF, RegVal1: uint16(0x0c), Reg2: chip8.RegV2, RegVal2: uint16(0xF9), RegExpected: uint8(0x0), ExpectedOF: uint8(0)},
	}

	ch := chip8.Chip8{}

	for _, tc := range testTable {
		t.Run(tc.Name, func(t *testing.T) {
			ch.Init(chip8.Chip_8)

			expectedPC := uint16(0x0206)

			ch.MovRegVal(tc.Reg1, tc.RegVal1)
			ch.MovRegVal(tc.Reg2, tc.RegVal2)
			ch.SubRegReg(tc.Reg1, tc.Reg2)

			assert.Equal(t, tc.RegExpected, ch.Reg.V[tc.Reg1])
			assert.Equal(t, tc.ExpectedOF, ch.Reg.V[0x0F])
			assert.Equal(t, expectedPC, ch.Reg.PC)

		})
	}
}

func TestSubNegRegReg(t *testing.T) {

	testTable := []struct {
		Name        string
		Reg1        chip8.Register
		RegVal1     uint16
		Reg2        chip8.Register
		RegVal2     uint16
		RegExpected uint8
		ExpectedOF  uint8
	}{
		{Name: "V1V2", Reg1: chip8.RegV1, RegVal1: uint16(0x3c), Reg2: chip8.RegV2, RegVal2: uint16(0x19), RegExpected: uint8(0xDD), ExpectedOF: uint8(0)},
		{Name: "V1V2OF", Reg1: chip8.RegV1, RegVal1: uint16(0x0c), Reg2: chip8.RegV2, RegVal2: uint16(0xF9), RegExpected: uint8(0xED), ExpectedOF: uint8(1)},
		{Name: "V1V2TwoZero", Reg1: chip8.RegV1, RegVal1: uint16(0x00), Reg2: chip8.RegV2, RegVal2: uint16(0x00), RegExpected: uint8(0x0), ExpectedOF: uint8(1)},
		{Name: "V1V2SubtoZero", Reg1: chip8.RegV1, RegVal1: uint16(0x85), Reg2: chip8.RegV2, RegVal2: uint16(0x85), RegExpected: uint8(0), ExpectedOF: uint8(1)},
		{Name: "VFNoCF", Reg1: chip8.RegVF, RegVal1: uint16(0x3c), Reg2: chip8.RegV2, RegVal2: uint16(0x19), RegExpected: uint8(0x00), ExpectedOF: uint8(0)},
		{Name: "VFCF", Reg1: chip8.RegVF, RegVal1: uint16(0x0c), Reg2: chip8.RegV2, RegVal2: uint16(0xF9), RegExpected: uint8(0x01), ExpectedOF: uint8(1)},
	}

	ch := chip8.Chip8{}

	for _, tc := range testTable {
		t.Run(tc.Name, func(t *testing.T) {
			ch.Init(chip8.Chip_8)

			expectedPC := uint16(0x0206)

			ch.MovRegVal(tc.Reg1, tc.RegVal1)
			ch.MovRegVal(tc.Reg2, tc.RegVal2)
			ch.SubNegRegReg(tc.Reg1, tc.Reg2)

			assert.Equal(t, tc.RegExpected, ch.Reg.V[tc.Reg1])
			assert.Equal(t, tc.ExpectedOF, ch.Reg.V[0x0F])
			assert.Equal(t, expectedPC, ch.Reg.PC)

		})
	}
}

func TestOr(t *testing.T) {

	testTable := []struct {
		Name        string
		Reg1        chip8.Register
		RegVal1     uint16
		Reg2        chip8.Register
		RegVal2     uint16
		RegExpected uint8
	}{
		{Name: "V1V2_1", Reg1: chip8.RegV1, RegVal1: uint16(0b01010101), Reg2: chip8.RegV2, RegVal2: uint16(0b10101010), RegExpected: uint8(0xFF)},
		{Name: "V1V2_2", Reg1: chip8.RegV1, RegVal1: uint16(0b00010001), Reg2: chip8.RegV2, RegVal2: uint16(0b00001111), RegExpected: uint8(0x1F)},
		{Name: "V1V2_3", Reg1: chip8.RegV1, RegVal1: uint16(0xa0), Reg2: chip8.RegV2, RegVal2: uint16(0xaa), RegExpected: uint8(0xaa)},
	}

	ch := chip8.Chip8{}

	for _, tc := range testTable {
		t.Run(tc.Name, func(t *testing.T) {
			ch.Init(chip8.Chip_8)
			ch.Reg.V[0x0F] = 0xF0 // some random values to check CF register reset

			expectedPC := uint16(0x0206)

			ch.MovRegVal(tc.Reg1, tc.RegVal1)
			ch.MovRegVal(tc.Reg2, tc.RegVal2)
			ch.Or(tc.Reg1, tc.Reg2)

			assert.Equal(t, tc.RegExpected, ch.Reg.V[tc.Reg1])
			assert.Equal(t, uint8(0x00), ch.Reg.V[0x0F])
			assert.Equal(t, expectedPC, ch.Reg.PC)
		})
	}
}

func TestXor(t *testing.T) {

	testTable := []struct {
		Name        string
		Reg1        chip8.Register
		RegVal1     uint16
		Reg2        chip8.Register
		RegVal2     uint16
		RegExpected uint8
	}{
		{Name: "V1V2_1", Reg1: chip8.RegV1, RegVal1: uint16(0b01010101), Reg2: chip8.RegV2, RegVal2: uint16(0b10101010), RegExpected: uint8(0xFF)},
		{Name: "V1V2_2", Reg1: chip8.RegV1, RegVal1: uint16(0b00010001), Reg2: chip8.RegV2, RegVal2: uint16(0b00001111), RegExpected: uint8(0x1E)},
		{Name: "V1V2_3", Reg1: chip8.RegV1, RegVal1: uint16(0xF0), Reg2: chip8.RegV2, RegVal2: uint16(0xFF), RegExpected: uint8(0x0F)},
	}

	ch := chip8.Chip8{}

	for _, tc := range testTable {
		t.Run(tc.Name, func(t *testing.T) {
			ch.Init(chip8.Chip_8)
			ch.Reg.V[0x0F] = 0xF0 // some random value to check CF register reset
			expectedPC := uint16(0x0206)

			ch.MovRegVal(tc.Reg1, tc.RegVal1)
			ch.MovRegVal(tc.Reg2, tc.RegVal2)
			ch.Xor(tc.Reg1, tc.Reg2)

			assert.Equal(t, tc.RegExpected, ch.Reg.V[tc.Reg1])
			assert.Equal(t, expectedPC, ch.Reg.PC)
			assert.Equal(t, uint8(0x00), ch.Reg.V[0x0F])
		})
	}
}

func TestShiftR(t *testing.T) {

	testTable := []struct {
		Name        string
		RegDst      chip8.Register
		RegSrc      chip8.Register
		RegVal      uint16
		RegExpected uint8
		ExpectedOF  uint8
	}{
		{Name: "V1_1", RegDst: chip8.RegV1, RegSrc: chip8.RegV2, RegVal: uint16(0b01010101), RegExpected: uint8(0b00101010), ExpectedOF: uint8(1)},
		{Name: "V1_2", RegDst: chip8.RegV1, RegSrc: chip8.RegV2, RegVal: uint16(0b00000001), RegExpected: uint8(0x0), ExpectedOF: uint8(1)},
		{Name: "V1_3", RegDst: chip8.RegV1, RegSrc: chip8.RegV2, RegVal: uint16(0xF0), RegExpected: uint8(0x78), ExpectedOF: uint8(0)},
		{Name: "VF_NoOF", RegDst: chip8.RegVF, RegSrc: chip8.RegV2, RegVal: uint16(0xF0), RegExpected: uint8(0x00), ExpectedOF: uint8(0)},
		{Name: "VF_OF", RegDst: chip8.RegVF, RegVal: uint16(0xF1), RegExpected: uint8(0x01), ExpectedOF: uint8(1)},
	}

	testTable_NotChip_8 := []struct {
		Name        string
		RegDst      chip8.Register
		RegSrc      chip8.Register
		RegVal      uint16
		RegExpected uint8
		ExpectedOF  uint8
	}{
		{Name: "V1_1", RegDst: chip8.RegV1, RegSrc: chip8.RegV2, RegVal: uint16(0b01010101), RegExpected: uint8(0b00101010), ExpectedOF: uint8(1)},
		{Name: "V1_2", RegDst: chip8.RegV1, RegSrc: chip8.RegV2, RegVal: uint16(0b00000001), RegExpected: uint8(0x0), ExpectedOF: uint8(1)},
		{Name: "V1_3", RegDst: chip8.RegV1, RegSrc: chip8.RegV2, RegVal: uint16(0xF0), RegExpected: uint8(0x78), ExpectedOF: uint8(0)},
		{Name: "VF_NoOF", RegDst: chip8.RegVF, RegSrc: chip8.RegV2, RegVal: uint16(0xF0), RegExpected: uint8(0x00), ExpectedOF: uint8(0)},
		{Name: "VF_OF", RegDst: chip8.RegVF, RegVal: uint16(0xF1), RegExpected: uint8(0x01), ExpectedOF: uint8(1)},
	}

	ch := chip8.Chip8{}

	for _, tc := range testTable {
		t.Run(tc.Name, func(t *testing.T) {
			ch.Init(chip8.Chip_8)

			expectedPC := uint16(0x0204)

			ch.MovRegVal(tc.RegSrc, tc.RegVal)
			ch.ShiftR(tc.RegDst, tc.RegSrc)

			assert.Equal(t, tc.RegVal, uint16(ch.Reg.V[tc.RegSrc]))
			assert.Equal(t, tc.RegExpected, ch.Reg.V[tc.RegDst])
			assert.Equal(t, tc.ExpectedOF, ch.Reg.V[0x0F])
			assert.Equal(t, expectedPC, ch.Reg.PC)
		})
	}

	for _, tc := range testTable_NotChip_8 {
		t.Run(tc.Name, func(t *testing.T) {
			ch.Init(chip8.XO_Chip)

			expectedPC := uint16(0x0206)
			// Non-Chip8 versions only work on X register (dst in this case)
			ch.MovRegVal(tc.RegDst, tc.RegVal)
			ch.MovRegVal(tc.RegSrc, uint16(0xAA)) // some random value to test not changed reg
			ch.ShiftR(tc.RegDst, tc.RegSrc)

			assert.Equal(t, uint8(0xAA), ch.Reg.V[tc.RegSrc])
			assert.Equal(t, tc.RegExpected, ch.Reg.V[tc.RegDst])
			assert.Equal(t, tc.ExpectedOF, ch.Reg.V[0x0F])
			assert.Equal(t, expectedPC, ch.Reg.PC)
		})
	}
}

func TestShiftL(t *testing.T) {

	testTable := []struct {
		Name        string
		RegDst      chip8.Register
		RegSrc      chip8.Register
		RegVal      uint16
		RegExpected uint8
		ExpectedOF  uint8
	}{
		{Name: "V1_1", RegDst: chip8.RegV1, RegSrc: chip8.RegV2, RegVal: uint16(0b01010101), RegExpected: uint8(0b10101010), ExpectedOF: uint8(0)},
		{Name: "V1_2_OF", RegDst: chip8.RegV1, RegSrc: chip8.RegV2, RegVal: uint16(0b11010101), RegExpected: uint8(0xaa), ExpectedOF: uint8(1)},
		{Name: "V1_3", RegDst: chip8.RegV1, RegSrc: chip8.RegV2, RegVal: uint16(0b00000001), RegExpected: uint8(0x2), ExpectedOF: uint8(0)},
		{Name: "V1_4_OF", RegDst: chip8.RegV1, RegSrc: chip8.RegV2, RegVal: uint16(0xF0), RegExpected: uint8(0xE0), ExpectedOF: uint8(1)},
		{Name: "VF_OF", RegDst: chip8.RegVF, RegSrc: chip8.RegV2, RegVal: uint16(0xF0), RegExpected: uint8(0x01), ExpectedOF: uint8(1)},
		{Name: "VF_NoOF", RegDst: chip8.RegVF, RegSrc: chip8.RegV2, RegVal: uint16(0x70), RegExpected: uint8(0x00), ExpectedOF: uint8(0)},
	}

	testTable_NotChip_8 := []struct {
		Name        string
		RegDst      chip8.Register
		RegSrc      chip8.Register
		RegVal      uint16
		RegExpected uint8
		ExpectedOF  uint8
	}{
		{Name: "V1_1", RegDst: chip8.RegV1, RegSrc: chip8.RegV2, RegVal: uint16(0b01010101), RegExpected: uint8(0b10101010), ExpectedOF: uint8(0)},
		{Name: "V1_2_OF", RegDst: chip8.RegV1, RegSrc: chip8.RegV2, RegVal: uint16(0b11010101), RegExpected: uint8(0xaa), ExpectedOF: uint8(1)},
		{Name: "V1_3", RegDst: chip8.RegV1, RegSrc: chip8.RegV2, RegVal: uint16(0b00000001), RegExpected: uint8(0x2), ExpectedOF: uint8(0)},
		{Name: "V1_4_OF", RegDst: chip8.RegV1, RegSrc: chip8.RegV2, RegVal: uint16(0xF0), RegExpected: uint8(0xE0), ExpectedOF: uint8(1)},
		{Name: "VF_OF", RegDst: chip8.RegVF, RegSrc: chip8.RegV2, RegVal: uint16(0xF0), RegExpected: uint8(0x01), ExpectedOF: uint8(1)},
		{Name: "VF_NoOF", RegDst: chip8.RegVF, RegSrc: chip8.RegV2, RegVal: uint16(0x70), RegExpected: uint8(0x00), ExpectedOF: uint8(0)},
	}

	ch := chip8.Chip8{}

	for _, tc := range testTable {
		t.Run(tc.Name, func(t *testing.T) {
			ch.Init(chip8.Chip_8)

			expectedPC := uint16(0x0204)

			ch.MovRegVal(tc.RegSrc, tc.RegVal)
			ch.ShiftL(tc.RegDst, tc.RegSrc)

			assert.Equal(t, tc.RegVal, uint16(ch.Reg.V[tc.RegSrc]))
			assert.Equal(t, tc.RegExpected, ch.Reg.V[tc.RegDst])
			assert.Equal(t, tc.ExpectedOF, ch.Reg.V[0x0F])
			assert.Equal(t, expectedPC, ch.Reg.PC)
		})
	}

	for _, tc := range testTable_NotChip_8 {
		t.Run(tc.Name, func(t *testing.T) {
			ch.Init(chip8.XO_Chip)

			expectedPC := uint16(0x0206)

			ch.MovRegVal(tc.RegSrc, uint16(0xAA))
			ch.MovRegVal(tc.RegDst, tc.RegVal)
			ch.ShiftL(tc.RegDst, tc.RegSrc)

			assert.Equal(t, uint8(0xAA), ch.Reg.V[tc.RegSrc])
			assert.Equal(t, tc.RegExpected, ch.Reg.V[tc.RegDst])
			assert.Equal(t, tc.ExpectedOF, ch.Reg.V[0x0F])
			assert.Equal(t, expectedPC, ch.Reg.PC)
		})
	}
}

func TestBcdReg(t *testing.T) {

	testTable := []struct {
		Name    string
		Reg     chip8.Register
		BcdVal  uint16
		MemAddr uint16
		Val1    uint8
		Val2    uint8
		Val3    uint8
	}{
		{Name: "V1_179", Reg: chip8.RegV1, BcdVal: uint16(179), MemAddr: uint16(0x220), Val1: uint8(1), Val2: uint8(7), Val3: uint8(9)},
		{Name: "V1_140", Reg: chip8.RegV1, BcdVal: uint16(140), MemAddr: uint16(0x220), Val1: uint8(1), Val2: uint8(4), Val3: uint8(0)},
		{Name: "V1_205", Reg: chip8.RegV1, BcdVal: uint16(205), MemAddr: uint16(0x220), Val1: uint8(2), Val2: uint8(0), Val3: uint8(5)},
		{Name: "V1_69", Reg: chip8.RegV1, BcdVal: uint16(69), MemAddr: uint16(0x220), Val1: uint8(0), Val2: uint8(6), Val3: uint8(9)},
		{Name: "V1_200", Reg: chip8.RegV1, BcdVal: uint16(200), MemAddr: uint16(0x220), Val1: uint8(2), Val2: uint8(0), Val3: uint8(0)},
		{Name: "V1_2", Reg: chip8.RegV1, BcdVal: uint16(2), MemAddr: uint16(0x220), Val1: uint8(0), Val2: uint8(0), Val3: uint8(2)},
		{Name: "V1_30", Reg: chip8.RegV1, BcdVal: uint16(30), MemAddr: uint16(0x220), Val1: uint8(0), Val2: uint8(3), Val3: uint8(0)},
		{Name: "V1_0", Reg: chip8.RegV1, BcdVal: uint16(0), MemAddr: uint16(0x220), Val1: uint8(0), Val2: uint8(0), Val3: uint8(0)},
	}

	ch := chip8.Chip8{}

	for _, tc := range testTable {
		t.Run(tc.Name, func(t *testing.T) {
			ch.Init(chip8.Chip_8)

			expectedPC := uint16(0x0206)

			ch.MovRegVal(chip8.RegI, tc.MemAddr)
			ch.MovRegVal(tc.Reg, tc.BcdVal)
			ch.BcdReg(tc.Reg)

			assert.Equal(t, tc.Val1, ch.Memory[ch.Reg.I+uint16(0)])
			assert.Equal(t, tc.Val2, ch.Memory[ch.Reg.I+uint16(1)])
			assert.Equal(t, tc.Val3, ch.Memory[ch.Reg.I+uint16(2)])

			assert.Equal(t, expectedPC, ch.Reg.PC)
		})
	}

	t.Run("BcdOpCode", func(t *testing.T) {
		expectedPC := uint16(0x0206)
		ch.Init(chip8.Chip_8)
		ch.ProcessCmd(0xA220) // MOV I, 0x0220
		ch.ProcessCmd(0x61B3) // MOV V1, 0xB3 (179)
		ch.ProcessCmd(0xF133) // BCD V1

		assert.Equal(t, uint8(1), ch.Memory[ch.Reg.I+uint16(0)])
		assert.Equal(t, uint8(7), ch.Memory[ch.Reg.I+uint16(1)])
		assert.Equal(t, uint8(9), ch.Memory[ch.Reg.I+uint16(2)])

		assert.Equal(t, expectedPC, ch.Reg.PC)
	})
}
