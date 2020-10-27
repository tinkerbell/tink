## Hello Contributors!

Thx for your interest!
We're so glad you're here.

### Important Resources

#### bugs: [https://github.com/tinkerbell/tink/issues](https://github.com/tinkerbell/tink/issues)

### Code of Conduct

Available via [https://github.com/tinkerbell/tink/blob/master/.github/CODE_OF_CONDUCT.md](https://github.com/tinkerbell/tink/blob/master/.github/CODE_OF_CONDUCT.md)

### Environment Details

[https://github.com/tinkerbell/tink/blob/master/Makefile](https://github.com/tinkerbell/tink/blob/master/Makefile)

### How to Submit Change Requests

Please submit change requests and / or features via [Issues](https://github.com/tinkerbell/tink/issues).
There's no guarantee it'll be changed, but you never know until you try.
We'll try to add comments as soon as possible, though.

### How to Report a Bug

Bugs are problems in code, in the functionality of an application or in its UI design; you can submit them through [Issues](https://github.com/tinkerbell/tink/issues).

## Code Style Guides

#### Protobuf

Please ensure protobuf related files are generated along with _any_ change to a protobuf file.
CI will enforce this, but its best to commit the generated files along with the protobuf changes in the same commit.
Handling of protobuf deps and generating the go files are both handled by the [protoc.sh](./protos/protoc.sh) script.
Both go & protoc are required by protoc.sh, these are both installed and used if using nix-shell.
