package commands

import (
	"context"
	"strconv"
	"time"

	"github.com/Niceblueman/goispappd/internal/config"
	"github.com/Niceblueman/goispappd/internal/exec"
)

var InformCommands = map[string]func(exec *exec.Executor, sshhost *string) (*exec.CommandResult, error){
	"Device.DeviceInfo.ManagementServer.URL": func(_ *exec.Executor, sshhost *string) (*exec.CommandResult, error) {
		// ManagementServer is from the yml config
		cfg, err := config.LoadConfig()
		if err != nil {
			return nil, err
		}
		return &exec.CommandResult{
			Success: true,
			Raw:     []byte(cfg.ACSURL),
		}, nil
	},
	"Device.DeviceInfo.ManagementServer.Username": func(_ *exec.Executor, sshhost *string) (*exec.CommandResult, error) {
		// ManagementServer is from the yml config
		cfg, err := config.LoadConfig()
		if err != nil {
			return nil, err
		}
		return &exec.CommandResult{
			Success: true,
			Raw:     []byte(cfg.Username),
		}, nil
	},
	"Device.DeviceInfo.ManagementServer.Password": func(_ *exec.Executor, sshhost *string) (*exec.CommandResult, error) {
		// ManagementServer is from the yml config
		cfg, err := config.LoadConfig()
		if err != nil {
			return nil, err
		}
		return &exec.CommandResult{
			Success: true,
			Raw:     []byte(cfg.Password),
		}, nil
	},
	"Device.DeviceInfo.ManagementServer.PeriodicInformEnable": func(_ *exec.Executor, sshhost *string) (*exec.CommandResult, error) {
		// ManagementServer static
		return &exec.CommandResult{
			Success: true,
			Raw:     []byte("1"),
		}, nil
	},
	"Device.DeviceInfo.ManagementServer.PeriodicInformInterval": func(_ *exec.Executor, sshhost *string) (*exec.CommandResult, error) {
		// ManagementServer is from the yml config
		cfg, err := config.LoadConfig()
		if err != nil {
			return nil, err
		}
		// Convert the interval to seconds
		interval := int(cfg.PeriodicInterval.Seconds())
		return &exec.CommandResult{
			Success: true,
			Raw:     []byte(strconv.Itoa(interval)),
		}, nil
	},
	"Device.OutsideIPAddress": func(exec *exec.Executor, sshhost *string) (*exec.CommandResult, error) {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
		defer cancel()
		_cmd := `curl -s "https://ip.longshot-router.com/json" 2>/dev/null | grep -o '"realIp":"[^"]*"' | cut -d':' -f2 | tr -d '"' || echo ""`
		if sshhost != nil {
			return exec.SSHExecute(ctx, *sshhost, _cmd)
		}
		return exec.Execute(ctx, _cmd)
	},
	"Device.DeviceInfo.ProvisioningCode": func(_ *exec.Executor, sshhost *string) (*exec.CommandResult, error) {
		cfg, err := config.LoadConfig()
		if err != nil {
			return nil, err
		}
		return &exec.CommandResult{
			Success: true,
			Raw:     []byte(cfg.ProvisioningCode),
		}, nil
	},

	"Device.DeviceInfo.Manufacturer": func(exec *exec.Executor, sshhost *string) (*exec.CommandResult, error) {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
		defer cancel()
		_cmd := `cat /etc/device_info 2>/dev/null | grep "DEVICE_MANUFACTURER" | cut -f 2 -d '=' | sed -e "s/['\"]//g" -e "s/[]:@/?#[!$&()*+,;=]/_/g" | head -n1 | tr -d '\r\n' || echo "OpenWrt"`
		if sshhost != nil {
			return exec.SSHExecute(ctx, *sshhost, _cmd)
		}
		return exec.Execute(ctx, _cmd)
	},
	"Device.DeviceInfo.ManufacturerOUI": func(exec *exec.Executor, sshhost *string) (*exec.CommandResult, error) {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
		defer cancel()
		_cmd := `cat /sys/class/net/eth0/address 2>/dev/null | cut -c 1-8 | tr -d ':' | tr '[:lower:]' '[:upper:]' || echo "000000"`
		if sshhost != nil {
			return exec.SSHExecute(ctx, *sshhost, _cmd)
		}
		return exec.Execute(ctx, _cmd)
	},
	"Device.DeviceInfo.ManufacturerURL": func(exec *exec.Executor, sshhost *string) (*exec.CommandResult, error) {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
		defer cancel()
		_cmd := `cat /etc/device_info 2>/dev/null | grep "DEVICE_MANUFACTURER_URL" | cut -f 2 -d '=' | sed -e "s/['\"]//g" | head -n1 | tr -d '\r\n' || echo "https://openwrt.org/"`
		if sshhost != nil {
			return exec.SSHExecute(ctx, *sshhost, _cmd)
		}
		return exec.Execute(ctx, _cmd)
	},
	"Device.DeviceInfo.ModelName": func(exec *exec.Executor, sshhost *string) (*exec.CommandResult, error) {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
		defer cancel()
		_cmd := `cat /etc/device_info 2>/dev/null | grep "DEVICE_PRODUCT" | cut -f 2 -d '=' | sed -e "s/['\"]//g" -e "s/[]:@/?#[!$&()*+,;=]/_/g" | head -n1 | tr -d '\r\n' || cat /tmp/board.json 2>/dev/null | grep "\"name\"" | cut -f 4 -d '"' | tr -d '\r\n' || echo "DR5332"`
		if sshhost != nil {
			return exec.SSHExecute(ctx, *sshhost, _cmd)
		}
		return exec.Execute(ctx, _cmd)
	},
	"Device.DeviceInfo.Description": func(exec *exec.Executor, sshhost *string) (*exec.CommandResult, error) {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
		defer cancel()
		_cmd := `cat /tmp/board.json 2>/dev/null | grep "\"name\"" | cut -f 4 -d '"' | tr -d '\r\n' || echo "Qualcomm Technologies, Inc. IPQ5332/AP-MI01.2"`
		if sshhost != nil {
			return exec.SSHExecute(ctx, *sshhost, _cmd)
		}
		return exec.Execute(ctx, _cmd)
	},
	"Device.DeviceInfo.ProductClass": func(exec *exec.Executor, sshhost *string) (*exec.CommandResult, error) {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
		defer cancel()
		_cmd := `cat /tmp/board.json 2>/dev/null | grep "\"id\"" | cut -f 4 -d '"' | tr -d '\r\n' || echo "qcom,ipq5332-ap-mi01.2"`
		if sshhost != nil {
			return exec.SSHExecute(ctx, *sshhost, _cmd)
		}
		return exec.Execute(ctx, _cmd)
	},
	"Device.DeviceInfo.SerialNumber": func(exec *exec.Executor, sshhost *string) (*exec.CommandResult, error) {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
		defer cancel()
		_cmd := `cat /proc/cpuinfo 2>/dev/null | grep "Serial" | cut -f 2 -d ':' | tr -d ' \r\n' || uci get system.@system[0].serial 2>/dev/null | tr -d '\r\n' || echo "Unknown"`
		if sshhost != nil {
			return exec.SSHExecute(ctx, *sshhost, _cmd)
		}
		return exec.Execute(ctx, _cmd)
	},
	"Device.DeviceInfo.SpecVersion": func(exec *exec.Executor, sshhost *string) (*exec.CommandResult, error) {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
		defer cancel()
		_cmd := `cat /etc/openwrt_release 2>/dev/null | grep "DISTRIB_RELEASE" | cut -f 2 -d '=' | tr -d '"' | tr -d '\r\n' || echo "Unknown"`
		if sshhost != nil {
			return exec.SSHExecute(ctx, *sshhost, _cmd)
		}
		return exec.Execute(ctx, _cmd)
	},
	"Device.DeviceInfo.HardwareVersion": func(exec *exec.Executor, sshhost *string) (*exec.CommandResult, error) {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
		defer cancel()
		_cmd := `cat /etc/device_info 2>/dev/null | grep "DEVICE_REVISION" | cut -f 2 -d '=' | sed -e "s/['\"]//g" -e "s/[]:@/?#[!$&()*+,;=]/_/g" | head -n1 | tr -d '\r\n' || echo "v0"`
		if sshhost != nil {
			return exec.SSHExecute(ctx, *sshhost, _cmd)
		}
		return exec.Execute(ctx, _cmd)
	},
	"Device.DeviceInfo.SoftwareVersion": func(exec *exec.Executor, sshhost *string) (*exec.CommandResult, error) {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
		defer cancel()
		_cmd := `cat /etc/openwrt_version 2>/dev/null | tr -d '\r\n' || cat /etc/os-release 2>/dev/null | grep "VERSION=" | cut -f 2 -d '=' | tr -d '"' | tr -d '\r\n' || echo "Unknown"`
		if sshhost != nil {
			return exec.SSHExecute(ctx, *sshhost, _cmd)
		}
		return exec.Execute(ctx, _cmd)
	},
	"Device.DeviceInfo.UpTime": func(exec *exec.Executor, sshhost *string) (*exec.CommandResult, error) {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
		defer cancel()
		_cmd := `cat /proc/uptime 2>/dev/null | cut -f 1 -d ' ' | cut -f 1 -d '.' | tr -d '\r\n' || echo "0"`
		if sshhost != nil {
			return exec.SSHExecute(ctx, *sshhost, _cmd)
		}
		return exec.Execute(ctx, _cmd)
	},
	"Device.DeviceInfo.VendorConfigFileNumberOfEntries": func(exec *exec.Executor, sshhost *string) (*exec.CommandResult, error) {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
		defer cancel()
		_cmd := `ls /etc/config 2>/dev/null | wc -l | tr -d '\r\n' || echo "0"`
		if sshhost != nil {
			return exec.SSHExecute(ctx, *sshhost, _cmd)
		}
		return exec.Execute(ctx, _cmd)
	},
	"Device.DeviceInfo.MemoryStatus.Total": func(exec *exec.Executor, sshhost *string) (*exec.CommandResult, error) {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
		defer cancel()
		_cmd := `cat /proc/meminfo 2>/dev/null | grep "MemTotal" | awk '{print $2}' | tr -d '\r\n' || echo "0"`
		if sshhost != nil {
			return exec.SSHExecute(ctx, *sshhost, _cmd)
		}
		return exec.Execute(ctx, _cmd)
	},
	"Device.DeviceInfo.MemoryStatus.Free": func(exec *exec.Executor, sshhost *string) (*exec.CommandResult, error) {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
		defer cancel()
		_cmd := `cat /proc/meminfo 2>/dev/null | grep "MemFree" | awk '{print $2}' | tr -d '\r\n' || echo "0"`
		if sshhost != nil {
			return exec.SSHExecute(ctx, *sshhost, _cmd)
		}
		return exec.Execute(ctx, _cmd)
	},
	"Device.DeviceInfo.ProcessStatus.CPUUsage": func(exec *exec.Executor, sshhost *string) (*exec.CommandResult, error) {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
		defer cancel()
		_cmd := `top -bn1 2>/dev/null | grep "%Cpu(s)" | awk '{print int($2 + $4)}' | tr -d '\r\n' || echo "0"`
		if sshhost != nil {
			return exec.SSHExecute(ctx, *sshhost, _cmd)
		}
		return exec.Execute(ctx, _cmd)
	},
	"Device.DeviceInfo.VendorConfigFile.1.Name": func(exec *exec.Executor, sshhost *string) (*exec.CommandResult, error) {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
		defer cancel()
		_cmd := `ls /etc/config 2>/dev/null | head -n1 | tail -n1 | tr -d '\r\n' || echo "config"`
		if sshhost != nil {
			return exec.SSHExecute(ctx, *sshhost, _cmd)
		}
		return exec.Execute(ctx, _cmd)
	},
	"Device.DeviceInfo.VendorConfigFile.1.Description": func(exec *exec.Executor, sshhost *string) (*exec.CommandResult, error) {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
		defer cancel()
		_cmd := `ls /etc/config 2>/dev/null | head -n1 | tail -n1 | xargs -I {} sh -c 'echo "Configuration file for {}" | tr -d '\r\n'' || echo "Configuration file"`
		if sshhost != nil {
			return exec.SSHExecute(ctx, *sshhost, _cmd)
		}
		return exec.Execute(ctx, _cmd)
	},
	"Device.DeviceInfo.VendorConfigFile.1.UseForBackupRestore": func(exec *exec.Executor, sshhost *string) (*exec.CommandResult, error) {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
		defer cancel()
		_cmd := `ls /etc/config 2>/dev/null | head -n1 | tail -n1 | grep -E "^(system|network|firewall|dhcp|wireless)$" >/dev/null && echo "true" || echo "false"`
		if sshhost != nil {
			return exec.SSHExecute(ctx, *sshhost, _cmd)
		}
		return exec.Execute(ctx, _cmd)
	},
	"Device.DeviceInfo.VendorConfigFile.2.Index": func(exec *exec.Executor, sshhost *string) (*exec.CommandResult, error) {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
		defer cancel()
		_cmd := `echo "2"`
		if sshhost != nil {
			return exec.SSHExecute(ctx, *sshhost, _cmd)
		}
		return exec.Execute(ctx, _cmd)
	},
	"Device.DeviceInfo.VendorConfigFile.2.Name": func(exec *exec.Executor, sshhost *string) (*exec.CommandResult, error) {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
		defer cancel()
		_cmd := `ls /etc/config 2>/dev/null | head -n2 | tail -n1 | tr -d '\r\n' || echo "config"`
		if sshhost != nil {
			return exec.SSHExecute(ctx, *sshhost, _cmd)
		}
		return exec.Execute(ctx, _cmd)
	},
	"Device.DeviceInfo.VendorConfigFile.2.Description": func(exec *exec.Executor, sshhost *string) (*exec.CommandResult, error) {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
		defer cancel()
		_cmd := `ls /etc/config 2>/dev/null | head -n2 | tail -n1 | xargs -I {} sh -c 'echo "Configuration file for {}" | tr -d '\r\n'' || echo "Configuration file"`
		if sshhost != nil {
			return exec.SSHExecute(ctx, *sshhost, _cmd)
		}
		return exec.Execute(ctx, _cmd)
	},
	"Device.DeviceInfo.VendorConfigFile.2.UseForBackupRestore": func(exec *exec.Executor, sshhost *string) (*exec.CommandResult, error) {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
		defer cancel()
		_cmd := `ls /etc/config 2>/dev/null | head -n2 | tail -n1 | grep -E "^(system|network|firewall|dhcp|wireless)$" >/dev/null && echo "true" || echo "false"`
		if sshhost != nil {
			return exec.SSHExecute(ctx, *sshhost, _cmd)
		}
		return exec.Execute(ctx, _cmd)
	},
	"Device.DeviceInfo.VendorConfigFile.3.Index": func(exec *exec.Executor, sshhost *string) (*exec.CommandResult, error) {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
		defer cancel()
		_cmd := `echo "3"`
		if sshhost != nil {
			return exec.SSHExecute(ctx, *sshhost, _cmd)
		}
		return exec.Execute(ctx, _cmd)
	},
	"Device.DeviceInfo.VendorConfigFile.3.Name": func(exec *exec.Executor, sshhost *string) (*exec.CommandResult, error) {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
		defer cancel()
		_cmd := `ls /etc/config 2>/dev/null | head -n3 | tail -n1 | tr -d '\r\n' || echo "config"`
		if sshhost != nil {
			return exec.SSHExecute(ctx, *sshhost, _cmd)
		}
		return exec.Execute(ctx, _cmd)
	},
	"Device.DeviceInfo.VendorConfigFile.3.Description": func(exec *exec.Executor, sshhost *string) (*exec.CommandResult, error) {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
		defer cancel()
		_cmd := `ls /etc/config 2>/dev/null | head -n3 | tail -n1 | xargs -I {} sh -c 'echo "Configuration file for {}" | tr -d '\r\n'' || echo "Configuration file"`
		if sshhost != nil {
			return exec.SSHExecute(ctx, *sshhost, _cmd)
		}
		return exec.Execute(ctx, _cmd)
	},
	"Device.DeviceInfo.VendorConfigFile.3.UseForBackupRestore": func(exec *exec.Executor, sshhost *string) (*exec.CommandResult, error) {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
		defer cancel()
		_cmd := `ls /etc/config 2>/dev/null | head -n3 | tail -n1 | grep -E "^(system|network|firewall|dhcp|wireless)$" >/dev/null && echo "true" || echo "false"`
		if sshhost != nil {
			return exec.SSHExecute(ctx, *sshhost, _cmd)
		}
		return exec.Execute(ctx, _cmd)
	},
	"Device.DeviceInfo.VendorConfigFile.4.Index": func(exec *exec.Executor, sshhost *string) (*exec.CommandResult, error) {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
		defer cancel()
		_cmd := `echo "4"`
		if sshhost != nil {
			return exec.SSHExecute(ctx, *sshhost, _cmd)
		}
		return exec.Execute(ctx, _cmd)
	},
	"Device.DeviceInfo.VendorConfigFile.4.Name": func(exec *exec.Executor, sshhost *string) (*exec.CommandResult, error) {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
		defer cancel()
		_cmd := `ls /etc/config 2>/dev/null | head -n4 | tail -n1 | tr -d '\r\n' || echo "config"`
		if sshhost != nil {
			return exec.SSHExecute(ctx, *sshhost, _cmd)
		}
		return exec.Execute(ctx, _cmd)
	},
	"Device.DeviceInfo.VendorConfigFile.4.Description": func(exec *exec.Executor, sshhost *string) (*exec.CommandResult, error) {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
		defer cancel()
		_cmd := `ls /etc/config 2>/dev/null | head -n4 | tail -n1 | xargs -I {} sh -c 'echo "Configuration file for {}" | tr -d '\r\n'' || echo "Configuration file"`
		if sshhost != nil {
			return exec.SSHExecute(ctx, *sshhost, _cmd)
		}
		return exec.Execute(ctx, _cmd)
	},
	"Device.DeviceInfo.VendorConfigFile.4.UseForBackupRestore": func(exec *exec.Executor, sshhost *string) (*exec.CommandResult, error) {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
		defer cancel()
		_cmd := `ls /etc/config 2>/dev/null | head -n4 | tail -n1 | grep -E "^(system|network|firewall|dhcp|wireless)$" >/dev/null && echo "true" || echo "false"`
		if sshhost != nil {
			return exec.SSHExecute(ctx, *sshhost, _cmd)
		}
		return exec.Execute(ctx, _cmd)
	},
	"Device.DeviceInfo.VendorConfigFile.5.Index": func(exec *exec.Executor, sshhost *string) (*exec.CommandResult, error) {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
		defer cancel()
		_cmd := `echo "5"`
		if sshhost != nil {
			return exec.SSHExecute(ctx, *sshhost, _cmd)
		}
		return exec.Execute(ctx, _cmd)
	},
	"Device.DeviceInfo.VendorConfigFile.5.Name": func(exec *exec.Executor, sshhost *string) (*exec.CommandResult, error) {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
		defer cancel()
		_cmd := `ls /etc/config 2>/dev/null | head -n5 | tail -n1 | tr -d '\r\n' || echo "config"`
		if sshhost != nil {
			return exec.SSHExecute(ctx, *sshhost, _cmd)
		}
		return exec.Execute(ctx, _cmd)
	},
	"Device.DeviceInfo.VendorConfigFile.5.Description": func(exec *exec.Executor, sshhost *string) (*exec.CommandResult, error) {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
		defer cancel()
		_cmd := `ls /etc/config 2>/dev/null | head -n5 | tail -n1 | xargs -I {} sh -c 'echo "Configuration file for {}" | tr -d '\r\n'' || echo "Configuration file"`
		if sshhost != nil {
			return exec.SSHExecute(ctx, *sshhost, _cmd)
		}
		return exec.Execute(ctx, _cmd)
	},
	"Device.DeviceInfo.VendorConfigFile.5.UseForBackupRestore": func(exec *exec.Executor, sshhost *string) (*exec.CommandResult, error) {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
		defer cancel()
		_cmd := `ls /etc/config 2>/dev/null | head -n5 | tail -n1 | grep -E "^(system|network|firewall|dhcp|wireless)$" >/dev/null && echo "true" || echo "false"`
		if sshhost != nil {
			return exec.SSHExecute(ctx, *sshhost, _cmd)
		}
		return exec.Execute(ctx, _cmd)
	},
}
