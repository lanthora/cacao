export default {
    auth: {
        requestFailed: "请求失败，请检查网络后重试。"
    },
    login: {
        noAccount: "还没有账号？",
        username: '用户名',
        password: '密码',
        login: '登录',
        inputUsername: '请输入用户名',
        inputPassword: '请输入密码',
        or: '或者'
    },
    register: {
        confirmPassword: "确认密码",
        inputConfirmPassword: "请再次输入密码",
        passwordMismatch: "两次输入的密码不一致",
        haveAccount: "已有账号？",
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
        edit: '编辑',
        editTitle: '修改虚拟 IP',
        save: '保存',
        cancel: '取消',
        addressRange: '网络地址范围（CIDR）',
        editHint: '仅对动态获取 IP 的客户端生效，固定 IP 客户端仍使用自身配置。在线设备保存后会断线重连，离线设备下次连接时生效。',
        invalidIP: '请输入所属网络内有效的 IPv4 主机地址，不能使用网络地址或广播地址。',
        ipConflict: '该 IP 已被同一网络中的其他设备占用（包括离线设备）。',
        notExists: '设备不存在或无权修改，请刷新设备列表。',
        updateFailed: '修改失败，请稍后重试。',
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
        design: {
            workspace: "CACAO / 管理控制台",
            heading: "系统偏好设置",
            description: "管理访问规则、清理策略与地理数据库，让日常维护更轻松。",
            preferences: "基础设置",
            registerDescription: "控制新用户的注册权限与频率",
            cleanDescription: "按不活跃时间维护用户列表",
            autoSaved: "更改后自动保存",
            cleanConfirm: "确认清理达到不活跃阈值的普通用户？",
            cancel: "取消",
            databaseHeading: "地理数据库",
            databaseDescription: "保持定位数据新鲜，为两种数据库独立设置更新策略。",
            hourly: "每小时检查更新",
            ipdbDescription: "IPIP 地理位置数据库",
            mmdbDescription: "MaxMind 地理位置数据库",
            enabled: "自动更新已开启",
            disabled: "自动更新已关闭",
            updating: "正在更新",
            updateError: "操作失败",
            rawFile: "支持 HTTP(S) 原始数据库文件，不支持压缩包。",
            disableHelp: "设为 0 或清空地址可关闭自动更新",
            unsaved: "有未保存的更改",
            savedState: "配置已保存",
            lookupOrder: "定位查询顺序",
            lookupHelp: "新版本用于后续定位查询。已有设备的位置将在正常活动时刷新。",
            loadFailed: "暂时无法加载数据库配置",
            retry: "重新加载"
        },
        geoDB: {
            url: "数据库下载地址",
            interval: "自动更新周期",
            hours: "小时",
            help: "0 或空地址表示关闭自动更新；下载内容须为原始 IPDB / MMDB 文件。",
            lastSuccess: "上次成功更新时间",
            never: "尚未更新",
            actions: "数据库更新",
            save: "保存配置",
            update: "立即更新",
            savedHelp: "立即更新使用已保存的配置；修改后请先保存。",
            priority: "定位优先级：IPDB → MMDB → ipinfo。更新后用于后续定位查询，已有设备记录不会立即刷新。",
            failed: "数据库操作失败",
            invalidInterval: "请输入整数小时数",
            saved: "配置已保存",
            updated: "数据库更新成功"
        },
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
    components: {
        sider: {
            statistics: '统计',
            network: '网络',
            device: '设备',
            route: '路由',
            user: '用户',
            setting: '设置',
            logout: '退出'
        },
        footer: {
            copyright: 'Cacao © {year}'
        }
    }
} 
