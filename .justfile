set quiet := true
binary_name := "benoit"

[private]
default:
    just -l

# Build the binary.
build:
    go build -o {{ binary_name }}

# Clean up logs and the built binary.
clean:
    rm *.txt {{ binary_name }}

# Run the code without building via 'go run'
run inputfile:
    go run main.go run -i {{ inputfile }}
