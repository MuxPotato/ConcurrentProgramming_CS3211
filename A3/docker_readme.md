# Docker Runtime (Optional Tool)
To aid you in your testing **locally**, we are providing a simple, lightweight [Docker](https://en.wikipedia.org/wiki/Docker_(software)) image and scripts for you to be able to control `ulimit` and other [resource limits](https://docs.docker.com/engine/containers/resource_constraints/) such as memory and cpu, if needed.  

- You will not be able to use Docker on the PDC cluster.  

- :warning: The docker runtime will NOT be used during grading.   

**We will not entertain debugging requests for the docker runtime and scripts**.

## Pre-requisites 

You need to install docker on your machine.
- [Windows install](https://docs.docker.com/desktop/setup/install/windows-install/)
  - [Enable Docker for WSL2](https://docs.docker.com/desktop/features/wsl/)
  - You will probably need to have Docker Desktop running for docker to be active in your WSL shell environment 
- [MacOS install](https://docs.docker.com/desktop/setup/install/mac-install/)
- [Linux install](https://docs.docker.com/engine/install/)  

The following instructions will also assume you are running docker on a shell that can run `.sh` scripts, such as WSL or bash.

### Disclaimer

The following scripts and containers were tested on a Windows 10 machine and Ubuntu machine.  

You may have to tweak the `dockerfile` configuration and pull different images if the script does not work.  

You may need to `chmod +x` the shell scripts if you get permission denied errors.  
 
Testing with extreme levels of clients *may* still result in your program encountering connection issues. However, the docker runtime should be resilient enough to handle the client limits stated in the FAQ.

## Building

Choose a name for your image. Something like `cs3211-a3-docker`. Your image name will be referred to as `<img_name>` in this document.

Your `dockerfile` and `docker_build.sh` should be in the root Rust project directory, i.e. same directory level as `Cargo.toml`.

In `Cargo.toml`, your project name must be "cs3211_assignment_3".

To build the image, simply:  
- `$ . docker_build.sh <img_name>` or
- `$ ./docker_build.sh <img_name>`  

## Running  

Your `docker_run.sh` should be in the root Rust project directory, i.e. same directory level as `Cargo.toml`.

- `$ . docker_run.sh <img_name> <port> <seed> <clients> <messages>` OR
- `$ ./docker_run.sh <img_name> <port> <seed> <clients> <messages>`

If you want to impose [resource limits](https://docs.docker.com/engine/containers/resource_constraints/) on the runtime, you will need to edit the script.

## Cleanup  

To remove your image:  
`$ docker image rm <img_name>`