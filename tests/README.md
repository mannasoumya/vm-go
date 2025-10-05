## Running Tests

```shell
$ cd ${vm-go-home} # go to root of repository
$ python3 test.py -h
Usage: python test.py [OPTIONS]

OPTIONS:
   -i  (str)  : Run a specific test file in `examples` folder
   -h  (bool) : Print this help and exit

If no arguments are passed, all tests are run
# Run a specific test file
$ python3 test.py -i hello_world.vasm
---------------------------
Executing: hello_world.vasm
Testing  : hello_world.vasm...Ok
---------------------------
```
