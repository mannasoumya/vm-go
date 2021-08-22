package main

import (
	"fmt"
	"io/ioutil"
	"strings"
	"strconv"
	"flag"
	"bufio"
	"os"
	"unicode"
	"math"
)

const STACK_CAPACITY = 1024
const PROGRAM_CAPACITY = 1024
const LABEL_CAPACITY = 1024
const UNRESOLVED_JUMPS_CAPACITY = 1024
var debug bool
const MaxUint = ^uint(0)
const MaxInt = int(MaxUint >> 1) 
const MinInt = -MaxInt - 1
const MinFloat = math.SmallestNonzeroFloat64

type Value_Holder struct{
	int64holder   int64
	float64holder float64
	pointer       string
}

type VM struct {
	stack_size   int64
	STACK        [STACK_CAPACITY]Value_Holder

	PROGRAM      [PROGRAM_CAPACITY]Inst
	inst_ptr     int64
	program_size int64
	
	vm_halt      int64
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

type Unresolved_Jump struct {
	unresolved_jmp_addr   int64
	unresolved_jump_label string
	unresolved_jmp_line   int64
}

type Unresolved_Jumps struct {
	unresolved_jump_arr   [UNRESOLVED_JUMPS_CAPACITY]Unresolved_Jump
	unresolved_jumps_size int64 
}

var unrslvdjmps_g Unresolved_Jumps

func check_err(e error) {
    if e != nil {
        panic(e)
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
		panic("Not enough values to add")
	}
	vm.STACK[vm.stack_size-2].int64holder = vm.STACK[vm.stack_size-1].int64holder + vm.STACK[vm.stack_size-2].int64holder
	vm.stack_size -= 1
	vm.inst_ptr += 1
}

func subi(vm *VM) {
	if vm.stack_size < 2 {
		panic("Not enough values to subtract")
	}
	vm.STACK[vm.stack_size-2].int64holder = vm.STACK[vm.stack_size-2].int64holder - vm.STACK[vm.stack_size-1].int64holder
	vm.stack_size -= 1
	vm.inst_ptr += 1
}

func muli(vm *VM) {
	if vm.stack_size < 2 {
		panic("Not enough values to multiply")
	}
	vm.STACK[vm.stack_size-2].int64holder = vm.STACK[vm.stack_size-2].int64holder * vm.STACK[vm.stack_size-1].int64holder
	vm.stack_size -= 1
	vm.inst_ptr += 1
}

func divi(vm *VM) {
	if vm.stack_size < 2 {
		panic("Not enough values to divide")
	}
	if vm.STACK[vm.stack_size-1].int64holder == 0 {
		print_stack(vm, true)
		panic("Zero Division Error")
	}
	vm.STACK[vm.stack_size-2].int64holder = vm.STACK[vm.stack_size-2].int64holder / vm.STACK[vm.stack_size-1].int64holder
	vm.stack_size -= 1
	vm.inst_ptr += 1
}

func peek(vm *VM) Value_Holder {
	if vm.stack_size == 0 {
		panic("Empty Stack")
	}
	return vm.STACK[vm.stack_size-1]
}

func jmp(vm *VM, inst Inst) {
	if inst.Operand.int64holder < 0 {
		panic("Wrong Jump Instruction. Underflow")
	}
	if inst.Operand.int64holder >= vm.program_size {
		panic("Wrong Jump Instruction. Overflow")
	}
	vm.inst_ptr = inst.Operand.int64holder
}

func nop(vm *VM) {
	vm.inst_ptr += 1
}

func halt(vm *VM) {
	vm.vm_halt = 1
}

func ret(vm *VM) {
	vm.inst_ptr = vm.STACK[vm.stack_size - 1].int64holder
	vm.stack_size -= 1;
}
// All Operands are initialized as Value_Holder{int64holder: MinInt , float64holder: MinFloat}
func get_operand_type_by_name(operand Value_Holder) string {
	if operand.float64holder != float64(math.SmallestNonzeroFloat64) {
		return "float64"
	}
	
	if operand.int64holder != int64(MinInt) {
		return "int64"
	}

	panic("Pointers/Strings Not Implemented Yes")
}

func reset_operand_except(operand *Value_Holder, name string) {
	switch name {
		case "int64":
			operand = &Value_Holder{float64holder: math.SmallestNonzeroFloat64}
		case "float64":
			operand = &Value_Holder{int64holder: int64(MinInt)}
	}
}

func dup(vm *VM, inst Inst) {
	if vm.stack_size >= STACK_CAPACITY {
		panic("Stack Overflow")
	}
	
	if (vm.stack_size - inst.Operand.int64holder <= 0) {
		panic("Stack Underflow")
	}
	
	
	inst_name_to_be_assigned := get_operand_type_by_name(vm.STACK[vm.stack_size - 1 - inst.Operand.int64holder])
	if inst_name_to_be_assigned == "float64" {
		reset_operand_except(&vm.STACK[vm.stack_size],"float64")
		vm.STACK[vm.stack_size].float64holder = vm.STACK[vm.stack_size - 1 - inst.Operand.int64holder].float64holder
	}
	if  inst_name_to_be_assigned == "int64" {
		reset_operand_except(&vm.STACK[vm.stack_size],"int64")
		vm.STACK[vm.stack_size].int64holder = vm.STACK[vm.stack_size - 1 - inst.Operand.int64holder].int64holder
	}
	
	vm.stack_size += 1;
	vm.inst_ptr += 1;
}

func execute_inst(vm *VM, inst Inst) {
	if vm.inst_ptr >= vm.program_size {
		fmt.Printf("Instruction : %s : %d\n", inst.Name, inst.Operand)
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
	case "ADDI":
		addi(vm)
	case "SUBI":
		subi(vm)
	case "MULI":
		muli(vm)
	case "DIVI":
		divi(vm)
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

func print_stack(vm *VM, reverse bool) {
	if vm.stack_size < 0 {
		panic("ERROR: Stack Underflow")
	}
	
	fmt.Println("---- STACK BEG ----")
	if reverse == true {
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
			fmt.Printf("%s : %+v \n", vm.PROGRAM[i].Name, vm.PROGRAM[i].Operand)
		case "ADDI":
			fmt.Printf("%s \n", vm.PROGRAM[i].Name)
		case "SUBI":
			fmt.Printf("%s \n", vm.PROGRAM[i].Name)
		case "MULI":
			fmt.Printf("%s \n", vm.PROGRAM[i].Name)
		case "DIVI":
			fmt.Printf("%s \n", vm.PROGRAM[i].Name)
		case "JMP":
			fmt.Printf("%s : %+v \n", vm.PROGRAM[i].Name, vm.PROGRAM[i].Operand)
		case "HALT":
			fmt.Printf("%s \n", vm.PROGRAM[i].Name)
		case "NOP":
			fmt.Printf("%s \n", vm.PROGRAM[i].Name)
		case "RET":
			fmt.Printf("%s \n", vm.PROGRAM[i].Name)
		case "DUP":
			fmt.Printf("%s : %+v \n", vm.PROGRAM[i].Name, vm.PROGRAM[i].Operand)
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
	if debug { fmt.Println() }
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
	if halt_flag == false {
		if halt_panic {
			print_program_trace(vm,true)
			panic("No `HALT` instruction in PROGRAM")
		}
	}
	if debug { fmt.Println() }
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

func check_if_label_and_push_to_label_table(vm *VM, lt *Label_Table, s string) bool {
	if string(s[len(s)-1]) == ":" {
		label_name := string(s[:len(s)-1])
		lt.labels[lt.table_size] = Label{Name: label_name, addr: vm.program_size}
		lt.table_size += 1
		return true
	}
	return false
}

func find_label_in_label_table(lt Label_Table, label_name string) int64 {
	for i := int64(0); i<lt.table_size; i++ {
		if lt.labels[i].Name == label_name {
			return lt.labels[i].addr
		}
	}
	return -1
}

func push_to_unresolved_jump_table(vm *VM, unrslvdjmps *Unresolved_Jumps, label_name string, line_number int64) {
	tmp_uj := Unresolved_Jump{unresolved_jmp_addr: vm.program_size, unresolved_jump_label: label_name, unresolved_jmp_line: line_number}
	unrslvdjmps.unresolved_jump_arr[unrslvdjmps.unresolved_jumps_size] = tmp_uj
	unrslvdjmps.unresolved_jumps_size += 1
}

func load_program_from_file(vm *VM, file_path string, halt_panic bool) {
	dat, err := ioutil.ReadFile(file_path)
	check_err(err)
	file_content := string(dat)
	lines := strings.Split(strings.ReplaceAll(file_content, "\r\n", "\n"), "\n")
	instruction_count := 0
	halt_flag := false
	if debug { fmt.Println() }
	for i:=0; i<len(lines) ; i++ {
		line := strings.Trim(process_comment(strings.Trim(lines[i], " ")), " ")
		if line != "" {
			line_split_by_space := strings.Split(line, " ")
			label_check := check_if_label_and_push_to_label_table(vm, &lt_g, line_split_by_space[0])
			if label_check {
				continue
			}
			inst_name := strings.ToUpper(line_split_by_space[0])

			switch inst_name {
			
			case "PUSH":
				if len(line_split_by_space) > 2 {
					fmt.Printf("File : %s\n", file_path)
					fmt.Printf("Too Many Args or Extra Spaces: Invalid Syntax near line %d : %s\n", (i+1), line)
					panic("Syntax Error")
				}
				if len(line_split_by_space) == 1 {
					fmt.Printf("File : %s\n", file_path)
					fmt.Printf("Missing Arguments: Invalid Syntax near line %d : %s\n", (i+1), line)
					panic("Syntax Error")
				}
				operand , err := strconv.Atoi(line_split_by_space[1])
				check_err(err)
				vm.PROGRAM[vm.program_size].Name = "PUSH"
				vm.PROGRAM[vm.program_size].Operand.int64holder = int64(operand)
				
			
			case "ADDI":
				if len(line_split_by_space) > 1 {
					fmt.Printf("File : %s\n", file_path)
					fmt.Printf("Syntax Error: Invalid Syntax near line %d : %s\n", (i+1), line)
					panic("Syntax Error")
				}
				vm.PROGRAM[vm.program_size] = Inst{Name: "ADDI"}
				
				
			case "SUBI":
				if len(line_split_by_space) > 1 {
					fmt.Printf("File : %s\n", file_path)
					fmt.Printf("Syntax Error: Invalid Syntax near line %d : %s\n", (i+1), line)
					panic("Syntax Error")
				}
				vm.PROGRAM[vm.program_size] = Inst{Name: "SUBI"}
				
				
			case "MULI":
				if len(line_split_by_space) > 1 {
					fmt.Printf("File : %s\n", file_path)
					fmt.Printf("Syntax Error: Invalid Syntax near line %d : %s\n", (i+1), line)
					panic("Syntax Error")
				}
				vm.PROGRAM[vm.program_size] = Inst{Name: "MULI"}
				
				
			case "DIVI":
				if len(line_split_by_space) > 1 {
					fmt.Printf("File : %s\n", file_path)
					fmt.Printf("Syntax Error: Invalid Syntax near line %d : %s\n", (i+1), line)
					panic("Syntax Error")
				}
				vm.PROGRAM[vm.program_size] = Inst{Name: "DIVI"}
				
				
			case "JMP":
				if len(line_split_by_space) > 2 {
					fmt.Printf("File : %s\n", file_path)
					fmt.Printf("Too Many Args or Extra Spaces: Invalid Syntax near line %d : %s\n", (i+1), line)
					panic("Syntax Error")
				}
				if len(line_split_by_space) == 1 {
					fmt.Printf("File : %s\n", file_path)
					fmt.Printf("Missing Arguments: Invalid Syntax near line %d : %s\n", (i+1), line)
					panic("Syntax Error")
				}
				temp_s := line_split_by_space[1]
				r := []rune(string(temp_s[0]))
				if unicode.IsDigit(r[0]) {
					operand , err := strconv.Atoi(line_split_by_space[1])
					check_err(err)
					vm.PROGRAM[vm.program_size].Name = "JMP"
					vm.PROGRAM[vm.program_size].Operand.int64holder = int64(operand)
				} else {
					vm.PROGRAM[vm.program_size].Name = "JMP"
					push_to_unresolved_jump_table(vm, &unrslvdjmps_g, temp_s, int64((i+1)))
				}
				
			case "HALT":
				if len(line_split_by_space) > 1 {
					fmt.Printf("File : %s\n", file_path)
					fmt.Printf("Syntax Error: Invalid Syntax near line %d : %s\n", (i+1), line)
					panic("Syntax Error")
				}
				halt_flag = true
				vm.PROGRAM[vm.program_size] = Inst{Name: "HALT"}
				
				
			case "NOP":
				if len(line_split_by_space) > 1 {
					fmt.Printf("File : %s\n", file_path)
					fmt.Printf("Syntax Error: Invalid Syntax near line %d : %s\n", (i+1), line)
					panic("Syntax Error")
				}
				vm.PROGRAM[vm.program_size] = Inst{Name: "NOP"}
				
				
			case "RET":
				if len(line_split_by_space) > 1 {
					fmt.Printf("File : %s\n", file_path)
					fmt.Printf("Syntax Error: Invalid Syntax near line %d : %s\n", (i+1), line)
					panic("Syntax Error")
				}
				vm.PROGRAM[vm.program_size] = Inst{Name: "RET"}
				
				
			case "DUP":
				if len(line_split_by_space) > 2 {
					fmt.Printf("File : %s\n", file_path)
					fmt.Printf("Too Many Args or Extra Spaces: Invalid Syntax near line %d : %s\n", (i+1), line)
					panic("Syntax Error")
				}
				if len(line_split_by_space) == 1 {
					fmt.Printf("File : %s\n", file_path)
					fmt.Printf("Missing Arguments: Invalid Syntax near line %d : %s\n", (i+1), line)
					panic("Syntax Error")
				}
				operand , err := strconv.Atoi(line_split_by_space[1])
				check_err(err)
				vm.PROGRAM[vm.program_size] = Inst{Name: "DUP", Operand: Value_Holder{int64holder: int64(operand)}}
				
			default:
				fmt.Printf("File : %s\n", file_path)
				fmt.Printf("Syntax Error: Unknown Instruction near line %d : %s\n",(i+1), line)
				panic("Unknown Instruction")
			}
			vm.program_size += 1
			instruction_count += 1
			if instruction_count >= PROGRAM_CAPACITY {
				fmt.Printf("File : %s\n", file_path)
				fmt.Printf("Number of Instructions is greater than PROGRAM CAPACITY = %d", PROGRAM_CAPACITY)
				panic("Overflow")
			}
			if debug {
				fmt.Printf("Loaded Instruction: %s : %+v \n", vm.PROGRAM[vm.program_size-1].Name, vm.PROGRAM[vm.program_size-1].Operand)
			}
		}
	}

	for i:=int64(0); i < unrslvdjmps_g.unresolved_jumps_size; i++ {
		ind := find_label_in_label_table(lt_g,unrslvdjmps_g.unresolved_jump_arr[i].unresolved_jump_label)
		if ind == -1 {
			error_line_number := unrslvdjmps_g.unresolved_jump_arr[i].unresolved_jmp_line
			fmt.Printf("Unknown Label near line %d : `%s` \n", error_line_number, unrslvdjmps_g.unresolved_jump_arr[i].unresolved_jump_label)
			panic("Unknown Label")
		}
		vm.PROGRAM[unrslvdjmps_g.unresolved_jump_arr[i].unresolved_jmp_addr].Operand.int64holder = ind
	}

	if halt_flag == false {
		if halt_panic {
			print_program_trace(vm,true)
			panic("No `HALT` instruction in PROGRAM")
		}
	}
	if debug { fmt.Println() }
}

func execute_program(vm *VM, limit int) {
	if vm.program_size == 0 {
		panic("No instruction to execute.. Load Program first")
	}
	counter := 0
	for (vm.vm_halt != 1 && counter < limit) {
		if debug {
			print_stack(vm, true)
			fmt.Printf("IP : %d\n", vm.inst_ptr)
			fmt.Printf("STEP(%d) Instruction to be executed : `%s : %+v`\n", (counter+1), vm.PROGRAM[vm.inst_ptr].Name, vm.PROGRAM[vm.inst_ptr].Operand)
			prompt_for_debug()
		}
		execute_inst(vm, vm.PROGRAM[vm.inst_ptr])
		counter += 1
	}
}

func init_all(initial_stack *[STACK_CAPACITY]Value_Holder, initial_program *[PROGRAM_CAPACITY]Inst) {
	for i :=0; i<STACK_CAPACITY; i++ {
		initial_stack[i] = Value_Holder{int64holder: int64(MinInt), float64holder: float64(MinFloat), pointer: ""}
	}
	for i :=0; i<PROGRAM_CAPACITY; i++ {
		initial_program[i] = Inst{Name: "", Operand: Value_Holder{int64holder: int64(MinInt), float64holder: float64(MinFloat), pointer: ""}}
	}
}

func main() {
	var initial_stack [STACK_CAPACITY]Value_Holder
	var initial_program [PROGRAM_CAPACITY]Inst
	init_all(&initial_stack,&initial_program)

	lt_g = Label_Table{}
	unrslvdjmps_g = Unresolved_Jumps{}

	vm_g := VM{STACK: initial_stack, PROGRAM: initial_program}
	
	// var prgm = []Inst {
	// 	Inst{Name: "PUSH", Operand: 10},
	// 	Inst{Name: "PUSH", Operand: 10},
	// 	Inst{Name: "PUSH", Operand: 10},
	// 	Inst{Name: "PUSH", Operand: 20},
	// 	Inst{Name: "ADD"},
	// 	Inst{Name: "MUL"},
	// 	Inst{Name: "NOP"},
	// 	Inst{Name: "PUSH", Operand: 10},
	// 	Inst{Name: "SUB", Operand: 10},
	// 	Inst{Name: "HALT"},
	// }
	// program_size := len(prgm)
	file_path := flag.String("input", "", ".vasm FILE PATH")
	execution_limit_steps_inp := flag.Int("limit", 69, "Execution Limit Steps")
	debug_flg := flag.Bool("debug", false, "Enable Debugger")
	
	flag.Parse()
	
	debug = *debug_flg
	if *file_path == "" {
		fmt.Println("No input .vasm file is provided. Use '-h' option for help")
		os.Exit(0)
		// load_program_from_memory(&vm_g, prgm, program_size, true)
	} else {
		load_program_from_file(&vm_g, *file_path, false)
	}
	print_program_trace(&vm_g, true)
	execute_program(&vm_g, *execution_limit_steps_inp)
	print_stack(&vm_g, false)
}