import os
import sys
import subprocess

all_examples = [x for x in os.listdir("./examples") if x.endswith(".vasm")]

# all_examples = ["demo1.vasm"]
exit_code = 0
error_count = 0
for example_file in all_examples:
    print("---------------------------")
    print(f"Executing: {example_file}")
    process = subprocess.Popen(['./vm-go', '-i', f"examples/{example_file}"], stdout=subprocess.PIPE, stderr=subprocess.PIPE)
    stdout, stderr = process.communicate()
    file_content = open(f"./tests/expected/{example_file.removesuffix('.vasm')}.expected").read()
    try:
        print(f"Testing  : {example_file}",end="")
        assert stdout.decode("utf-8") == file_content, "TestFail"
        print("...Ok")
    except Exception as e:
        print("\n")
        print("stdout:")
        print()
        print(stdout.decode("utf-8"))
        error_count = error_count + 1

print("---------------------------")
if error_count > 0:
    exit_code = 1

print("-----\nSTATS\n-----")
print(f"Total  : {len(all_examples)}")
print(f"Passed : {len(all_examples) - error_count}")
print(f"Failed : {error_count}")

sys.exit(exit_code)
