# WebMine
A Minecraft web server panel made using Go and daisyUI.

## Build 
To build the project, you will need Go and Bun. Bun is used to build daisyUI, the CSS framework used in the project.

The istructions are for linux, so for Windows users, you need to adapt the commands.

First clone the repository:
```bash
git clone https://github.com/Skyfield1888/WebMine.git && cd WebMine/
```

Then, install the dependencies with Bun:
```bash
bun install
```

You should now be able to run the program using `go run .`
Note that the executable is not yet standalone.