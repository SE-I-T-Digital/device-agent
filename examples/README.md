Example Directory
The example directory contains everything needed to run the device agent and demonstrate its functionality.

Application Deployment Definition
The application deployment definition consists of three folders:
- install: Contains the application definition for deploying a new application.
- update: Contains the application definition for updating an already deployed application using the same file name as in the install folder.
- resources: Contains additional files (resources) required for the installation or update.

File Locations
All files must be copied to the appropriate locations:
- The application definition should be copied to the location specified by DeploymentDir defined in common.go.
- The resources should be copied to the location specified by AppData defined in common.go.