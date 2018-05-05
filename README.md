Secrets History
-

To find out if there's any commit with credentials on a given repository

## Usage
```sh
secrets-history -path=<path> [-from=<from> -to=<to>]
Options:
  -path         Path of of the local repository
  -from         Start commit to check [default: repository's head]
  -to           Start commit to check [default: repository's initial commit]
``` 

## Dependencies
[glide](https://github.com/Masterminds/glide)

## Test
```sh
go test
```

## Build
```sh
glide install
go build
```
