#! /usr/bin/env bash

# Note to future self: do NOT run this script! It has NOT been tested. This is
# only a crude recording of the original steps taken to setup the deployment
# user on the server. If you need to reproduce the environment, run the commands
# one by one and use caution!
#
# The following vars are referenced in this script:
# - $DEPLOY_USER        The name of the deployment user
# - $DEPLOY_SSH_KEY     The private deployment SSH key
# - $SERVER_DIR         The path to the directory of the server files
# - $SERVER_SERVICE     The name of the systemd service for the server

# Create deployment user
sudo adduser --disabled-password --gecos "" $DEPLOY_USER
# Set user shell to bash
sudo usermod --shell /bin/bash $DEPLOY_USER

# Add SSH private key (public key is used by the GitHub action)
sudo mkdir -p "/home/$DEPLOY_USER/.ssh"
sudo chmod 700 "/home/$DEPLOY_USER/.ssh"
sudo chown "$DEPLOY_USER:$DEPLOY_USER" "/home/$DEPLOY_USER/.ssh"
echo "$DEPLOY_SSH_KEY" | sudo tee "/home/$DEPLOY_USER/.ssh/authorized_keys"
sudo chmod 600 "/home/$DEPLOY_USER/.ssh/authorized_keys"
sudo chown "$DEPLOY_USER:$DEPLOY_USER" "/home/$DEPLOY_USER/.ssh/authorized_keys"

# Grant access to app dir
sudo chown -R "$DEPLOY_USER:$DEPLOY_USER" "$SERVER_DIR"
sudo chmod -R 755 $SERVER_DIR

# Grant ability to "sudo" for this one specific command only:
echo "$DEPLOY_USER ALL=(ALL) NOPASSWD: /bin/systemctl restart $SERVER_SERVICE" | sudo tee "/etc/sudoers.d/$DEPLOY_USER"

# SSH config for user
echo "Match User $DEPLOY_USER" | sudo tee -a /etc/ssh/sshd_config
echo "    PermitTTY yes" | sudo tee -a /etc/ssh/sshd_config
echo "    PermitTunnel no" | sudo tee -a /etc/ssh/sshd_config
echo "    AllowTcpForwarding no" | sudo tee -a /etc/ssh/sshd_config
echo "    X11Forwarding no" | sudo tee -a /etc/ssh/sshd_config
# Restart ssh for above changes to take effect
sudo systemctl restart ssh
