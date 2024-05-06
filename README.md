# Blockchain as a Security Layer in the Smart Grid

Smart grid self-healing applications may be faced with the challenges of emerging cyber-physical security threats. These can result in disruption to the applications' operations thereby affecting the power system reliability. Blockchain is one technology that has been deployed in several applications to offer security and bookkeeping. Blockchain can be deployed as a second-tier security mechanism to support time-critical self-healing operations in smart distribution grids ([Reference Paper](https://drive.google.com/file/d/1-Sww2ZEgU-wcpWiuROuD4X6lCf96uzVd/view?usp=sharing)). The main objective of this task is to design and implement a testbed to realize this architecture.  The testbed should mimic a smart grid self-healing application that utilizes blockchain to achieve the second-tier security layer. 

## Activities

1. Review literature on blockchain, blockchain use in the smart grid, and self-healing applications.

2. Using a proposed model, design and implement a testbed to study the behavior of blockchain deployed on a self-healing smart grid network.

3. Collect performance data and analyze it.
   

## System Overview

- RDSO (Main Power Supply): The primary source of power for the grid.
- DSO1 (Consumers): The entities consuming power from the grid.
- IPP (Independent Power Producers): Alternative power sources activated when RDSO is offline.
- IED 1 (RDSO IED): Communicates the status of the RDSO Circuit Breaker.
- IED 2 (IPP IED): Communicates the status of the IPP Circuit Breaker.

### Diagram

![Diagram](https://github.com/camaison/final-year-project-smart-grid-security/blob/main/images/Use_Case_Diagram.jpg?raw=true)

### IED Communication and Functionality

#### IED 1 (RDSO IED):

Monitors and communicates the status of the RDSO.

##### Functionality

- Toggles its status (circuit breaker open/closed) at regular intervals.
- Publishes this status using GOOSE messaging.
- Subscribes and Receives updates from IED2, reflecting the system's current state.

#### IED 2 (IPP IED):

Monitors and communicates the status of the IPP.

##### Functionality

- Subscribes and Receives GOOSE messages from IED1 indicating the status of the RDSO Circuit Breaker.
- Updates its status based on RDSO's condition.
- Publishes its status (circuit breaker open/closed) using GOOSE messaging.

## Setup Instructions

After setting up both virtual machines running linux debian 11 and in the linux terminals of both machines:
### 1. Install the Pre-requisite Hyperledger Fabric Software and Fabric Samples
``` 
https://hyperledger-fabric.readthedocs.io/en/release-2.5/prereqs.html 
https://hyperledger-fabric.readthedocs.io/en/release-2.5/install.html
```

### 2. Clone this repository
In the same directory as the fabric samples, clone this repository. 

Run `git clone https://github.com/camaison/blockchain-security-layer-in-smart-grid.git`

NB: Decide which Virtual Machine will run IPP_IED and which will run RDSO_IED and stick to that.

It should look like this:
```
/fabric-samples
/blockchain-security-layer-in-smart-grid
install-fabric.sh
```

### 3. Change directory to the project folder

Run `cd blockchain-security-layer-in-smart-grid/GOOSE_Client`

### 4. Clone libiec61850 library

Run `git clone https://github.com/mz-automation/libiec61850.git`

## Usage Instructions

In the directory containing the script file for each IED (IPP_IED or RDSO_IED) open the terminal making sure it points to that path. Then:
### 1. Update Packages

Run `sudo apt update` to update packages.

### 2. Install make 

Run `sudo apt install make` to install make.

### 3. Compile the script

Run `make` to compile the script.

### 4. Run the script

Run `sudo ./{name of compiled script}` to run the code.

Example: `sudo ./ipp`
