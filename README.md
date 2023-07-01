# COSO
Containerization Oriented Standard Operations


*In Italian, the word "coso" is an informal term that is used to refer to a thing, object, or an unspecified item. It is often used when the specific name or term for something is not known or forgotten. It can also be used as a placeholder word when referring to a person whose name you can't remember or don't want to mention. In English, "coso" can be loosely translated as "thingy," "whatchamacallit," or "doodad."*

## Status: WIP

- 30/06/23 - project start

## Aim

COSO is mainly a study project to apply teachings on Linux containers internals.
In the long run, the project should evolve in a Docker-like CLI capable of reading configurations files and spin up containers according to the given instructions.

## Setup

Right now (01/07/23) some areas of COSO are not handled directly by the main binary executable.

 - Filesystem: a minimal Alpine distribution for x86_64 architecture is provided and used as lower layer
 - Networking: the configuration of the necessary devices used to route traffic from the standard namespace to the newly created one(s) is handled by a separate bynary: *cosonet*

You can run  `make fs-setup`  and  `make net-setup`  to configure the above.

Or just `make run` and follow the error messages :)


## Running coso

IF the setup has been successfull, you should be able to run COSO with `make run`.

# Customisazation

COSO expects a path to a root filesystem, to use as lower layer, and to an executable which can handle network devices creation.
COSO will execute the following call when trying to setup network devices for the new process (running in a separate namespace):

`<path to the executable> -pid <pid of the child process>`

You can modify the 2 above mentioned paths with the following flags

| Flag | Type | Default | Meaning
| :---:|:--:|:--:|:--|
| rootfs | string | /tmp/coso/rootfs | path to the root filesystem |
| network | string | /usr/local/bin/cosonet | path to the executable which will handle the setup of  network devices |


