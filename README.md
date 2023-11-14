# Invitation Leader Election Algorithm

Implementation of the Invitation algorithm. 

### Configuring the peers

The configuration .yaml should have the following structure:

```yaml
net_port: <At what port should the service run on each of the peers>

#List of peers where the service is being run
peers: 
  - peer_name: <Name of the peer, should be a positive integer>
    net_name: <url,IP,NetworkID of the peer>
```

It is important to also set the **name** variable, it can be eiter set on the config.yaml or as an environment variable.

### Running the example

1. Execute the following command to build the image of a container that runs the algorithm
```bash
  docker build -t invitation -f ./Dockerfile .
```
3. Execute the following command to run the amount of peers that are defined at docker-compose-dev.yaml
```bash
  docker compose -f ./docker-compose-dev.yaml up
```
4. Execute the following command to clean up
```bash
  docker compose -f ./docker-compose-dev.yaml down
```
