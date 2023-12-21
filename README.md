# FINAL YEAR PROJECT

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
