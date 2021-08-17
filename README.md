# Stack Based Virtual Machine in GO

## Currently Supported Instructions
- **PUSH**
- **ADD**
- **SUB**
- **MUL**
- **DIV**
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
#### Executing in Virtual Machine from .vasm file
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
# Adding Last two of the stack
ADD
# Duplicating 
DUP 0
DUP 1
# Loop
JMP 2
```
#### Output
```console
---- PROGRAM TRACE BEG ----
JMP : 2
DUP : 1
DUP : 0
ADD
PUSH : 1
PUSH : 0
---- PROGRAM TRACE END ----

---- STACK TOP ----
131072
65536
32768
16384
8192
4096
2048
1024
512
256
128
64
32
16
8
4
2
1
---- STACK END ----

```

#### Sample Program

```go
func main() {
	var initial_stack [STACK_CAPACITY]int // Initialise STACK
	var initial_program [PROGRAM_CAPACITY]Inst // Initialise PROGRAM
	// Define PROGRAM
	var prgm = []Inst {
		Inst{Name: "PUSH", Operand: 10},
		Inst{Name: "PUSH", Operand: 10},
		Inst{Name: "PUSH", Operand: 10},
		Inst{Name: "PUSH", Operand: 20},
		Inst{Name: "ADD"},
		Inst{Name: "MUL"},
		Inst{Name: "NOP"},
		Inst{Name: "PUSH", Operand: 10},
		Inst{Name: "SUB", Operand: 10},
		Inst{Name: "HALT"},
	}
	program_size := len(prgm)
	// Set Execution Limit
	execution_limit_steps := 69 // How many times to execute (useful for non halting Virtual Machines)
	// Initialise Virtual Machine 'vm_g'
	vm_g := VM{STACK: initial_stack, PROGRAM: initial_program}
	// Load above PROGRAM 'prgm' into the Virtual Machine 'vm_g'
	load_program_from_memory(&vm_g, prgm, program_size, true)
	// Execute PROGRAM in Virtual Machine 'vm_g'
	execute_program(&vm_g, execution_limit_steps)
	// Dump STACK to stdout
	print_stack(&vm_g)
	// Dump PROGRAM inst to stdout
	print_program_trace(&vm_g, true)
}
```

#### Output 

```console
---- STACK TOP ----
290
10
---- STACK END ----

---- PROGRAM TRACE BEG ----
HALT
SUB
PUSH : 10
NOP
MUL
ADD
PUSH : 20
PUSH : 10
PUSH : 10
PUSH : 10
---- PROGRAM TRACE END ----

```
