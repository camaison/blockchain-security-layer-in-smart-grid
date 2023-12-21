# FINAL YEAR PROJECT

## Description

Simulating GOOSE communication between Intelligent Electronic Devices(IEDs) in response to substation events in a Smart Distribution Grid using the IEC 61850 Protocol and Virtual Machines.

## System Overview

### Components

- RDSO (Main Power Supply): The primary source of power for the grid.
- DSO1 (Consumers): The entities consuming power from the grid.
- IPP (Independent Power Producers): Alternative power sources activated when RDSO is offline.
- IED 1 (RDSO IED): Communicates the status of the RDSO Circuit Breaker.
- IED 2 (IPP IED): Communicates the status of the IPP Circuit Breaker.

### Diagram

![Diagram]([images\Use_Case_Diagram.jpg](https://github.com/camaison/final-year-project-smart-grid-security/blob/main/images/Use_Case_Diagram.jpg?raw=true))

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

### 1. Clone this repository

Run `git clone https://github.com/camaison/final-year-project-smart-grid-security.git`

NB: Decide which Virtual Machine will run IPP_IED and which will run RDSO_IED and stick to that.

### 2. Change directory to the project folder

Run `cd final-year-project-smart-grid-security`

### 3. Clone libiec61850 library

Run `git clone https://github.com/mz-automation/libiec61850.git`

## Usage Instructions

In the directory containing the script file for each IED (IPP_IED or RDSO_IED) open the terminal making sure it points to that path. Then:

### 1. Compile the script

Run `make` to compile the script.

### 2. Run the script

Run `sudo ./{name of compiled script}` to run the code.

Example: `sudo ./ipp`
