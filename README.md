# pmigrate
Linux Process Migration Utilies (Written in Go)

pmigrate is a simple suite of two utility command, *pfrez* and *pthaw*. 

1. [pfrez](https://github.com/tarndt/pmigrate/tree/master/pfrez) (process freeze): Capture the state of a process
2. [pthaw](https://github.com/tarndt/pmigrate/tree/master/pthaw) (process thaw): Restore a process to execute from saved state

These utilities are written in Go (with a small C helper program called [pload](https://github.com/tarndt/pmigrate/tree/master/pthaw/pload)), share most of the same code base and rely solely on user-space facilities with no requirement for loading kernel modules or patching.

The state captured by pfrez can be streamed to stdout to be composed with other operations, explicitly serialized to a file or sent to a remote cooperating pthaw process (optional compression and encryption to minimize and protect program state in transit). Reciprocally pthaw is the utility that consumes captured state and restores it to execution while supervising. There is some overhead for restored process due to the need to intercept system-calls that reference specific local resources (such as file handles or PIDs) and remap them to match the new execution environment.

This project was originally built as a component of my Masters degree in Software Engineering, and a paper discussing design concerns as well as outlining design and implementation details, and a road-map for future improvement can be found [here](https://github.com/tarndt/pmigrate/blob/master/ProcessMigrationPaper.pdf).

### Further work

If you are interested in using these tools for a serious task or would like to participate contining the work started here, please contact me and I would be happy to bring this project out of hibernation and collaborate with you.

### Usage

Usage of pfrez: 
```
   -compress string 
    	Compression mode: none | gzip | flate | snappy (default "none") 
  -debug 
    	Debug: true | false, if enabled outgoing data will be displayed 
  -dest string 
    	Output sink: stdout | tcp|udp:host:port | unix:socketpath | snapshot-filepath (default "stdout") 
  -dial-timeout duration 
    	Optional: Duration to wait for socket level connection to be established 
  -encrypt string 
    	Encryption mode: none | AES-CFB|AES-CTR|AES-OFB:keypath (default "none") 
  -halt 
    	Halt the target process after state capture and transmission is complete 
  -pid int 
    	PID of process to be frozen (default -1) 
  -write-timeout duration 
    	Optional: Duration to wait transmitting data to an active stream before timing out-compress string 
    	Compression mode: none | gzip | flate | snappy (default "none") 
  -debug 
    	Debug: true | false, if enabled incoming data will be displayed 
  -dest string 
    	Output sink: stdout | tcp|udp:host:port | unix:socketpath | snapshot-filepath (default "stdout") 
  -encrypt string 
    	Encryption mode: none | AES-CFB|AES-CTR|AES-OFB:keypath (default "none") 
  -pid int 
    	PID of process to be frozen (default -1) 
```

Usage of pthaw:
```
    -debug 
    	Debug: true | false, if enabled incomming data will be displayed 
  -keydir string 
    	Optional: Directory containing decryption keys 
  -loader string 
    	Optional: Alternate path to loader executable 
  -read-timeout duration 
    	Optional: Duration to wait for incomming data on an active stream before timing out 
  -src string 
    	Input source: stdin | tcp|udp:port | unix:socketpath | snapshot-filepath (default "stdin") 
```

A very simple usage example:

Start our target process, [countforever](https://github.com/tarndt/pmigrate/blob/master/testprogs/countforever.c) which increments and prints forever:
```
user@system:~/testdir$ ./countforever 
0 
1 
2 
3 
...
```

Capture process state and send  it to stdout :
```
user@system:~/testdir$ sudo ./pfrez -pid=`pgrep countforever` > demo.snap
```

Now, restore process from above snapshot (via stdin) :
```
user@system:~/testdir$ ./pthaw < demo.snap 
.Attaching... .Attached. 
Loading Registers... Loaded. 
Resuming process... 
32218008 
32218009 
32218010 
32218011 
32218012 
32218013 
... 
```
