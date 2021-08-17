package main

import (
	"fmt"
    "io/ioutil"
	"strings"
	"strconv"
	"flag"
)

const STACK_CAPACITY = 1024
const PROGRAM_CAPACITY = 1024

type VM struct {
	stack_size   int
	STACK        [STACK_CAPACITY]int

	PROGRAM      [PROGRAM_CAPACITY]Inst
	inst_ptr     int
	program_size int
	
	vm_halt      int
}

type Inst struct {
	Name    string
	Operand int
}

func check_err(e error) {
    if e != nil {
        panic(e)
    }
}

func push(vm *VM, inst Inst) {
	vm.STACK[vm.stack_size] = inst.Operand
	vm.stack_size += 1
	vm.inst_ptr += 1
}

func add(vm *VM) {
	if vm.stack_size < 2 {
		panic("Not enough values to add")
	}
	vm.STACK[vm.stack_size-2] = vm.STACK[vm.stack_size-1] + vm.STACK[vm.stack_size-2]
	vm.stack_size -= 1
	vm.inst_ptr += 1
}

func sub(vm *VM) {
	if vm.stack_size < 2 {
		panic("Not enough values to subtract")
	}
	vm.STACK[vm.stack_size-2] = vm.STACK[vm.stack_size-2] - vm.STACK[vm.stack_size-1]
	vm.stack_size -= 1
	vm.inst_ptr += 1
}

func mul(vm *VM) {
	if vm.stack_size < 2 {
		panic("Not enough values to multiply")
	}
	vm.STACK[vm.stack_size-2] = vm.STACK[vm.stack_size-2] * vm.STACK[vm.stack_size-1]
	vm.stack_size -= 1
	vm.inst_ptr += 1
}

func div(vm *VM) {
	if vm.stack_size < 2 {
		panic("Not enough values to divide")
	}
	if vm.STACK[vm.stack_size-1] == 0 {
		print_stack(vm)
		panic("Zero Division Error")
	}
	vm.STACK[vm.stack_size-2] = vm.STACK[vm.stack_size-2] / vm.STACK[vm.stack_size-1]
	vm.stack_size -= 1
	vm.inst_ptr += 1
}

func peek(vm *VM) int {
	if vm.stack_size == 0 {
		panic("Empty Stack")
	}
	return vm.STACK[vm.stack_size-1]
}

func jmp(vm *VM, inst Inst) {
	if inst.Operand < 0 {
		panic("Wrong Jump Instruction. Underflow")
	}
	if inst.Operand >= vm.program_size {
		panic("Wrong Jump Instruction. Overflow")
	}
	vm.inst_ptr = inst.Operand
}

func nop(vm *VM) {
	vm.inst_ptr += 1
}

func halt(vm *VM) {
	vm.vm_halt = 1
}

func ret(vm *VM) {
	vm.inst_ptr = vm.STACK[vm.stack_size - 1]
	vm.stack_size -= 1;
}

func dup(vm *VM, inst Inst) {
	if vm.stack_size >= STACK_CAPACITY {
		panic("Stack Overflow");
	}
	
	if (vm.stack_size - inst.Operand <= 0) {
		panic("Stack Underflow");
	}

	vm.STACK[vm.stack_size] = vm.STACK[vm.stack_size - 1 - inst.Operand];
	vm.stack_size += 1;
	vm.inst_ptr += 1;
}

func execute_inst(vm *VM, inst Inst) {
	if vm.inst_ptr >= vm.program_size {
		panic("Illegal Instruction Access")
	}
	if vm.stack_size < 0 {
		panic("Stack Underflow")
	}
	if vm.stack_size > STACK_CAPACITY {
		panic("Stack Overflow")
	}
	switch inst.Name {
	case "PUSH":
		push(vm, inst)
	case "ADD":
		add(vm)
	case "SUB":
		sub(vm)
	case "MUL":
		mul(vm)
	case "DIV":
		div(vm)
	case "JMP":
		jmp(vm, inst)
	case "HALT":
		halt(vm)
	case "NOP":
		nop(vm)
	case "RET":
		ret(vm)
	case "DUP":
		dup(vm, inst)
	default:
		panic("Unknown Instruction")
	}
	// vm.PROGRAM[vm.inst_ptr] =  inst
	// vm.inst_ptr += 1

}

func print_stack(vm *VM) {
	if vm.stack_size < 0 {
		panic("ERROR: Stack Underflow")
	}
	fmt.Println("---- STACK TOP ----")
	for i := vm.stack_size - 1; i >= 0; i-- {
		fmt.Println(vm.STACK[i])
	}
	fmt.Println("---- STACK END ----")
	fmt.Println()
}

func print_program_trace(vm *VM, banner bool) {
	if vm.program_size == 0 {
		panic("Empty Program")
	}
	if vm.program_size >= PROGRAM_CAPACITY {
		panic("Overflow: vm.program_size >= PROGRAM_CAPACITY")
	}
	if banner {
		fmt.Println("---- PROGRAM TRACE BEG ----")
	}
	
	for i := vm.program_size - 1; i >= 0; i-- {
		switch vm.PROGRAM[i].Name {
		case "PUSH":
			fmt.Printf("%s : %d \n", vm.PROGRAM[i].Name, vm.PROGRAM[i].Operand)
		case "ADD":
			fmt.Printf("%s \n", vm.PROGRAM[i].Name)
		case "SUB":
			fmt.Printf("%s \n", vm.PROGRAM[i].Name)
		case "MUL":
			fmt.Printf("%s \n", vm.PROGRAM[i].Name)
		case "DIV":
			fmt.Printf("%s \n", vm.PROGRAM[i].Name)
		case "JMP":
			fmt.Printf("%s : %d \n", vm.PROGRAM[i].Name, vm.PROGRAM[i].Operand)
		case "HALT":
			fmt.Printf("%s \n", vm.PROGRAM[i].Name)
		case "NOP":
			fmt.Printf("%s \n", vm.PROGRAM[i].Name)
		case "RET":
			fmt.Printf("%s \n", vm.PROGRAM[i].Name)
		case "DUP":
			fmt.Printf("%s : %d \n", vm.PROGRAM[i].Name, vm.PROGRAM[i].Operand)
		default:
			panic("Unknown Instruction")
		}
	}
	if banner {
		fmt.Println("---- PROGRAM TRACE END ----")
		fmt.Println()
	}
}

func load_program_from_memory(vm *VM, program []Inst, program_size int, halt_panic bool) {
	if program_size > PROGRAM_CAPACITY {
		panic("Overflow")
	}

	halt_flag := false
	for i := 0; i < program_size; i++ {
		if program[i].Name == "HALT" {
			halt_flag = true
		}
		vm.PROGRAM[vm.program_size] = program[i]
		vm.program_size += 1 
	}
	if halt_flag == false {
		if halt_panic {
			print_program_trace(vm,true)
			panic("No `HALT` instruction in PROGRAM")
		}
	}
}


func load_program_from_file(vm *VM, file_path string, halt_panic bool) {
	dat, err := ioutil.ReadFile(file_path)
	check_err(err)
    file_content := string(dat)
	lines := strings.Split(strings.ReplaceAll(file_content, "\r\n", "\n"), "\n")
	instruction_count := 0
	halt_flag := false
	for i:=0; i<len(lines) ; i++ {
		line := strings.Trim(lines[i]," ")
		if line != "" {
			line_split_by_space := strings.Split(line, " ")
			inst_name := strings.ToUpper(line_split_by_space[0])
			switch inst_name {
			
			case "PUSH":
				if len(line_split_by_space) > 2 {
					fmt.Printf("File : %s\n", file_path)
					fmt.Printf("Too Many Args or Extra Spaces: Invalid Syntax near line %d : %s\n", (i+1), line)
					panic("Syntax Error")
				}
				operand , err := strconv.Atoi(line_split_by_space[1])
				check_err(err)
				vm.PROGRAM[vm.program_size] = Inst{Name: "PUSH", Operand: operand}
				vm.program_size += 1
			
			case "ADD":
				if len(line_split_by_space) > 1 {
					fmt.Printf("File : %s\n", file_path)
					fmt.Printf("Syntax Error: Invalid Syntax near line %d : %s\n", (i+1), line)
					panic("Syntax Error")
				}
				vm.PROGRAM[vm.program_size] = Inst{Name: "ADD"}
				vm.program_size += 1
				
			case "SUB":
				if len(line_split_by_space) > 1 {
					fmt.Printf("File : %s\n", file_path)
					fmt.Printf("Syntax Error: Invalid Syntax near line %d : %s\n", (i+1), line)
					panic("Syntax Error")
				}
				vm.PROGRAM[vm.program_size] = Inst{Name: "SUB"}
				vm.program_size += 1
				
			case "MUL":
				if len(line_split_by_space) > 1 {
					fmt.Printf("File : %s\n", file_path)
					fmt.Printf("Syntax Error: Invalid Syntax near line %d : %s\n", (i+1), line)
					panic("Syntax Error")
				}
				vm.PROGRAM[vm.program_size] = Inst{Name: "MUL"}
				vm.program_size += 1
				
			case "DIV":
				if len(line_split_by_space) > 1 {
					fmt.Printf("File : %s\n", file_path)
					fmt.Printf("Syntax Error: Invalid Syntax near line %d : %s\n", (i+1), line)
					panic("Syntax Error")
				}
				vm.PROGRAM[vm.program_size] = Inst{Name: "DIV"}
				vm.program_size += 1
				
			case "JMP":
				if len(line_split_by_space) > 2 {
					fmt.Printf("File : %s\n", file_path)
					fmt.Printf("Too Many Args or Extra Spaces: Invalid Syntax near line %d : %s\n", (i+1), line)
					panic("Syntax Error")
				}
				operand , err := strconv.Atoi(line_split_by_space[1])
				check_err(err)
				vm.PROGRAM[vm.program_size] = Inst{Name: "JMP", Operand: operand}
				vm.program_size += 1
				
			case "HALT":
				if len(line_split_by_space) > 1 {
					fmt.Printf("File : %s\n", file_path)
					fmt.Printf("Syntax Error: Invalid Syntax near line %d : %s\n", (i+1), line)
					panic("Syntax Error")
				}
				halt_flag = true
				vm.PROGRAM[vm.program_size] = Inst{Name: "HALT"}
				vm.program_size += 1
				
			case "NOP":
				if len(line_split_by_space) > 1 {
					fmt.Printf("File : %s\n", file_path)
					fmt.Printf("Syntax Error: Invalid Syntax near line %d : %s\n", (i+1), line)
					panic("Syntax Error")
				}
				vm.PROGRAM[vm.program_size] = Inst{Name: "NOP"}
				vm.program_size += 1
				
			case "RET":
				if len(line_split_by_space) > 1 {
					fmt.Printf("File : %s\n", file_path)
					fmt.Printf("Syntax Error: Invalid Syntax near line %d : %s\n", (i+1), line)
					panic("Syntax Error")
				}
				vm.PROGRAM[vm.program_size] = Inst{Name: "RET"}
				vm.program_size += 1
				
			case "DUP":
				if len(line_split_by_space) > 2 {
					fmt.Printf("File : %s\n", file_path)
					fmt.Printf("Too Many Args or Extra Spaces: Invalid Syntax near line %d : %s\n", (i+1), line)
					panic("Syntax Error")
				}
				operand , err := strconv.Atoi(line_split_by_space[1])
				check_err(err)
				vm.PROGRAM[vm.program_size] = Inst{Name: "DUP", Operand: operand}
				vm.program_size += 1
			default:
				fmt.Printf("File : %s\n", file_path)
				fmt.Printf("Syntax Error: Unknown Instruction near line %d : %s\n",(i+1), line)
				panic("Unknown Instruction")
			}
			instruction_count += 1
			if instruction_count >= PROGRAM_CAPACITY {
				fmt.Printf("File : %s\n", file_path)
				fmt.Printf("Number of Instructions is greater than PROGRAM CAPACITY = %d", PROGRAM_CAPACITY)
				panic("Overflow")
			}
		}	
	}
	if halt_flag == false {
		if halt_panic {
			print_program_trace(vm,true)
			panic("No `HALT` instruction in PROGRAM")
		}
	}
}

func execute_program(vm *VM, limit int) {
	if vm.program_size == 0 {
		panic("No instruction to execute.. Load Program first")
	}
	counter := 0
	for (vm.vm_halt != 1 && counter < limit) {
		// execute_inst(vm, vm.PROGRAM[counter % tot_len])
		execute_inst(vm, vm.PROGRAM[vm.inst_ptr])
		counter += 1
	}
}

func main() {
	var initial_stack [STACK_CAPACITY]int
	var initial_program [PROGRAM_CAPACITY]Inst
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
	
	var execution_limit_steps int
	vm_g := VM{STACK: initial_stack, PROGRAM: initial_program}
	
	file_path := flag.String("input", "", ".vasm FILE PATH")
	execution_limit_steps_inp := flag.Int("limit", 69, "Execution Limit Steps")
	
	flag.Parse()
	
	if *file_path == "" {
		execution_limit_steps = 69
		load_program_from_memory(&vm_g, prgm, program_size, true)
		execute_program(&vm_g, execution_limit_steps)
		print_stack(&vm_g)
		print_program_trace(&vm_g, true)
	} else {
		execution_limit_steps = *execution_limit_steps_inp
		load_program_from_file(&vm_g, *file_path, false)
		print_program_trace(&vm_g, true)
		execute_program(&vm_g, execution_limit_steps)
		print_stack(&vm_g)
	}
}
