package main

import (
	"fmt"
)

const STACK_CAPACITY = 1024
const PROGRAM_CAPACITY = 1024

type VM struct {
	stack_size int
	STACK      [STACK_CAPACITY]int
	PROGRAM    [PROGRAM_CAPACITY]Inst
	inst_ptr   int
	vm_halt    int
}

type Inst struct {
	Name    string
	Operand int
}

func push(vm *VM, inst Inst) {
	vm.STACK[vm.stack_size] = inst.Operand
	vm.stack_size += 1
}

func add(vm *VM) {
	if vm.stack_size < 2 {
		panic("Not enough values to add")
	}
	vm.STACK[vm.stack_size-2] = vm.STACK[vm.stack_size-1] + vm.STACK[vm.stack_size-2]
	vm.stack_size -= 1
}

func sub(vm *VM) {
	if vm.stack_size < 2 {
		panic("Not enough values to subtract")
	}
	vm.STACK[vm.stack_size-2] = vm.STACK[vm.stack_size-2] - vm.STACK[vm.stack_size-1]
	vm.stack_size -= 1
}

func mul(vm *VM) {
	if vm.stack_size < 2 {
		panic("Not enough values to multiply")
	}
	vm.STACK[vm.stack_size-2] = vm.STACK[vm.stack_size-2] * vm.STACK[vm.stack_size-1]
	vm.stack_size -= 1
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
	if inst.Operand >= vm.inst_ptr {
		panic("Wrong Jump Instruction. Overflow")
	}
	vm.inst_ptr = inst.Operand
}

func halt(vm *VM){
	vm.vm_halt = 1
}

func execute_inst(vm *VM, inst Inst) {
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
		{}
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
	if len(vm.PROGRAM) == 0 {
		panic("Empty Instruction Slice")
	}
	if banner {
		fmt.Println("---- PROGRAM TRACE BEG ----")
	}
	
	for i := len(vm.PROGRAM) - 1; i >= 0; i-- {
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
		vm.PROGRAM[vm.inst_ptr] = program[i]
		vm.inst_ptr += 1 
	}
	if halt_flag == false {
		if halt_panic {
			print_program_trace(vm,true)
			panic("No `HALT` instruction in PROGRAM")
		}
	}
}

func execute_program(vm *VM) {
	if vm.inst_ptr == 0 {
		panic("No instruction to execute.. Load Program first")
	}
	counter := 0
	tot_len := vm.inst_ptr
	for (vm.vm_halt != 1 && counter < tot_len) {
		execute_inst(vm, vm.PROGRAM[counter % tot_len])
		// fmt.Printf("\n[%s : %d]\n",vm.PROGRAM[counter % tot_len].Name , vm.PROGRAM[counter % tot_len].Operand)
		// print_stack(vm)
		counter = (counter + 1) % tot_len
	}
}

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
