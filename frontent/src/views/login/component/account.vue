<template>
	<el-form ref="formRef" size="large" class="login-content-form">
		<el-menu
			:default-active="activeIndex"
			class="el-menu-login"
			mode="horizontal"
			@select="handleSelect"
		>
			<el-menu-item index="login">登录</el-menu-item>
			<el-menu-item index="regester">注册</el-menu-item>
		</el-menu>
		<el-form-item class="login-animation1" prop="username">
			<el-input type="text" placeholder="请输入用户名" clearable autocomplete="off">
			</el-input>
		</el-form-item>
		<el-form-item class="login-animation2" prop="password">
			<el-input type="password" placeholder="请输入用密码">
				<template #suffix>
					<i class="iconfont el-input__icon login-content-password icon-yincangmima"></i>
				</template>
			</el-input>
		</el-form-item>
		<el-form-item class="login-animation4">
			<el-button type="primary" class="login-content-submit" round @click="loginClick">
				<span>登录</span>
			</el-button>
		</el-form-item>
	</el-form>
</template>


<script lang="ts" setup>
import { reactive, ref } from 'vue'
import { errorMessage } from '@/utils/message';

//选择登录还是注册
const activeIndex = ref('login')
const handleSelect = (key: string, keyPath: string[]) => {
  console.log(key, keyPath)
}

//表单内容
const state = reactive({
	isShowPassword: false,
	ruleForm: {
		username: '',
		password: '',
	},
	loading: {
		signIn: false,
	},
});

// 表单校验
const loginClick = async () => {
	errorMessage("请填写登录信息")
};
</script>

<style scoped lang="scss">
.login-content-form {
	.el-menu-login {
		margin: 20px 0;
	}

	.login-content-submit {
		width: 100%;
		letter-spacing: 2px;
		font-weight: 300;
		margin-top: 15px;
	}
}
</style>
