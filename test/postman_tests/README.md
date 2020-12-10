# Summary of the Collection of Tests
Using Postman Runner/Newman, the collection will read the hardware id and operating system plan from provisioning_worker_info.csv on dfw2lab.

The first call will check to see if the hardware is in a "provisionable" state for the hardware
(assuming there will be reserved workers for automation) and the server is reachable with a 200 response code.
If not, the tests will exit. If the hardware is "provisionable", it will send the POST /devices call.

The next tests will try to reboot, rescue, and deprovision the server. It will first check to see if we can get the device id from the GET /hardware call. 
* If the state is "failed", the test will exit. 
* If the state is "provisioning", it will recheck the status in 30 seconds.
* If the state is "in_use", it will reboot.

# Important Notes
* The collection to provision servers only targets dfw2lab and t3.small.x86
* Environment variables and values will not work right out the box. You will need to tokens through the staff portal.
* Automation in the CI pipeline in future will require the collection to fetch X-Auth-Token and X-Consumer-Tokens.
* Considerations: Hardware reservation agreement so the tests are idempotent and repeatable for CI/CD Pipeline.

#  Refer to Wiki on How to Use Postman
https://wiki.corp.equinix.com/confluence/display/ES/Step-by-Step+Quick+Start+Postman+Guide

# Download and Import
1. Install Postman
2. Download the Collection and Environment Collection
3. Import the collection or environment collection files into Postman
