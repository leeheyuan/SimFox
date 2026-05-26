# SimFox Desktop Client

This Electron client is a local capability bridge for the SimFox private-cloud platform.

It should not execute cluster simulations directly. SUMO workloads belong on private-cloud worker machines. The desktop client only provides controlled native capabilities that browsers cannot provide well:

- Pick local files and folders
- Upload input packages to the platform API
- Download result artifacts to a chosen local folder
- Open local folders and files
- Store the private-cloud endpoint and local preferences
- Show desktop notifications
- Optionally run approved local validation tools

The renderer must call APIs exposed by `preload`; it should not use Node.js or arbitrary shell execution.
