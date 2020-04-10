# Setup the Provisioner machine with docker-compose.yml file

1. ### Setup git
    ```shell
   Ubuntu -  $ apt install -y git
   CentOS -  $ yum install -y git
    ```

2. ### Install docker and docker-compose as follows:
   ```shell
    $ curl -L get.docker.com | bash
    $ curl -L "https://github.com/docker/compose/releases/download/1.24.1/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
    $ chmod +x /usr/local/bin/docker-compose
    For CentOS only:
        $ systemctl start docker 
        $ echo "net.ipv4.ip_forward=1" >> /etc/sysctl.conf 
        $ systemctl restart network
   ```

3. ### Clone the tink repo in the $GOPATH
    ```shell
    $ git clone https://github.com/tinkerbell/tink.git
    $ cd tink/deploy
    ```

4. ### Provide the input details in "inputenv" file which also includes the name of network interface which you would like to configure.

5. ### Run the following command as "root" user
    ```shell
    $ source inputenv
    $ ./setup_with_docker_compose.sh
    ```
