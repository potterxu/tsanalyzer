This is a tool for analyzing ts

it is at starting phase of development with limited features

**pipe command**

this is targeted to be a high extensible command, by implementing different cells, user can use command line tool to achieve different works without adding new commands

**vbv command**

`vbv filename -p pcr -stream p1,p2,p3 filename`

is an alias for 

`pipe file_reader name=filename ! bytes_converter output_format=ts_packet ! vbv pcr=pcr pids=p1,-p2,p3 dir=filename.log`


to calculate (DTS-PCR) value of a specific pids
