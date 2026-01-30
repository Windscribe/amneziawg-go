# AmneziaWG Configuration Utility

A simple Go utility to configure AmneziaWG interfaces via UAPI socket.

## Building

```bash
go build -o awg-config
```

## Usage

### Start AmneziaWG (Terminal 1)

```bash
cd /Users/gindersingh/amneziawg-go
sudo ./amneziawg-go -f utun
```

Note the interface name that gets created (e.g., `utun7`).

### Configure the Interface (Terminal 2)

Replace `utun7` with your actual interface name from Terminal 1.

```bash
cd /Users/gindersingh/amneziawg-go/awg-config

# Step 1: Set IP address
sudo ifconfig utun7 100.64.171.56 100.64.171.56

# Step 2: Apply AmneziaWG configuration
sudo ./awg-config -i utun7 -c example-fixed.conf

# Step 3: Verify configuration
sudo ./awg-config -i utun7 -show

# Step 4 (Optional): Add routes
sudo route add -net 0.0.0.0/1 -interface utun7
sudo route add -net 128.0.0.0/1 -interface utun7
```

## Configuration File Format

The utility supports WireGuard INI-style configuration files:

```ini
[Interface]
Address = 100.64.171.56/32
DNS = 10.255.255.1
PrivateKey = <base64-encoded-key>
Jc = 3
Jmin = 50
Jmax = 1000
S1 = 29
S2 = 87
H1 = 97329410
H2 = 907372093
H3 = 1342305062
H4 = 1802786097
I1 = <obfuscation-spec>

[Peer]
PublicKey = <base64-encoded-key>
PreSharedKey = <base64-encoded-key>
Endpoint = SERVER_IP:PORT
AllowedIPs = 0.0.0.0/0
PersistentKeepalive = 25
```

### Supported Parameters

**Interface Section:**
- `PrivateKey` - Your device's private key (base64)
- `Jc`, `Jmin`, `Jmax` - Junk packet obfuscation
- `S1`, `S2`, `S3`, `S4` - Message padding sizes
- `H1`, `H2`, `H3`, `H4` - Message header values
- `I1`, `I2`, `I3`, `I4`, `I5` - Custom signature packets

**Peer Section:**
- `PublicKey` - Peer's public key (base64)
- `PreSharedKey` - Pre-shared key (base64, optional)
- `Endpoint` - Peer's IP:PORT
- `AllowedIPs` - IP ranges to route through tunnel
- `PersistentKeepalive` - Keepalive interval in seconds

**Note:** `Address` and `DNS` are ignored by this tool - set them via `ifconfig` and system DNS configuration.

## Commands

### Apply Configuration
```bash
sudo ./awg-config -i <interface> -c <config-file>
```

### Show Current Configuration
```bash
sudo ./awg-config -i <interface> -show
```

## Managing AmneziaWG

### Check Available Sockets
```bash
ls -la /var/run/amneziawg/
```

### Stop All Instances
```bash
# Remove sockets (causes processes to exit)
sudo rm -f /var/run/amneziawg/*.sock

# Or kill processes directly
sudo killall amneziawg-go
```

### Check Interface Status
```bash
ifconfig utun7

# Look for:
# - flags=8051<UP,POINTOPOINT,RUNNING,MULTICAST>
# - inet 100.64.171.56 --> 100.64.171.56
```

### Verify Handshake
```bash
sudo ./awg-config -i utun7 -show

# Look for:
# - last_handshake_time_sec (should be non-zero)
# - tx_bytes and rx_bytes (traffic counters)
```

## Troubleshooting

### "Network is unreachable" when adding routes
- Check if IP address is set: `ifconfig utun7`
- Verify handshake: `sudo ./awg-config -i utun7 -show`
- Check Terminal 1 for error messages

### "dial unix /var/run/amneziawg/utunX.sock: no such file"
- Make sure amneziawg-go is running in Terminal 1
- Check socket exists: `ls -la /var/run/amneziawg/`
- Use correct interface name matching the socket file

### Configuration fails with "errno=-22"
- Verify config file format (no line breaks in long values like I1)
- Check that base64 keys are valid
- Look for the debug output showing the UAPI config being sent

### No handshake (last_handshake_time_sec=0)
- Check endpoint is correct and reachable
- Verify keys match between client and server
- Check firewall isn't blocking UDP traffic
- Ensure server is running and configured correctly

## Key Conversion

The utility automatically converts base64-encoded keys (WireGuard format) to hex (UAPI format). You can use standard WireGuard key format in your config files.
