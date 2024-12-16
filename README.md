# kubevirt-ip-helper client

This client tool helps managing the kubevirt-ip-helper application.

## Building the tool

Execute the go build command to build the tool:
```SH
go build -o kihctl .
```

## Usage

Make sure the kubeconfig of the KubeVirt cluster is used or point the KUBECONFIG environment variable to it, for example:
```SH
export KUBECONFIG=<PATH_TO_KUBECONFIG_FILE>
```

Execute the vmnetcfg tool like:
```SH
./kihctl <command> [object namespace] [object name]
```

Commands:

* ippool-list: lists all IPPool objects
* ippool-show \<name>: show the IPPool configuration and status
* vmnetcfg-list: lists all VirtualMachineNetworkConfig objects 
* vmnetcfg-clear-status \<namespace> \<name>: clears the status of a VirtualMachineNetworkConfig object (in case of errors and this needs to be cleared)
* vmnetcfg-reset \<namespace> \<name>: resets the VirtualMachineNetworkConfig object network configuration (in case you want to allocate a new IP)
* vmnetcfg-cleanup: cleans up VirtualMachineNetworkConfig object orphans (this is interactive with a backup option)

# License

Copyright (c) 2024 Joey Loman <joey@binbash.org>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

[http://www.apache.org/licenses/LICENSE-2.0](http://www.apache.org/licenses/LICENSE-2.0)

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
