import os
import sys
import subprocess

test_file = ""
all = False

all_examples = [x for x in os.listdir("./examples") if x.endswith(".vasm")]

ignore_direct_execution = ["consts.vasm"]
exit_code               = 0
error_count             = 0
FIXED_DASH_BLOCK_SIZE   = 16

def parse_arguments(arr, argument, bool=False, verbose=False):
    for i, val in enumerate(arr):
        if val.replace("-", "") == argument:
            if bool:
                if verbose:
                    print(f"{argument} : True")
                return True
            if i+1 == len(arr):
                if verbose:
                    print(f":ERROR: No Value Passed for Argument: `{argument}`")
                raise Exception("NoValueForArgument")
            if verbose:
                print(f"{argument} : {arr[i+1]}")
            return arr[i+1]
    if verbose:
        print(f"ERROR: Argument '{argument}' not found")
    raise Exception("ArgumentNotFound")

def usage(exit_code):
    print(f"Usage: python {sys.argv[0]} [OPTIONS]")
    print("\nOPTIONS:")
    print("   -i  (str)  : Run a specific test file in `examples` folder")
    print("   -h  (bool) : Print this help and exit")
    print()
    print("If no arguments are passed, all tests are run")
    if exit_code != None:
        sys.exit(exit_code)


def run_test(file_):
    global error_count
    dashes_count = len(file_)+FIXED_DASH_BLOCK_SIZE
    print("-"*dashes_count)
    print(f"Executing: {file_}")
    process        = subprocess.Popen(['./vm-go', '-i', f"examples/{file_}"], stdout=subprocess.PIPE, stderr=subprocess.PIPE)
    stdout, stderr = process.communicate()
    file_content   = open(f"./tests/expected/{file_.removesuffix('.vasm')}.expected").read()
    try:
        print(f"Testing  : {file_}",end="")
        assert stdout.decode("utf-8") == file_content, "TestFail"
        print("...Ok")
        print("-"*dashes_count)
    except Exception as e:
        print("\n")
        print("stdout:")
        print()
        print(stdout.decode("utf-8"))
        error_count = error_count + 1

if __name__ == "__main__":
    try:
        if parse_arguments(sys.argv, 'h', True):
            usage(0)
    except Exception as e:
        pass
    try:
        test_file = parse_arguments(sys.argv,'i')
    except Exception as e:
        pass

    if test_file:
        run_test(test_file)
        sys.exit(exit_code)

    for example_file in all_examples:
        if example_file in ignore_direct_execution:
            continue
        run_test(example_file)

    if error_count > 0:
        exit_code = 1

    print("-----\nSTATS\n-----")
    print(f"Total  : {len(all_examples) - len(ignore_direct_execution)}")
    print(f"Passed : {len(all_examples) - len(ignore_direct_execution) - error_count}")
    print(f"Failed : {error_count}")

    sys.exit(exit_code)
