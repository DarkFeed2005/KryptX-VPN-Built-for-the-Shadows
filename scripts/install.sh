#!/bin/bash

# KryptX VPN Installation Script

set -e

INSTALL_DIR="/usr/local/bin"
CONFIG_DIR="/etc/kryptx"
SERVICE_DIR="/etc/systemd/system"

echo "Installing KryptX VPN..."

# Check if running as root
if [[ $EUID -eq 0 ]]; then
   echo "Don't run this script as root. Use sudo when prompted."
   exit 1
fi

# Check dependencies
echo "Checking dependencies..."
if ! command -v wireguard &> /dev/null; then
    echo "WireGuard not found. Installing..."
    case "$(uname -s)" in
        Linux*)
            if command -v apt &> /dev/null; then
                sudo apt update && sudo apt install -y wireguard
            elif command -v yum &> /dev/null; then
                sudo yum install -y wireguard-tools
            else
                echo "Please install WireGuard manually"
                exit 1
            fi
            ;;
        Darwin*)
            if command -v brew &> /dev/null; then
                brew install wireguard-tools
            else
                echo "Please install Homebrew and WireGuard manually"
                exit 1
            fi
            ;;
        *)
            echo "Unsupported OS. Please install WireGuard manually."
            exit 1
            ;;
    esac
fi

# Create directories
echo "Creating directories..."
sudo mkdir -p "$CONFIG_DIR"
sudo mkdir -p "$CONFIG_DIR/keys"

# Copy binary
echo "Installing binary..."
if [[ -f "build/kryptx" ]]; then
    sudo cp build/kryptx "$INSTALL_DIR/"
    sudo chmod +x "$INSTALL_DIR/kryptx"
else
    echo "Binary not found. Please run 'make build' first."
    exit 1
fi

# Copy configuration
if [[ -f "configs/client.yaml" ]]; then
    sudo cp configs/client.yaml "$CONFIG_DIR/"
    echo "Default configuration copied to $CONFIG_DIR/"
fi

# Set permissions
sudo chown -R root:root "$CONFIG_DIR"
sudo chmod 700 "$CONFIG_DIR"
sudo chmod 600 "$CONFIG_DIR"/*.yaml 2>/dev/null || true

# Create systemd service (Linux only)
if [[ -d "$SERVICE_DIR" ]]; then
    echo "Creating systemd service..."
    sudo tee "$SERVICE_DIR/kryptx.service" > /dev/null <<EOF
[Unit]
Description=KryptX VPN Client
After=network.target

[Service]
Type=simple
User=root
ExecStart=$INSTALL_DIR/kryptx -config=$CONFIG_DIR/client.yaml -gui=false
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
EOF

    sudo systemctl daemon-reload
    echo "Systemd service created. Enable with: sudo systemctl enable kryptx"
fi

# Create desktop entry (Linux only)
if command -v desktop-file-install &> /dev/null; then
    echo "Creating desktop entry..."
    cat > /tmp/kryptx.desktop <<EOF
[Desktop Entry]
Name=KryptX VPN
Comment=Secure VPN Client
Exec=$INSTALL_DIR/kryptx
Icon=network-vpn
Terminal=false
Type=Application
Categories=Network;Security;
EOF
    
    sudo desktop-file-install /tmp/kryptx.desktop
    rm /tmp/kryptx.desktop
fi

echo ""
echo "Installation complete!"
echo ""
echo "Configuration file: $CONFIG_DIR/client.yaml"
echo "Edit the configuration file and add your VPN server details."
echo ""
echo "Usage:"
echo "  GUI mode: kryptx"
echo "  CLI mode: kryptx -gui=false"
echo "  Service:  sudo systemctl start kryptx"
echo ""
echo "For help: kryptx -help"