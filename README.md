# Stack Based Virtual Machine in GO

## Quick Start

```console
> go build main.go
> .\main.exe
```

### Sample Program

```go
func main() {
    var initial [STACK_CAPACITY]int
    var initial_inst []Inst
    vm_g := VM{stack_size: 0, STACK: initial, PROGRAM: initial_inst, inst_ptr: -1}
    push_inst(&vm_g, Inst{Name: "PUSH", Value: 10, Is_Operand:true})
    push_inst(&vm_g, Inst{Name: "PUSH", Value: 10, Is_Operand:true})
    push_inst(&vm_g, Inst{Name: "PUSH", Value: 10, Is_Operand:true})
    push_inst(&vm_g, Inst{Name: "PUSH", Value: 20, Is_Operand:true})
    print_stack(&vm_g)
    push_inst(&vm_g, Inst{Name: "ADD", Value: 0, Is_Operand:true})
    print_stack(&vm_g)
    push_inst(&vm_g, Inst{Name: "MUL", Value: 0, Is_Operand:true})
    print_stack(&vm_g)
    push_inst(&vm_g, Inst{Name: "PUSH", Value: 10, Is_Operand:true})
    print_stack(&vm_g)
    push_inst(&vm_g, Inst{Name: "SUB", Value: 10, Is_Operand:true})
    print_stack(&vm_g)
    print_program_trace(&vm_g, true)
}
```

### Output 

```console

---- STACK TOP ----
20
10
10
10
---- STACK END ----

---- STACK TOP ----
30
10
10
---- STACK END ----

---- STACK TOP ----
300
10
---- STACK END ----

---- STACK TOP ----
10
300
10
---- STACK END ----

---- STACK TOP ----
290
10
---- STACK END ----

---- PROGRAM TRACE BEG ----
SUB
PUSH : 10
MUL
ADD
PUSH : 20
PUSH : 10
PUSH : 10
PUSH : 10
---- PROGRAM TRACE END ----
```
