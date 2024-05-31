package main

import (
	"fmt"

	"github.com/veandco/go-sdl2/sdl"

	"github.com/brus-fabrika/chip8/chip8"
)

const (
	SCREEN_WIDTH        = 640
	SCREEN_HEIGHT       = 320
	FRAMERATE           = 30.0
	USE_FIXED_FRAMERATE = false
)

var displayTest = []uint8{0xa2, 0x0a, 0x61, 0x00, 0x62, 0x0a, 0xd1, 0x25, 0x12, 0x08, 0xf0, 0x90, 0xf0, 0x90, 0xf0, 0x00}

func main() {
	e := Engine{}

	if err := e.Init(); err != nil {
		e.Destroy()
		panic(err)
	}
	defer e.Destroy()

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
	//chip.Execute()
	//chip.DisplayDump()

	running := true
	for running {
		for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
			switch event.(type) {
			case *sdl.QuitEvent:
				println("Quit")
				running = false
			}
		}

		cmd := uint16(chip.Memory[int(chip.Reg.PC)])<<8 + uint16(chip.Memory[int(chip.Reg.PC+1)])
		chip.ProcessCmd(cmd)

		for y := 0; y < chip8.DISPLAY_HEIGHT; y++ {
			for x := 0; x < chip8.DISPLAY_WIDTH; x++ {
				if chip.DisplayBuffer[x+y*chip8.DISPLAY_WIDTH] {
					rect := sdl.Rect{X: int32(x * 10), Y: int32(y * 10), W: 10, H: 10}
					e.Renderer.SetDrawColor(0, 200, 0, 255)
					e.Renderer.FillRect(&rect)
					e.Renderer.SetDrawColor(200, 0, 0, 255)
					e.Renderer.DrawRect(&rect)

				}
			}
		}

		e.Renderer.Present()

		sdl.Delay(500)
	}
}
