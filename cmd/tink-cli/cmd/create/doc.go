//
// Package create provides a reusable implementation of the Create command
// for the tink cli. The Create command creates resources. It is designed
// to be extendible and usable across resources.
// In order to be reusable across resources it can't reference by itself any of
// them. It has to take the concrete procedures for each specific resource from
// the outside.
// The common features that the command has:
// * It can take input from stdin, file or from flags.
// * Only from stdin or from file, not both at the same time
// * Flags can be specified eather way even with stdin or file and they
//   override the content of the file
// * The output format is the same no matter the resource
//
package create
