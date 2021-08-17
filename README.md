# Stack Based Virtual Machine in GO

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

### Currently Supported Instructions
- **PUSH**
- **ADD**
- **SUB**
- **MUL**
- **DIV**
- **JMP**
- **HALT**
- **NOP**

### Sample Program

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

### Output 

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
