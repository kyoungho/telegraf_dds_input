# Connext DDS Input Plugin

Connext DDS input plugin consumes data over DDS by creating DDS readers defined in XML App Creation configurations. This plugin converts DDS data to JSON data, and then adds them to a Telegraf output plugin. 

# Install Instructions

Download the repo

    $ git clone git@github.com:kyoungho/dds_input.git

build the "dds_input" binary

    $ go build -o dds_input cmd/main.go
    
 (if you're using windows, you'll want to give it an .exe extension)
 
    go build -o dds_input.exe cmd/main.go

You should be able to call this from telegraf now using execd:

```
[[inputs.execd]]
  command = ["/path/to/dds_input_binary"]
  signal = "none"
  
# sample output: write metrics to stdout
[[outputs.file]]
  files = ["stdout"]
```
