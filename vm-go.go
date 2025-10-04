package main

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"errors"
	"flag"
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"
)

const STACK_CAPACITY = 1024
const PROGRAM_CAPACITY = 1024
const LABEL_CAPACITY = 1024
const UNRESOLVED_JUMPS_CAPACITY = 1024
const INCLUDE_FILE_PATH_CAPACITY = 1024

var debug bool

const MaxUint = ^uint(0)
const MaxInt = int(MaxUint >> 1)
const MinInt = -MaxInt - 1
const MinFloat = math.SmallestNonzeroFloat64

var Inst_ARR = []string{
	"push",
	"addi",
	"subi",
	"muli",
	"divi",
	"addf",
	"subf",
	"mulf",
	"divf",
	"jmp",
	"halt",
	"nop",
	"ret",
	"dup",
	"swap",
	"call",
	"drop",
	"jmp_if",
	"not",
	"eqi",
	"eqf",
	"print",
	"ignore_halt",
}
var Constant_Mapping_int = make(map[string]int64)
var Constant_Mapping_float = make(map[string]float64)
var Constant_Mapping_string = make(map[string]string)

type Value_Holder struct {
	int64holder   int64
	float64holder float64
	pointer       string
}

type VM struct {
	stack_size int64
	STACK      [STACK_CAPACITY]Value_Holder

	PROGRAM      [PROGRAM_CAPACITY]Inst
	inst_ptr     int64
	program_size int64

	vm_halt int64
}

type Inst struct {
	Name    string
	Operand Value_Holder
}

type Label struct {
	Name string
	addr int64
}

type Label_Table struct {
	labels     [LABEL_CAPACITY]Label
	table_size int64
}

var lt_g Label_Table

type Deferred_Operand struct {
	deferred_oprnd_addr  int64
	deferred_oprnd_label string
	deferred_oprnd_line  int64
}

type Deferred_Operands struct {
	deferred_operand_arr   [UNRESOLVED_JUMPS_CAPACITY]Deferred_Operand
	deferred_operands_size int64
}

var deferredoprnds_g Deferred_Operands
var include_file_path_array [INCLUDE_FILE_PATH_CAPACITY]string
var include_file_path_array_size = 0

func check_err(e error) {
	if e != nil {
		panic(e)
	}
}

func exit_with_one(message string) {
	if debug {
		panic(message)
	} else {
		fmt.Println(message)
		os.Exit(1)
	}
}

func assert_runtime(cond bool, message string) {
	if !cond {
		fmt.Println("Runtime Assertion Error")
		panic(message)
	}
}

func prompt_for_debug() {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("\n-> Press Enter")
	_, _, err := reader.ReadRune()
	check_err(err)
	fmt.Println()
}
func push(vm *VM, inst Inst) {
	vm.STACK[vm.stack_size] = inst.Operand
	vm.stack_size += 1
	vm.inst_ptr += 1
}

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

func operand_type_check(op Value_Holder, expected_name string) bool {
	return get_operand_type_by_name(op) == expected_name
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

// All Operands are initialized as Value_Holder{int64holder: MinInt , float64holder: MinFloat, pointer:""}
func get_operand_type_by_name(operand Value_Holder) string {
	if operand.float64holder != float64(math.SmallestNonzeroFloat64) {
		return "float64"
	}
	if operand.int64holder != int64(MinInt) {
		return "int64"
	}
	if operand.pointer != "" {
		return "pointer"
	}
	panic("Unreachable")
}

func reset_operand_except(operand *Value_Holder, name string) {
	_ = operand
	switch name {
	case "int64":
		operand = &Value_Holder{float64holder: math.SmallestNonzeroFloat64}
	case "float64":
		operand = &Value_Holder{int64holder: int64(MinInt)}
	case "pointer":
		operand = &Value_Holder{pointer: ""}
	}

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

func execute_inst(vm *VM, inst Inst) {
	if vm.inst_ptr >= vm.program_size {
		fmt.Printf("Instruction : %s : %v\n", inst.Name, inst.Operand)
		exit_with_one("Illegal Instruction Access")
	}
	if vm.stack_size < 0 {
		exit_with_one("Stack Underflow")
	}
	if vm.stack_size > STACK_CAPACITY {
		exit_with_one("Stack Overflow")
	}
	switch inst.Name {
	case "PUSH":
		push(vm, inst)
	case "ADDI":
		addi(vm)
	case "SUBI":
		subi(vm)
	case "MULI":
		muli(vm)
	case "DIVI":
		divi(vm)
	case "EQI":
		eqi(vm)
	case "ADDF":
		addf(vm)
	case "SUBF":
		subf(vm)
	case "MULF":
		mulf(vm)
	case "DIVF":
		divf(vm)
	case "EQF":
		eqf(vm)
	case "JMP":
		jmp(vm, inst)
	case "JMP_IF":
		jmp_if(vm, inst)
	case "HALT":
		halt(vm)
	case "NOT":
		not(vm)
	case "NOP":
		nop(vm)
	case "DROP":
		drop(vm)
	case "RET":
		ret(vm)
	case "CALL":
		call(vm, inst)
	case "DUP":
		dup(vm, inst)
	case "SWAP":
		swap(vm, inst)
	case "PRINT":
		print(vm)
	case "DEFINE":
		nop(vm)
	case "INCLUDE":
		nop(vm)
	case "IGNORE_HALT":
		nop(vm)
	default:
		panic("Unknown Instruction")
	}
	// vm.PROGRAM[vm.inst_ptr] =  inst
	// vm.inst_ptr += 1

}

func print_stack(vm *VM, reverse bool) {
	if vm.stack_size < 0 {
		exit_with_one("ERROR: Stack Underflow")
	}

	fmt.Println("---- STACK BEG ----")
	if reverse {
		for i := vm.stack_size - 1; i >= 0; i-- {
			fmt.Println(vm.STACK[i])
		}
	} else {
		for i := int64(0); i < vm.stack_size; i++ {
			fmt.Println(vm.STACK[i])
		}
	}
	fmt.Println("---- STACK END ----")
	fmt.Println()
}

func print_program_trace(vm *VM, banner bool) {
	if vm.program_size == 0 {
		exit_with_one("Empty Program")
	}
	if vm.program_size >= PROGRAM_CAPACITY {
		exit_with_one("Overflow: vm.program_size >= PROGRAM_CAPACITY")
	}
	if banner {
		fmt.Println("---- PROGRAM TRACE BEG ----")
	}

	for i := vm.program_size - 1; i >= 0; i-- {
		switch vm.PROGRAM[i].Name {
		case "PUSH":
			fmt.Printf("%s : %+v \n", vm.PROGRAM[i].Name, vm.PROGRAM[i].Operand)
		case "ADDI":
			fmt.Printf("%s \n", vm.PROGRAM[i].Name)
		case "SUBI":
			fmt.Printf("%s \n", vm.PROGRAM[i].Name)
		case "MULI":
			fmt.Printf("%s \n", vm.PROGRAM[i].Name)
		case "DIVI":
			fmt.Printf("%s \n", vm.PROGRAM[i].Name)
		case "EQI":
			fmt.Printf("%s \n", vm.PROGRAM[i].Name)
		case "ADDF":
			fmt.Printf("%s \n", vm.PROGRAM[i].Name)
		case "SUBF":
			fmt.Printf("%s \n", vm.PROGRAM[i].Name)
		case "MULF":
			fmt.Printf("%s \n", vm.PROGRAM[i].Name)
		case "DIVF":
			fmt.Printf("%s \n", vm.PROGRAM[i].Name)
		case "EQF":
			fmt.Printf("%s \n", vm.PROGRAM[i].Name)
		case "JMP":
			fmt.Printf("%s : %+v \n", vm.PROGRAM[i].Name, vm.PROGRAM[i].Operand)
		case "JMP_IF":
			fmt.Printf("%s : %+v \n", vm.PROGRAM[i].Name, vm.PROGRAM[i].Operand)
		case "CALL":
			fmt.Printf("%s : %+v \n", vm.PROGRAM[i].Name, vm.PROGRAM[i].Operand)
		case "HALT":
			fmt.Printf("%s \n", vm.PROGRAM[i].Name)
		case "NOP":
			fmt.Printf("%s \n", vm.PROGRAM[i].Name)
		case "NOT":
			fmt.Printf("%s \n", vm.PROGRAM[i].Name)
		case "DROP":
			fmt.Printf("%s \n", vm.PROGRAM[i].Name)
		case "RET":
			fmt.Printf("%s \n", vm.PROGRAM[i].Name)
		case "DUP":
			fmt.Printf("%s : %+v \n", vm.PROGRAM[i].Name, vm.PROGRAM[i].Operand)
		case "SWAP":
			fmt.Printf("%s : %+v \n", vm.PROGRAM[i].Name, vm.PROGRAM[i].Operand)
		case "PRINT":
			fmt.Printf("%s \n", vm.PROGRAM[i].Name)
		case "DEFINE":
			fmt.Printf("%s \n", vm.PROGRAM[i].Name)
		case "INCLUDE":
			fmt.Printf("%s \n", vm.PROGRAM[i].Name)
		case "IGNORE_HALT":
			fmt.Printf("%s \n", vm.PROGRAM[i].Name)
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
		exit_with_one("Overflow")
	}
	if debug {
		fmt.Println()
	}
	halt_flag := false
	for i := 0; i < program_size; i++ {
		if program[i].Name == "HALT" {
			halt_flag = true
		}
		vm.PROGRAM[vm.program_size] = program[i]
		vm.program_size += 1
		if debug {
			fmt.Printf("Loaded Instruction: %s : %+v\n", vm.PROGRAM[vm.program_size-1].Name, vm.PROGRAM[vm.program_size-1].Operand)
		}
	}
	if !halt_flag {
		if halt_panic {
			print_program_trace(vm, true)
			exit_with_one("No `HALT` instruction in PROGRAM")
		}
	}
	if debug {
		fmt.Println()
	}
}

func process_comment(line string) string {
	if line == "" {
		return line
	}
	if string(line[0]) == "#" {
		return ""
	}
	last_index := strings.LastIndex(line, "#")
	if last_index > 0 {
		return string(line[0:last_index])
	}
	return line
}

func chk_if_tok_is_inst(token string) bool {
	assert_runtime(len(Inst_ARR) == 23, "Number of Instructions have changed")
	for _, el := range Inst_ARR {
		if el == strings.ToLower(token) {
			return true
		}
	}
	return false
}

func check_if_label_and_push_to_label_table(vm *VM, lt *Label_Table, s string, line_number int, file_path string) (bool, string) {
	tmp_slice := strings.Split(s, ":")
	if string(tmp_slice[0]) == s {
		return false, s
	}
	label_name := strings.Trim(string(tmp_slice[0]), " ")
	if chk_if_tok_is_inst(label_name) {
		fmt.Printf("File : %s\n", file_path)
		fmt.Printf("ERROR: Error near line %d : `%s`\n", line_number, s)
		exit_with_one("Label Cannot be an Instruction")
	}
	lt.labels[lt.table_size] = Label{Name: label_name, addr: vm.program_size}
	lt.table_size += 1
	return true, strings.Trim(strings.Join(tmp_slice[1:], " "), " ")
}

func find_label_in_label_table(lt Label_Table, label_name string) int64 {
	for i := int64(0); i < lt.table_size; i++ {
		if lt.labels[i].Name == label_name {
			return lt.labels[i].addr
		}
	}
	return -1
}

func push_to_deferred_operand_table(vm *VM, unrslvdjmps *Deferred_Operands, label_name string, line_number int64) {
	tmp_do := Deferred_Operand{deferred_oprnd_addr: vm.program_size, deferred_oprnd_label: label_name, deferred_oprnd_line: line_number}
	unrslvdjmps.deferred_operand_arr[unrslvdjmps.deferred_operands_size] = tmp_do
	unrslvdjmps.deferred_operands_size += 1
}

func report_error(err error, line_number int, error_string string, file_path string, use_line_number bool) {
	if err != nil {
		if use_line_number {
			fmt.Printf("File : %s\n", file_path)
			fmt.Printf("ERROR: Error near line %d : %s\n", line_number, error_string)
		}
		fmt.Println(err)
		os.Exit(1)
	}
}

func peek_next_token(line string) (string, int) {
	s := strings.Split(line, " ")
	pos := 0
	found := false
	for _, el := range s {
		if el != "" {
			found = true
		}
		if found {
			return el, pos
		}
		if len(el) == 0 {
			pos = pos + 1
		} else {
			pos = pos + len(el) - 1
		}
	}
	return "", -1
}

func parse_and_load_define_operands(strs_to_process string, file_path string) (err error) {
	var_name, pos := peek_next_token(strs_to_process)
	if pos == -1 {
		err = errors.New("MissingArguments")
		return err
	}
	_, found_int := Constant_Mapping_int[var_name]
	_, found_float := Constant_Mapping_float[var_name]
	_, found_string := Constant_Mapping_string[var_name]
	if found_int || found_float || found_string {
		fmt.Printf("File : %s\n", file_path)
		fmt.Printf("Constant `%s` is already defined.\n", var_name)
		exit_with_one("Redifinition of Constant")
	}
	runes := []rune(strs_to_process)
	space_pos := pos + len(var_name)
	if space_pos >= len(runes) {
		err = errors.New("MissingArguments")
		return err
	}
	// This below check is not needed... need to rethink
	if runes[pos+len(var_name)] != ' ' {
		err = errors.New("MissingArguments")
		return err
	}
	var operand string
	for i := space_pos; i < len(runes); i++ {
		operand = operand + string(runes[i])
	}
	operand = strings.Trim(operand, " ")
	parsed_str, str_parse_err := parse_string_literal(operand)
	if str_parse_err == nil {
		Constant_Mapping_string[var_name] = parsed_str
		return nil
	}

	if strings.Contains(operand, ".") || strings.Contains(operand, "e") {
		operand, err := strconv.ParseFloat(operand, 64)
		if err != nil {
			return err
		}
		Constant_Mapping_float[var_name] = float64(operand)
	} else {
		operand, err := strconv.Atoi(operand)
		if err != nil {
			return err
		}
		Constant_Mapping_int[var_name] = int64(operand)
	}

	return nil
}

func parse_string_literal(str string) (string, error) {
	str = strings.Trim(str, " ")
	var err error
	var result string
	runes := []rune(str)
	initial_failures := (runes[0] != '"' || runes[len(runes)-1] != '"')

	if initial_failures {
		err = errors.New("InvalidStringLiteral")
		return "", err
	}
	for i := 1; i < len(runes)-1; i++ {
		result = result + string(runes[i])
	}
	return result, nil
}

func load_program_from_file(vm *VM, file_path string, halt_panic bool) {
	dat, err := os.ReadFile(file_path)
	check_err(err)
	file_content := string(dat)
	lines := strings.Split(strings.ReplaceAll(file_content, "\r\n", "\n"), "\n")
	instruction_count := 0
	halt_flag := false
	ignore_halt := false
	if debug {
		fmt.Println()
	}

	lex_and_parse_program(lines, file_path, vm, ignore_halt, &halt_flag, &instruction_count)

	for i := int64(0); i < deferredoprnds_g.deferred_operands_size; i++ {
		ind := find_label_in_label_table(lt_g, deferredoprnds_g.deferred_operand_arr[i].deferred_oprnd_label)
		if ind == -1 {
			error_line_number := deferredoprnds_g.deferred_operand_arr[i].deferred_oprnd_line
			fmt.Printf("Unknown Label near line %d : `%s` \n", error_line_number, deferredoprnds_g.deferred_operand_arr[i].deferred_oprnd_label)
			exit_with_one("Unknown Label")
		}
		vm.PROGRAM[deferredoprnds_g.deferred_operand_arr[i].deferred_oprnd_addr].Operand.int64holder = ind
	}

	if !halt_flag {
		if halt_panic {
			print_program_trace(vm, true)
			exit_with_one("No `HALT` instruction in PROGRAM")
		}
	}
	if debug {
		fmt.Println()
	}
}

func compile_program_to_binary(vm *VM, file_path string) {
	output_file_path := strings.ReplaceAll(file_path, ".vasm", ".vm")
	if vm.program_size == 0 {
		exit_with_one("Empty Program.. Cannot Compile to binary")
	}
	file, err := os.Create(output_file_path)
	if err != nil {
		exit_with_one("Cannot Open file to write")
	}
	// type inst_data struct {
	// 	Name_tmp          string
	// 	int64holder_tmp   int64
	// 	float64holder_tmp float64
	// 	pointer_tmp       string
	// }
	fmt.Printf("Compiling '%s' to '%s'\n", file_path, output_file_path)
	defer file.Close()
	for i := int64(0); i < vm.program_size; i++ {
		buf := new(bytes.Buffer)
		var data = []interface{}{
			[]byte(vm.PROGRAM[i].Name),
			int64(vm.PROGRAM[i].Operand.int64holder),
			float64(vm.PROGRAM[i].Operand.float64holder),
			[]byte(vm.PROGRAM[i].Operand.pointer),
		}
		for _, v := range data {
			err := binary.Write(buf, binary.LittleEndian, v)
			if err != nil {
				fmt.Println("binary.Write failed:", err)
				panic(err)
			}
		}
		writeNextBytes(file, buf.Bytes())

	}
	fmt.Println("Binary Written To:", output_file_path)
	fmt.Println()
}

func writeNextBytes(file *os.File, bytes []byte) {
	_, err := file.Write(bytes)
	if err != nil {
		exit_with_one("Cannot write to file")
	}
}

func execute_program(vm *VM, limit int) {
	if vm.program_size == 0 {
		exit_with_one("No instruction to execute.. Load Program first")
	}
	counter := 0
	for vm.vm_halt != 1 && counter < limit {
		if debug {
			print_stack(vm, true)
			fmt.Printf("IP : %d\n", vm.inst_ptr)
			fmt.Printf("STEP(%d) Instruction to be executed : `%s : %+v`\n", (counter + 1), vm.PROGRAM[vm.inst_ptr].Name, vm.PROGRAM[vm.inst_ptr].Operand)
			prompt_for_debug()
		}
		execute_inst(vm, vm.PROGRAM[vm.inst_ptr])
		counter += 1
	}
}

func init_all(initial_stack *[STACK_CAPACITY]Value_Holder, initial_program *[PROGRAM_CAPACITY]Inst) {
	for i := 0; i < STACK_CAPACITY; i++ {
		initial_stack[i] = Value_Holder{int64holder: int64(MinInt), float64holder: float64(MinFloat), pointer: ""}
	}
	for i := 0; i < PROGRAM_CAPACITY; i++ {
		initial_program[i] = Inst{Name: "", Operand: Value_Holder{int64holder: int64(MinInt), float64holder: float64(MinFloat), pointer: ""}}
	}
	Constant_Mapping_int = make(map[string]int64)
	Constant_Mapping_float = make(map[string]float64)
	Constant_Mapping_string = make(map[string]string)
}

func main() {
	var initial_stack [STACK_CAPACITY]Value_Holder
	var initial_program [PROGRAM_CAPACITY]Inst
	init_all(&initial_stack, &initial_program)

	lt_g = Label_Table{}
	deferredoprnds_g = Deferred_Operands{}

	vm_g := VM{STACK: initial_stack, PROGRAM: initial_program}

	file_path := flag.String("i", "", ".vasm FILE PATH")
	execution_limit_steps_inp := flag.Int("limit", 69, "Execution Limit Steps")
	debug_flg := flag.Bool("debug", false, "Enable Debugger")
	compile_flg := flag.Bool("compile", false, "Compile VASM to native Binary .vm")

	flag.Parse()

	debug = *debug_flg
	if *file_path == "" {
		fmt.Println("No input .vasm file is provided. Use '-h' option for help")
		os.Exit(0)
	}
	load_program_from_file(&vm_g, *file_path, false)
	if debug {
		print_program_trace(&vm_g, true)
	}
	if *compile_flg {
		compile_program_to_binary(&vm_g, *file_path)
		if include_file_path_array_size > 0 {
			for i := 0; i < include_file_path_array_size; i++ {
				var initial_stack_include [STACK_CAPACITY]Value_Holder
				var initial_program_include [PROGRAM_CAPACITY]Inst
				init_all(&initial_stack_include, &initial_program_include)
				vm_temp := VM{STACK: initial_stack_include, PROGRAM: initial_program_include}
				load_program_from_file(&vm_temp, include_file_path_array[i], false)
				compile_program_to_binary(&vm_temp, include_file_path_array[i])
			}
		}
	}
	execute_program(&vm_g, *execution_limit_steps_inp)
	// print_stack(&vm_g, false)
}
