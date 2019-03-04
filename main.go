package main

import (
	"bufio"
	"fmt"
	"math/rand"
	"os"

	"github.com/veandco/go-sdl2/sdl"
)

/**************************************
 *
 *
 *    SETUP THE CPU
 *
 *
 *************************************/
// setting up the cpu.
// to store current opcode
var opcode uint16

// stores the Chip-8 memeory
var memory = make([]byte, 4096)

//CPU registers
var V = make([]byte, 16)

// index register
var I uint16

// program counter
var pc uint16

/*
   How the CHIP-8 Memory map is:
   0x000-0x1FF - Chip 8 interpreter (contains font set in emu)
   0x050-0x0A0 - Used for the built in 4x5 pixel font set (0-F)
   0x200-0xFFF - Program ROM and work RAM
*/
// store the screen;
var gfx [32][64]uint8
var draw_flag = false

// timer registers that count at 60Hz, when set above zero they count down
// to zero;
var delay_timer uint16
var sound_timer uint16

// setup the stack
var stack = make([]uint16, 16)
var sp uint16

// Chip 8 has a HEX based keypad (0x0-0xF)
var key = make([]byte, 16)

// font set
var chip8_fontset = []byte{
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

func main() {

	if len(os.Args) < 2 {
		fmt.Println("Please pass in a rom file...")
		os.Exit(0)
	}

	fileArg := os.Args[1]

	/**************************************
	 *
	 *
	 *    INIT THE CPU
	 *
	 *
	 *************************************/

	pc = 0x200
	opcode = 0
	I = 0
	sp = 0

	// Clear display
	// Clear stack
	// Clear registers V0-VF
	// Clear memory

	// load fontset

	for i := 0; i < 80; i++ {
		memory[i] = chip8_fontset[i]
	}

	// reset timers
	fmt.Printf("Running File: %v", fileArg)
	// load file into memory;
	f, err := os.Open(fileArg)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	defer f.Close()

	var rbyte []byte = make([]byte, 1)
	reader := bufio.NewReader(f)
	var j = 0
	for {
		// read byte by byte
		_, err := reader.Read(rbyte)
		// check for end of file.
		if err != nil {
			if err.Error() == "EOF" {
				break
			} else {
				fmt.Println(err.Error())
				os.Exit(1)
			}
		}
		// load into memroy;
		memory[j+512] = rbyte[0]
		j += 1
	}

	/**************************************
	 *
	 *
	 *    SETUP THE DRAWING STUFF
	 *
	 *
	 *************************************/
	if err := sdl.Init(sdl.INIT_EVERYTHING); err != nil {
		panic(err)
	}
	defer sdl.Quit()

	window, err := sdl.CreateWindow("Gate - Chip8 Emulator In Golang", sdl.WINDOWPOS_UNDEFINED, sdl.WINDOWPOS_UNDEFINED,
		640, 320, sdl.WINDOW_SHOWN)
	if err != nil {
		panic(err)
	}
	defer window.Destroy()

	surface, err := window.GetSurface()
	if err != nil {
		panic(err)
	}
	//surface.FillRect(nil, 0)

	//rect := sdl.Rect{0, 0, 200, 200}
	//surface.FillRect(&rect, 0xffff0000)

	// before we update the surface, we should probably
	window.UpdateSurface()

	running := true
	/**************************************
	 *
	 *
	 *    MAIN EMULATION/RENDER LOOP
	 *
	 *
	 *************************************/
	for running {

		// emulate one cycle
		emulateCycle()
		// update screen if needed
		if draw_flag {
			draw(surface)
		}
		// set key press state

		for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
			switch ev := event.(type) {
			case *sdl.QuitEvent:
				println("Quit")
				running = false
				break
			case *sdl.KeyboardEvent:
				if ev.Type == sdl.KEYUP {
					switch ev.Keysym.Sym {
					case sdl.K_1:
						keyPress(0x1, false)
					case sdl.K_2:
						keyPress(0x2, false)
					case sdl.K_3:
						keyPress(0x3, false)
					case sdl.K_4:
						keyPress(0xC, false)
					case sdl.K_q:
						keyPress(0x4, false)
					case sdl.K_w:
						keyPress(0x5, false)
					case sdl.K_e:
						keyPress(0x6, false)
					case sdl.K_r:
						keyPress(0xD, false)
					case sdl.K_a:
						keyPress(0x7, false)
					case sdl.K_s:
						keyPress(0x8, false)
					case sdl.K_d:
						keyPress(0x9, false)
					case sdl.K_f:
						keyPress(0xE, false)
					case sdl.K_z:
						keyPress(0xA, false)
					case sdl.K_x:
						keyPress(0x0, false)
					case sdl.K_c:
						keyPress(0xB, false)
					case sdl.K_v:
						keyPress(0xF, false)
					}
				} else if ev.Type == sdl.KEYDOWN {
					switch ev.Keysym.Sym {
					case sdl.K_1:
						keyPress(0x1, true)
					case sdl.K_2:
						keyPress(0x2, true)
					case sdl.K_3:
						keyPress(0x3, true)
					case sdl.K_4:
						keyPress(0xC, true)
					case sdl.K_q:
						keyPress(0x4, true)
					case sdl.K_w:
						keyPress(0x5, true)
					case sdl.K_e:
						keyPress(0x6, true)
					case sdl.K_r:
						keyPress(0xD, true)
					case sdl.K_a:
						keyPress(0x7, true)
					case sdl.K_s:
						keyPress(0x8, true)
					case sdl.K_d:
						keyPress(0x9, true)
					case sdl.K_f:
						keyPress(0xE, true)
					case sdl.K_z:
						keyPress(0xA, true)
					case sdl.K_x:
						keyPress(0x0, true)
					case sdl.K_c:
						keyPress(0xB, true)
					case sdl.K_v:
						keyPress(0xF, true)
					}
				}
				break
			}
		}
		/*
			surface.FillRect(nil, 0)
			i += 1
			if i > 200 {
				i = 0
			}
		*/
		//rect := sdl.Rect{0, 0, int32(i), int32(i / 2)}
		//surface.FillRect(&rect, 0xffff0000)

		// before we update the surface, we should probably
		window.UpdateSurface()

		sdl.Delay(1000 / 60)
	}

}

func keyPress(ky uint8, press bool) {
	if press {
		key[ky] = 1
	} else {
		key[ky] = 0
	}
}

func emulateCycle() {
	// fetch opcode
	opcode = uint16(memory[pc])<<8 | uint16(memory[pc+1])

	//fmt.Printf("OPCODE: 0x%02X%02X\tPC: 0x%04X\tI: 0x%04X\nV:\t", memory[pc], memory[pc+1], pc, I)
	// decode opcode
	// opcode & 0xF000
	switch opcode & 0xF000 {
	case 0x0000:
		switch opcode & 0x000F {
		case 0x0000: // 0x00E0: Clears the screen
			// execute opcode
			// 00E0 - CLS
			// Clear the display.
			for j := 0; j < len(gfx); j++ {
				for i := 0; i < len(gfx[j]); i++ {
					gfx[j][i] = 0x0
				}
			}
			pc += 2
			break
		case 0x000E: // 0x00EE returns from subroutine
			// exec
			// 00EE - RET
			// Return from a subroutine.
			sp--
			pc = stack[sp]
			pc += 2
			break
		default:
			fmt.Printf("Unknown opcode: 0x%X\n", opcode)
			break
		}
		break
	case 0x1000:
		// Jump to location nnn.
		// The interpreter sets the program counter to nnn.
		pc = opcode & 0x0FFF
		break
	case 0x2000: // calls the subroutine at address NNN
		stack[sp] = pc
		sp += 1
		pc = opcode & 0x0FFF
		break
	case 0x3000:
		// Skip next instruction if Vx = kk.
		if V[(opcode&0x0F00)>>8] == byte(opcode&0x00FF) {
			pc += 4
		} else {
			pc += 2
		}
		break
	case 0x4000:
		// Skip next instruction if Vx != kk.
		if V[(opcode&0x0F00)>>8] != byte(opcode&0x00FF) {
			pc += 4
		} else {
			pc += 2
		}
		break
	case 0x5000:
		// Skip next instruction if Vx = Vy.
		if V[(opcode&0x0F00)>>8] == V[(opcode&0x00F0)>>4] {
			pc += 4
		} else {
			pc += 2
		}
		break
	case 0x6000:
		// Set Vx = kk.
		// the interpreter puts the value kk into register Vx.
		V[(opcode&0x0F00)>>8] = byte(opcode & 0x00FF)
		pc += 2
		break
	case 0x7000:
		// set Vx = Vx + kk
		V[(opcode&0x0F00)>>8] = V[(opcode&0x0F00)>>8] + byte(opcode&0x00FF)
		pc += 2
		break
	case 0x8000:
		switch opcode & 0x000F {
		case 0x0000:
			//Set Vx = Vy.
			V[(opcode&0x0F00)>>8] = V[(opcode&0x00F0)>>4]
			pc += 2
			break
		case 0x0001:
			//Set Vx = Vx OR Vy.
			V[(opcode&0x0F00)>>8] = V[(opcode&0x0F00)>>8] | V[(opcode&0x00F0)>>4]
			pc += 2
			break
		case 0x0002:
			//Set Vx = Vx AND Vy.
			V[(opcode&0x0F00)>>8] = V[(opcode&0x0F00)>>8] & V[(opcode&0x00F0)>>4]
			pc += 2
			break
		case 0x0003:
			//Set Vx = Vx XOR Vy.
			V[(opcode&0x0F00)>>8] = V[(opcode&0x0F00)>>8] ^ V[(opcode&0x00F0)>>4]
			pc += 2
			break
		case 0x0004:
			// Set Vx = Vx + Vy, set VF = carry.
			V[0xF] = 0
			if V[(opcode&0x00F0)>>4] > 0xFF-V[(opcode&0x0F00)>>8] {
				V[0xF] = 1
			}
			V[(opcode&0x0F00)>>8] += V[(opcode&0x00F0)>>4]
			pc += 2
			break
		case 0x0005:
			//Set Vx = Vx - Vy, set VF = NOT borrow.
			if V[(opcode&0x0F00)>>8] < V[(opcode&0x00F0)>>4] {
				V[0xF] = 1
			} else {
				V[0xF] = 0
			}
			V[(opcode&0x0F00)>>8] -= V[(opcode&0x00F0)>>4]
			pc += 2
			break
		case 0x0006:
			//Set Vx = Vx SHR 1.
			// If the least-significant bit of Vx is 1, then VF is set to 1,
			// otherwise 0. Then Vx is divided by 2.
			V[0xF] = V[(opcode&0x0F00)>>8] & 0x1
			V[(opcode&0x0F00)>>8] /= 2
			pc += 2
			break
		case 0x0007:
			//Set Vx = Vy - Vx, set VF = NOT borrow.
			if V[(opcode&0x00F0)>>4] < V[(opcode&0x0F00)>>8] {
				V[0xF] = 1
			} else {
				V[0xF] = 0
			}
			V[(opcode&0x0F00)>>8] = V[(opcode&0x00F0)>>4] - V[(opcode&0x0F00)>>8]
			pc += 2
			break
		case 0x000E:
			// If the most-significant bit of Vx is 1, then VF is set to 1,
			// otherwise to 0. Then Vx is multiplied by 2.
			// shifts the byte 7 places so we get the most significant bit
			V[0xF] = V[(opcode&0x0F00)>>8] >> 7
			V[(opcode&0x0F00)>>8] *= 2
			pc += 2
			break
		default:
			fmt.Printf("Unimplemented Opcode: 0x%02X\n", opcode)
			break
		}
		break
	case 0x9000:
		// Skip next instruction if Vx != Vy.
		// The values of Vx and Vy are compared, and if they are not equal,
		// the program counter is increased by 2.
		if V[(opcode&0x0F00)>>8] != V[(opcode&0x00F0)>>4] {
			pc += 4
		} else {
			pc += 2
		}
		break
	case 0xA000: // ANN: Sets I to address NNN
		I = opcode & 0x0FFF
		pc += 2
		break
	case 0xB000:
		pc = uint16(opcode&0x0FFF) + uint16(V[0x0])
		break
	case 0xC000:
		//Set Vx = random byte AND kk.
		V[(opcode&0x0F00)>>8] = uint8(rand.Intn(256)) & uint8(opcode&0x00FF)
		pc += 2
		break
	/*
	 *	Not gonna lie, I couldn't understand the Painting/Drawing function at all
	 *	tried following the logic, but i don't get it, so I took the code from
	 *	https://github.com/skatiyar/go-chip8/blob/master/emulator/emulator.go#L198
	 *
	 */
	case 0xD000:
		x := V[(opcode&0x0F00)>>8]
		y := V[(opcode&0x00F0)>>4]
		h := opcode & 0x000F
		V[0xF] = 0
		for j := uint16(0); j < h; j++ {
			sprite := memory[I+j]
			for i := uint16(0); i < 8; i++ {
				if (sprite & (0x80 >> i)) != 0 {
					if gfx[y+uint8(j)][x+uint8(i)] == 1 {
						V[0xF] = 1
					}
					gfx[y+uint8(j)][x+uint8(i)] ^= 1
				}
			}
		}
		draw_flag = true
		pc += 2
		break
	case 0xE000:
		{
			switch opcode & 0x00FF {
			case 0x009E:
				// Skip next instruction if key with the value of Vx is not pressed.
				if key[V[(opcode&0x0F00)>>8]] == 1 {
					pc += 4
				} else {
					pc += 2
				}
				break
			case 0x00A1:
				// Skip next instruction if key with the value of Vx is not pressed.
				if key[V[(opcode&0x0F00)>>8]] != 1 {
					pc += 4
				} else {
					pc += 2
				}
				break
			default:
				fmt.Printf("Unimplemented Opcode: 0x%02X\n", opcode)
				break
			}
			break
		}
	case 0xF000:
		switch opcode & 0x00FF {
		case 0x0007:
			// Set Vx = delay timer value.
			// The value of DT is placed into Vx.
			V[opcode&0x0F00>>8] = byte(delay_timer)
			pc += 2
			break
		case 0x000A:
			// Wait for a key press, store the value of the key in Vx.
			pressed := false
			for i := 0; i < len(key); i++ {
				if key[i] != 0 {
					V[(opcode&0x0F00)>>8] = uint8(i)
					pressed = true
				}
			}
			if pressed == false {
				return
			}
			pc += 2
			break
		case 0x0015:
			// Set delay timer = Vx.
			delay_timer = opcode & 0x0F00
			pc += 2
			break
		case 0x0018:
			// Set sound timer = Vx.
			sound_timer = uint16(V[(opcode&0x0F00)>>8])
			pc += 2
			break
		case 0x0020:
			fmt.Printf("Unimplemented Opcode: 0x%02X does not appear to exist in spec\n", opcode)
			pc += 2
			break
		case 0x001E:
			// The values of I and Vx are added, and the results are stored in I.
			if I+uint16(V[(opcode&0x0F00)>>8]) > 0xFFF {
				V[0xF] = 1
			} else {
				V[0xF] = 0
			}
			I += uint16(V[(opcode&0x0F00)>>8])
			pc += 2
			break
		case 0x0029:
			// Set I = location of sprite for digit Vx.
			I = uint16(V[opcode&0x0F00>>8] * 0x5)
			pc += 2
			break
		case 0x0033:
			// Store BCD representation of Vx in memory locations I, I+1, and I+2.
			// The interpreter takes the decimal value of Vx, and places the hundreds
			// digit in memory at location in I, the tens digit at location I+1, and
			// the ones digit at location I+2.
			bcd := V[opcode&0x0F00>>8]
			memory[I] = bcd / 100
			memory[I+1] = ((bcd & 0x00F0) / 10) % 10
			memory[I+2] = ((bcd & 0x000F) % 100) / 10
			pc += 2
			break
		case 0x0055:
			// Store registers V0 through Vx in memory starting at location I.
			for i := 0; i < len(V); i++ {
				memory[I+uint16(i)] = V[i]
			}
			pc += 2
			break
		case 0x0065:
			untilVx := (opcode & 0x0F00) >> 8
			for i := 0; i < int(untilVx); i++ {
				V[i] = memory[I+uint16(i)]
			}
			pc += 2
			break
		case 0x0090:
			fmt.Printf("Unimplemented Opcode: 0x%02X does not appear to exist in spec\n", opcode)
			pc += 2
			break
		default:
			fmt.Printf("Unimplemented Opcode: 0x%02X\n", opcode)
			break
		}
		break
	default:
		fmt.Printf("Unimplemented Opcode: 0x%02X\n", opcode)
		break
	}
	// update timers
	if delay_timer > 0 {
		delay_timer -= 1
	}

	if sound_timer > 0 {
		if sound_timer == 1 {
			fmt.Println("BEEP")
			// Should produce an actual beep/
		}
		sound_timer -= 1
	}
}

func draw(screenSurface *sdl.Surface) {
	bit_size := 10

	for j := 0; j < len(gfx); j++ {
		for i := 0; i < len(gfx[j]); i++ {
			var cl = sdl.MapRGB(screenSurface.Format, 255, 255, 255)
			if gfx[j][i] != 1 {
				cl = sdl.MapRGB(screenSurface.Format, 0, 0, 0)
			}

			screenSurface.FillRect(&sdl.Rect{
				X: int32(i * bit_size),
				Y: int32(j * bit_size),
				W: int32(bit_size),
				H: int32(bit_size)}, cl)
		}

	}

	draw_flag = false
}
