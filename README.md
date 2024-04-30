# NuvlaEdge

This is a comprehensive solution for NuvlaEdge devices. The project is primarily written in Go.

Please note that this project is still under development, and some functionalities may not be fully implemented.


## Supported platforms

| 	        | amd64   	 | arm64   	 | arm        	 |
|----------|-----------|-----------|--------------|
| Linux  	 | Service 	 | Service 	 | Service    	 |
| Darwin 	 | Manual  	 | Manual  	 | No Support 	 |

The nuvlaedge-cli installer will install nuvlaedge as a service in linux platforms. For other platforms, the binary must be started manually.

## Usage

At this moment, the recommended way to use nuvlaedge-go is installing it as a service on the device. To use in linux platforms:

```shell
# Get the installer script
$ curl -fsSL https://raw.githubusercontent.com/nuvlaedge/nuvlaedge-go/main/installer/get-ne.sh -o get-ne.sh

# The installer script will downlaod the installer cli
$ sh ./get-ne.sh

# Install nuvlaedge as super user
$ sudo ./nuvlaedge-cli install --service --uuid=<device-uuid>
```

The latest command will place the files in the next locations:
- binary: /usr/local/bin/nuvlaedge
- configuration: /etc/nuvlaedge/template.toml
- nuvlaedge.service: /etc/systemd/system/nuvlaedge.service

Then, it will enable and start the nuvlaedge service.

At the same time, the cli will set the proper configuration in the template.toml file. This file is a template, 
so it can be modified to fit the device needs.




## Contributing

Contributions are welcome. Feel free to submit a pull request or open an issue.

## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.


