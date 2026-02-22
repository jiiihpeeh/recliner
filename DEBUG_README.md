# ReCLIner Debug Mode

ReCLIner supports a debug mode that provides real-time debug output through a socket-based system.

## Usage

### Method 1: Separate Terminals

1. In one terminal, run the debug client:
   ```bash
   ./debug_client
   ```

2. In another terminal, run ReCLIner with debug mode:
   ```bash
   ./recliner --debug
   ```

### Method 2: Automated Script

Run the provided script that opens debug output in a new terminal window:
```bash
./debug_run.sh
```

## How it Works

- When `--debug` flag is used, ReCLIner starts a Unix socket server at `/tmp/recliner_debug.sock`
- The debug client connects to this socket and displays real-time debug output
- Debug messages include timestamps and are also written to `/tmp/recliner_debug.log`
- Multiple debug clients can connect simultaneously

## Debug Output

The debug output includes:
- Application startup and initialization
- Terminal setup and mode changes
- Rendering cycles
- User input events
- Component lifecycle events
- Error conditions and recovery

## Building

The debug functionality is built into the main ReCLIner binary. The debug client is built separately:

```bash
# Build main application
go build -o recliner main.go

# Build debug client
go build -o debug_client ./cmd/debug
```