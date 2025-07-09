FROM scratch

# Add the OpenWrt rootfs (ext4 format)
ADD bin/targets/bcm27xx/bcm2710/openwrt-23.05.4-bcm27xx-bcm2710-rootfs.tar.gz /

# Install dependencies
RUN opkg update && \
    opkg install \
    libubus \
    libuci \
    libubox \
    libroxml \
    libjson-c \
    libcurl \
    curl

# Enable SSH
RUN echo "config dropbear" >> /etc/config/dropbear && \
    echo "    option PasswordAuth 'on'" >> /etc/config/dropbear && \
    echo "    option RootPasswordAuth 'on'" >> /etc/config/dropbear && \
    echo "    option Port '22'" >> /etc/config/dropbear

# Create /app directory
RUN mkdir /app

# Set root password (change 'password' to your desired password)
RUN echo -e "password\npassword" | passwd root

# Start dropbear (SSH) on boot
RUN /etc/init.d/dropbear enable

# Entrypoint
CMD ["/sbin/init"]