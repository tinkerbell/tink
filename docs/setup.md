# Setup the Provisioner machine with docker-compose.yml file

Execute the following commands to bring up the Tinkerbell stack and setup the provisioner:
```shell
$ wget https://raw.githubusercontent.com/infracloudio/tink/deploy_stack/setup.sh && chmod +x setup.sh
$ ./setup.sh
```

All the environment variables are kept in the `envrc` file, which is generated from the setup itself. It is advised that you keep all the environment variables in the same file. It is important to note that if you execute the setup the again, a new `envrc` will be generated. However, the current environment configuration be saved as `envrc.bak`.
