# fsized

A small utility written in Golang, showing file size distribution details
in a given directory. For this purpose, it groups all files found within
given path (recursively) by size ranges bounded by powers of 2, i.e.
each group consists of files 2^i (including) ... 2^(i+1) (not including)
bytes long. The following stats are calculated per each file group:

 - number of files in group (whose size lies within given range)
 - total size of files in group
 - total and average number of file system allocation units occupied
   by files in group
 - total and average overhead (in bytes) on files in group

Allocation units and overhead are calculated based on allocation unit
size, which can be specified with optional `--block` commandline switch,
e.g.: `--block=16276` or `--block=8k`. If `--block` is not provided,
default value of 4096 bytes is assumed for calculations.
