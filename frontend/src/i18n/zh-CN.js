export default {
    login: {
        username: '用户名',
        password: '密码',
        login: '登录',
        register: '注册',
        registerNow: '现在注册',
        inputUsername: '请输入用户名',
        inputPassword: '请输入密码',
        or: '或者'
    },
    register: {
        username: '用户名',
        password: '密码',
        register: '注册',
        inputUsername: '请输入用户名',
        inputPassword: '请输入密码'
    },
    loading: {
        title: '加载中'
    },
    user: {
        title: '用户',
        subtitle: '用户信息',
        username: '用户名',
        changePassword: '修改密码',
        oldPassword: '旧密码',
        newPassword: '新密码',
        update: '更新',
        inputOldPassword: '请输入旧密码',
        inputNewPassword: '请输入新密码'
    },
    network: {
        title: '网络',
        subtitle: '创建和管理私有网络',
        add: '添加',
        edit: '编辑',
        delete: '删除',
        netName: '网络名称',
        password: '密码',
        dhcp: 'DHCP',
        broadcast: '广播',
        lease: '租期',
        action: '操作',
        enable: '启用',
        disable: '禁用',
        modalTitle: '网络配置',
        inputNetname: '请输入网络名称',
        inputPassword: '请输入密码',
        inputDhcp: '请输入DHCP配置',
        inputLease: '请输入租期'
    },
    device: {
        title: '设备',
        subtitle: '查看和管理设备',
        delete: '删除',
        confirmDelete: '确定要删除这个设备吗？',
        yes: '是',
        no: '否',
        columns: {
            hostname: '主机名',
            network: '网络',
            ip: 'IP地址',
            country: '国家',
            region: '地区',
            rx: '接收',
            tx: '发送',
            online: '在线',
            os: '操作系统',
            version: '版本',
            lastActiveTime: '最后活跃时间',
            action: '操作'
        },
        status: {
            online: '在线',
            offline: '离线'
        }
    },
    route: {
        title: '路由',
        subtitle: '多局域网络互联',
        add: '添加',
        delete: '删除',
        modalTitle: '路由配置',
        columns: {
            network: '网络',
            devAddr: '设备地址',
            devMask: '设备掩码',
            dstAddr: '目标地址',
            dstMask: '目标掩码',
            nextHop: '下一跳',
            priority: '优先级',
            action: '操作'
        },
        placeholder: {
            network: '选择网络',
            devAddr: '请输入设备地址',
            devMask: '请输入设备掩码',
            dstAddr: '请输入目标地址',
            dstMask: '请输入目标掩码',
            nextHop: '请输入下一跳',
            priority: '请输入优先级'
        }
    },
    statistics: {
        title: '统计',
        subtitle: '用户统计',
        columns: {
            net: '网络数',
            device: '设备数',
            rx: '接收流量',
            tx: '发送流量'
        }
    },
    adminUser: {
        title: '用户',
        subtitle: '用户管理',
        create: '创建',
        update: '更新',
        delete: '删除',
        confirmDelete: '确定要删除这个用户吗？',
        yes: '是',
        no: '否',
        columns: {
            username: '用户名',
            role: '角色',
            network: '网络数',
            device: '设备数',
            rx: '接收流量',
            tx: '发送流量',
            lastActiveTime: '最后活跃时间',
            action: '操作'
        },
        placeholder: {
            username: '用户名',
            password: '密码'
        }
    },
    adminSetting: {
        title: '设置',
        subtitle: '系统配置',
        register: {
            title: '注册设置',
            allowed: '允许注册',
            interval: '注册间隔',
            intervalUnit: '分钟'
        },
        userClean: {
            title: '用户清理',
            auto: '自动清理用户',
            threshold: '不活跃用户阈值',
            thresholdUnit: '天',
            manual: '手动清理',
            clean: '清理',
            success: '清理成功'
        }
    },
    adminLicense: {
        title: '许可证',
        subtitle: '许可证信息',
        renew: '续期',
        columns: {
            licenseId: '许可证ID',
            description: '描述',
            expire: '过期时间',
            action: '操作'
        }
    },
    components: {
        sider: {
            statistics: '统计',
            network: '网络',
            device: '设备',
            route: '路由',
            user: '用户',
            setting: '设置',
            license: '许可证',
            logout: '退出'
        },
        footer: {
            copyright: 'Cacao © 2024'
        }
    }
} 
