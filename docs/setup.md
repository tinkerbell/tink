# Setup the Provisioner

### Prerequisites
 - The setup must be executed as a privileged or root user.
 - The setup downloads about 1.5GB of assets, therefore, it is advised to have a minimum of 10GB disk space available before you begin.

### Interactive Mode
Execute the following commands to bring up the Tinkerbell stack with an interactive setup:
```shell
$ wget https://raw.githubusercontent.com/tinkerbell/tink/master/setup.sh && chmod +x setup.sh
$ ./setup.sh
```

### Declarative Mode
You can also execute the setup in declarative mode.
In order to do so, define the following environment variables (examples here):
```shell
export TB_INTERFACE=network-interface         # enp1s0f1
export TB_NETWORK=network-with-cidr           # 192.168.1.0/29
export TB_IPADDR=provisioner-ip-address       # 192.168.1.1
export TB_REGUSER=registry-username           # admin
```

Now, you can execute the setup with the following command:
```shell
$ curl https://raw.githubusercontent.com/tinkerbell/tink/master/setup.sh | bash
```

### Good to know
 - All the environment variables are kept in the `envrc` file, which is generated from the setup itself.
 - It is advised that you keep all the environment variables in the same file.
 - It is important to note that if you execute the setup the again, a new `envrc` will be generated.
   However, the existing environment configuration be saved as `envrc.bak`.
 - The setup removes all the `.tar.gz` files downloaded in process.

### For Packet Environment

The script was tested with:
 - Server class: `c3.small.x86`
 - Operating System: Ubuntu 18.04 and CentOS 7
 - Region: Amsterdam, NL (AMS1)
 - ENV variables:
 ```shell
  export TB_INTERFACE=enp1s0f1
  export TB_NETWORK=192.168.1.0/29
  export TB_IPADDR=192.168.1.1
  export TB_REGUSER=admin
 ```
 - Command:
 ```shell
 $ curl https://raw.githubusercontent.com/tinkerbell/tink/master/setup.sh | bash
 ```

### What's Next

Once you have the provisioner setup successfully, you can try executing your first workflow.
Follow the steps described in [here](hello-world.md) to say "Hello World!" with a workflow.

