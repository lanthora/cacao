export default {
    auth: {
        requestFailed: "Request failed. Check your connection and try again."
    },
    login: {
        noAccount: "Don't have an account?",
        username: 'Username',
        password: 'Password',
        login: 'Log in',
        inputUsername: 'Please input your username',
        inputPassword: 'Please input your password',
        or: 'Or'
    },
    register: {
        confirmPassword: "Confirm password",
        inputConfirmPassword: "Please enter your password again",
        passwordMismatch: "Passwords do not match",
        haveAccount: "Already have an account?",
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
        design: {
            workspace: "CACAO / ADMINISTRATION",
            heading: "Workspace settings",
            description: "Manage access, housekeeping, and location data from one place.",
            preferences: "General preferences",
            registerDescription: "Control access and registration frequency",
            cleanDescription: "Keep your user list up to date",
            autoSaved: "Changes save automatically",
            cleanConfirm: "Remove normal users who meet the inactivity threshold?",
            cancel: "Cancel",
            databaseHeading: "Location databases",
            databaseDescription: "Keep location data current with independent update schedules.",
            hourly: "Checked every hour",
            ipdbDescription: "IPIP location database",
            mmdbDescription: "MaxMind location database",
            enabled: "Auto-update on",
            disabled: "Auto-update off",
            updating: "Updating",
            updateError: "Action failed",
            rawFile: "Use a raw database file over HTTP(S). Archives are not supported.",
            disableHelp: "Set to 0 or clear the URL to disable updates",
            unsaved: "Unsaved changes",
            savedState: "Settings saved",
            lookupOrder: "Lookup priority",
            lookupHelp: "New versions apply to future lookups. Existing device locations refresh through normal activity.",
            loadFailed: "Unable to load database settings",
            retry: "Try again"
        },
        geoDB: {
            url: "Database download URL",
            interval: "Automatic update interval",
            hours: "hours",
            help: "0 or an empty URL disables automatic updates. The URL must return a raw IPDB / MMDB file.",
            lastSuccess: "Last successful update",
            never: "Never updated",
            actions: "Database update",
            save: "Save settings",
            update: "Update now",
            savedHelp: "Updates use saved settings. Save your changes first.",
            priority: "Location priority: IPDB → MMDB → ipinfo. Updates apply to future lookups; existing device records are not refreshed immediately.",
            failed: "Database operation failed",
            invalidInterval: "Enter a whole number of hours",
            saved: "Settings saved",
            updated: "Database updated"
        },
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
    components: {
        sider: {
            statistics: 'Statistics',
            network: 'Network',
            device: 'Device',
            route: 'Route',
            user: 'User',
            setting: 'Setting',
            logout: 'Logout'
        },
        footer: {
            copyright: 'Cacao © {year}'
        }
    }
} 
