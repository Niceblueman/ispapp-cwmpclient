# SSH Test Scripts for Device Testing

This directory contains bash scripts for testing commands on remote devices via SSH.

## Available Scripts

### 1. `ssh-test.sh` - Password-based SSH Testing

This script uses password authentication to connect to remote devices.

**Prerequisites:**
- `sshpass` must be installed on your system
- Ubuntu/Debian: `sudo apt-get install sshpass`
- CentOS/RHEL: `sudo yum install sshpass`
- macOS: `brew install sshpass`

**Configuration:**
Edit the constants at the top of the script:
```bash
SSH_HOST="192.168.1.1"          # Target device IP or hostname
SSH_PORT="22"                   # SSH port (default: 22)
SSH_USER="root"                 # SSH username
SSH_PASSWORD="password"         # SSH password
SSH_TIMEOUT="10"                # Connection timeout in seconds
```

### 2. `ssh-test-key.sh` - SSH Key-based Testing

This script uses SSH key authentication (more secure, no password required).

**Prerequisites:**
- SSH key pair must be generated: `ssh-keygen -t rsa -b 2048`
- Public key must be added to the device's `~/.ssh/authorized_keys` file

**Configuration:**
Edit the constants at the top of the script:
```bash
SSH_HOST="192.168.1.1"          # Target device IP or hostname
SSH_PORT="22"                   # SSH port (default: 22)
SSH_USER="root"                 # SSH username
SSH_KEY_PATH="~/.ssh/id_rsa"    # Path to SSH private key
SSH_TIMEOUT="10"                # Connection timeout in seconds
```

## Usage Examples

### Basic Usage
```bash
# Show help
./scripts/ssh-test.sh --help

# Test SSH connection only
./scripts/ssh-test.sh --check-connection

# Show current configuration
./scripts/ssh-test.sh --config
```

### Running Tests
```bash
# Run predefined device tests
./scripts/ssh-test.sh --test

# Execute a single command
./scripts/ssh-test.sh -c "uname -a"

# Enter interactive mode
./scripts/ssh-test.sh --interactive
```

### Interactive Mode
In interactive mode, you can execute commands one by one:
```
ssh-test> uname -a
ssh-test> ps aux | grep cwmp
ssh-test> exit
```

## Predefined Tests

Both scripts include the following predefined tests:

1. **System Information**
   - OS version and kernel information
   - System uptime

2. **Network Information**
   - IP addresses and network interfaces
   - Routing table

3. **System Resources**
   - Memory usage
   - Disk space usage

4. **Process Information**
   - Running processes

5. **CWMP/ISPApp Specific Tests**
   - CWMP-related processes
   - ISPApp-related processes  
   - Configuration files

## Security Notes

- **Password Authentication**: The password is stored in plain text in the script. Use this only for testing environments.
- **SSH Key Authentication**: More secure option. The private key should be properly protected (chmod 600).
- **Host Key Checking**: Both scripts disable strict host key checking for ease of use in testing environments.

## Troubleshooting

### Common Issues

1. **Connection Refused**
   - Check if SSH service is running on the target device
   - Verify the IP address and port number

2. **Permission Denied**
   - Check username and password/key
   - Ensure SSH key is properly configured

3. **Command Not Found (sshpass)**
   - Install sshpass package (password-based script only)

4. **Key Authentication Failed**
   - Ensure SSH key is added to authorized_keys on the target device
   - Check SSH key permissions (should be 600)

### Testing OpenWrt/LEDE Devices

For OpenWrt/LEDE devices, typical settings are:
- Default IP: `192.168.1.1`
- Default user: `root`
- Default SSH port: `22`
- Password: Set during initial setup

### Testing the Scripts

You can test the scripts locally first:
```bash
# Test on localhost
SSH_HOST="localhost" ./scripts/ssh-test-key.sh --test
```

## Integration with ISPApp CWMP Client

These scripts are designed to work with the ISPApp CWMP client project and can be used to:
- Test CWMP daemon status on remote devices
- Monitor configuration changes
- Verify connectivity and system health
- Debug deployment issues

The predefined tests include specific checks for CWMP and ISPApp processes and configuration files.
