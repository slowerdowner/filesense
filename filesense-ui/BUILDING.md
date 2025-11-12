# Building the Filesense UI

This document outlines the steps required to build the `filesense-ui` application on a fresh Linux (Ubuntu/Debian) environment.

## Prerequisites

The following system dependencies are required to build the application:

- Go (version 1.18 or later)
- Node.js (version 16 or later)
- npm (version 8 or later)
- Wails CLI (version 2 or later)

## System Dependencies

The Wails framework relies on a number of system libraries to build the application. On Ubuntu/Debian, you can install these with the following command:

```bash
sudo apt-get update
sudo apt-get install -y libgtk-3-dev libwebkit2gtk-4.1-dev
```

**Note:** The build process specifically requires `webkit2gtk-4.0`. On newer systems where this version is not available, you may need to create a symbolic link to trick the build system into using version 4.1:

```bash
sudo ln -s /usr/lib/x86_64-linux-gnu/pkgconfig/webkit2gtk-4.1.pc /usr/lib/x86_64-linux-gnu/pkgconfig/webkit2gtk-4.0.pc
```

## Building the Application

Once the prerequisites and system dependencies are installed, you can build the application with the following command:

```bash
wails build
```

This will create a `filesense-ui` executable in the `build/bin` directory.
