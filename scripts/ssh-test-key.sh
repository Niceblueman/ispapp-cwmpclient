#!/bin/bash

# SSH Test Script for Device Testing (SSH Key Authentication)
# This script allows testing commands on a remote device via SSH
# using SSH key authentication (no password required)

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# SSH Connection Constants - EDIT THESE VALUES
SSH_HOST="192.168.1.1"          # Target device IP or hostname
SSH_PORT="22"                   # SSH port (default: 22)
SSH_USER="root"                 # SSH username
SSH_KEY_PATH="~/.ssh/id_rsa"    # Path to SSH private key
SSH_TIMEOUT="10"                # Connection timeout in seconds

# Function to print colored output
print_status() { echo -e "${BLUE}[INFO]${NC} $1"; }
print_success() { echo -e "${GREEN}[SUCCESS]${NC} $1"; }
print_warning() { echo -e "${YELLOW}[WARNING]${NC} $1"; }
print_error() { echo -e "${RED}[ERROR]${NC} $1"; }

# Function to check if SSH key exists
check_ssh_key() {
    local key_path=$(eval echo "$SSH_KEY_PATH")
    if [[ ! -f "$key_path" ]]; then
        print_error "SSH key not found at: $key_path"
        print_status "Please ensure your SSH key is available or use the password-based script"
        print_status "To generate an SSH key: ssh-keygen -t rsa -b 2048"
        exit 1
    fi
}

# Function to test SSH connection
test_ssh_connection() {
    print_status "Testing SSH connection to $SSH_USER@$SSH_HOST:$SSH_PORT..."
    
    if ssh -o StrictHostKeyChecking=no \
       -o ConnectTimeout="$SSH_TIMEOUT" \
       -i "$SSH_KEY_PATH" \
       -p "$SSH_PORT" "$SSH_USER@$SSH_HOST" \
       'echo "SSH connection successful"' >/dev/null 2>&1; then
        print_success "SSH connection established successfully"
        return 0
    else
        print_error "Failed to establish SSH connection"
        print_status "Make sure your SSH key is added to the device's authorized_keys"
        return 1
    fi
}

# Function to execute command on remote device
execute_remote_command() {
    local command="$1"
    local show_output="${2:-true}"
    
    print_status "Executing: $command"
    
    if [[ "$show_output" == "true" ]]; then
        ssh -o StrictHostKeyChecking=no \
            -o ConnectTimeout="$SSH_TIMEOUT" \
            -i "$SSH_KEY_PATH" \
            -p "$SSH_PORT" "$SSH_USER@$SSH_HOST" \
            "$command"
    else
        ssh -o StrictHostKeyChecking=no \
            -o ConnectTimeout="$SSH_TIMEOUT" \
            -i "$SSH_KEY_PATH" \
            -p "$SSH_PORT" "$SSH_USER@$SSH_HOST" \
            "$command" >/dev/null 2>&1
    fi
    
    local exit_code=$?
    if [[ $exit_code -eq 0 ]]; then
        print_success "Command executed successfully"
    else
        print_error "Command failed with exit code: $exit_code"
    fi
    
    return $exit_code
}

# Function to run predefined device tests
run_device_tests() {
    print_status "Running predefined device tests..."
    echo ""
    
    # Test 1: System information
    print_status "=== System Information ==="
    execute_remote_command "uname -a"
    execute_remote_command "cat /etc/os-release 2>/dev/null || cat /etc/openwrt_release 2>/dev/null || echo 'OS info not available'"
    execute_remote_command "uptime"
    echo ""
    
    # Test 2: Network information
    print_status "=== Network Information ==="
    execute_remote_command "ip addr show | grep -E '^[0-9]+:|inet '"
    execute_remote_command "ip route show | head -10"
    echo ""
    
    # Test 3: System resources
    print_status "=== System Resources ==="
    execute_remote_command "free -h 2>/dev/null || free"
    execute_remote_command "df -h | head -10"
    echo ""
    
    # Test 4: Process information
    print_status "=== Running Processes ==="
    execute_remote_command "ps aux | head -10"
    echo ""
    
    # Test 5: CWMP/ISPApp specific tests (if applicable)
    print_status "=== CWMP/ISPApp Tests ==="
    execute_remote_command "ps aux | grep -i cwmp || echo 'No CWMP processes found'"
    execute_remote_command "ps aux | grep -i ispapp || echo 'No ISPApp processes found'"
    execute_remote_command "ls -la /etc/config/ | grep -E 'cwmp|ispapp' || echo 'No CWMP/ISPApp config files found'"
    echo ""
}

# Function to run interactive command mode
interactive_mode() {
    print_status "Entering interactive command mode..."
    print_status "Type 'exit' or 'quit' to return to main menu"
    print_status "Type 'help' for available commands"
    echo ""
    
    while true; do
        echo -n "ssh-test> "
        read -r user_command
        
        case "$user_command" in
            "exit"|"quit")
                print_status "Exiting interactive mode"
                break
                ;;
            "help")
                echo "Available commands:"
                echo "  exit/quit  - Exit interactive mode"
                echo "  help       - Show this help"
                echo "  test       - Run predefined tests"
                echo "  Any other command will be executed on the remote device"
                ;;
            "test")
                run_device_tests
                ;;
            "")
                continue
                ;;
            *)
                execute_remote_command "$user_command"
                ;;
        esac
        echo ""
    done
}

# Function to show help
show_help() {
    cat << EOF
SSH Test Script for Device Testing (SSH Key Authentication)

Usage: $0 [options] [command]

Options:
    -h, --help          Show this help message
    -c, --command CMD   Execute single command and exit
    -t, --test          Run predefined device tests
    -i, --interactive   Enter interactive command mode
    --check-connection  Test SSH connection only
    --config            Show current configuration

Configuration:
    SSH_HOST:     $SSH_HOST
    SSH_PORT:     $SSH_PORT
    SSH_USER:     $SSH_USER
    SSH_KEY_PATH: $SSH_KEY_PATH
    SSH_TIMEOUT:  $SSH_TIMEOUT seconds

Examples:
    $0 --test                           # Run predefined tests
    $0 -c "uname -a"                    # Execute single command
    $0 --interactive                    # Enter interactive mode
    $0 --check-connection               # Test connection only

To modify connection settings, edit the constants at the top of this script.

Note: This script uses SSH key authentication. Make sure your SSH key is:
1. Present at the specified path
2. Added to the device's ~/.ssh/authorized_keys file

EOF
}

# Function to show current configuration
show_config() {
    print_status "Current SSH Configuration:"
    echo "  Host:     $SSH_HOST"
    echo "  Port:     $SSH_PORT"
    echo "  User:     $SSH_USER"
    echo "  Key Path: $SSH_KEY_PATH"
    echo "  Timeout:  $SSH_TIMEOUT seconds"
}

# Main execution
main() {
    # Check SSH key
    check_ssh_key
    
    # Parse command line arguments
    while [[ $# -gt 0 ]]; do
        case $1 in
            -h|--help)
                show_help
                exit 0
                ;;
            -c|--command)
                if [[ -n "$2" ]]; then
                    test_ssh_connection && execute_remote_command "$2"
                    exit $?
                else
                    print_error "Command option requires an argument"
                    exit 1
                fi
                ;;
            -t|--test)
                test_ssh_connection && run_device_tests
                exit $?
                ;;
            -i|--interactive)
                test_ssh_connection && interactive_mode
                exit $?
                ;;
            --check-connection)
                test_ssh_connection
                exit $?
                ;;
            --config)
                show_config
                exit 0
                ;;
            *)
                print_error "Unknown option: $1"
                show_help
                exit 1
                ;;
        esac
        shift
    done
    
    # If no arguments provided, show help
    show_help
}

# Run main function with all arguments
main "$@"
