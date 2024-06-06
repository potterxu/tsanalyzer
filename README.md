# This is a tool for analyzing ts

The main purpose is to use a pipeline of cells to process ts

## Pipeline
### Pipe command
`tsanalyzer pipe` command will set up a pipeline with multiple connected cells to execute

### Pipeline syntax
#### Cell description
The following command declares a cell with type `cellname`
`cellname prop1=val1 prop2=val2 ...`
#### Cell connection
At current phase of development, only 1-in-1-out is supported

`!` symbol is used as a connection indication:
```
// setup a pipeline cell1 -> cell2 -> ...
cell1 prop1=val1 ! cell2 prop2=val2 prop3=val3 ! ...
```

## Cell
cell is the basic processing unit in the architecture
### Available cells
| name                                | description                      |
| ----------------------------------- | -------------------------------- |
| [file_reader](#file_reader)         | read from file                   |
| [file_writer](#file_writer)         | write to file                    |
| [bytes_converter](#bytes_converter) | convert bytes to specific format |
| [vbv](#vbv)                         | processing vbv data              |
| [mcast_reader](#mcast_reader)       | read from udp multicast          |

### file_reader
### file_writer
### bytes_converter
### vbv
### mcast_reader

## Alias
you can always use the pipe command to set up a customized pipeline, while the tool will also provide some alias commands for some use case

the alias command is simply use some flags to build up a preset pipeline for processing

### Available alias
| name        | description       |
| ----------- | ----------------- |
| [vbv](#vbv) | calculate DTS-PCR |
| [cap](#cap) | capture multicast |

### vbv
`vbv filename -p pcrPid -stream pid1,pid2,pid3 -plot filename`

is an alias for 

`pipe file_reader name=filename ! bytes_converter output_format=ts_packet ! vbv pcr=pcrPid pids=pid1,pid2,pid3 dir=filename.log plot=true`

the above pipeline will
* calculate (DTS-PCR) value for pid1 pid2 and pid3 
* store the result in `filename.log` directory
* plot the result and store in the same directory

### cap
`vbv -t ts -o out.ts eth0 239.1.1.1:1111`

is an alias for 

`pipe mcast_reader intf=eth0 addr=239.1.1.1.1111 ! file_writer name=out.ts`

the above pipeline will
* capture the multicast 239.1.1.1:1111@eth0
* extract udp payload and store to out.ts
