# Second NIC BGP Configuration Guide

This document covers configuring a second network interface (USB Ethernet adapter) on the OCP SNO 1050ti node for BGP networking with MetalLB.

## Overview

The `ocp-sno-1050ti` node needs a second NIC configured on the BGP network (192.168.19.0/24) to enable MetalLB BGP mode for advanced LoadBalancer service routing.

## Hardware Information

**USB Ethernet Adapter**: ASIX AX88179 USB 3.0 Gigabit Ethernet
- **Interface**: `enp0s20f0u2` (or similar, check with `ip addr`)
- **Network**: 192.168.19.0/24
- **Gateway**: 192.168.19.1
- **Purpose**: BGP peering for MetalLB

## Prerequisites

- Physical access or SSH to the OCP SNO node
- Root or sudo access
- Network information:
  - IP address to assign (e.g., 192.168.19.20)
  - Subnet mask: 255.255.255.0 (/24)
  - Gateway: 192.168.19.1
  - BGP peer information (if known)

## Step 1: Identify the Interface

**On the OCP SNO node** (via SSH or console):

```bash
# List all network interfaces
ip addr show

# Or use nmcli
nmcli device status

# Check USB devices
lsusb | grep -i asix

# Check dmesg for interface name
dmesg | grep -i "ax88179\|enp0s20f0u2"
```

**Expected output**: Interface name like `enp0s20f0u2` or `enp0s20f0u1`

## Step 2: Configure Network Interface

### Option A: Using nmcli (Recommended for RHEL CoreOS)

```bash
# Connect to the node
oc debug node/ocp-sno-1050ti
chroot /host

# Check current interfaces
nmcli connection show

# Create new connection for second NIC
nmcli connection add \
  type ethernet \
  con-name "bgp-nic" \
  ifname enp0s20f0u2 \
  ipv4.addresses 192.168.19.20/24 \
  ipv4.gateway 192.168.19.1 \
  ipv4.dns "8.8.8.8 8.8.4.4" \
  ipv4.method manual \
  ipv6.method disabled

# Activate the connection
nmcli connection up "bgp-nic"

# Verify configuration
ip addr show enp0s20f0u2
ip route show
```

### Option B: Using NetworkManager Configuration File

**Create persistent configuration**:

```bash
# On the node (chroot /host)
cat > /etc/NetworkManager/system-connections/bgp-nic.nmconnection <<EOF
[connection]
id=bgp-nic
type=ethernet
interface-name=enp0s20f0u2

[ethernet]

[ipv4]
method=manual
addresses=192.168.19.20/24
gateway=192.168.19.1
dns=8.8.8.8;8.8.4.4

[ipv6]
method=disabled
EOF

# Set permissions
chmod 600 /etc/NetworkManager/system-connections/bgp-nic.nmconnection

# Reload NetworkManager
systemctl reload NetworkManager

# Activate connection
nmcli connection up bgp-nic
```

## Step 3: Verify Connectivity

```bash
# Check interface is up
ip link show enp0s20f0u2

# Check IP address is assigned
ip addr show enp0s20f0u2 | grep "inet "

# Test gateway connectivity
ping -c 3 192.168.19.1

# Test BGP peer connectivity (if known)
ping -c 3 <bgp-peer-ip>
```

## Step 4: Configure MetalLB for BGP Mode

**Note**: This is done in the `post-ocp-deployment.sh` script, but configuration details:

### MetalLB BGP Configuration

```yaml
apiVersion: metallb.io/v1beta2
kind: BGPPeer
metadata:
  name: bgp-peer
  namespace: metallb-system
spec:
  myASN: 64500  # Your ASN
  peerASN: 64501  # Peer ASN
  peerAddress: 192.168.19.1  # BGP peer IP
  peerPort: 179
  holdTime: 90s
  routerID: 192.168.19.20  # IP of second NIC
  nodeSelectors:
  - matchLabels:
      kubernetes.io/hostname: ocp-sno-1050ti
---
apiVersion: metallb.io/v1beta1
kind: IPAddressPool
metadata:
  name: bgp-pool
  namespace: metallb-system
spec:
  addresses:
  - 192.168.19.100-192.168.19.150
---
apiVersion: metallb.io/v1beta1
kind: BGPAdvertisement
metadata:
  name: bgp-advertisement
  namespace: metallb-system
spec:
  ipAddressPools:
  - bgp-pool
  communities:
  - 64500:100
```

## Step 5: Make Configuration Persistent

**For RHEL CoreOS**, network configuration should be persistent, but verify:

```bash
# Check connection is saved
nmcli connection show bgp-nic

# Verify it's set to auto-connect
nmcli connection modify bgp-nic connection.autoconnect yes

# Test reboot persistence (if safe)
# reboot
```

## Step 6: Update OpenShift Node Configuration

**If needed**, update node network configuration:

```bash
# Check node network status
oc get node ocp-sno-1050ti -o yaml | grep -A 10 "addresses:"

# The second NIC should appear in node addresses after configuration
```

## Troubleshooting

### Interface Not Detected

```bash
# Check USB device
lsusb | grep -i asix

# Check kernel module
lsmod | grep ax88179

# Load module if needed
modprobe ax88179_178a

# Check dmesg for errors
dmesg | tail -50
```

### Network Not Working

```bash
# Check interface status
ip link show enp0s20f0u2

# Check routing
ip route show

# Check NetworkManager status
systemctl status NetworkManager

# Check connection status
nmcli connection show bgp-nic
```

### BGP Not Peering

```bash
# Check MetalLB speaker logs
oc logs -n metallb-system -l app=metallb,component=speaker

# Verify BGP peer configuration
oc get bgppeer -n metallb-system -o yaml

# Test BGP connectivity
# (May need to check firewall rules)
```

## Security Considerations

- **Firewall**: Ensure firewall rules allow BGP (port 179) and required traffic
- **Network Isolation**: Consider network policies for BGP network
- **Access Control**: Limit who can modify network configuration

## Integration with post-ocp-deployment.sh

The `post-ocp-deployment.sh` script will:

1. **Check for second NIC**: Verify interface exists
2. **Configure MetalLB BGP mode**: If second NIC is configured, use BGP instead of ARP
3. **Validate BGP connectivity**: Test BGP peering

**Script Logic**:
```bash
# Pseudo-code
if [ -n "$(ip addr show enp0s20f0u2 2>/dev/null)" ]; then
  echo "Second NIC detected, configuring MetalLB for BGP mode"
  # Use BGP configuration
else
  echo "No second NIC found, using ARP mode"
  # Use ARP/L2 configuration
fi
```

## Next Steps

1. **Configure second NIC** (this guide)
2. **Run post-ocp-deployment.sh** (will detect and use BGP mode)
3. **Verify BGP peering** in MetalLB logs
4. **Test LoadBalancer services** get BGP-advertised IPs

## References

- [MetalLB BGP Configuration](https://metallb.universe.tf/configuration/bgp/)
- [NetworkManager Documentation](https://access.redhat.com/documentation/en-us/red_hat_enterprise_linux/8/html/configuring_and_managing_networking/getting-started-with-networkmanager_configuring-and-managing-networking)
- [RHEL CoreOS Networking](https://docs.openshift.com/container-platform/latest/post_installation_configuration/machine-configuration-tasks.html)

