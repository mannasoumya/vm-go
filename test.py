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

def print_(message, end="\n", quiet=False):
    if not quiet:
        print(message, end=end)

def usage(exit_code):
    print(f"Usage: python {sys.argv[0]} [OPTIONS]")
    print("\nOPTIONS:")
    print("   -i  (str)  : Run a specific test file in `examples` folder")
    print("   -q  (bool) : Enable quiet mode")
    print("   -h  (bool) : Print this help and exit")
    print()
    print("If no arguments are passed, all tests are run")
    if exit_code != None:
        sys.exit(exit_code)

def print_test_error(stdout, stderr, expected, quiet):
    print_("Expected:", quiet=quiet)
    print_(expected, quiet=quiet)
    print_("Actual:", quiet=quiet)
    print_(stdout.decode("utf-8"), quiet=quiet)
    if stderr:
        print_("Error:", quiet=quiet)
        print_(stderr.decode("utf-8"), quiet=quiet)

def run_test(file_, quiet):
    if not file_.endswith(".vasm"):
        file_ = file_ + ".vasm"
    global error_count
    dashes_count = len(file_) + FIXED_DASH_BLOCK_SIZE

    print_("-" * dashes_count, quiet=quiet)
    print_(f"Executing: {file_}", quiet=quiet)

    process        = subprocess.Popen(['./vm-go', '-i', f"examples/{file_}"], stdout=subprocess.PIPE, stderr=subprocess.PIPE)
    stdout, stderr = process.communicate()
    file_content   = open(f"./tests/expected/{file_.removesuffix('.vasm')}.expected").read()

    try:
        print_(f"Testing  : {file_}", end="", quiet=quiet)
        assert stdout.decode("utf-8") == file_content, "TestFail"
        print_("...Ok", quiet=quiet)
        print_("-" * dashes_count, quiet=quiet)
    except Exception as e:
        error_count = error_count + 1
        print_("...Failed", quiet=quiet)
        print_("\n", quiet=quiet)
        print_test_error(stdout, stderr, file_content, quiet)
        print_("-" * dashes_count, quiet=quiet)

if __name__ == "__main__":
    quiet = False
    try:
        if parse_arguments(sys.argv, 'h', True):
            usage(0)
    except Exception as e:
        pass

    try:
        if parse_arguments(sys.argv, 'q', True):
            quiet = True
    except Exception as e:
        pass

    try:
        test_file = parse_arguments(sys.argv, 'i')
    except Exception as e:
        pass

    if test_file:
        run_test(test_file, quiet)
        sys.exit(exit_code)

    for example_file in all_examples:
        if example_file in ignore_direct_execution:
            continue
        run_test(example_file, quiet)

    if error_count > 0:
        exit_code = 1

    print("-----\nSTATS\n-----")
    print(f"Total  : {len(all_examples) - len(ignore_direct_execution)}")
    print(f"Passed : {len(all_examples) - len(ignore_direct_execution) - error_count}")
    print(f"Failed : {error_count}")

    sys.exit(exit_code)
