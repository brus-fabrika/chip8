package main

import (
	"fmt"
	"math"

	"github.com/veandco/go-sdl2/sdl"

	"github.com/brus-fabrika/chip8/chip8"
)

const (
	SCREEN_WIDTH         = 640
	SCREEN_HEIGHT        = 320
	SCREEN_FG_COLOR      = 0x00C800
	SCREEN_FG_COLOR2     = 0xC80000
	SCREEN_BG_COLOR      = 0x0
	USE_FIXED_FRAMERATE  = true
	FRAMERATE            = 60
	INSTRUCTIONS_PER_SEC = 500
)

//var displayTest = []uint8{0xa2, 0x0a, 0x61, 0x00, 0x62, 0x0a, 0xd1, 0x25, 0x12, 0x08, 0xf0, 0x90, 0xf0, 0x90, 0xf0, 0x00}

// flags are OK
//var romFile = ".\\bin\\4-flags.ch8"

// corax is OK
//var romFile = ".\\bin\\3-corax+.ch8"

// quirks for Chip_8 are OK except display wait
//var romFile = ".\\bin\\5-quirks.ch8"

//var romFile = ".\\bin\\6-keypad.ch8"

var romFile = ".\\bin\\tetris.ch8"

func main() {
	e := Engine{}

	if err := e.Init(); err != nil {
		e.Destroy()
		panic(err)
	}
	defer e.Destroy()

	e.Window.SetTitle(romFile)

	println("Hello from CHIP8")
	println("Memory layout:")
	fmt.Printf("\tUser memory      : 0x%x - 0x%x\n", chip8.MEMORY_USER, chip8.MEMORY_STACK-1)
	fmt.Printf("\tStack area       : 0x%x - 0x%0x\n", chip8.MEMORY_STACK, chip8.MEMORY_INT_AREA-1)
	fmt.Printf("\tInterpreter  area: 0x%x - 0x%0x\n", chip8.MEMORY_INT_AREA, chip8.MEMORY_REG_AREA-1)
	fmt.Printf("\tRegisters        : 0x%x - 0x%0x\n", chip8.MEMORY_REG_AREA, chip8.MEMORY_DISPLAY-1)
	fmt.Printf("\tUser memory start: 0x%x - 0x%0x\n", chip8.MEMORY_DISPLAY, chip8.MEMORY_SIZE-1)
	chip := chip8.Chip8{}
	chip.Init(chip8.Chip_8)
	//chip.LoadRomFromFile(".\\bin\\IbmLogo.ch8")
	chip.LoadRomFromFile(romFile)
	//chip.LoadRomFromData(displayTest)
	chip.MemoryDump(0x0200, 0x0600)
	//chip.Execute()
	//chip.DisplayDump()

	perfFreq := float64(sdl.GetPerformanceFrequency())

	for chip.State.Running {
		// clear the current keyboard state
		//chip.ClearKeyboard()

		HandleEvent(&chip)

		if !chip.State.Running {
			// if running disabled - exit processing cycle
			break
		}

		if chip.State.Paused {
			// we still need to make a delay, otherwise huge cpu consumption in PollEvent
			sdl.Delay(50)
			continue
		}

		start := sdl.GetPerformanceCounter()

		for i := 0; i < int(INSTRUCTIONS_PER_SEC/FRAMERATE); i++ {
			cmd := uint16(chip.Memory[int(chip.Reg.PC)])<<8 + uint16(chip.Memory[int(chip.Reg.PC+1)])
			chip.ProcessCmd(cmd)
		}

		end := sdl.GetPerformanceCounter()
		elapsed := float64(end-start) / perfFreq * 1000.0

		if USE_FIXED_FRAMERATE {
			if 1000.0/FRAMERATE > elapsed {
				sdl.Delay(uint32(math.Floor(1000.0/FRAMERATE - elapsed)))
			}
		}

		UpdateDisplay(&e, &chip)
		UpdateTimer(&chip)
	}
}

func UpdateTimer(chip *chip8.Chip8) {
	if chip.Reg.T0 > 0 {
		chip.Reg.T0--
	}

	if chip.Reg.T1 > 0 {
		chip.Reg.T1--
	}
}

func UpdateDisplay(e *Engine, chip *chip8.Chip8) {
	var fg_r uint8 = (SCREEN_FG_COLOR & 0xFF0000) >> 16
	var fg_g uint8 = (SCREEN_FG_COLOR & 0x00FF00) >> 8
	var fg_b uint8 = (SCREEN_FG_COLOR & 0x0000FF)

	var fg2_r uint8 = (SCREEN_FG_COLOR2 & 0xFF0000) >> 16
	var fg2_g uint8 = (SCREEN_FG_COLOR2 & 0x00FF00) >> 8
	var fg2_b uint8 = (SCREEN_FG_COLOR2 & 0x0000FF)

	var bg_r uint8 = (SCREEN_BG_COLOR & 0xFF0000) >> 16
	var bg_g uint8 = (SCREEN_BG_COLOR & 0x00FF00) >> 8
	var bg_b uint8 = (SCREEN_BG_COLOR & 0x0000FF)

	for y := 0; y < chip8.DISPLAY_HEIGHT; y++ {
		for x := 0; x < chip8.DISPLAY_WIDTH; x++ {
			rect := sdl.Rect{X: int32(x * 10), Y: int32(y * 10), W: 10, H: 10}
			if chip.DisplayBuffer[x+y*chip8.DISPLAY_WIDTH] {
				e.Renderer.SetDrawColor(fg_r, fg_g, fg_b, 255)
				e.Renderer.FillRect(&rect)
				e.Renderer.SetDrawColor(fg2_r, fg2_g, fg2_b, 255)
				e.Renderer.DrawRect(&rect)

			} else {
				// we should clear the "pixel" otherwise, or it won't be re-drawn never ever
				e.Renderer.SetDrawColor(bg_r, bg_g, bg_b, 255)
				e.Renderer.FillRect(&rect)
			}
		}
	}

	e.Renderer.Present()
}

func HandleEvent(chip *chip8.Chip8) {
	for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
		switch event := event.(type) {
		case *sdl.QuitEvent:
			println("Quit")
			chip.State.Running = false
		case *sdl.KeyboardEvent:
			HandleKeyboardEvent(chip, event)
		}
	}
}

func HandleKeyboardEvent(chip *chip8.Chip8, event *sdl.KeyboardEvent) {
	if event.Type == sdl.KEYUP {
		// clear the current keyboard state
		// but only on key up
		//chip.ClearKeyboard()
		switch event.Keysym.Sym {
		case sdl.K_1:
			chip.Keyboard[0x01] = false
		case sdl.K_2:
			chip.Keyboard[0x02] = false
		case sdl.K_3:
			chip.Keyboard[0x03] = false
		case sdl.K_4:
			chip.Keyboard[0x0C] = false
		case sdl.K_q:
			chip.Keyboard[0x04] = false
		case sdl.K_w:
			chip.Keyboard[0x05] = false
		case sdl.K_e:
			chip.Keyboard[0x06] = false
		case sdl.K_r:
			chip.Keyboard[0x0D] = false
		case sdl.K_a:
			chip.Keyboard[0x07] = false
		case sdl.K_s:
			chip.Keyboard[0x08] = false
		case sdl.K_d:
			chip.Keyboard[0x09] = false
		case sdl.K_f:
			chip.Keyboard[0x0E] = false
		case sdl.K_z:
			chip.Keyboard[0x0A] = false
		case sdl.K_x:
			chip.Keyboard[0x00] = false
		case sdl.K_c:
			chip.Keyboard[0x0B] = false
		case sdl.K_v:
			chip.Keyboard[0x0F] = false
		}
	}
	if event.Type == sdl.KEYDOWN {
		switch event.Keysym.Sym {
		case sdl.K_ESCAPE:
			println("Quit")
			chip.State.Running = false
		case sdl.K_SPACE:
			chip.State.Paused = !chip.State.Paused
			//if !paused {
			//frameCounter = 0
			//currentFrameCounter = 0
			//}
		}
	}
	if event.Type == sdl.KEYDOWN {
		switch event.Keysym.Sym {
		case sdl.K_1:
			chip.Keyboard[0x01] = true
		case sdl.K_2:
			chip.Keyboard[0x02] = true
		case sdl.K_3:
			chip.Keyboard[0x03] = true
		case sdl.K_4:
			chip.Keyboard[0x0C] = true
		case sdl.K_q:
			chip.Keyboard[0x04] = true
		case sdl.K_w:
			chip.Keyboard[0x05] = true
		case sdl.K_e:
			chip.Keyboard[0x06] = true
		case sdl.K_r:
			chip.Keyboard[0x0D] = true
		case sdl.K_a:
			chip.Keyboard[0x07] = true
		case sdl.K_s:
			chip.Keyboard[0x08] = true
		case sdl.K_d:
			chip.Keyboard[0x09] = true
		case sdl.K_f:
			chip.Keyboard[0x0E] = true
		case sdl.K_z:
			chip.Keyboard[0x0A] = true
		case sdl.K_x:
			chip.Keyboard[0x00] = true
		case sdl.K_c:
			chip.Keyboard[0x0B] = true
		case sdl.K_v:
			chip.Keyboard[0x0F] = true
		}
	}
}
