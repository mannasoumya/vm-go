package main

import "fmt"

func push(vm *VM, inst Inst) {
	vm.STACK[vm.stack_size] = inst.Operand
	vm.stack_size += 1
	vm.inst_ptr += 1
}

func peek(vm *VM) Value_Holder {
	if vm.stack_size == 0 {
		exit_with_one("Empty Stack")
	}
	return vm.STACK[vm.stack_size-1]
}

func jmp(vm *VM, inst Inst) {
	if inst.Operand.int64holder < 0 {
		exit_with_one("Wrong Jump Instruction. Underflow")
	}
	if inst.Operand.int64holder >= vm.program_size {
		exit_with_one("Wrong Jump Instruction. Overflow")
	}
	vm.inst_ptr = inst.Operand.int64holder
}

func nop(vm *VM) {
	vm.inst_ptr += 1
}

func print(vm *VM) {
	if vm.stack_size < 1 {
		exit_with_one("Not enough values on the stack to print")
	}
	type_of_operand := get_operand_type_by_name(vm.STACK[vm.stack_size-1])
	switch type_of_operand {
	case "int64":
		fmt.Printf("%d\n", vm.STACK[vm.stack_size-1].int64holder)
	case "float64":
		fmt.Printf("%f\n", vm.STACK[vm.stack_size-1].float64holder)
	default:
		fmt.Printf("%s\n", vm.STACK[vm.stack_size-1].pointer)
	}
	vm.stack_size -= 1
	vm.inst_ptr += 1
}

func print_asc(vm *VM) {
	if vm.stack_size < 1 {
		exit_with_one("Not enough values on the stack to print")
	}
	type_of_operand := get_operand_type_by_name(vm.STACK[vm.stack_size-1])
	switch type_of_operand {
	case "int64":
		fmt.Printf("%s", string(rune(vm.STACK[vm.stack_size-1].int64holder)))
	default:
		fmt.Printf("\nERROR: Runtime error: Instruction `%s` failed: Expected type `integer` on top of stack but found `%s`\n", vm.PROGRAM[vm.inst_ptr].Name, type_of_operand)
	}
	vm.stack_size -= 1
	vm.inst_ptr += 1
}

func halt(vm *VM) {
	vm.vm_halt = 1
}

func drop(vm *VM) {
	if vm.stack_size < 1 {
		exit_with_one("STACK UNDERFLOW")
	}
	vm.stack_size -= 1
	vm.inst_ptr += 1
}

func ret(vm *VM) {
	if vm.stack_size < 1 {
		exit_with_one("Stack Underflow")
	}
	vm.inst_ptr = vm.STACK[vm.stack_size-1].int64holder
	vm.stack_size -= 1
}

func call(vm *VM, inst Inst) {
	if vm.stack_size >= STACK_CAPACITY {
		exit_with_one("Stack Overflow")
	}
	reset_operand_except(&vm.STACK[vm.stack_size], "int64")
	vm.STACK[vm.stack_size].int64holder = vm.inst_ptr + 1
	vm.stack_size += 1
	vm.inst_ptr = inst.Operand.int64holder
}

func dup(vm *VM, inst Inst) {
	if vm.stack_size >= STACK_CAPACITY {
		exit_with_one("Stack Overflow")
	}

	if vm.stack_size-inst.Operand.int64holder <= 0 {
		exit_with_one("Stack Underflow")
	}

	inst_name_to_be_assigned := get_operand_type_by_name(vm.STACK[vm.stack_size-1-inst.Operand.int64holder])
	if inst_name_to_be_assigned == "float64" {
		reset_operand_except(&vm.STACK[vm.stack_size], "float64")
		vm.STACK[vm.stack_size].float64holder = vm.STACK[vm.stack_size-1-inst.Operand.int64holder].float64holder
	}
	if inst_name_to_be_assigned == "int64" {
		reset_operand_except(&vm.STACK[vm.stack_size], "int64")
		vm.STACK[vm.stack_size].int64holder = vm.STACK[vm.stack_size-1-inst.Operand.int64holder].int64holder
	}

	vm.stack_size += 1
	vm.inst_ptr += 1
}

func swap(vm *VM, inst Inst) {
	if inst.Operand.int64holder >= vm.stack_size {
		exit_with_one("Stack Underflow")
	}
	a := vm.stack_size - 1
	b := vm.stack_size - 1 - inst.Operand.int64holder
	t := vm.STACK[a]
	vm.STACK[a] = vm.STACK[b]
	vm.STACK[b] = t
	vm.inst_ptr += 1
}

func jmp_if(vm *VM, inst Inst) {
	if inst.Operand.int64holder >= vm.program_size {
		exit_with_one("Wrong Jump_If Instruction. Overflow")
	}
	if vm.stack_size < 1 {
		exit_with_one("Wrong Jump_If Instruction. Underflow")
	}
	tmp_chk := operand_type_check(vm.STACK[vm.stack_size-1], "int64") && inst.Operand.int64holder != 0
	if tmp_chk {
		vm.inst_ptr = inst.Operand.int64holder
	} else {
		vm.inst_ptr += 1
	}
	vm.stack_size -= 1
}

func not(vm *VM) {
	if vm.stack_size < 0 {
		exit_with_one("Stack Underflow")
	}
	tmp_chk := operand_type_check(vm.STACK[vm.stack_size-1], "int64")
	if tmp_chk && vm.STACK[vm.stack_size-1].int64holder != 0 {
		vm.STACK[vm.stack_size-1].int64holder = 0
	} else {
		vm.STACK[vm.stack_size-1].int64holder = 1
	}
	vm.inst_ptr += 1
}
