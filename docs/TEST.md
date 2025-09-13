# Testing

This project uses the standard Go testing tools. The `go test` command is used to run the tests.

## Running Tests

To run all the tests in the project, navigate to the root directory of the project and run the following command:

```sh
go test ./...
```

This command will discover and run all test files (files ending in `_test.go`) in the current directory and all its subdirectories.

### Example Output

You should see output similar to the following, indicating that the tests have passed:

```
?   	logseq_gen	[no test files]
?   	logseq_gen/internal/cmd	[no test files]
ok  	logseq_gen/internal/config	0.123s
ok  	logseq_gen/internal/generator	0.456s
```

The `?` indicates packages without test files.
The `ok` indicates that the tests for that package passed successfully.
