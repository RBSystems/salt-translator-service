# salt-translator-service
Translates salt events and pushes them to an ELK stack

The address of the logstash shipper must be specified with an environment variable called ```ELASTIC_API_EVENTS```. 

The subscription to salt depends on ```SALT_MASTER_ADDRESS```, ```SALT_EVENT_USERNAME```, and ```SALT_EVENT_PASSWORD```.

Installation Instructions:
Copy the salt-translator.service file and place it in /etc/systemd/system/
After copying the service, make sure permissions are set so owner and group have read/write permissiongs
```chmod 664 /etc/systemd/system/salt-translator.service```
Copy the binary down from Github into /usr/bin/salt-translator-service
```wget -O https://github.com/byuoitav/salt-translator-service/releases/download/v0.1.0/salt-translator-service-linux```
Change the name from salt-translator-service-linux to salt-translator-service
After copying the binary and changing the name, make sure permissions are set so owner has full permissions and both group and other have read and execute
```chmod 755 /etc/systemd/system/salt-translator.service```
Set the environmental variables by creating /etc/environment and set the following variables:
```ELASTIC_API_EVENTS='http://elk-stack.example.com:5546' ```
```SALT_MASTER_ADDRESS='https://localhost:8000'```
```SALT_EVENT_USERNAME='user'```
```SALT_EVENT_PASSWORD='userpassword' ```
Change permissions so that owner has read/write and group and other have read access
```chmod 644 /etc/environment```
