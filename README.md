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

#eg 

secrets-history -path=../my-repository -to=3d724e662e1f2289d9309c4bc92b2566529337c9
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
