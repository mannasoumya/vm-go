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
    	var initial [STACK_CAPACITY]int
	var initial_inst [PROGRAM_CAPACITY]Inst
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
	
	vm_g := VM{stack_size: 0, STACK: initial, PROGRAM: initial_inst, inst_ptr: 0}
	load_program_from_memory(&vm_g, prgm, program_size, true)
	execute_program(&vm_g)
	print_stack(&vm_g)
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
