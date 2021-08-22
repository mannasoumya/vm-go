# Stack Based Virtual Machine in GO

## Currently Supported Instructions
- **PUSH**
- **ADDI**
- **SUBI**
- **MULI**
- **DIVI**
- **JMP**
- **HALT**
- **NOP**
- **RET**
- **DUP**


## Quick Start

```console
> go build main.go
> .\main.exe
```
#### Alternatively on Powershell (Build with Optimizations)
```console
> .\go_build.ps1 .\main.go
> .\main.exe
```
#### Running With Step Debugging
```console
> .\go_build.ps1 .\main.go
> .\main.exe -debug
```
#### Executing in Virtual Machine from .vasm file

See [examples](./examples) folder for more .vasm Examples
```console
> .\go_build.ps1 .\main.go
> .\main.exe -input .\examples\powers_of_two.vasm -limit 71
see .\main.exe -h for help on input paramenters
```
**'vasm'** Instruction Set which generated powers of 2 : [powers_of_two.vasm](./examples/powers_of_two.vasm)
```asm
# Pushing initial
PUSH 0
PUSH 1
# Starting Loop
loop1:
    ADDI		# Adding Last two of the stack
    DUP 0		# Duplicating 
    DUP 1		# Duplicating 
    JMP loop1
```
#### Output
```console
---- PROGRAM TRACE BEG ----
JMP : {int64holder:2 float64holder:5e-324 pointer:}
DUP : {int64holder:1 float64holder:0 pointer:}
DUP : {int64holder:0 float64holder:0 pointer:}
ADDI
PUSH : {int64holder:1 float64holder:5e-324 pointer:}
PUSH : {int64holder:0 float64holder:5e-324 pointer:}
---- PROGRAM TRACE END ----

---- STACK BEG ----
{1 5e-324 }
{2 5e-324 }
{4 5e-324 }
{8 5e-324 }
{16 5e-324 }
{32 5e-324 }
{64 5e-324 }
{128 5e-324 }
{256 5e-324 }
{512 5e-324 }
{1024 5e-324 }
{2048 5e-324 }
{4096 5e-324 }
{8192 5e-324 }
{16384 5e-324 }
{32768 5e-324 }
{65536 5e-324 }
{65536 5e-324 }
{65536 5e-324 }
---- STACK END ----