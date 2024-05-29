# Pipe Command
## Syntax
Now it is a initial version of the pipe command

* Now only 1-in-1-out Cell is supported
### Cell
> Use `cellname [key=value...]` to create a cell in the pipe

### Connection
> Use `!` to connect multiple cells in sequence

## Example
```
// a file copy operation example
tsanalyzer pipe filereader name=inputfile ! filewriter name=outputfile
```

