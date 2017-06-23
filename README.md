# salt-translator-service
Translates salt events and pushes them to an ELK stack

The address of the logstash shipper must be specified with an environment variable called ```ELASTIC_API_EVENTS```. 

The subscription to salt depends on ```SALT_MASTER_ADDRESS```, ```SALT_EVENT_USERNAME```, and ```SALT_EVENT_PASSWORD```.
