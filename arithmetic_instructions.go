package main

func addi(vm *VM) {
	if vm.stack_size < 2 {
		exit_with_one("Not enough values to add")
	}
	if !(operand_type_check(vm.STACK[vm.stack_size-1], "int64") && operand_type_check(vm.STACK[vm.stack_size-2], "int64")) {
		print_stack(vm, true)
		exit_with_one("Invalid Type: Explicitly Push Operands as Int for Integer Operands")
	}
	vm.STACK[vm.stack_size-2].int64holder = vm.STACK[vm.stack_size-1].int64holder + vm.STACK[vm.stack_size-2].int64holder
	vm.stack_size -= 1
	vm.inst_ptr += 1
}

func subi(vm *VM) {
	if vm.stack_size < 2 {
		exit_with_one("Not enough values to subtract")
	}
	if !(operand_type_check(vm.STACK[vm.stack_size-1], "int64") && operand_type_check(vm.STACK[vm.stack_size-2], "int64")) {
		print_stack(vm, true)
		exit_with_one("Invalid Type: Explicitly Push Operands as Int for Integer Operands")
	}
	vm.STACK[vm.stack_size-2].int64holder = vm.STACK[vm.stack_size-2].int64holder - vm.STACK[vm.stack_size-1].int64holder
	vm.stack_size -= 1
	vm.inst_ptr += 1
}

func muli(vm *VM) {
	if vm.stack_size < 2 {
		exit_with_one("Not enough values to multiply")
	}
	if !(operand_type_check(vm.STACK[vm.stack_size-1], "int64") && operand_type_check(vm.STACK[vm.stack_size-2], "int64")) {
		print_stack(vm, true)
		exit_with_one("Invalid Type: Explicitly Push Operands as Int for Integer Operands")
	}
	vm.STACK[vm.stack_size-2].int64holder = vm.STACK[vm.stack_size-2].int64holder * vm.STACK[vm.stack_size-1].int64holder
	vm.stack_size -= 1
	vm.inst_ptr += 1
}

func divi(vm *VM) {
	if vm.stack_size < 2 {
		exit_with_one("Not enough values to divide")
	}
	if vm.STACK[vm.stack_size-1].int64holder == 0 {
		print_stack(vm, true)
		exit_with_one("Zero Division Error")
	}
	if !(operand_type_check(vm.STACK[vm.stack_size-1], "int64") && operand_type_check(vm.STACK[vm.stack_size-2], "int64")) {
		print_stack(vm, true)
		exit_with_one("Invalid Type: Explicitly Push Operands as Int for Integer Operands")
	}
	vm.STACK[vm.stack_size-2].int64holder = vm.STACK[vm.stack_size-2].int64holder / vm.STACK[vm.stack_size-1].int64holder
	vm.stack_size -= 1
	vm.inst_ptr += 1
}

func eqi(vm *VM) {
	if vm.stack_size < 2 {
		exit_with_one("Not enough values for equality")
	}
	if !(operand_type_check(vm.STACK[vm.stack_size-1], "int64") && operand_type_check(vm.STACK[vm.stack_size-2], "int64")) {
		print_stack(vm, true)
		exit_with_one("Invalid Type: Explicitly Push Operands as Int for Integer Equality")
	}
	b := (vm.STACK[vm.stack_size-2].int64holder == vm.STACK[vm.stack_size-1].int64holder)
	if b {
		vm.STACK[vm.stack_size-2].int64holder = 1
	} else {
		vm.STACK[vm.stack_size-2].int64holder = 0
	}
	vm.stack_size -= 1
	vm.inst_ptr += 1
}

func addf(vm *VM) {
	if vm.stack_size < 2 {
		exit_with_one("Not enough values to add")
	}
	if !(operand_type_check(vm.STACK[vm.stack_size-1], "float64") && operand_type_check(vm.STACK[vm.stack_size-2], "float64")) {
		print_stack(vm, true)
		exit_with_one("Invalid Type: 'Implicit Conversion' to Float Not Yet Supported. Explicitly Push Operands as Float")
	}
	vm.STACK[vm.stack_size-2].float64holder = vm.STACK[vm.stack_size-1].float64holder + vm.STACK[vm.stack_size-2].float64holder
	vm.stack_size -= 1
	vm.inst_ptr += 1
}

func subf(vm *VM) {
	if vm.stack_size < 2 {
		exit_with_one("Not enough values to subtract")
	}
	if !(operand_type_check(vm.STACK[vm.stack_size-1], "float64") && operand_type_check(vm.STACK[vm.stack_size-2], "float64")) {
		print_stack(vm, true)
		exit_with_one("Invalid Type: 'Implicit Conversion' to Float Not Yet Supported. Explicitly Push Operands as Float")
	}
	vm.STACK[vm.stack_size-2].float64holder = vm.STACK[vm.stack_size-2].float64holder - vm.STACK[vm.stack_size-1].float64holder
	vm.stack_size -= 1
	vm.inst_ptr += 1
}

func mulf(vm *VM) {
	if vm.stack_size < 2 {
		exit_with_one("Not enough values to multiply")
	}
	if !(operand_type_check(vm.STACK[vm.stack_size-1], "float64") && operand_type_check(vm.STACK[vm.stack_size-2], "float64")) {
		print_stack(vm, true)
		exit_with_one("Invalid Type: 'Implicit Conversion' to Float Not Yet Supported. Explicitly Push Operands as Float")
	}
	vm.STACK[vm.stack_size-2].float64holder = vm.STACK[vm.stack_size-2].float64holder * vm.STACK[vm.stack_size-1].float64holder
	vm.stack_size -= 1
	vm.inst_ptr += 1
}

func divf(vm *VM) {
	if vm.stack_size < 2 {
		exit_with_one("Not enough values to divide")
	}
	if vm.STACK[vm.stack_size-1].float64holder == 0.0 {
		print_stack(vm, true)
		exit_with_one("Zero Division Error")
	}
	if !(operand_type_check(vm.STACK[vm.stack_size-1], "float64") && operand_type_check(vm.STACK[vm.stack_size-2], "float64")) {
		print_stack(vm, true)
		exit_with_one("Invalid Type: 'Implicit Conversion' to Float Not Yet Supported. Explicitly Push Operands as Float")
	}
	vm.STACK[vm.stack_size-2].float64holder = vm.STACK[vm.stack_size-2].float64holder / vm.STACK[vm.stack_size-1].float64holder
	vm.stack_size -= 1
	vm.inst_ptr += 1
}

func eqf(vm *VM) {
	if vm.stack_size < 2 {
		exit_with_one("Not enough values for equality")
	}
	if !(operand_type_check(vm.STACK[vm.stack_size-1], "float64") && operand_type_check(vm.STACK[vm.stack_size-2], "float64")) {
		print_stack(vm, true)
		exit_with_one("Invalid Type: Explicitly Push Operands as Float for Float Equality")
	}
	b := (vm.STACK[vm.stack_size-2].float64holder == vm.STACK[vm.stack_size-1].float64holder)
	reset_operand_except(&vm.STACK[vm.stack_size-2], "int64")
	if b {
		vm.STACK[vm.stack_size-2].int64holder = 1
	} else {
		vm.STACK[vm.stack_size-2].int64holder = 0
	}
	vm.stack_size -= 1
	vm.inst_ptr += 1
}
