package main

import (
	"fmt"
)

const STACK_CAPACITY = 1024

type VM struct {
	stack_size int
	STACK      [STACK_CAPACITY]int
	PROGRAM    []Inst
	inst_ptr   int
}
type Inst struct {
	Name       string
	Value      int
	Is_Operand bool
}

func push(vm *VM, inst Inst) {
	vm.STACK[vm.stack_size] = inst.Value
	vm.stack_size += 1
}

func add(vm *VM, inst Inst) {
	if vm.stack_size < 2 {
		panic("Not enough values to add")
	}
	vm.STACK[vm.stack_size-2] = vm.STACK[vm.stack_size-1] + vm.STACK[vm.stack_size-2]
	vm.stack_size -= 1
}
func sub(vm *VM, inst Inst) {
	if vm.stack_size < 2 {
		panic("Not enough values to subtract")
	}
	vm.STACK[vm.stack_size-2] = vm.STACK[vm.stack_size-2] - vm.STACK[vm.stack_size-1]
	vm.stack_size -= 1
}
func mul(vm *VM, inst Inst) {
	if vm.stack_size < 2 {
		panic("Not enough values to multiply")
	}
	vm.STACK[vm.stack_size-2] = vm.STACK[vm.stack_size-2] * vm.STACK[vm.stack_size-1]
	vm.stack_size -= 1
}

func peek(vm *VM) int {
	if vm.stack_size == 0 {
		panic("Empty Stack")
	}
	return vm.STACK[vm.stack_size-1]
}

func push_inst(vm *VM, inst Inst) {
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
		add(vm, inst)
	case "SUB":
		sub(vm, inst)
	case "MUL":
		mul(vm, inst)
	default:
		panic("Unknown Instruction")
	}
	vm.PROGRAM = append(vm.PROGRAM, inst)
	vm.inst_ptr += 1

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
			fmt.Printf("%s : %d \n", vm.PROGRAM[i].Name, vm.PROGRAM[i].Value)
		case "ADD":
			fmt.Printf("%s \n", vm.PROGRAM[i].Name)
		case "SUB":
			fmt.Printf("%s \n", vm.PROGRAM[i].Name)
		case "MUL":
			fmt.Printf("%s \n", vm.PROGRAM[i].Name)
		}
	}
	if banner {
		fmt.Println("---- PROGRAM TRACE END ----")
	}
}
func main() {
	var initial [STACK_CAPACITY]int
	var initial_inst []Inst
	vm_g := VM{stack_size: 0, STACK: initial, PROGRAM: initial_inst, inst_ptr: -1}
	push_inst(&vm_g, Inst{Name: "PUSH", Value: 10, Is_Operand: true})
	push_inst(&vm_g, Inst{Name: "PUSH", Value: 10, Is_Operand: true})
	push_inst(&vm_g, Inst{Name: "PUSH", Value: 10, Is_Operand: true})
	push_inst(&vm_g, Inst{Name: "PUSH", Value: 20, Is_Operand: true})
	print_stack(&vm_g)
	push_inst(&vm_g, Inst{Name: "ADD", Value: 0, Is_Operand: true})
	print_stack(&vm_g)
	// print_program_trace(&vm_g)
	push_inst(&vm_g, Inst{Name: "MUL", Value: 0, Is_Operand: true})
	print_stack(&vm_g)
	push_inst(&vm_g, Inst{Name: "PUSH", Value: 10, Is_Operand: true})
	print_stack(&vm_g)
	push_inst(&vm_g, Inst{Name: "SUB", Value: 10, Is_Operand: true})
	print_stack(&vm_g)
	print_program_trace(&vm_g, true)

	// fmt.Println(vm_g)
}
