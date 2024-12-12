export default {
    login: {
        username: 'Username',
        password: 'Password',
        login: 'Log in',
        register: 'Register',
        registerNow: 'register now',
        inputUsername: 'Please input your username',
        inputPassword: 'Please input your password',
        or: 'Or'
    },
    register: {
        username: 'Username',
        password: 'Password',
        register: 'Register',
        inputUsername: 'Please input your username',
        inputPassword: 'Please input your password'
    },
    loading: {
        title: 'Loading'
    },
    user: {
        title: 'User',
        subtitle: 'user information',
        username: 'User Name',
        changePassword: 'Change Password',
        oldPassword: 'Old Password',
        newPassword: 'New Password',
        update: 'Update',
        inputOldPassword: 'Please input old password',
        inputNewPassword: 'Please input new password'
    },
    network: {
        title: 'Network',
        subtitle: 'create and manage private networks',
        add: 'Add',
        edit: 'Edit',
        delete: 'Delete',
        netName: 'Net Name',
        password: 'Password',
        dhcp: 'DHCP',
        broadcast: 'Broadcast',
        lease: 'Lease',
        action: 'Action',
        enable: 'Enable',
        disable: 'Disable',
        modalTitle: 'Network',
        inputNetname: 'Please input network name',
        inputPassword: 'Please input password',
        inputDhcp: 'Please input DHCP configuration',
        inputLease: 'Please input lease time'
    },
    device: {
        title: 'Device',
        subtitle: 'view and manage devices',
        delete: 'Delete',
        confirmDelete: 'Are you sure delete this device?',
        yes: 'Yes',
        no: 'No',
        columns: {
            hostname: 'Host Name',
            network: 'Network',
            ip: 'IP',
            country: 'Country',
            region: 'Region',
            rx: 'RX',
            tx: 'TX',
            online: 'Online',
            os: 'OS',
            version: 'Version',
            lastActiveTime: 'Last Active At',
            action: 'Action'
        },
        status: {
            online: 'true',
            offline: 'false'
        }
    },
    route: {
        title: 'Route',
        subtitle: 'multiple local area network networking',
        add: 'Add',
        delete: 'Delete',
        modalTitle: 'Route',
        columns: {
            network: 'Network',
            devAddr: 'Device Address',
            devMask: 'Device Mask',
            dstAddr: 'Destination Address',
            dstMask: 'Destination Mask',
            nextHop: 'Next Hop',
            priority: 'Priority',
            action: 'Action'
        },
        placeholder: {
            network: 'Network',
            devAddr: 'Device Address',
            devMask: 'Device Mask',
            dstAddr: 'Destination Address',
            dstMask: 'Destination Mask',
            nextHop: 'Next Hop',
            priority: 'Priority'
        }
    },
    statistics: {
        title: 'Statistics',
        subtitle: 'user statistics',
        columns: {
            net: 'Net',
            device: 'Device',
            rx: 'RX',
            tx: 'TX'
        }
    },
    adminUser: {
        title: 'User',
        subtitle: 'user management',
        create: 'Create',
        update: 'Update',
        delete: 'Delete',
        confirmDelete: 'Are you sure delete this user?',
        yes: 'Yes',
        no: 'No',
        columns: {
            username: 'Username',
            role: 'Role',
            network: 'Network',
            device: 'Device',
            rx: 'RX',
            tx: 'TX',
            lastActiveTime: 'Last Active At',
            action: 'Action'
        },
        placeholder: {
            username: 'Username',
            password: 'Password'
        }
    },
    adminSetting: {
        title: 'Setting',
        subtitle: 'system configuration',
        register: {
            title: 'Registration Allowed',
            allowed: 'Registration Allowed',
            interval: 'Registration Interval',
            intervalUnit: 'mins'
        },
        userClean: {
            title: 'Auto Clean User',
            auto: 'Auto Clean User',
            threshold: 'Inactive User Threshold',
            thresholdUnit: 'days',
            manual: 'Manual Clean',
            clean: 'Clean',
            success: 'success'
        }
    },
    adminLicense: {
        title: 'License',
        subtitle: 'license information',
        renew: 'Renew',
        columns: {
            licenseId: 'License ID',
            description: 'Description',
            expire: 'Expire',
            action: 'Action'
        }
    },
    components: {
        sider: {
            statistics: 'Statistics',
            network: 'Network',
            device: 'Device',
            route: 'Route',
            user: 'User',
            setting: 'Setting',
            license: 'License',
            logout: 'Logout'
        },
        footer: {
            copyright: 'Cacao © 2024'
        }
    }
} 
