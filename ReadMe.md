## Current bugs
Based on 3-corax+ test:
- [x] Program debug prints cmd address increased (next one, not current)
- [x] incorrect char positioning (not sure yet why) - set/get Register were missing V7 and V8 (:
- [ ] 8xx7 and 8xxE ops give errors on test scren
- [ ] v8 test gives error - something with 8bits in V registers


## Chip8 memory layout
4096 bytes total (0x0000 - 0x0fff)

| Start Address | End Address | Description |
| --------------- | --------------- | --------------- |
| 0x0000 | 0x0fff | 4096 RAM bytes in total |
| 0x0000 | 0x01ff | Chip8 interpreter |
| 0x0200 | 0x0e9f | User program space, 3328 bytes in total for 4096 RAM |
| 0x0ea0 | 0x0ecf | Chip8 subroutine call stack, 48 bytes |
| 0x0ed0 | 0x0eef | Interpreter work area, 24 bytes |
| 0x0ef0 | 0x0eff | General purpose registers, V0-VF |
| 0x0f00 | 0x0fff | 256 RAM area for display refresh |

## Commands
|Status| Test    | Code | Short Desc   | Impl Function              | Description |
|------|---------|------|--------------|----------------------------|-------------|
| [ ]  | [ ]/[ ] | 0MMM | MCALL        |                            | Machine (OS) subroutine call |
| [x]  | [ ]/[ ] | 00E0 | CLS          | ClearScreen()              | Clear screen |
| [x]  | [ ]/[ ] | DXYN | DRAW N,VX,VY | DisplayAt(VX, VY, N)       | Draw byte pattern at pos VX,VY |
| [x]  | [x]/[ ] | 1MMM | JMP @0MMM    | Jump(MMM)                  | Unconditional jump to address |
| [x]  | [x]/[ ] | BMMM | JMPV @0MMM   | JumpV(MMM)                 | Unconditional jump to (V0 + address)|
| [x]  | [x]/[ ] | 2MMM | CALL @0MMM   | Call(MMM)                  | Subroutine call            |
| [x]  | [x]/[ ] | 00EE | RET          | Ret()                      | Return from subroutine call|
| [x]  | [x]/[ ] | 3XKK | SEQ VX, KK   | SkipEqualVal(VX, KK)       | Skip next command if VX == KK |
| [x]  | [x]/[ ] | 4XKK | SNE VX, KK   | SkipNotEqualVal(VX, KK)    | Skip next commend if VX != KK |
| [x]  | [x]/[ ] | 5XY0 | SEQ VX, VY   | SkipEqualReg(VX, VY)       | Skip next command if VX == VY |
| [x]  | [x]/[ ] | 9XY0 | SNE VX, VY   | SkipNotEqualReg(VX,VY)     | Skip next command if VX != VY |
| [x]  | [x]/[ ] | EX9E | SK  VX       | SkipKeyPressedAt(VX)       | Skip next command if VX == Hex Key (LSD) |
| [x]  | [x]/[ ] | EXA1 | SNK VX       | SkipKeyNotPressedAt(VX)    | Skip next command if VX != Hex Key (LSD) |
| [x]  | [x]/[ ] | 6XKK | MOV VX, KK   | MovRegVal(VX, KK)          | Set VX = KK |
| [x]  | [ ]/[ ] | CXKK | RND VX, KK   | MovRegRnd(VX, KK)          | Set VX = Rnd with KK as mask |
| [x]  | [x]/[ ] | 7XKK | ADD VX, KK   | AddRegVal(VX, KK)          | Set VX = VX + KK |
| [x]  | [x]/[ ] | 8X00 | MOV VX, VY   | MovRegReg(VX, VY)          | Set VX = VY |
| [x]  | [x]/[ ] | 8X01 | OR  VX, VY   | Or(VX, VY)                 | Set VX = VX OR VY (VF mod) |
| [x]  | [x]/[ ] | 8X02 | AND VX, VY   | And(VX, VY)                | Set VX = VX AND VY (VF mod) |
| [x]  | [x]/[ ] | 8X03 | XOR VX, VY   | Xor(VX, VY)                | Set VX = VX XOR VY |
| [x]  | [x]/[ ] | 8X04 | ADD VX, VY   | AddRegReg(VX, VY)          | Set VX = VX + VY (VF mod) |
| [x]  | [x]/[ ] | 8X05 | SUB VX, VY   | SubRegReg(VX, VY)          | Set VX = VX - VY (VF mod)|
| [x]  | [x]/[ ] | 8X06 | RSH VX       | ShiftR(VX)                 | Set VX = VX>>1 (VF mod) |
| [x]  | [x]/[ ] | 8X07 | SUB VX, VY   | SubRegReg(VY, VX)          | Set VX = VY - VX (VF mod) |
| [x]  | [x]/[ ] | 8X0E | LSH VX       | ShiftL(VX)                 | Set VX = VX<<1 (VF mod) |
| [x]  | [x]/[ ] | AMMM | MOV I, 0MMM  | MovRegVal(I, MMM)          | Set I = 0MMM |
| [x]  | [ ]/[ ] | FX07 | MOV VX, T0   | MovRegVal(VX, T0)          | Set VX = T0 current timer value |
| [x]  | [ ]/[ ] | FX0A | KEY VX       | GetKeyReg(VX)              | Set VX = Hex Key digit |
| [x]  | [ ]/[ ] | FX15 | MOV T0, VX   | MovRegVal(T0, VX)          | Set T0 = VX |
| [x]  | [ ]/[ ] | FX18 | MOV T1, VX   | MovRegVal(T1, VX)          | Set T1 = VX |
| [x]  | [ ]/[ ] | FX1E | ADD I, VX    | MovRegVal(I, I + getRegister(VX)) | Set I = I + VX |
| [x]  | [ ]/[ ] | FX29 | STC VX       | SetCharReg(VX)             | ??? |
| [x]  | [x]/[x] | FX33 | BCD VX       | BcdReg(VX)                 | Set MI = 3 dec digit of VX (I not updated) |
| [x]  | [ ]/[ ] | FX55 | CAM VX       | CopyRegToMem(VX)           | Set MI = V0:VX (I = I + X + 1) |
| [x]  | [ ]/[ ] | FX66 | CAR VX       | CopyMemToReg(VX)           | Set V0:VX = MI (I = I + X + 1) |


## Todo

	- [x] Add LoadRomFromData to load into user program space from data array
    - [x] Add video memory buffer (array of bools) to draw in
    - [ ] Make unit test coverage above 80%
    - [ ] When all commands added, separate cmd printing info from execution
    - [ ] Add SDL to render display
