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

## MacOs usage

For MacOS, the binary must be started manually and the configuration file edit accordingly. 
Both the binary and the configuration file can be found in the release page (.

```shell
export NE_RELEASE=v0.7.3
# Download configuration file
$ curl -fsSL https://github.com/nuvlaedge/nuvlaedge-go/releases/download/$NE_RELEASE/template.toml -o template.toml
# Download binary
$ curl -fsSL https://github.com/nuvlaedge/nuvlaedge-go/releases/download/$NE_RELEASE/nuvlaedge-darwin-amd64-v0.7.3 -o nuvlaedge

# Configure the nuvlaedge uuid inside the template.toml file
$ export NUVLAEDGE_SETTINGS=/path/to/template.toml
$ ./nuvlaedge
# Or if you want to run it in detached mode
$ nohup ./nuvlaedge > /dev/null 2>&1 &

```




## Contributing

Contributions are welcome. Feel free to submit a pull request or open an issue.

## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.


