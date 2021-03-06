# Stack Based Virtual Machine in GO

##### VASM - Assmebly Language for the Virtual Machine
## Currently Supported Instructions in VASM 

*(Only 64-bit Architecture Supported)*
- **DEFINE**
- **INCLUDE**
- **PUSH**
- **ADDI**
- **SUBI**
- **MULI**
- **DIVI**
- **ADDF**
- **SUBF**
- **MULF**
- **DIVF**
- **JMP**
- **HALT**
- **IGNORE_HALT**
- **NOP**
- **RET**
- **DUP**
- **SWAP**
- **CALL**
- **DROP**
- **JMP_IF**
- **NOT**
- **EQI**
- **EQF**
- **PRINT**


(See [Instruction Help](#instruction-help))
## Quick Start

```console
> go build vm-go.go
> .\vm-go.exe
```
#### Alternatively on Linux (Build with Optimizations)
```console
> chmod +x build.sh
> ./build.sh
> ./vm-go
```
#### Alternatively on Powershell (Build with Optimizations)
```console
> .\go_build.ps1 .\vm-go.go
> .\vm-go.exe
```
#### Running With Step Debugging and Also Compiling to Byte Code .vm
```console
> .\go_build.ps1 .\vm-go.go
> .\vm-go.exe -i .\examples\fib.vasm -debug -compile
see generated .\examples\fib.vm 
```
#### Executing in Virtual Machine from .vasm file

See [examples](./examples) folder for more .vasm Examples
```console
> .\go_build.ps1 .\vm-go.go
> .\vm-go.exe -i .\examples\powers_of_two.vasm -limit 71
see .\vm-go.exe -h for help on input paramenters
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
{131072 5e-324 }
---- STACK END ----
```

#### Calculating 'e' from [e.vasm](./examples/e.vasm)

```console
> .\vm-go.exe -i .\examples\e.vasm -limit 120
```
#### Output
```console
---- PROGRAM TRACE BEG ----
JMP : {int64holder:3 float64holder:5e-324 pointer:}
SWAP : {int64holder:2 float64holder:0 pointer:}
SWAP : {int64holder:1 float64holder:0 pointer:}
MULF
SWAP : {int64holder:2 float64holder:0 pointer:}
DUP : {int64holder:0 float64holder:0 pointer:}
ADDF
PUSH : {int64holder:-9223372036854775808 float64holder:1 pointer:}
SWAP : {int64holder:2 float64holder:0 pointer:}
ADDF
DIVF
DUP : {int64holder:2 float64holder:0 pointer:}
PUSH : {int64holder:-9223372036854775808 float64holder:1 pointer:}
PUSH : {int64holder:-9223372036854775808 float64holder:1 pointer:}
PUSH : {int64holder:-9223372036854775808 float64holder:1 pointer:}
PUSH : {int64holder:-9223372036854775808 float64holder:1 pointer:}
---- PROGRAM TRACE END ----

---- STACK BEG ----
{-9223372036854775808 10 }
{-9223372036854775808 3.6288e+06 }
{-9223372036854775808 2.7182815255731922 }  <- Approximation of 'e' using Taylor series around 0 and value of x=1
---- STACK END ----
```

## Instruction Help

- **DEFINE** (Constant Name) (Operand Any) : Define Constants 
- **INCLUDE** (Operand File Path String) : Include file path and load program from include path to VM 
- **PUSH** (Operand Any) : Push Operand to Stack 
- **ADDI** : Add top two operands as integers and push it back to top of Stack
- **SUBI** : Subtract top two operands as integers and push it back to top of Stack
- **MULI** : Multiply top two operands as integers and push it back to top of Stack
- **DIVI** : Divide top two operands as integers and push it back to top of Stack
- **ADDF** : Add top two operands as float and push it back to top of Stack
- **SUBF** : Subtract top two operands as float and push it back to top of Stack
- **MULF** : Multiply top two operands as float and push it back to top of Stack
- **DIVF** : Divide top two operands as float and push it back to top of Stack
- **JMP** (Operand Int / Label) : Jump to label or location within Stack
- **HALT** : Halt the Virtual Machine
- **IGNORE_HALT** : Ignore any HALT instruction in the file
- **NOP** : Perform No Operation in Stack
- **RET** : Point Instruction Pointer to top of Satck
- **DUP** (Operand Int) : Duplicate Operand from operand location of stack and push it back to top of Stack
- **SWAP** (Operand Int) : Swap Values of operand location and top of stack (*can be used as an accumulator*) 
- **CALL** (Operand Int / Label) : Jump to label or funcall within Stack
- **DROP** : Remove Value from top of Stack
- **JMP_IF** (Operand Int) : Jump If int64 is not 0
- **NOT** : !0 -> 0 and 0 -> 1 on the top of the Stack
- **EQI** : Perform Integer Equality on top two values of stack, consumes them and push 1 if equal and 0 otherwise on top of Stack
- **EQF** : Perform Float Equality on top two values of stack, consumes them and push 1 if equal and 0 otherwise on top of Stack
- **PRINT** : Prints the top of the stack and consumes it
