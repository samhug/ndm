## cfgbak

### About

Automates the process of backing up network device configurations.
This tool reads device information from a configuration file, connects
to each device via SSH, and initiates a TFTP upload back to a build-in
TFTP server.

Device classes are configured with an expect macro written in Javascript.
Support for additional devices can easily be added by writing a simple
expect macro.

The sample config has example expect macros for Cisco ISRs, HP
switches, and the Ubiquiti EdgeRouter.

### Instructions

- Build the tool
```
go get .
go build
```

- Create a device backup spec:
```
# Copy the sample
cp config.example.conf config.conf
# Edit as needed
vim config.conf
```

- Run the backup
```
./cfgbak --config config.conf
```
