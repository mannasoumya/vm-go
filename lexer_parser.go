package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"unicode"
)

func lex_and_parse_program(lines []string, file_path string, vm *VM, ignore_halt bool, halt_flag *bool, instruction_count *int) {
	for i := 0; i < len(lines); i++ {
		line := strings.Trim(process_comment(strings.Trim(lines[i], " ")), " ")
		line = strings.Trim(line, "\t")
		if line != "" {
			label_check, new_line := check_if_label_and_push_to_label_table(vm, &lt_g, line, (i + 1), file_path)
			line_split_by_space := strings.Split(new_line, " ")
			if label_check && strings.Trim(new_line, " ") == "" {
				continue
			}

			inst_name := strings.ToUpper(line_split_by_space[0])

			switch inst_name {

			case "PUSH":
				if len(line_split_by_space) > 2 {
					fmt.Printf("File : %s\n", file_path)
					fmt.Printf("Too Many Args or Extra Spaces: Invalid Syntax near line %d : %s\n", (i + 1), line)
					exit_with_one("Syntax Error")
				}
				if len(line_split_by_space) == 1 {
					fmt.Printf("File : %s\n", file_path)
					fmt.Printf("Missing Arguments: Invalid Syntax near line %d : %s\n", (i + 1), line)
					exit_with_one("Syntax Error")
				}
				unknown_op := strings.Trim(line_split_by_space[1], " ")

				if x, found_int := Constant_Mapping_int[unknown_op]; found_int {
					vm.PROGRAM[vm.program_size].Operand.int64holder = int64(x)
				} else if x, found_float := Constant_Mapping_float[unknown_op]; found_float {
					vm.PROGRAM[vm.program_size].Operand.float64holder = x
				} else if x, found_str := Constant_Mapping_string[unknown_op]; found_str {
					vm.PROGRAM[vm.program_size].Operand.pointer = x
				} else if strings.Contains(unknown_op, ".") {
					operand, err := strconv.ParseFloat(unknown_op, 64)
					report_error(err, (i + 1), line, file_path, true)
					vm.PROGRAM[vm.program_size].Operand.float64holder = operand
				} else if strings.Contains(unknown_op, "e") {
					operand, err := strconv.ParseFloat(unknown_op, 64)
					report_error(err, (i + 1), line, file_path, true)
					vm.PROGRAM[vm.program_size].Operand.float64holder = operand
				} else {
					operand, err := strconv.Atoi(unknown_op)
					if err != nil {
						fmt.Printf("File : %s\n", file_path)
						fmt.Printf("ERROR: Error near line %d : %s\n", (i + 1), line)
						fmt.Printf("Failed Parsing Operand. `%s` is not defined\n", unknown_op)
						os.Exit(1)
					}
					vm.PROGRAM[vm.program_size].Operand.int64holder = int64(operand)
				}
				vm.PROGRAM[vm.program_size].Name = "PUSH"

			case "ADDI":
				if len(line_split_by_space) > 1 {
					fmt.Printf("File : %s\n", file_path)
					fmt.Printf("Syntax Error: Invalid Syntax near line %d : %s\n", (i + 1), line)
					exit_with_one("Syntax Error")
				}
				vm.PROGRAM[vm.program_size] = Inst{Name: "ADDI"}

			case "SUBI":
				if len(line_split_by_space) > 1 {
					fmt.Printf("File : %s\n", file_path)
					fmt.Printf("Syntax Error: Invalid Syntax near line %d : %s\n", (i + 1), line)
					exit_with_one("Syntax Error")
				}
				vm.PROGRAM[vm.program_size] = Inst{Name: "SUBI"}

			case "MULI":
				if len(line_split_by_space) > 1 {
					fmt.Printf("File : %s\n", file_path)
					fmt.Printf("Syntax Error: Invalid Syntax near line %d : %s\n", (i + 1), line)
					exit_with_one("Syntax Error")
				}
				vm.PROGRAM[vm.program_size] = Inst{Name: "MULI"}

			case "DIVI":
				if len(line_split_by_space) > 1 {
					fmt.Printf("File : %s\n", file_path)
					fmt.Printf("Syntax Error: Invalid Syntax near line %d : %s\n", (i + 1), line)
					exit_with_one("Syntax Error")
				}
				vm.PROGRAM[vm.program_size] = Inst{Name: "DIVI"}

			case "EQI":
				if len(line_split_by_space) > 1 {
					fmt.Printf("File : %s\n", file_path)
					fmt.Printf("Syntax Error: Invalid Syntax near line %d : %s\n", (i + 1), line)
					exit_with_one("Syntax Error")
				}
				vm.PROGRAM[vm.program_size] = Inst{Name: "EQI"}

			case "ADDF":
				if len(line_split_by_space) > 1 {
					fmt.Printf("File : %s\n", file_path)
					fmt.Printf("Syntax Error: Invalid Syntax near line %d : %s\n", (i + 1), line)
					exit_with_one("Syntax Error")
				}
				vm.PROGRAM[vm.program_size] = Inst{Name: "ADDF"}

			case "SUBF":
				if len(line_split_by_space) > 1 {
					fmt.Printf("File : %s\n", file_path)
					fmt.Printf("Syntax Error: Invalid Syntax near line %d : %s\n", (i + 1), line)
					exit_with_one("Syntax Error")
				}
				vm.PROGRAM[vm.program_size] = Inst{Name: "SUBF"}

			case "MULF":
				if len(line_split_by_space) > 1 {
					fmt.Printf("File : %s\n", file_path)
					fmt.Printf("Syntax Error: Invalid Syntax near line %d : %s\n", (i + 1), line)
					exit_with_one("Syntax Error")
				}
				vm.PROGRAM[vm.program_size] = Inst{Name: "MULF"}

			case "DIVF":
				if len(line_split_by_space) > 1 {
					fmt.Printf("File : %s\n", file_path)
					fmt.Printf("Syntax Error: Invalid Syntax near line %d : %s\n", (i + 1), line)
					exit_with_one("Syntax Error")
				}
				vm.PROGRAM[vm.program_size] = Inst{Name: "DIVF"}

			case "EQF":
				if len(line_split_by_space) > 1 {
					fmt.Printf("File : %s\n", file_path)
					fmt.Printf("Syntax Error: Invalid Syntax near line %d : %s\n", (i + 1), line)
					exit_with_one("Syntax Error")
				}
				vm.PROGRAM[vm.program_size] = Inst{Name: "EQF"}

			case "JMP":
				if len(line_split_by_space) > 2 {
					fmt.Printf("File : %s\n", file_path)
					fmt.Printf("Too Many Args or Extra Spaces: Invalid Syntax near line %d : %s\n", (i + 1), line)
					exit_with_one("Syntax Error")
				}
				if len(line_split_by_space) == 1 {
					fmt.Printf("File : %s\n", file_path)
					fmt.Printf("Missing Arguments: Invalid Syntax near line %d : %s\n", (i + 1), line)
					exit_with_one("Syntax Error")
				}
				temp_s := line_split_by_space[1]
				r := []rune(string(temp_s[0]))
				if unicode.IsDigit(r[0]) {
					operand, err := strconv.Atoi(line_split_by_space[1])
					check_err(err)
					vm.PROGRAM[vm.program_size].Name = "JMP"
					vm.PROGRAM[vm.program_size].Operand.int64holder = int64(operand)
				} else {
					vm.PROGRAM[vm.program_size].Name = "JMP"
					push_to_deferred_operand_table(vm, &deferredoprnds_g, temp_s, int64((i + 1)))
				}

			case "DEFINE":
				runes := []rune(strings.Trim(line, " "))
				first_space := -1
				for iter := 0; iter < len(runes); iter++ {
					if runes[iter] == ' ' {
						first_space = iter
						break
					}
				}
				if first_space == -1 {
					fmt.Printf("File : %s\n", file_path)
					fmt.Printf("Missing Arguments: Invalid Syntax near line %d : %s\n", (i + 1), line)
					exit_with_one("Syntax Error")
				}
				strs_to_process := ""

				for iter := first_space + 1; iter < len(runes); iter++ {
					strs_to_process = strs_to_process + string(runes[iter])
				}

				strs_to_process = strings.Trim(strs_to_process, " ")
				curr_err := parse_and_load_define_operands(strs_to_process, file_path)
				if curr_err != nil {
					if curr_err.Error() == "MissingArguments" {
						fmt.Printf("File : %s\n", file_path)
						fmt.Printf("Missing Arguments: Invalid Syntax near line %d : %s\n", (i + 1), line)
						exit_with_one("Syntax Error")
					} else {
						report_error(curr_err, (i + 1), line, file_path, true)
					}
				}
				vm.PROGRAM[vm.program_size] = Inst{Name: "DEFINE"}

			case "JMP_IF":
				if len(line_split_by_space) > 2 {
					fmt.Printf("File : %s\n", file_path)
					fmt.Printf("Too Many Args or Extra Spaces: Invalid Syntax near line %d : %s\n", (i + 1), line)
					exit_with_one("Syntax Error")
				}
				if len(line_split_by_space) == 1 {
					fmt.Printf("File : %s\n", file_path)
					fmt.Printf("Missing Arguments: Invalid Syntax near line %d : %s\n", (i + 1), line)
					exit_with_one("Syntax Error")
				}
				temp_s := line_split_by_space[1]
				r := []rune(string(temp_s[0]))
				if unicode.IsDigit(r[0]) {
					operand, err := strconv.Atoi(line_split_by_space[1])
					report_error(err, (i + 1), line, file_path, true)
					vm.PROGRAM[vm.program_size].Name = "JMP_IF"
					vm.PROGRAM[vm.program_size].Operand.int64holder = int64(operand)
				} else {
					vm.PROGRAM[vm.program_size].Name = "JMP_IF"
					push_to_deferred_operand_table(vm, &deferredoprnds_g, temp_s, int64((i + 1)))
				}

			case "CALL":
				if len(line_split_by_space) > 2 {
					fmt.Printf("File : %s\n", file_path)
					fmt.Printf("Too Many Args or Extra Spaces: Invalid Syntax near line %d : %s\n", (i + 1), line)
					exit_with_one("Syntax Error")
				}
				if len(line_split_by_space) == 1 {
					fmt.Printf("File : %s\n", file_path)
					fmt.Printf("Missing Arguments: Invalid Syntax near line %d : %s\n", (i + 1), line)
					exit_with_one("Syntax Error")
				}
				temp_s := line_split_by_space[1]
				r := []rune(string(temp_s[0]))
				if unicode.IsDigit(r[0]) {
					operand, err := strconv.Atoi(line_split_by_space[1])
					check_err(err)
					vm.PROGRAM[vm.program_size].Name = "CALL"
					vm.PROGRAM[vm.program_size].Operand.int64holder = int64(operand)
				} else {
					vm.PROGRAM[vm.program_size].Name = "CALL"
					push_to_deferred_operand_table(vm, &deferredoprnds_g, temp_s, int64((i + 1)))
				}

			case "SWAP":
				if len(line_split_by_space) > 2 {
					fmt.Printf("File : %s\n", file_path)
					fmt.Printf("Too Many Args or Extra Spaces: Invalid Syntax near line %d : %s\n", (i + 1), line)
					exit_with_one("Syntax Error")
				}
				if len(line_split_by_space) == 1 {
					fmt.Printf("File : %s\n", file_path)
					fmt.Printf("Missing Arguments: Invalid Syntax near line %d : %s\n", (i + 1), line)
					exit_with_one("Syntax Error")
				}
				operand, err := strconv.Atoi(line_split_by_space[1])
				report_error(err, (i + 1), line, file_path, true)
				vm.PROGRAM[vm.program_size] = Inst{Name: "SWAP", Operand: Value_Holder{int64holder: int64(operand)}}

			case "HALT":
				if len(line_split_by_space) > 1 {
					fmt.Printf("File : %s\n", file_path)
					fmt.Printf("Syntax Error: Invalid Syntax near line %d : %s\n", (i + 1), line)
					exit_with_one("Syntax Error")
				}
				if ignore_halt {
					vm.PROGRAM[vm.program_size] = Inst{Name: "NOP"}
				} else {
					halt_flag = new(bool)
					*halt_flag = true
					vm.PROGRAM[vm.program_size] = Inst{Name: "HALT"}
				}

			case "IGNORE_HALT":
				if len(line_split_by_space) > 1 {
					fmt.Printf("File : %s\n", file_path)
					fmt.Printf("Syntax Error: Invalid Syntax near line %d : %s\n", (i + 1), line)
					exit_with_one("Syntax Error")
				}
				ignore_halt = true
				vm.PROGRAM[vm.program_size] = Inst{Name: "NOP"}

			case "NOP":
				if len(line_split_by_space) > 1 {
					fmt.Printf("File : %s\n", file_path)
					fmt.Printf("Syntax Error: Invalid Syntax near line %d : %s\n", (i + 1), line)
					exit_with_one("Syntax Error")
				}
				vm.PROGRAM[vm.program_size] = Inst{Name: "NOP"}

			case "NOT":
				if len(line_split_by_space) > 1 {
					fmt.Printf("File : %s\n", file_path)
					fmt.Printf("Syntax Error: Invalid Syntax near line %d : %s\n", (i + 1), line)
					exit_with_one("Syntax Error")
				}
				vm.PROGRAM[vm.program_size] = Inst{Name: "NOT"}

			case "DROP":
				if len(line_split_by_space) > 1 {
					fmt.Printf("File : %s\n", file_path)
					fmt.Printf("Syntax Error: Invalid Syntax near line %d : %s\n", (i + 1), line)
					exit_with_one("Syntax Error")
				}
				vm.PROGRAM[vm.program_size] = Inst{Name: "DROP"}

			case "RET":
				if len(line_split_by_space) > 1 {
					fmt.Printf("File : %s\n", file_path)
					fmt.Printf("Syntax Error: Invalid Syntax near line %d : %s\n", (i + 1), line)
					exit_with_one("Syntax Error")
				}
				vm.PROGRAM[vm.program_size] = Inst{Name: "RET"}

			case "PRINT":
				if len(line_split_by_space) > 1 {
					fmt.Printf("File : %s\n", file_path)
					fmt.Printf("Syntax Error: Invalid Syntax near line %d : %s\n", (i + 1), line)
					exit_with_one("Syntax Error")
				}
				vm.PROGRAM[vm.program_size] = Inst{Name: "PRINT"}

			case "DUP":
				if len(line_split_by_space) > 2 {
					fmt.Printf("File : %s\n", file_path)
					fmt.Printf("Too Many Args or Extra Spaces: Invalid Syntax near line %d : %s\n", (i + 1), line)
					exit_with_one("Syntax Error")
				}
				if len(line_split_by_space) == 1 {
					fmt.Printf("File : %s\n", file_path)
					fmt.Printf("Missing Arguments: Invalid Syntax near line %d : %s\n", (i + 1), line)
					exit_with_one("Syntax Error")
				}
				operand, err := strconv.Atoi(line_split_by_space[1])
				report_error(err, (i + 1), line, file_path, true)
				vm.PROGRAM[vm.program_size] = Inst{Name: "DUP", Operand: Value_Holder{int64holder: int64(operand)}}

			case "INCLUDE":
				if len(line_split_by_space) > 2 {
					fmt.Printf("File : %s\n", file_path)
					fmt.Printf("Too Many Args or Extra Spaces: Invalid Syntax near line %d : %s\n", (i + 1), line)
					exit_with_one("Syntax Error")
				}
				if len(line_split_by_space) == 1 {
					fmt.Printf("File : %s\n", file_path)
					fmt.Printf("Missing Arguments: Invalid Syntax near line %d : %s\n", (i + 1), line)
					exit_with_one("Syntax Error")
				}
				include_file_path, err := parse_string_literal(line_split_by_space[1])
				report_error(err, (i + 1), line, file_path, true)
				include_file_path_array[include_file_path_array_size] = include_file_path
				include_file_path_array_size += 1
				load_program_from_file(vm, include_file_path, false)
				vm.PROGRAM[vm.program_size] = Inst{Name: "INCLUDE"}

			default:
				fmt.Printf("File : %s\n", file_path)
				fmt.Printf("Syntax Error: Unknown Instruction near line %d : %s\n", (i + 1), line)
				exit_with_one("Unknown Instruction")
			}
			vm.program_size += 1
			*instruction_count += 1
			if *instruction_count >= PROGRAM_CAPACITY {
				fmt.Printf("File : %s\n", file_path)
				fmt.Printf("Number of Instructions is greater than PROGRAM CAPACITY = %d", PROGRAM_CAPACITY)
				exit_with_one("Overflow")
			}
			if debug {
				fmt.Printf("Loaded Instruction: %s : %+v \n", vm.PROGRAM[vm.program_size-1].Name, vm.PROGRAM[vm.program_size-1].Operand)
			}
		}
	}
}
