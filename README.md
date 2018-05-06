Secrets History
-

To find out if there's any commit with credentials on a given repository

## Install
```sh
go get github.com/paulojean/secrets-history
```

## Usage
```sh
Usage of secrets-history:
  -credential-patterns string
    	json file to use custom patterns on search
  -default-patterns
    	use default pattern credentials (default true)
  -from string
    	start commit to search, ie: newest commit too look
  -path string
    	path to a local git project
  -to string
    	final commit to search, ie: oldest commit to look

#eg 

secrets-history -path=../my-repository -to=3d724e662e1f2289d9309c4bc92b2566529337c9 -credential-patterns='such_patterns.json'
```

## Build

### Dependencies
[glide](https://github.com/Masterminds/glide)

```sh
glide install
go build
```


## Test
```sh
go test
```