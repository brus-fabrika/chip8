package main

import (
	"fmt"

	"github.com/brus-fabrika/chip8/chip8"
)

var displayTest = []uint8{0xa2, 0x0a, 0x61, 0x00, 0x62, 0x0a, 0xd1, 0x25, 0x12, 0x08, 0xf0, 0x90, 0xf0, 0x90, 0xf0, 0x00}

func main() {
	println("Hello from CHIP8")
	println("Memory layout:")
	fmt.Printf("\tUser memory      : 0x%x - 0x%x\n", chip8.MEMORY_USER, chip8.MEMORY_STACK-1)
	fmt.Printf("\tStack area       : 0x%x - 0x%0x\n", chip8.MEMORY_STACK, chip8.MEMORY_INT_AREA-1)
	fmt.Printf("\tInterpreter  area: 0x%x - 0x%0x\n", chip8.MEMORY_INT_AREA, chip8.MEMORY_REG_AREA-1)
	fmt.Printf("\tRegisters        : 0x%x - 0x%0x\n", chip8.MEMORY_REG_AREA, chip8.MEMORY_DISPLAY-1)
	fmt.Printf("\tUser memory start: 0x%x - 0x%0x\n", chip8.MEMORY_DISPLAY, chip8.MEMORY_SIZE-1)
	chip := chip8.Chip8{}
	chip.Init()
	chip.LoadRomFromFile(".\\bin\\IbmLogo.ch8")
	//chip.LoadRomFromData(displayTest)
	chip.MemoryDump(0x0200, 0x0300)
	chip.Execute()
	chip.DisplayDump()
}
