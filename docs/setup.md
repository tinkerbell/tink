# Setup the Packet Workflow Environment with Terraform on packet provider

 - Clone the `tink` repository for latest code:
```shell
$ git clone https://github.com/tinkerbell/tink.git
$ cd tink/terraform
```

 - Update the `input.tf` file with desired values 
 - Add your Packet `auth_token` in `input.tf`
 - Run the following commands
```shell
$ terraform init
$ terraform apply
``` 

The above commands will create a complete setup with `tf-provisioner` and `tf-worker` machines for the `packet` provider on which you can run any workflow. As an output it returns the IP address of the provisioner and MAC address of the worker machine.


***_Note_ :*** The default names of machines created by Terraform are `tf-provisioner` and `tf-worker`. If you prefer other names, you need to replace `tf-provisioner` and `tf-worker` with the new ones at all places in `main.tf`.


# Setup the Provisioner machine with docker-compose.yml file

## Install git and git lfs as follows

1. ### Setup git and git lfs
    ```shell
    $ sudo apt install -y git  
    $ wget https://github.com/git-lfs/git-lfs/releases/download/v2.9.0/git-lfs-linux-amd64-v2.9.0.tar.gz  
    $ tar -C /usr/local/bin -xzf git-lfs-linux-amd64-v2.9.0.tar.gz  
    $ rm git-lfs-linux-amd64-v2.9.0.tar.gz  
    $ git lfs install  

2. ### Setup go
    ```shell
    $ wget https://dl.google.com/go/go1.13.9.linux-amd64.tar.gz
    $ tar -C /usr/local -xzf go1.13.9.linux-amd64.tar.gz go/
    $ rm go1.12.13.linux-amd64.tar.gz

3. ### Set GOPATH
    ```shell
    $ echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
    $ echo 'export GOPATH=$GOPATH:$HOME/go' >> ~/.bashrc
    $ echo 'export PATH=$PATH:$GOPATH' >> ~/.bashrc
    $ source ~/.bashrc

4. ### Install docker and docker-compose as follows:
   ```shell
    $ curl -L get.docker.com | bash
    $ curl -L "https://github.com/docker/compose/releases/download/1.24.1/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
    $ chmod +x /usr/local/bin/docker-compose

5. ### Clone the tink repo in the $GOPATH
    ```shell
    $ mkdir -p ~/go/src/github.com/tinkerbell
    $ cd ~/go/src/github.com/tinkerbell
    $ git clone https://github.com/tinkerbell/tink.git
    $ cd tink

6. ### Provide the input details in "inputenv" file

7. ### Run the following command
    ```
    $ sudo ./setup_with_docker_compose.sh
